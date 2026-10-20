<?php
if (!defined('ABSPATH')) { exit; }

final class Fieldnotes_ChatGPT {
    const OPTION = 'fieldnotes_chatgpt_settings';
    const GROUP = 'fieldnotes_chatgpt_group';
    const COOKIE = 'fieldnotes_chat_access';
    const PAGE_SLUG = 'chatgpt';
    const SHORTCODE = 'fieldnotes_chatgpt';

    private static $instance = null;

    public static function instance() {
        if (null === self::$instance) {
            self::$instance = new self();
        }
        return self::$instance;
    }

    public static function activate() {
        $page = get_page_by_path(self::PAGE_SLUG, OBJECT, 'page');
        if (!$page) {
            $page_id = wp_insert_post(array(
                'post_type' => 'page',
                'post_status' => 'publish',
                'post_title' => __('ChatGPT 对话', 'fieldnotes-chatgpt'),
                'post_name' => self::PAGE_SLUG,
                'post_content' => '[' . self::SHORTCODE . ']',
            ));
            if (!is_wp_error($page_id) && $page_id) {
                self::set_theme_template($page_id);
                update_option('fieldnotes_chatgpt_page_id', (int) $page_id, false);
            }
        } else {
            $page_id = (int) $page->ID;
            if (!has_shortcode($page->post_content, self::SHORTCODE)) {
                wp_update_post(array(
                    'ID' => $page_id,
                    'post_content' => trim($page->post_content . "\n\n[" . self::SHORTCODE . "]"),
                ));
            }
            self::set_theme_template($page_id);
            update_option('fieldnotes_chatgpt_page_id', $page_id, false);
        }
        flush_rewrite_rules(false);
    }

    private static function set_theme_template($page_id) {
        if ('fieldnotes-journal' === get_option('stylesheet')) {
            update_post_meta($page_id, '_wp_page_template', 'templates/page-chatgpt.php');
        }
    }

    private function __construct() {
        add_action('init', array($this, 'register_shortcode'));
        add_action('wp_enqueue_scripts', array($this, 'enqueue_assets'), 20);
        add_action('template_redirect', array($this, 'prevent_page_caching'), 0);
        add_action('admin_menu', array($this, 'admin_menu'));
        add_action('admin_init', array($this, 'register_settings'));
        add_action('admin_enqueue_scripts', array($this, 'enqueue_admin_assets'));
        add_action('wp_ajax_fieldnotes_chatgpt_diagnose', array($this, 'ajax_diagnose'));
        add_filter('body_class', array($this, 'body_classes'));
        add_filter('wp_nav_menu_items', array($this, 'append_navigation_item'), 15, 2);
        add_filter('the_content', array($this, 'deduplicate_chat_output'), PHP_INT_MAX);
        add_action('wp_ajax_fieldnotes_chatgpt_unlock', array($this, 'ajax_unlock'));
        add_action('wp_ajax_nopriv_fieldnotes_chatgpt_unlock', array($this, 'ajax_unlock'));
        add_action('wp_ajax_fieldnotes_chatgpt_chat', array($this, 'ajax_chat'));
        add_action('wp_ajax_nopriv_fieldnotes_chatgpt_chat', array($this, 'ajax_chat'));
        add_action('wp_ajax_fieldnotes_chatgpt_lock', array($this, 'ajax_lock'));
        add_action('wp_ajax_nopriv_fieldnotes_chatgpt_lock', array($this, 'ajax_lock'));
        foreach (array('upload', 'remove_file', 'approve', 'reset') as $action) {
            add_action('wp_ajax_fieldnotes_chatgpt_' . $action, array($this, 'ajax_' . $action));
            add_action('wp_ajax_nopriv_fieldnotes_chatgpt_' . $action, array($this, 'ajax_' . $action));
        }
    }

    private function defaults() {
        return array(
            'api_key' => '',
            'password_hash' => '',
            'model' => 'gpt-6-luna',
            'instructions' => '你是一个清晰、友好、可靠的中文助手。请先直接回答，再根据需要补充简短说明；遇到不确定的信息要明确说明。',
            'max_output_tokens' => 1200,
            'allow_http_testing' => 1,
            'request_timeout' => 60,
            'uploads_enabled' => 1,
            'max_upload_mb' => 10,
            'audio_enabled' => 1,
            'transcription_model' => 'gpt-transcribe',
            'mcp_servers' => array(),
        );
    }

    private function settings() {
        $stored = get_option(self::OPTION, array());
        return wp_parse_args(is_array($stored) ? $stored : array(), $this->defaults());
    }

    private function http_testing_allowed($settings = null) {
        if (defined('FIELDNOTES_CHATGPT_ALLOW_HTTP')) return (bool) FIELDNOTES_CHATGPT_ALLOW_HTTP;
        $settings = is_array($settings) ? $settings : $this->settings();
        return !empty($settings['allow_http_testing']);
    }

    private function transport_allowed() {
        return is_ssl() || $this->http_testing_allowed();
    }

    private function chat_page() {
        $page = get_post((int) get_option('fieldnotes_chatgpt_page_id'));
        if (!$page || $page->post_type !== 'page' || $page->post_status !== 'publish' || !has_shortcode($page->post_content, self::SHORTCODE)) {
            $page = get_page_by_path(self::PAGE_SLUG, OBJECT, 'page');
        }
        return $page && $page->post_type === 'page' && $page->post_status === 'publish' && has_shortcode($page->post_content, self::SHORTCODE) ? $page : null;
    }

