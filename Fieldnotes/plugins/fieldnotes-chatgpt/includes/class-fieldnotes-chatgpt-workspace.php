<?php
if (!defined('ABSPATH')) { exit; }

/** Server-owned file references and stateless Responses/MCP continuations. */
final class Fieldnotes_ChatGPT_Workspace {
    const TTL = 3600;
    private $settings;
    private $key;
    private $owner;

    public function __construct($settings, $key, $owner) {
        $this->settings = $settings;
        $this->key = $key;
        $this->owner = $owner;
    }

    private function error($code, $message, $status = 400) {
        return new WP_Error($code, __($message, 'fieldnotes-chatgpt'), array('status' => $status));
    }

    private function token() { return bin2hex(random_bytes(16)); }
    private function valid_token($value) { return is_string($value) && preg_match('/^[a-f0-9]{32}$/D', $value); }
    private function read($kind, $token) {
        if (!$this->valid_token($token)) return false;
        $value = get_transient('fnchat_' . $kind . '_' . $token);
        return is_array($value) && isset($value['owner'], $value['expires']) && hash_equals($this->owner, $value['owner']) && $value['expires'] > time() ? $value : false;
    }
    private function write($kind, $token, $value) {
        $value['owner'] = $this->owner;
        return set_transient('fnchat_' . $kind . '_' . $token, $value, max(1, $value['expires'] - time()));
    }

    /** Only fixed OpenAI endpoints are accepted. Secrets never enter the client response. */
    private function api($path, $method = 'POST', $body = null, $content_type = 'application/json') {
        if (!preg_match('~^/(?:responses|audio/transcriptions|files(?:/file-[a-zA-Z0-9_-]+)?)$~D', $path)) return $this->error('endpoint_invalid', '请求地址无效。');
        $args = array('method' => $method, 'timeout' => $method === 'DELETE' ? 2 : min(120, max(15, (int) $this->settings['request_timeout'])), 'redirection' => 0, 'sslverify' => true, 'limit_response_size' => 2097152,
            'headers' => array('Authorization' => 'Bearer ' . $this->key, 'Accept' => 'application/json'));
        if ($body !== null) {
            $args['headers']['Content-Type'] = $content_type;
            $args['body'] = $content_type === 'application/json' ? wp_json_encode($body) : $body;
            $args['data_format'] = 'body';
        }
        $response = wp_remote_request('https://api.openai.com/v1' . $path, $args);
        if (is_wp_error($response)) {
            $issue = Fieldnotes_ChatGPT_Connection::transport_error($response, $this->key);
            return $this->error($issue['code'], $issue['message'], in_array($issue['code'], array('request_timeout', 'connection_timeout'), true) ? 504 : 502);
        }
        $status = (int) wp_remote_retrieve_response_code($response);
        $data = json_decode(wp_remote_retrieve_body($response), true);
        if ($method === 'DELETE' && $status === 404) return array('deleted' => true);
        if ($status < 200 || $status >= 300 || !is_array($data)) {
            $issue = Fieldnotes_ChatGPT_Connection::http_error($status, $data, $this->key);
            return $this->error($issue['code'], $issue['message'], 502);
        }
        return $data;
    }

    private function multipart($path, $fields, $file, $mime, $name) {
        $boundary = 'fnchat-' . $this->token();
        $body = '';
        foreach ($fields as $field => $value) $body .= '--' . $boundary . "\r\nContent-Disposition: form-data; name=\"" . $field . "\"\r\n\r\n" . $value . "\r\n";
        // A deliberately plain transport filename prevents header injection and filename encoding issues.
        $body .= '--' . $boundary . "\r\nContent-Disposition: form-data; name=\"file\"; filename=\"" . $name . "\"\r\nContent-Type: " . $mime . "\r\n\r\n";
        $bytes = file_get_contents($file);
        if ($bytes === false) return $this->error('file_unreadable', '无法读取上传的文件。');
        $body .= $bytes . "\r\n--" . $boundary . "--\r\n";
        return $this->api($path, 'POST', $body, 'multipart/form-data; boundary=' . $boundary);
    }

