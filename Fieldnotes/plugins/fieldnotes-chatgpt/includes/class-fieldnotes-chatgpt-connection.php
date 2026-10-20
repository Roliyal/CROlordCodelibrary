<?php
if (!defined('ABSPATH')) exit;

/** Fixed official endpoint, safe error descriptions, and an admin-only metadata probe. */
final class Fieldnotes_ChatGPT_Connection {
    public static function clean_detail($text, $api_key = '') {
        $text = is_scalar($text) ? (string) $text : '';
        if ($api_key !== '') $text = str_replace($api_key, '[redacted]', $text);
        $text = preg_replace('/Bearer\s+[^\s\"\'<>]+/i', 'Bearer [redacted]', $text);
        $text = preg_replace('/\bsk-[a-zA-Z0-9_-]+/', '[redacted]', $text);
        $text = preg_replace('~(https?://)[^/\s@]+@~i', '$1[redacted]@', $text);
        $text = trim(wp_strip_all_tags($text));
        return function_exists('mb_substr') ? mb_substr($text, 0, 600, 'UTF-8') : substr($text, 0, 600);
    }

    private static function issue($code, $message, $hint, $technical, $status = 0) {
        return array('code' => $code, 'message' => __($message, 'fieldnotes-chatgpt'), 'hint' => __($hint, 'fieldnotes-chatgpt'), 'technical' => $technical, 'http_status' => (int) $status);
    }

    public static function transport_error($error, $api_key = '') {
        $raw = (string) $error->get_error_message();
        $detail = self::clean_detail($raw, $api_key);
        $wp_code = (string) $error->get_error_code();
        if (preg_match('/cURL error 5\b|could not resolve proxy/i', $raw)) {
            return self::issue('proxy_resolution_failed', '服务器无法解析配置的代理地址，请管理员检查代理设置。', '检查 WordPress/PHP 的代理主机与 DNS；确认 PHP 使用的配置与终端一致。', $detail);
        }
        if (preg_match('/cURL error 6\b|could not resolve host|getaddrinfo|php_network_getaddresses|name or service not known|temporary failure in name resolution/i', $raw)) {
            return self::issue('dns_failed', '服务器无法解析 OpenAI 域名，请管理员检查 DNS。', '在 WordPress 实际运行环境检查 api.openai.com 的域名解析，再检查服务器 DNS 配置。', $detail);
        }
        if (preg_match('/cURL error (60|77)\b|certificate verify failed|SSL certificate problem|unable to get local issuer|certificate has expired/i', $raw)) {
            return self::issue('certificate_failed', '服务器证书校验失败，请管理员检查 CA 证书与系统时间。', '更新操作系统 CA 证书与 PHP/cURL 的 CA 配置，核对系统时间；保持证书校验开启。', $detail);
        }
        if (preg_match('/cURL error 35\b|SSL connect error|TLS handshake|SSL routines|schannel/i', $raw)) {
            return self::issue('tls_failed', '服务器与 OpenAI 的 TLS 连接未能建立。', '检查 PHP/cURL/OpenSSL 版本、TLS 协商，以及出口网关是否干预 HTTPS。', $detail);
        }
        if (preg_match('/(?:connection|connect) timed out|connection timeout/i', $raw)) {
            return self::issue('connection_timeout', '服务器到 OpenAI 的连接建立超时。', '超时发生在 DNS、TCP 或 TLS 连接阶段，尚未取得 API HTTP 响应。先检查服务器或容器出口；延长对话生成时限不能修复无法连通的网络。', $detail);
        }
        if (preg_match('/cURL error 28\b|timed out|timeout/i', $raw)) {
            return self::issue('request_timeout', '等待 OpenAI 响应超时，请稍后重试。', '检查服务器出口与延迟；可调整插件请求时限，同时核对 PHP 和反向代理的执行时限。', $detail);
        }
        if ($wp_code === 'http_request_not_executed' || preg_match('/blocked requests through HTTP/i', $raw)) {
            return self::issue('wordpress_http_blocked', 'WordPress 阻止了外部请求，请管理员检查站点配置。', '检查 WP_HTTP_BLOCK_EXTERNAL、WP_ACCESSIBLE_HOSTS 和拦截外部请求的安全插件。', $detail);
        }
        if (preg_match('/cURL error 7\b|failed to connect|connection refused|no route to host|network is unreachable/i', $raw)) {
            return self::issue('connection_failed', '服务器到 OpenAI 的网络连接未能建立。', '检查服务器出站 443、网络出口、主机防火墙，以及 WordPress 所在容器的网络。', $detail);
        }
        return self::issue('connection_failed', '服务器与 OpenAI 的请求未能完成，请管理员检测连接。', '先使用后台连接检测获取原因，再核对 WordPress/PHP 的网络、代理与证书配置。', $detail);
    }