    private function api_key() {
        if (defined('FIELDNOTES_CHATGPT_API_KEY') && trim((string) FIELDNOTES_CHATGPT_API_KEY) !== '') {
            return trim((string) FIELDNOTES_CHATGPT_API_KEY);
        }
        $environment_key = getenv('OPENAI_API_KEY');
        if (is_string($environment_key) && '' !== trim($environment_key)) {
            return trim($environment_key);
        }
        $settings = $this->settings();
        return trim((string) $settings['api_key']);
    }

    private function api_key_source() {
        if (defined('FIELDNOTES_CHATGPT_API_KEY') && trim((string) FIELDNOTES_CHATGPT_API_KEY) !== '') return 'wp-config.php';
        $key = getenv('OPENAI_API_KEY');
        if (is_string($key) && trim($key) !== '') return 'OPENAI_API_KEY';
        return $this->api_key() !== '' ? __('后台设置', 'fieldnotes-chatgpt') : __('未设置', 'fieldnotes-chatgpt');
    }

    public function body_classes($classes) {
        $post = get_queried_object();
        if (is_singular() && $post instanceof WP_Post && has_shortcode($post->post_content, self::SHORTCODE)) $classes[] = 'fnchat-page';
        return array_values(array_unique($classes));
    }

    public function enqueue_admin_assets($hook) {
        if ($hook !== 'settings_page_fieldnotes-chatgpt') return;
        wp_enqueue_style('fieldnotes-chatgpt-admin', FIELDNOTES_CHATGPT_URL . 'assets/css/admin.css', array(), FIELDNOTES_CHATGPT_VERSION);
        wp_enqueue_script('fieldnotes-chatgpt-admin', FIELDNOTES_CHATGPT_URL . 'assets/js/admin.js', array(), FIELDNOTES_CHATGPT_VERSION, true);
        wp_localize_script('fieldnotes-chatgpt-admin', 'fieldnotesChatGPTAdmin', array('ajaxUrl' => admin_url('admin-ajax.php', 'relative'), 'nonce' => wp_create_nonce('fieldnotes_chatgpt_diagnose'), 'working' => __('正在检测…', 'fieldnotes-chatgpt'), 'error' => __('检测请求未完成，请刷新后台重试。', 'fieldnotes-chatgpt')));
    }

    public function ajax_diagnose() {
        if (!current_user_can('manage_options')) wp_send_json_error(array('message' => __('仅管理员可以检测连接。', 'fieldnotes-chatgpt')), 403);
        $nonce = isset($_POST['nonce']) && is_string($_POST['nonce']) ? sanitize_text_field(wp_unslash($_POST['nonce'])) : '';
        if (!wp_verify_nonce($nonce, 'fieldnotes_chatgpt_diagnose')) wp_send_json_error(array('message' => __('检测页面已过期，请刷新后重试。', 'fieldnotes-chatgpt')), 403);
        $settings = $this->settings();
        wp_send_json_success(Fieldnotes_ChatGPT_Connection::diagnose($this->api_key(), $settings['model'], $this->api_key_source(), $settings['request_timeout']));
    }

    public function register_shortcode() {
        add_shortcode(self::SHORTCODE, array($this, 'render_shortcode'));
    }

    public function enqueue_assets() {
        if (!is_singular()) { return; }
        $post = get_queried_object();
        if (!$post instanceof WP_Post || !has_shortcode($post->post_content, self::SHORTCODE)) { return; }

        wp_enqueue_style('fieldnotes-chatgpt', FIELDNOTES_CHATGPT_URL . 'assets/css/chat.css', array(), FIELDNOTES_CHATGPT_VERSION);
        wp_enqueue_script('fieldnotes-chatgpt-media', FIELDNOTES_CHATGPT_URL . 'assets/js/media.js', array(), FIELDNOTES_CHATGPT_VERSION, true);
        wp_enqueue_script('fieldnotes-chatgpt', FIELDNOTES_CHATGPT_URL . 'assets/js/chat.js', array('fieldnotes-chatgpt-media'), FIELDNOTES_CHATGPT_VERSION, true);
        $settings = $this->settings();
        wp_localize_script('fieldnotes-chatgpt', 'fieldnotesChatGPT', array(
            'ajaxUrl' => admin_url('admin-ajax.php', 'relative'),
            'nonce' => wp_create_nonce('fieldnotes_chatgpt'),
            'uploadLimit' => min(20 * 1048576, (int) $settings['max_upload_mb'] * 1048576, wp_max_upload_size()),
            'uploadsEnabled' => !empty($settings['uploads_enabled']),
            'audioEnabled' => !empty($settings['audio_enabled']),
            'servers' => Fieldnotes_ChatGPT_Workspace::public_servers($settings),
            'strings' => array(
                'working' => __('正在连接…', 'fieldnotes-chatgpt'),
                'send' => __('发送', 'fieldnotes-chatgpt'),
                'networkError' => __('连接暂时失败，请稍后重试。', 'fieldnotes-chatgpt'),
                'empty' => __('先写点内容再发送。', 'fieldnotes-chatgpt'),
                'locked' => __('口令错误或访问已过期，请重试。', 'fieldnotes-chatgpt'),
                'replyError' => __('这条消息暂时没有发送成功，请稍后重试。', 'fieldnotes-chatgpt'),
                'welcomeTitle' => __('有什么想聊的？', 'fieldnotes-chatgpt'),
                'welcomeCopy' => __('对话只保留在当前浏览器页面中，刷新或关闭后会清空。', 'fieldnotes-chatgpt'),
                'copied' => __('已复制', 'fieldnotes-chatgpt'),
                'copy' => __('复制回复', 'fieldnotes-chatgpt'),
                'enter' => __('进入对话', 'fieldnotes-chatgpt'),
                'passwordRequired' => __('请先输入访问口令。', 'fieldnotes-chatgpt'),
                'copyFailed' => __('复制未完成，请选择文字复制。', 'fieldnotes-chatgpt'),
                'thinking' => __('正在思考…', 'fieldnotes-chatgpt'),
                'retry' => __('草稿与附件已保留，可以再次发送。', 'fieldnotes-chatgpt'),
            ),
        ));
    }