    public static function public_servers($settings) {
        $out = array();
        foreach ((array) ($settings['mcp_servers'] ?? array()) as $server) {
            if (empty($server['enabled']) || empty($server['url']) || empty($server['label'])) continue;
            $out[] = array('label' => $server['label'], 'title' => $server['title'] ?: $server['label'], 'host' => (string) wp_parse_url($server['url'], PHP_URL_HOST));
        }
        return $out;
    }

    public static function valid_server_url($url) {
        $parts = wp_parse_url($url);
        if (!is_array($parts) || ($parts['scheme'] ?? '') !== 'https' || empty($parts['host']) || isset($parts['user']) || isset($parts['pass']) || isset($parts['fragment'])) return false;
        $host = strtolower(trim($parts['host'], '[]'));
        if ($host === 'localhost' || strpos($host, '.') === false && !filter_var($host, FILTER_VALIDATE_IP) || preg_match('/\.(?:localhost|local|internal)$/D', $host)) return false;
        if (filter_var($host, FILTER_VALIDATE_IP) && !filter_var($host, FILTER_VALIDATE_IP, FILTER_FLAG_NO_PRIV_RANGE | FILTER_FLAG_NO_RES_RANGE)) return false;
        return strlen($url) <= 2048 && !preg_match('/[\x00-\x20]/', $url);
    }

    private function tools($labels) {
        $tools = array();
        foreach ((array) ($this->settings['mcp_servers'] ?? array()) as $server) {
            if (empty($server['enabled']) || !in_array($server['label'], $labels, true) || !self::valid_server_url($server['url'])) continue;
            $tool = array('type' => 'mcp', 'server_label' => $server['label'], 'server_url' => $server['url'], 'require_approval' => 'always');
            if (!empty($server['token'])) $tool['authorization'] = $server['token'];
            if (!empty($server['allowed_tools'])) $tool['allowed_tools'] = $server['allowed_tools'];
            $tools[] = $tool;
        }
        if (count($tools) !== count($labels)) return $this->error('mcp_changed', '工具设置已变更，请重新选择连接的工具。');
        return $tools;
    }

    private function tool_signature($tools) { return hash('sha256', wp_json_encode($tools)); }

    private function call_allowed($request, $tools) {
        foreach ($tools as $tool) {
            if ($tool['server_label'] !== ($request['server_label'] ?? $request['server'] ?? '')) continue;
            return empty($tool['allowed_tools']) || in_array($request['name'] ?? '', $tool['allowed_tools'], true);
        }
        return false;
    }