    public static function http_error($status, $body, $api_key = '') {
        $error = is_array($body) && isset($body['error']) && is_array($body['error']) ? $body['error'] : array();
        $code = isset($error['code']) && is_scalar($error['code']) ? (string) $error['code'] : '';
        $type = isset($error['type']) && is_scalar($error['type']) ? (string) $error['type'] : '';
        $detail = self::clean_detail($error['message'] ?? '', $api_key);
        if ($detail === '') $detail = 'HTTP ' . (int) $status;
        if ((int) $status === 401) return self::issue('authentication_failed', 'OpenAI 拒绝了身份验证，请管理员核对 API Key 与项目权限。', '核对当前生效 Key 的来源、有效性、项目权限与项目/组织 IP 白名单。', $detail, $status);
        if ((int) $status === 403 && preg_match('/unsupported_country|unsupported_region|region_not_supported/', $code . ' ' . $type)) {
            return self::issue('region_not_supported', '当前请求地区不受 OpenAI API 支持，请管理员核对服务地区。', '核对 OpenAI 官方支持地区，以及实际服务器出口所在地区。', $detail, $status);
        }
        if ((int) $status === 403) return self::issue('permission_denied', '当前 API Key 没有这项请求的访问权限。', '核对项目、Key 的端点权限、模型权限及地区限制；模型读取与对话生成的权限分别判断。', $detail, $status);
        if ((int) $status === 404) return self::issue('model_unavailable', '模型不存在或当前 API Key 无权使用，请管理员核对模型 ID。', '确认模型 ID 与项目可访问模型一致，不要把网页产品名称当作 API 模型 ID。', $detail, $status);
        if ((int) $status === 429) {
            $billing = array('insufficient_quota', 'credit_balance_exhausted', 'billing_hard_limit_reached', 'organization_spend_limit_exceeded', 'project_spend_limit_exceeded', 'organization_usage_limit_exceeded');
            if (in_array($code, $billing, true) || $type === 'insufficient_quota') return self::issue('quota_exhausted', 'OpenAI API 额度或用量预算已达到限制，请管理员检查账单。', '检查 API 余额及组织/项目用量上限；更换页面口令或连续重试不会恢复额度。', $detail, $status);
            return self::issue('rate_limited', 'OpenAI API 请求受到速率限制，请稍后再试。', '按返回的 Retry-After 等待，再检查项目速率限制。', $detail, $status);
        }
        if ((int) $status >= 500) return self::issue('upstream_unavailable', 'OpenAI 服务暂时不可用，请稍后重试。', '稍后重试，并核对 OpenAI 服务状态。', $detail, $status);
        if ((int) $status >= 300 && (int) $status < 400) return self::issue('unexpected_redirect', '服务器收到意外跳转，请管理员检查网络出口。', '官方 API 请求不跟随跳转；检查代理登录页或出口网关重定向。', $detail, $status);
        if (!is_array($body)) return self::issue('invalid_response', '服务器收到非 API 格式的响应，请管理员检查网络出口。', '检查网关、代理或安全设备是否返回了 HTML 页面；不要关闭证书校验。', $detail, $status);
        return self::issue('request_rejected', 'OpenAI 未接受本次请求，请管理员核对模型与请求配置。', '查看此处的 API 错误详情，核对模型兼容性、端点权限与参数。', $detail, $status);
    }

    public static function diagnose($api_key, $model, $key_source, $request_timeout) {
        $report = array('ok' => false, 'key_source' => $key_source, 'model' => $model, 'request_timeout' => (int) $request_timeout, 'diagnostic_timeout' => 20, 'checked_at' => gmdate('c'), 'scope' => 'model_read', 'http_status' => 0, 'duration_ms' => 0, 'request_id' => '', 'proxy_configured' => defined('WP_PROXY_HOST') && (string) WP_PROXY_HOST !== '' && defined('WP_PROXY_PORT') && (int) WP_PROXY_PORT > 0);
        if ($api_key === '') return array_merge($report, self::issue('api_key_missing', '尚未读取到 API Key。', '先保存 API Key，再检测；wp-config.php 与 PHP 环境变量优先于后台设置。', ''));
        $started = microtime(true);
        $response = wp_remote_get('https://api.openai.com/v1/models/' . rawurlencode($model), array('timeout' => 20, 'redirection' => 0, 'sslverify' => true, 'limit_response_size' => 1048576, 'headers' => array('Authorization' => 'Bearer ' . $api_key, 'Accept' => 'application/json')));
        $report['duration_ms'] = (int) round((microtime(true) - $started) * 1000);
        if (is_wp_error($response)) return array_merge($report, self::transport_error($response, $api_key));
        $status = (int) wp_remote_retrieve_response_code($response);
        $body = json_decode(wp_remote_retrieve_body($response), true);
        $report['http_status'] = $status;
        $report['request_id'] = substr(preg_replace('/[^a-zA-Z0-9_-]/', '', self::clean_detail(wp_remote_retrieve_header($response, 'x-request-id'), $api_key)), 0, 128);
        if ($status === 200 && is_array($body) && !empty($body['id']) && is_string($body['id'])) {
            return array_merge($report, array('ok' => true, 'code' => 'model_read_ok', 'message' => __('已连接 OpenAI，当前模型信息可读取。', 'fieldnotes-chatgpt'), 'hint' => __('本检测不生成回复，不验证生成权限、API 余额或额度；请回到对话页发送一条消息完成验证。', 'fieldnotes-chatgpt'), 'technical' => ''));
        }
        return array_merge($report, self::http_error($status, $body, $api_key));
    }
}