    public function prevent_page_caching() {
        $post = get_queried_object();
        if (is_singular() && $post instanceof WP_Post && has_shortcode($post->post_content, self::SHORTCODE)) {
            if (!defined('DONOTCACHEPAGE')) { define('DONOTCACHEPAGE', true); }
            nocache_headers();
            if (!headers_sent()) {
                header('Cache-Control: private, no-store, no-cache, must-revalidate, max-age=0', true);
            }
        }
    }

    public function render_shortcode() {
        $page = $this->chat_page();
        $page_id = $page ? (int) $page->ID : 0;
        $settings = $this->settings();
        $password_ready = !empty($settings['password_hash']);
        $authorized = $password_ready && $page_id && $this->has_access($page_id);
        $api_ready = '' !== $this->api_key();
        ob_start();
        include FIELDNOTES_CHATGPT_DIR . 'views/chat.php';
        return '<!--fieldnotes-chatgpt:start-->' . ob_get_clean() . '<!--fieldnotes-chatgpt:end-->';
    }

    /**
     * Keep one chat in each content result without changing the stored page.
     * No request-wide render flag: SEO or preview rendering must not consume
     * the chat that the actual page will subsequently display.
     */
    public function deduplicate_chat_output($content) {
        if (!is_string($content) || strpos($content, '<!--fieldnotes-chatgpt:start-->') === false) return $content;
        $seen = false;
        $result = preg_replace_callback('~<!--fieldnotes-chatgpt:start-->.*?<!--fieldnotes-chatgpt:end-->~s', function ($match) use (&$seen) {
            if ($seen) return '';
            $seen = true;
            return $match[0];
        }, $content);
        return is_string($result) ? $result : $content;
    }

    public function admin_menu() {
        add_options_page(__('ChatGPT 接入', 'fieldnotes-chatgpt'), __('ChatGPT 接入', 'fieldnotes-chatgpt'), 'manage_options', 'fieldnotes-chatgpt', array($this, 'render_settings_page'));
    }

    public function register_settings() {
        register_setting(self::GROUP, self::OPTION, array(
            'type' => 'array',
            'sanitize_callback' => array($this, 'sanitize_settings'),
            'default' => $this->defaults(),
        ));
    }