    /** The PHP upload temp file is never moved to the public WordPress media directory. */
    public function upload($file, $meta = array()) {
        if (empty($this->settings['uploads_enabled'])) return $this->error('uploads_disabled', '附件上传暂未开启。', 403);
        if (!is_array($file) || !isset($file['error'], $file['tmp_name'], $file['name']) || !is_scalar($file['error']) || !is_string($file['tmp_name']) || !is_string($file['name'])) return $this->error('file_missing', '请选择一个文件。');
        if ((int) $file['error'] !== UPLOAD_ERR_OK) return $this->error('upload_failed', '文件未完整上传。请核对服务器上传大小限制后重试。', 413);
        if (!is_uploaded_file($file['tmp_name'])) return $this->error('upload_invalid', '上传文件无效。');
        $size = (int) filesize($file['tmp_name']);
        $limit = min(20 * 1048576, (int) $this->settings['max_upload_mb'] * 1048576, wp_max_upload_size());
        if ($size < 1 || $size > $limit) return $this->error('file_too_large', '文件超过本站上传大小限制。', 413);
        $name = sanitize_file_name($file['name']);
        $ext = strtolower(pathinfo($name, PATHINFO_EXTENSION));
        $images = array('jpg' => 'image/jpeg', 'jpeg' => 'image/jpeg', 'png' => 'image/png', 'webp' => 'image/webp', 'gif' => 'image/gif');
        $documents = array('pdf' => 'application/pdf', 'txt' => 'text/plain', 'md' => 'text/markdown', 'csv' => 'text/csv', 'tsv' => 'text/tab-separated-values', 'json' => 'application/json', 'doc' => 'application/msword', 'docx' => 'application/vnd.openxmlformats-officedocument.wordprocessingml.document', 'ppt' => 'application/vnd.ms-powerpoint', 'pptx' => 'application/vnd.openxmlformats-officedocument.presentationml.presentation', 'xls' => 'application/vnd.ms-excel', 'xlsx' => 'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet');
        $audio = array('mp3' => 'audio/mpeg', 'mpga' => 'audio/mpeg', 'mpeg' => 'audio/mpeg', 'm4a' => 'audio/mp4', 'wav' => 'audio/wav', 'webm' => 'audio/webm');
        $kind = isset($images[$ext]) ? 'image' : (isset($documents[$ext]) ? 'document' : (isset($audio[$ext]) ? 'audio' : ''));
        if (!$kind) return $this->error('file_type', '此格式暂不支持，请使用图片、文档或音频。');
        if ($kind === 'audio' && empty($this->settings['audio_enabled'])) return $this->error('audio_disabled', '音频转写暂未开启。');
        $mime = $kind === 'image' ? $images[$ext] : ($kind === 'document' ? $documents[$ext] : $audio[$ext]);
        $detected = function_exists('mime_content_type') ? (string) mime_content_type($file['tmp_name']) : '';
        if ($kind === 'image') {
            $info = @getimagesize($file['tmp_name']);
            if (!$info || ($info['mime'] ?? '') !== $mime || $info[0] * $info[1] > 40000000) return $this->error('image_invalid', '图片格式无效或尺寸过大，请重新导出后上传。');
        } elseif ($ext === 'pdf' && strpos((string) file_get_contents($file['tmp_name'], false, null, 0, 5), '%PDF-') !== 0) {
            return $this->error('document_invalid', 'PDF 文件格式无效。');
        } elseif (in_array($ext, array('docx', 'pptx', 'xlsx'), true) && file_get_contents($file['tmp_name'], false, null, 0, 2) !== 'PK') {
            return $this->error('document_invalid', 'Office 文档格式无效。');
        } elseif ($detected && preg_match('~(?:x-php|x-executable|x-sharedlib|x-dosexec|x-msdownload|image/svg)~i', $detected)) {
            return $this->error('file_type', '此文件内容不属于支持的格式。');
        }
        $record = array('expires' => time() + self::TTL, 'name' => $name, 'kind' => $kind, 'size' => $size, 'file_id' => '', 'parts' => array());
        if ($kind === 'audio') {
            $data = $this->multipart('/audio/transcriptions', array('model' => $this->settings['transcription_model'], 'response_format' => 'json'), $file['tmp_name'], $mime, 'audio.' . $ext);
            if (is_wp_error($data)) return $data;
            $text = isset($data['text']) && is_string($data['text']) ? trim($data['text']) : '';
            if ($text === '') return $this->error('transcription_empty', '没有识别到语音，请更换清晰的音频。', 502);
            if (strlen($text) > 60000) return $this->error('transcription_long', '录音转写过长，请拆分后上传。', 413);
            $record['transcript'] = $text;
            $record['parts'][] = array('type' => 'input_text', 'text' => '音频「' . $name . '」的语音转写（不包含音色、音乐或其他声学分析）：' . "\n" . $text);
        } else {
            $data = $this->multipart('/files', array('purpose' => 'user_data', 'expires_after[anchor]' => 'created_at', 'expires_after[seconds]' => self::TTL), $file['tmp_name'], $mime, 'attachment.' . $ext);
            if (is_wp_error($data)) return $data;
            if (empty($data['id']) || !is_string($data['id']) || !preg_match('/^file-[a-zA-Z0-9_-]+$/D', $data['id'])) return $this->error('upload_response', '附件服务没有返回有效文件标识。', 502);
            $record['file_id'] = $data['id'];
            $record['expires'] = min($record['expires'], (int) ($data['expires_at'] ?? $record['expires']));
            $record['parts'][] = array('type' => 'input_text', 'text' => '附件名称：' . $name);
            $record['parts'][] = $kind === 'image' ? array('type' => 'input_image', 'file_id' => $data['id'], 'detail' => 'auto') : array('type' => 'input_file', 'file_id' => $data['id']);
            if (($meta['kind'] ?? '') === 'video_frames' && $kind === 'image') {
                $record['kind'] = 'video';
                $record['name'] = sanitize_file_name((string) ($meta['source_name'] ?? $name));
                $record['parts'][] = array('type' => 'input_text', 'text' => '视频「' . $record['name'] . '」的四张等间隔关键帧拼图，时间标注在画面内。仅分析这些画面，不推断完整视频内容、动作连续性或声音。');
            }
        }
        $ref = $this->token();
        if (!$this->write('file', $ref, $record)) {
            if ($record['file_id']) $this->api('/files/' . $record['file_id'], 'DELETE');
            return $this->error('storage_failed', '附件状态未能保存，请重试。', 503);
        }
        return array('ref' => $ref, 'name' => $record['name'], 'kind' => $record['kind'], 'size' => $size, 'expires' => $record['expires'], 'transcript' => $record['transcript'] ?? '');
    }