    public function sanitize_settings($input) {
        $old = $this->settings();
        $input = is_array($input) ? wp_unslash($input) : array();
        $new = $old;
        $new['allow_http_testing'] = !empty($input['allow_http_testing']) ? 1 : 0;

        $key = isset($input['api_key']) && is_scalar($input['api_key']) ? trim((string) $input['api_key']) : '';
        $password = isset($input['page_password']) && is_scalar($input['page_password']) ? (string) $input['page_password'] : '';
        if (!is_ssl() && !$this->http_testing_allowed($new) && ('' !== $key || '' !== $password)) {
            add_settings_error(self::OPTION, 'https_required', __('请先启用 HTTPS，再保存 API Key 或页面口令。', 'fieldnotes-chatgpt'), 'error');
            $key = '';
            $password = '';
        }
        if (!empty($input['clear_api_key'])) {
            $new['api_key'] = '';
        } elseif ('' !== $key) {
            $new['api_key'] = sanitize_text_field($key);
        }

        if ('' !== $password) {
            $length = function_exists('mb_strlen') ? mb_strlen($password, 'UTF-8') : strlen($password);
            if ($length < 8) {
                add_settings_error(self::OPTION, 'password_short', __('页面口令至少需要 8 个字符。', 'fieldnotes-chatgpt'), 'error');
            } else {
                $new['password_hash'] = password_hash($password, PASSWORD_DEFAULT);
            }
        }

        $model = isset($input['model']) && is_scalar($input['model']) ? sanitize_text_field((string) $input['model']) : $old['model'];
        if (preg_match('/^[a-zA-Z0-9._-]{2,80}$/', $model)) {
            $new['model'] = $model;
        } else {
            add_settings_error(self::OPTION, 'model_invalid', __('模型 ID 格式无效，已保留原设置。', 'fieldnotes-chatgpt'), 'error');
        }

        if (isset($input['instructions'])) {
            $instructions = is_scalar($input['instructions']) ? sanitize_textarea_field((string) $input['instructions']) : $old['instructions'];
            if (strlen($instructions) > 6000) { $instructions = substr($instructions, 0, 6000); }
            $new['instructions'] = $instructions;
        }

        $tokens = isset($input['max_output_tokens']) && is_scalar($input['max_output_tokens']) ? absint($input['max_output_tokens']) : (int) $old['max_output_tokens'];
        $new['max_output_tokens'] = min(4096, max(256, $tokens));
        $timeout = isset($input['request_timeout']) && is_scalar($input['request_timeout']) ? absint($input['request_timeout']) : (int) $old['request_timeout'];
        $new['request_timeout'] = min(120, max(15, $timeout));
        $new['uploads_enabled'] = !empty($input['uploads_enabled']) ? 1 : 0;
        $new['audio_enabled'] = !empty($input['audio_enabled']) ? 1 : 0;
        $new['max_upload_mb'] = min(20, max(1, (int) ($input['max_upload_mb'] ?? $old['max_upload_mb'])));
        $transcription = isset($input['transcription_model']) && is_string($input['transcription_model']) ? trim($input['transcription_model']) : $old['transcription_model'];
        if (preg_match('/^[a-zA-Z0-9._-]{2,80}$/D', $transcription)) $new['transcription_model'] = $transcription;
        $servers = array();
        foreach (array_slice(is_array($input['mcp_servers'] ?? null) ? $input['mcp_servers'] : array(), 0, 4) as $row) {
            if (!is_array($row)) continue;
            $url = is_string($row['url'] ?? null) ? trim($row['url']) : '';
            $label = is_string($row['label'] ?? null) ? trim($row['label']) : '';
            if ($url === '' && $label === '') continue;
            $previous = array();
            foreach ((array) $old['mcp_servers'] as $entry) if ($entry['label'] === $label) $previous = $entry;
            if (!preg_match('/^[a-zA-Z0-9_-]{1,40}$/D', $label) || !Fieldnotes_ChatGPT_Workspace::valid_server_url($url) || isset($servers[$label])) {
                add_settings_error(self::OPTION, 'mcp_invalid', __('MCP 需要唯一英文标识与公网 HTTPS 地址，请检查后重新保存。', 'fieldnotes-chatgpt'), 'error');
                if ($previous) $servers[$label] = $previous;
                continue;
            }
            $token = is_string($row['token'] ?? null) ? trim($row['token']) : '';
            if (!$this->transport_allowed() && $token !== '') $token = '';
            if ($token === '' && empty($row['clear_token']) && ($previous['url'] ?? '') === $url) $token = $previous['token'] ?? '';
            $allowed = is_string($row['allowed_tools'] ?? null) ? preg_split('/[\s,，]+/u', $row['allowed_tools'], -1, PREG_SPLIT_NO_EMPTY) : array();
            $allowed = array_slice(array_values(array_unique(array_filter($allowed, function ($name) { return preg_match('/^[a-zA-Z0-9_.:\/-]{1,128}$/D', $name); }))), 0, 50);
            $servers[$label] = array('enabled' => !empty($row['enabled']) ? 1 : 0, 'label' => $label, 'title' => sanitize_text_field(is_scalar($row['title'] ?? null) ? (string) $row['title'] : $label), 'url' => $url, 'token' => preg_replace('/[\r\n]/', '', substr($token, 0, 8192)), 'allowed_tools' => $allowed);
        }
        $new['mcp_servers'] = array_values($servers);
        return $new;
    }

    public function render_settings_page() {
        if (!current_user_can('manage_options')) { return; }
        $settings = $this->settings();
        $page = $this->chat_page();
        $key_is_constant = in_array($this->api_key_source(), array('wp-config.php', 'OPENAI_API_KEY'), true);
        ?>
        <div class="wrap">
            <h1><?php esc_html_e('ChatGPT 接入', 'fieldnotes-chatgpt'); ?></h1>
            <?php settings_errors(); ?>
            <?php if (!is_ssl() && !$this->http_testing_allowed()) : ?><div class="notice notice-info"><p><?php esc_html_e('HTTP 测试已关闭。需要测试时，先勾选下方「允许 HTTP 测试」并保存，即可设置口令与 API Key。', 'fieldnotes-chatgpt'); ?></p></div><?php endif; ?>
            <p><?php esc_html_e('配置私密对话页面的访问口令与 OpenAI Responses API。页面地址：', 'fieldnotes-chatgpt'); ?>
                <?php if ($page) : ?><a href="<?php echo esc_url(get_permalink($page)); ?>" target="_blank" rel="noopener noreferrer"><?php echo esc_html(get_permalink($page)); ?></a><?php else : ?><code>/chatgpt/</code><?php endif; ?>
            </p>
            <form action="options.php" method="post">
                <?php settings_fields(self::GROUP); ?>
                <table class="form-table" role="presentation">
                    <tr>
                        <th scope="row"><?php esc_html_e('HTTP 测试', 'fieldnotes-chatgpt'); ?></th>
                        <td><label><input type="checkbox" name="<?php echo esc_attr(self::OPTION); ?>[allow_http_testing]" value="1" <?php checked($this->http_testing_allowed()); ?> <?php disabled(defined('FIELDNOTES_CHATGPT_ALLOW_HTTP')); ?>> <?php esc_html_e('允许 HTTP 测试', 'fieldnotes-chatgpt'); ?></label>
                            <p class="description"><?php esc_html_e('本测试版本默认开启，HTTP 下可保存配置、验证口令及对话。HTTP 不加密浏览器传输；测试结束后可关闭。服务端始终通过 HTTPS 连接 OpenAI。', 'fieldnotes-chatgpt'); ?></p>
                            <?php if (defined('FIELDNOTES_CHATGPT_ALLOW_HTTP')) : ?><p class="description"><?php esc_html_e('当前由 wp-config.php 中的 FIELDNOTES_CHATGPT_ALLOW_HTTP 控制。', 'fieldnotes-chatgpt'); ?></p><?php endif; ?>
                        </td>
                    </tr>
                    <tr>
                        <th scope="row"><label for="fnchat-page-password"><?php esc_html_e('页面访问口令', 'fieldnotes-chatgpt'); ?></label></th>
                        <td>
                            <?php if ($this->transport_allowed()) : ?>
                                <input id="fnchat-page-password" class="regular-text" type="password" name="<?php echo esc_attr(self::OPTION); ?>[page_password]" autocomplete="new-password" value="">
                                <p class="description"><?php echo !empty($settings['password_hash']) ? esc_html__('口令已设置。留空会保留当前口令；输入新口令可立即更换。', 'fieldnotes-chatgpt') : esc_html__('尚未设置口令。至少 8 个字符，设置后访客才能进入页面。', 'fieldnotes-chatgpt'); ?></p>
                            <?php else : ?>
                                <p><?php esc_html_e('启用 HTTPS 后可设置页面口令。', 'fieldnotes-chatgpt'); ?></p>
                            <?php endif; ?>
                        </td>
                    </tr>
                    <tr>
                        <th scope="row"><label for="fnchat-api-key"><?php esc_html_e('OpenAI API Key', 'fieldnotes-chatgpt'); ?></label></th>
                        <td>
                            <?php if (!$this->transport_allowed()) : ?>
                                <p><?php esc_html_e('启用 HTTPS 后可配置 API Key。', 'fieldnotes-chatgpt'); ?></p>
                            <?php elseif ($key_is_constant) : ?>
                                <p><strong><?php esc_html_e('已从服务器环境变量或 wp-config.php 读取。', 'fieldnotes-chatgpt'); ?></strong></p>
                            <?php else : ?>
                                <input id="fnchat-api-key" class="regular-text" type="password" name="<?php echo esc_attr(self::OPTION); ?>[api_key]" autocomplete="new-password" value="" spellcheck="false">
                                <p class="description"><?php echo '' !== $settings['api_key'] ? esc_html__('Key 已保存。留空会保留当前 Key；勾选清除可移除。', 'fieldnotes-chatgpt') : esc_html__('粘贴你的 OpenAI API Key。该 Key 保存在 WordPress 服务器端，不会输出到网页。', 'fieldnotes-chatgpt'); ?></p>
                                <label><input type="checkbox" name="<?php echo esc_attr(self::OPTION); ?>[clear_api_key]" value="1"> <?php esc_html_e('清除已保存的 API Key', 'fieldnotes-chatgpt'); ?></label>
                            <?php endif; ?>
                            <p class="description"><?php esc_html_e('也可在 wp-config.php 中定义 FIELDNOTES_CHATGPT_API_KEY，或给 PHP 设置 OPENAI_API_KEY 环境变量；服务器配置优先。', 'fieldnotes-chatgpt'); ?></p>
                        </td>
                    </tr>
                    <tr>
                        <th scope="row"><label for="fnchat-model"><?php esc_html_e('模型 ID', 'fieldnotes-chatgpt'); ?></label></th>
                        <td><input id="fnchat-model" class="regular-text" type="text" name="<?php echo esc_attr(self::OPTION); ?>[model]" value="<?php echo esc_attr($settings['model']); ?>" spellcheck="false"><p class="description"><?php esc_html_e('默认使用 gpt-6-luna。也可以填写你的 API 项目已开通的其他 Responses API 模型 ID。', 'fieldnotes-chatgpt'); ?></p></td>
                    </tr>
                    <tr>
                        <th scope="row"><label for="fnchat-instructions"><?php esc_html_e('助手系统提示词', 'fieldnotes-chatgpt'); ?></label></th>
                        <td><textarea id="fnchat-instructions" class="large-text" rows="5" maxlength="6000" name="<?php echo esc_attr(self::OPTION); ?>[instructions]"><?php echo esc_textarea($settings['instructions']); ?></textarea><p class="description"><?php esc_html_e('此指令由服务器发送，访客不能通过浏览器表单修改。', 'fieldnotes-chatgpt'); ?></p></td>
                    </tr>
                    <tr>
                        <th scope="row"><label for="fnchat-max-tokens"><?php esc_html_e('单次回答上限', 'fieldnotes-chatgpt'); ?></label></th>
                        <td><input id="fnchat-max-tokens" type="number" min="256" max="4096" step="64" name="<?php echo esc_attr(self::OPTION); ?>[max_output_tokens]" value="<?php echo esc_attr((string) $settings['max_output_tokens']); ?>"> <?php esc_html_e('tokens（256–4096）', 'fieldnotes-chatgpt'); ?></td>
                    </tr>
                    <tr>
                        <th scope="row"><label for="fnchat-timeout"><?php esc_html_e('请求时限', 'fieldnotes-chatgpt'); ?></label></th>
                        <td><input id="fnchat-timeout" type="number" min="15" max="120" name="<?php echo esc_attr(self::OPTION); ?>[request_timeout]" value="<?php echo esc_attr((string) $settings['request_timeout']); ?>"> <?php esc_html_e('秒（15–120，默认 60）', 'fieldnotes-chatgpt'); ?><p class="description"><?php esc_html_e('同时受 PHP 与反向代理的执行时限约束。请求超时不会自动重复提交。', 'fieldnotes-chatgpt'); ?></p></td>
                    </tr>
                </table>
                <?php include FIELDNOTES_CHATGPT_DIR . 'views/settings-workspace.php'; ?>
                <?php submit_button(__('保存设置', 'fieldnotes-chatgpt')); ?>
            </form>
            <section class="fnchat-diagnostics" aria-labelledby="fnchat-diagnostics-title">
                <h2 id="fnchat-diagnostics-title"><?php esc_html_e('连接检测', 'fieldnotes-chatgpt'); ?></h2>
                <p><?php esc_html_e('先保存设置，再检测。检测从 WordPress 服务器读取当前模型信息，不创建对话；成功不等于已验证生成权限或 API 额度。', 'fieldnotes-chatgpt'); ?></p>
                <button id="fnchat-diagnose" class="button button-secondary" type="button"><?php esc_html_e('检测连接', 'fieldnotes-chatgpt'); ?></button>
                <div id="fnchat-diagnostic-result" role="status" aria-live="polite"></div>
            </section>
            <hr>
            <p><?php esc_html_e('安全提示：对外开放前请先为网站启用 HTTPS。此页面有口令保护，但共享口令的访客会共用 API 用量。OpenAI API 按用量单独计费。', 'fieldnotes-chatgpt'); ?></p>
        </div>
        <?php
    }