    public function remove_file($ref) {
        $file = $this->read('file', $ref);
        if (!$file) return $this->error('file_expired', '附件已过期或不属于当前访问。', 404);
        if (!empty($file['used'])) return $this->error('file_in_use', '此附件已在对话中使用，请通过新对话一并清理。', 409);
        if ($file['file_id']) {
            $result = $this->api('/files/' . $file['file_id'], 'DELETE');
            if (is_wp_error($result)) {
                delete_transient('fnchat_file_' . $ref);
                return array('removed' => true, 'cleanup_pending' => true);
            }
        }
        delete_transient('fnchat_file_' . $ref);
        return array('removed' => true);
    }

    private function state($token) {
        if ($token === '') return array('owner' => $this->owner, 'expires' => time() + self::TTL, 'input' => array(), 'files' => array(), 'bytes' => 0, 'turns' => 0, 'pending' => array(), 'revision' => '', 'labels' => array());
        $state = $this->read('conversation', $token);
        return $state ?: $this->error('conversation_expired', '对话上下文已过期，请开启新对话；本次草稿会保留。', 409);
    }

    private function acquire($token) {
        $name = 'fnchat_busy_' . $token;
        $old = get_option($name);
        if (is_array($old) && ($old['until'] ?? 0) < time()) delete_option($name);
        return add_option($name, array('until' => time() + 180), '', false);
    }

    public function chat($conversation, $message, $refs, $labels) {
        if (!is_string($conversation) || ($conversation !== '' && !$this->valid_token($conversation)) || !is_string($message) || strlen($message) > 20000 || !is_array($refs) || count($refs) > 4 || !is_array($labels) || count($labels) > 4) return $this->error('message_invalid', '消息格式或附件数量无效。');
        foreach ($refs as $ref) if (!$this->valid_token($ref)) return $this->error('file_invalid', '附件标识无效。');
        if (trim($message) === '' && !$refs) return $this->error('message_empty', '写下想法，或添加一份附件。');
        foreach ($labels as $label) if (!is_string($label) || !preg_match('/^[a-zA-Z0-9_-]{1,40}$/D', $label)) return $this->error('tools_invalid', '工具选择无效。');
        $labels = array_values(array_unique($labels));
        $is_new = $conversation === '';
        $conversation = $conversation ?: $this->token();
        if (!$this->acquire($conversation)) return $this->error('request_busy', '上一条请求尚未完成，请稍候。', 409);
        try {
            $state = $this->state($is_new ? '' : $conversation);
            if (is_wp_error($state)) return $state;
            if (!empty($state['pending']) || !empty($state['uncertain'])) return $this->error('approval_pending', '请先处理工具确认，或开启新对话。', 409);
            if ($state['turns'] >= 24) return $this->error('conversation_full', '这段对话已经很长，请开启新对话。', 413);
            $parts = array(array('type' => 'input_text', 'text' => trim($message) ?: '请阅读并概括这些附件。'));
            foreach (array_unique($refs) as $ref) {
                $file = $this->read('file', $ref);
                if (!$file || (!empty($file['used']) && $file['used'] !== $conversation)) return $this->error('file_expired', '附件已过期或不属于当前对话，请重新添加。', 409);
                if (!in_array($ref, $state['files'], true)) { $state['files'][] = $ref; $state['bytes'] += $file['size']; }
                $state['expires'] = min($state['expires'], $file['expires']);
                $parts = array_merge($parts, $file['parts']);
            }
            if ($state['bytes'] > 40 * 1048576 || count($state['files']) > 12) return $this->error('attachments_full', '本段对话附件已达到总量限制，请开启新对话。', 413);
            $state['input'][] = array('role' => 'user', 'content' => $parts);
            $state['labels'] = $labels;
            $state['turns']++;
            return $this->complete($conversation, $state);
        } finally { delete_option('fnchat_busy_' . $conversation); }
    }