    public function append_navigation_item($items, $args) {
        if (!is_object($args) || empty($args->theme_location) || 'primary' !== $args->theme_location) { return $items; }
        $page = $this->chat_page();
        if (!$page || 'publish' !== $page->post_status) { return $items; }
        $url = get_permalink($page);
        if (false !== strpos($items, 'href="' . esc_url($url) . '"')) { return $items; }
        $label = function_exists('fieldnotes_section_label') ? fieldnotes_section_label('chatgpt') : get_the_title($page);
        return $items . '<li class="menu-item menu-item-chatgpt"><a href="' . esc_url($url) . '" title="' . esc_attr__('ChatGPT 对话', 'fieldnotes-chatgpt') . '">' . esc_html($label) . '</a></li>';
    }

    public function ajax_unlock() {
        $this->verify_nonce();
        if (!$this->transport_allowed()) {
            wp_send_json_error(array('message' => __('页面口令只会通过 HTTPS 验证，请先启用 HTTPS。', 'fieldnotes-chatgpt')), 403);
        }
        $settings = $this->settings();
        $page = $this->chat_page();
        if (!$page || empty($settings['password_hash'])) {
            wp_send_json_error(array('message' => __('此页面尚未完成口令设置。', 'fieldnotes-chatgpt')), 503);
        }

        $rate_key = $this->rate_key('unlock');
        $attempts = (int) get_transient($rate_key);
        if ($attempts >= 5) {
            wp_send_json_error(array('message' => __('尝试次数过多，请 15 分钟后再试。', 'fieldnotes-chatgpt')), 429);
        }

        $password = isset($_POST['password']) && is_string($_POST['password']) ? (string) wp_unslash($_POST['password']) : '';
        if (strlen($password) > 512 || !password_verify($password, $settings['password_hash'])) {
            if (0 === $attempts) { set_transient($rate_key, 1, 15 * MINUTE_IN_SECONDS); }
            else { set_transient($rate_key, $attempts + 1, 15 * MINUTE_IN_SECONDS); }
            wp_send_json_error(array('message' => __('口令不正确，请重试。', 'fieldnotes-chatgpt')), 403);
        }

        delete_transient($rate_key);
        $this->set_access_cookie((int) $page->ID, $settings['password_hash']);
        wp_send_json_success(array('message' => __('验证成功。', 'fieldnotes-chatgpt')));
    }