    public function approve($conversation, $revision, $decisions) {
        if (!$this->valid_token($conversation) || !$this->valid_token($revision) || !is_array($decisions)) return $this->error('approval_invalid', '工具确认格式无效。');
        if (!$this->acquire($conversation)) return $this->error('request_busy', '工具确认正在处理中，请稍候。', 409);
        try {
            $state = $this->state($conversation);
            if (is_wp_error($state)) return $state;
            if (empty($state['pending']) || !hash_equals($state['revision'], $revision) || !empty($state['uncertain']) || count($decisions) !== count($state['pending'])) return $this->error('approval_stale', '这次确认已经失效或已提交，请开启新对话。', 409);
            $tools = $this->tools($state['labels']);
            if (is_wp_error($tools) || !hash_equals($state['tool_signature'] ?? '', $this->tool_signature($tools))) return $this->error('mcp_changed', '工具配置已经变更，本次确认尚未提交。请开启新对话。', 409);
            foreach ($state['pending'] as $pending) {
                if (!$this->call_allowed($pending, $tools)) return $this->error('tool_not_allowed', '这项工具不在当前允许范围内，本次确认尚未提交。', 403);
                if (!array_key_exists($pending['id'], $decisions) || !is_bool($decisions[$pending['id']])) return $this->error('approval_invalid', '请逐项选择允许或拒绝。');
                $state['input'][] = array('type' => 'mcp_approval_response', 'approval_request_id' => $pending['id'], 'approve' => $decisions[$pending['id']]);
            }
            // Consume the decision before the network call. A timeout must not replay an action.
            $state['uncertain'] = true;
            if (!$this->write('conversation', $conversation, $state)) return $this->error('storage_failed', '无法保存工具确认，本次操作尚未提交。', 503);
            $result = $this->complete($conversation, $state);
            if (is_wp_error($result)) return $this->error('tool_result_unknown', '未能确认工具执行结果。请先到对应服务检查，再开启新对话，避免重复操作。', 502);
            return $result;
        } finally { delete_option('fnchat_busy_' . $conversation); }
    }