    public function ajax_chat() {
        if (array_key_exists('message', $_POST)) {
            $engine = $this->authorized_workspace('chat');
            $conversation = $this->post_string('conversation');
            $refs = json_decode($this->post_string('attachments', '[]'), true);
            $labels = json_decode($this->post_string('tools', '[]'), true);
            $this->send_workspace_result($engine->chat($conversation, $this->post_string('message'), $refs, $labels));
            return;
        }
        $this->verify_nonce();
        if (!$this->transport_allowed()) {
            wp_send_json_error(array('message' => __('对话只会通过 HTTPS 提交，请先启用 HTTPS。', 'fieldnotes-chatgpt')), 403);
        }
        $page = $this->chat_page();
        $settings = $this->settings();
        if (!$page || !$this->has_access((int) $page->ID)) {
            wp_send_json_error(array('message' => __('访问已锁定，请刷新页面后重新输入口令。', 'fieldnotes-chatgpt'), 'code' => 'session_expired'), 403);
        }

        $api_key = $this->api_key();
        if ('' === $api_key) {
            wp_send_json_error(array('message' => __('管理员尚未配置 OpenAI API Key。', 'fieldnotes-chatgpt')), 503);
        }

        $raw_messages = isset($_POST['messages']) && is_string($_POST['messages']) ? (string) wp_unslash($_POST['messages']) : '';
        if (strlen($raw_messages) > 65000) {
            wp_send_json_error(array('message' => __('对话内容过长，请清空后重新开始。', 'fieldnotes-chatgpt')), 413);
        }
        $messages = json_decode($raw_messages, true);
        if (!is_array($messages) || count($messages) < 1 || count($messages) > 12) {
            wp_send_json_error(array('message' => __('消息格式无效，请重新发送。', 'fieldnotes-chatgpt')), 400);
        }

        $input = array();
        $previous_role = '';
        foreach ($messages as $message) {
            if (!is_array($message) || !isset($message['role'], $message['content']) || !is_string($message['role']) || !is_string($message['content'])) {
                wp_send_json_error(array('message' => __('消息格式无效，请重新发送。', 'fieldnotes-chatgpt')), 400);
            }
            $role = sanitize_key($message['role']);
            $content = sanitize_textarea_field((string) $message['content']);
            if (!in_array($role, array('user', 'assistant'), true) || $role === $previous_role || '' === trim($content) || strlen($content) > 18000) {
                wp_send_json_error(array('message' => __('消息格式无效，请重新发送。', 'fieldnotes-chatgpt')), 400);
            }
            $input[] = array('role' => $role, 'content' => $content);
            $previous_role = $role;
        }
        if ('user' !== $input[0]['role'] || 'user' !== $previous_role) {
            wp_send_json_error(array('message' => __('最后一条消息需要来自访客。', 'fieldnotes-chatgpt')), 400);
        }

        $rate_key = $this->rate_key('chat');
        $requests = (int) get_transient($rate_key);
        if ($requests >= 20) {
            wp_send_json_error(array('message' => __('发送次数较多，请稍后再试。', 'fieldnotes-chatgpt')), 429);
        }
        if (0 === $requests) { set_transient($rate_key, 1, 10 * MINUTE_IN_SECONDS); }
        else { set_transient($rate_key, $requests + 1, 10 * MINUTE_IN_SECONDS); }

        $payload = array(
            'model' => $settings['model'],
            'instructions' => $settings['instructions'],
            'input' => $input,
            'max_output_tokens' => (int) $settings['max_output_tokens'],
            'store' => false,
        );
        $response = wp_remote_post('https://api.openai.com/v1/responses', array(
            'timeout' => min(120, max(15, (int) $settings['request_timeout'])),
            'redirection' => 0,
            'sslverify' => true,
            'limit_response_size' => 2097152,
            'headers' => array(
                'Authorization' => 'Bearer ' . $api_key,
                'Content-Type' => 'application/json',
            ),
            'body' => wp_json_encode($payload),
            'data_format' => 'body',
        ));

        if (is_wp_error($response)) {
            $issue = Fieldnotes_ChatGPT_Connection::transport_error($response, $api_key);
            wp_send_json_error(array('message' => $issue['message'], 'code' => $issue['code']), in_array($issue['code'], array('request_timeout', 'connection_timeout'), true) ? 504 : 502);
        }
        $status = (int) wp_remote_retrieve_response_code($response);
        $body = json_decode(wp_remote_retrieve_body($response), true);
        if ($status < 200 || $status >= 300 || !is_array($body)) {
            $issue = Fieldnotes_ChatGPT_Connection::http_error($status, $body, $api_key);
            wp_send_json_error(array('message' => $issue['message'], 'code' => $issue['code']), 502);
        }

        $text = isset($body['output_text']) && is_string($body['output_text']) ? trim($body['output_text']) : '';
        if ('' === $text && !empty($body['output']) && is_array($body['output'])) {
            foreach ($body['output'] as $item) {
                if (!is_array($item) || empty($item['content']) || !is_array($item['content'])) { continue; }
                foreach ($item['content'] as $part) {
                    if (is_array($part) && isset($part['type'], $part['text']) && 'output_text' === $part['type'] && is_string($part['text'])) { $text .= $part['text']; }
                }
            }
            $text = trim($text);
        }
        if ('' === $text) {
            wp_send_json_error(array('message' => __('模型没有返回文字，请稍后重试。', 'fieldnotes-chatgpt')), 502);
        }
        wp_send_json_success(array('text' => $text));
    }

    private function post_string($name, $default = '') {
        return isset($_POST[$name]) && is_string($_POST[$name]) ? wp_unslash($_POST[$name]) : $default;
    }

    private function authorized_workspace($kind) {
        $this->verify_nonce();
        $page = $this->chat_page();
        if (!$this->transport_allowed() || !$page || !$this->has_access((int) $page->ID)) wp_send_json_error(array('message' => __('访问已锁定，请重新验证口令。', 'fieldnotes-chatgpt'), 'code' => 'session_expired'), 403);
        $key = $this->api_key();
        if ($key === '') wp_send_json_error(array('message' => __('管理员尚未配置 OpenAI API Key。', 'fieldnotes-chatgpt'), 'code' => 'api_key_missing'), 503);
        $rate_key = $this->rate_key($kind);
        $count = (int) get_transient($rate_key);
        if ($count >= ($kind === 'upload' ? 30 : 40)) wp_send_json_error(array('message' => __('请求较多，请稍后再试。', 'fieldnotes-chatgpt'), 'code' => 'rate_limited'), 429);
        set_transient($rate_key, $count + 1, 10 * MINUTE_IN_SECONDS);
        $owner = hash_hmac('sha256', (int) $page->ID . '|' . $_COOKIE[self::COOKIE], wp_salt('auth'));
        return new Fieldnotes_ChatGPT_Workspace($this->settings(), $key, $owner);
    }

    private function send_workspace_result($result) {
        if (is_wp_error($result)) {
            $data = $result->get_error_data();
            wp_send_json_error(array('message' => $result->get_error_message(), 'code' => $result->get_error_code()), is_array($data) ? (int) ($data['status'] ?? 400) : 400);
        }
        wp_send_json_success($result);
    }

    public function ajax_upload() {
        $engine = $this->authorized_workspace('upload');
        $this->send_workspace_result($engine->upload($_FILES['file'] ?? null, array('kind' => $this->post_string('kind'), 'source_name' => $this->post_string('source_name'))));
    }

    public function ajax_remove_file() {
        $engine = $this->authorized_workspace('remove');
        $this->send_workspace_result($engine->remove_file($this->post_string('ref')));
    }

    public function ajax_approve() {
        $engine = $this->authorized_workspace('approve');
        $raw = $this->post_string('decisions');
        $this->send_workspace_result($engine->approve($this->post_string('conversation'), $this->post_string('revision'), strlen($raw) <= 10000 ? json_decode($raw, true) : null));
    }

    public function ajax_reset() {
        $engine = $this->authorized_workspace('reset');
        $this->send_workspace_result($engine->reset($this->post_string('conversation')));
    }

    public function ajax_lock() {
        $this->verify_nonce();
        $this->clear_access_cookie();
        wp_send_json_success(array('message' => __('页面已锁定。', 'fieldnotes-chatgpt')));
    }

    private function verify_nonce() {
        $nonce = isset($_POST['nonce']) && is_string($_POST['nonce']) ? sanitize_text_field(wp_unslash($_POST['nonce'])) : '';
        if (!wp_verify_nonce($nonce, 'fieldnotes_chatgpt')) {
            wp_send_json_error(array('message' => __('页面已过期，请刷新后重试。', 'fieldnotes-chatgpt')), 403);
        }
    }

    private function has_access($page_id) {
        $settings = $this->settings();
        if (empty($settings['password_hash']) || empty($_COOKIE[self::COOKIE]) || !is_string($_COOKIE[self::COOKIE])) { return false; }
        $token = sanitize_text_field(wp_unslash($_COOKIE[self::COOKIE]));
        if (!preg_match('/^([0-9]{10})\.([a-f0-9]{64})$/', $token, $matches)) { return false; }
        $expires = (int) $matches[1];
        if ($expires < time() || $expires > time() + 12 * HOUR_IN_SECONDS + 60) { return false; }
        $expected = hash_hmac('sha256', $expires . '|' . (int) $page_id . '|' . $settings['password_hash'], wp_salt('auth'));
        return hash_equals($expected, $matches[2]);
    }

    private function set_access_cookie($page_id, $password_hash) {
        $expires = time() + 12 * HOUR_IN_SECONDS;
        $signature = hash_hmac('sha256', $expires . '|' . (int) $page_id . '|' . $password_hash, wp_salt('auth'));
        $value = $expires . '.' . $signature;
        $this->write_cookie($value, $expires);
        $_COOKIE[self::COOKIE] = $value;
    }

    private function clear_access_cookie() {
        $this->write_cookie('', time() - HOUR_IN_SECONDS);
        unset($_COOKIE[self::COOKIE]);
    }

    private function write_cookie($value, $expires) {
        $path = defined('COOKIEPATH') && COOKIEPATH ? COOKIEPATH : '/';
        // The cookie must reach both the public page and admin-ajax.php.
        $ajax_path = wp_parse_url(admin_url('admin-ajax.php', 'relative'), PHP_URL_PATH);
        if (!is_string($ajax_path) || strpos($ajax_path, trailingslashit($path)) !== 0) $path = '/';
        $domain = defined('COOKIE_DOMAIN') && COOKIE_DOMAIN ? COOKIE_DOMAIN : '';
        setcookie(self::COOKIE, $value, array(
            'expires' => (int) $expires,
            'path' => $path,
            'domain' => $domain,
            'secure' => is_ssl(),
            'httponly' => true,
            'samesite' => 'Lax',
        ));
    }

    private function rate_key($kind) {
        $ip = isset($_SERVER['REMOTE_ADDR']) ? sanitize_text_field(wp_unslash($_SERVER['REMOTE_ADDR'])) : 'unknown';
        return 'fnchat_' . sanitize_key($kind) . '_' . substr(wp_hash($ip), 0, 32);
    }
}