    private function complete($conversation, $state) {
        $tools = $this->tools($state['labels']);
        if (is_wp_error($tools)) return $tools;
        $signature = $this->tool_signature($tools);
        if (isset($state['tool_signature']) && !hash_equals($state['tool_signature'], $signature)) return $this->error('mcp_changed', '工具配置已经变更，请开启新对话后重新选择。', 409);
        $state['tool_signature'] = $signature;
        if (strlen(wp_json_encode($state['input'])) > 1200000) return $this->error('conversation_full', '对话上下文已达到限制，请开启新对话。', 413);
        $payload = array('model' => $this->settings['model'], 'instructions' => $this->settings['instructions'], 'input' => $state['input'], 'max_output_tokens' => (int) $this->settings['max_output_tokens'], 'store' => false, 'include' => array('reasoning.encrypted_content'));
        if ($tools) $payload['tools'] = $tools;
        $data = $this->api('/responses', 'POST', $payload);
        if (is_wp_error($data)) return $data;
        if (!empty($data['error']) || ($data['status'] ?? '') === 'failed') return $this->error('response_failed', '本次生成未完成，请开启新对话后重试。', 502);
        $output = isset($data['output']) && is_array($data['output']) ? $data['output'] : array();
        $text = isset($data['output_text']) && is_string($data['output_text']) ? $data['output_text'] : '';
        $pending = array(); $activity = array();
        foreach ($output as $item) {
            if (!is_array($item) || empty($item['type'])) continue;
            if ($item['type'] === 'message' && !isset($data['output_text'])) foreach ((array) ($item['content'] ?? array()) as $part) {
                if (($part['type'] ?? '') === 'output_text' && is_string($part['text'] ?? null)) $text .= $part['text'];
                if (($part['type'] ?? '') === 'refusal' && is_string($part['refusal'] ?? null)) $text .= $part['refusal'];
            }
            if ($item['type'] === 'mcp_approval_request') {
                if (!is_string($item['id'] ?? null) || !is_string($item['arguments'] ?? null) || !in_array($item['server_label'] ?? '', $state['labels'], true) || !$this->call_allowed($item, $tools)) return $this->error('approval_invalid', '工具服务返回的确认信息无效或超出允许范围，请开启新对话。', 502);
                $host = '';
                foreach (self::public_servers($this->settings) as $server) if ($server['label'] === $item['server_label']) $host = $server['host'];
                $pending[] = array('id' => $item['id'], 'name' => sanitize_text_field($item['name'] ?? ''), 'server' => $item['server_label'], 'host' => $host, 'arguments' => $item['arguments']);
            }
            if ($item['type'] === 'mcp_call') $activity[] = array('server' => sanitize_text_field($item['server_label'] ?? ''), 'name' => sanitize_text_field($item['name'] ?? ''), 'ok' => empty($item['error']));
            if ($item['type'] === 'mcp_list_tools' && !empty($item['error'])) $activity[] = array('server' => sanitize_text_field($item['server_label'] ?? ''), 'name' => '读取工具列表', 'ok' => false);
        }
        if (!$pending && trim($text) === '' && !$activity) return $this->error('response_empty', '模型没有返回可显示的内容，请重试。', 502);
        if (!$output && trim($text) !== '') $output[] = array('role' => 'assistant', 'content' => $text);
        $state['input'] = array_merge($state['input'], $output);
        $state['pending'] = $pending;
        $state['revision'] = $pending ? $this->token() : '';
        $state['uncertain'] = false;
        if (!$this->write('conversation', $conversation, $state)) return $this->error('storage_failed', '生成已完成，但上下文未能保存，请开启新对话。', 503);
        foreach ($state['files'] as $ref) {
            $file = $this->read('file', $ref);
            if ($file) { $file['used'] = $conversation; $this->write('file', $ref, $file); }
        }
        return array('conversation' => $conversation, 'text' => trim($text), 'approvals' => $pending, 'revision' => $state['revision'], 'activity' => $activity, 'incomplete' => ($data['status'] ?? '') === 'incomplete');
    }

    public function reset($conversation) {
        if (!$this->valid_token($conversation)) return $this->error('conversation_invalid', '对话标识无效。');
        if (!$this->acquire($conversation)) return $this->error('request_busy', '请求尚未结束，请稍候。', 409);
        try {
            $state = $this->read('conversation', $conversation);
            if (!$state) return array('reset' => true, 'cleanup_pending' => false);
            $pending = false;
            foreach ($state['files'] as $ref) {
                $file = $this->read('file', $ref);
                if (!$file) continue;
                if ($file['file_id'] && is_wp_error($this->api('/files/' . $file['file_id'], 'DELETE'))) $pending = true;
                delete_transient('fnchat_file_' . $ref);
            }
            delete_transient('fnchat_conversation_' . $conversation);
            return array('reset' => true, 'cleanup_pending' => $pending);
        } finally { delete_option('fnchat_busy_' . $conversation); }
    }
}
