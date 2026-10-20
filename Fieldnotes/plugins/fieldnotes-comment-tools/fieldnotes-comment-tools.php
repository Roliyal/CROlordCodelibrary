<?php
/**
 * Plugin Name: Fieldnotes Comment Tools
 * Description: A quiet comment composer with email-first Gravatar, local fallback avatars and image emoji.
 * Version: 1.6.0
 * Requires at least: 6.2
 * Requires PHP: 7.4
 * License: GPL-2.0-or-later
 * Text Domain: fieldnotes-comment-tools
 */
if (!defined('ABSPATH')) { exit; }

define('FCT_FILE', __FILE__);
define('FCT_DIR', plugin_dir_path(__FILE__));
define('FCT_URL', plugin_dir_url(__FILE__));
define('FCT_VERSION', '1.6.0');
require_once FCT_DIR . 'includes/avatars.php';
require_once FCT_DIR . 'includes/moderation.php';

function fct_avatars() {
    return array(
        'traveler' => array(__('旅人', 'fieldnotes-comment-tools'), 'traveler.svg'),
        'cat' => array(__('小猫', 'fieldnotes-comment-tools'), 'cat.svg'),
        'mountain' => array(__('山野', 'fieldnotes-comment-tools'), 'mountain.svg'),
        'coffee' => array(__('咖啡', 'fieldnotes-comment-tools'), 'coffee.svg'),
        'sakura' => array(__('樱花', 'fieldnotes-comment-tools'), 'sakura.svg'),
        'book' => array(__('读书', 'fieldnotes-comment-tools'), 'book.svg'),
        'leaf' => array(__('绿叶', 'fieldnotes-comment-tools'), 'leaf.svg'),
        'gravatar' => array(__('邮箱头像', 'fieldnotes-comment-tools'), 'gravatar.svg'),
    );
}

function fct_enqueue_assets() {
    if (!is_singular() || (!comments_open() && !get_comments_number())) return;
    wp_enqueue_style('fieldnotes-comment-tools', FCT_URL . 'assets/comment-tools.css', array(), FCT_VERSION);
    wp_enqueue_script('fieldnotes-comment-tools', FCT_URL . 'assets/comment-tools.js', array(), FCT_VERSION, true);
    wp_localize_script('fieldnotes-comment-tools', 'fctAvatarConfig', array(
        'ajaxUrl' => wp_make_link_relative(admin_url('admin-ajax.php')),
        'nonce' => wp_create_nonce('fct_avatar_lookup'),
        'messages' => array(
            'idle' => __('头像，随邮箱而来。', 'fieldnotes-comment-tools'),
            'idleNote' => __('填写邮箱后，会自动匹配 Gravatar。', 'fieldnotes-comment-tools'),
            'loading' => __('正在寻找你的头像…', 'fieldnotes-comment-tools'),
            'loadingNote' => __('稍等片刻，也可以继续写下想法。', 'fieldnotes-comment-tools'),
            'found' => __('已使用邮箱头像', 'fieldnotes-comment-tools'),
            'foundNote' => __('无需另外选择头像。', 'fieldnotes-comment-tools'),
            'missing' => __('选一枚喜欢的头像', 'fieldnotes-comment-tools'),
            'missingNote' => __('这个邮箱还没有公开头像，也可以保留默认选择。', 'fieldnotes-comment-tools'),
            'unavailable' => __('头像可随邮箱同步', 'fieldnotes-comment-tools'),
            'unavailableNote' => __('暂未确认头像，留言不受影响。', 'fieldnotes-comment-tools'),
            'fallback' => __('本次使用备用头像', 'fieldnotes-comment-tools'),
            'fallbackNote' => __('更换邮箱后，会重新匹配 Gravatar。', 'fieldnotes-comment-tools'),
            'cachedNote' => __('已使用本站保存的邮箱头像。', 'fieldnotes-comment-tools'),
        ),
    ));
}
add_action('wp_enqueue_scripts', 'fct_enqueue_assets', 20);

function fct_comment_fields($fields) {
    if (is_user_logged_in()) {
        unset($fields['author'], $fields['email'], $fields['url']);
        return $fields;
    }
    $commenter = wp_get_current_commenter();
    $required = ' required aria-required="true"';
    $required_label = '<span class="fct-required-note">' . esc_html__('必填', 'fieldnotes-comment-tools') . '</span>';
    $fields['author'] = '<p class="comment-form-author fct-form-field"><label for="author">' . esc_html__('昵称', 'fieldnotes-comment-tools') . $required_label . '</label><input id="author" name="author" type="text" value="' . esc_attr($commenter['comment_author']) . '" placeholder="' . esc_attr__('怎么称呼你', 'fieldnotes-comment-tools') . '" maxlength="245" autocomplete="nickname"' . $required . '></p>';
    $fields['email'] = '<p class="comment-form-email fct-form-field"><label for="email">' . esc_html__('邮箱', 'fieldnotes-comment-tools') . $required_label . '</label><input id="email" name="email" type="email" value="' . esc_attr($commenter['comment_author_email']) . '" placeholder="' . esc_attr__('用于回复与头像，不会公开', 'fieldnotes-comment-tools') . '" maxlength="100" autocomplete="email" autocapitalize="off" spellcheck="false" aria-describedby="fct-email-note"' . $required . '></p>';
    unset($fields['url']);
    return $fields;
}
add_filter('comment_form_default_fields', 'fct_comment_fields', 20);

function fct_comment_defaults($defaults) {
    $defaults['comment_notes_before'] = '<p class="fct-privacy-note" id="fct-email-note">' . esc_html__('很高兴，在这里遇见你。邮箱仅用于评论与头像，不会公开。', 'fieldnotes-comment-tools') . '</p>';
    $defaults['comment_notes_after'] = '';
    $defaults['label_submit'] = __('留下评论', 'fieldnotes-comment-tools');
    return $defaults;
}
add_filter('comment_form_defaults', 'fct_comment_defaults', 20);

function fct_avatar_picker_markup() {
    $avatars = fct_avatars();
    $saved = isset($_COOKIE['fieldnotes_comment_avatar']) ? sanitize_key(wp_unslash($_COOKIE['fieldnotes_comment_avatar'])) : 'traveler';
    if (!isset($avatars[$saved]) || $saved === 'gravatar') $saved = 'traveler';
    $options = '';
    foreach ($avatars as $key => $avatar) {
        if ($key === 'gravatar') continue;
        $checked = $saved === $key;
        $options .= '<label class="fct-avatar-option' . ($checked ? ' is-selected' : '') . '"><input type="radio" name="fieldnotes_avatar_fallback" value="' . esc_attr($key) . '"' . checked($checked, true, false) . '><img src="' . esc_url(FCT_URL . 'assets/avatars/' . $avatar[1]) . '" alt="" width="44" height="44" loading="lazy"><span>' . esc_html($avatar[0]) . '</span></label>';
    }
    $markup = '<div class="fct-avatar-profile" data-fct-account="' . (is_user_logged_in() ? '1' : '0') . '" data-state="idle"><input type="hidden" name="fieldnotes_avatar" value="gravatar"><input type="hidden" name="fieldnotes_avatar_override" value="0">';
    $markup .= '<div class="fct-avatar-summary"><img class="fct-avatar-preview" src="' . esc_url(FCT_URL . 'assets/avatars/avatar-placeholder.svg') . '" alt="" width="44" height="44" data-fct-default="' . esc_url(FCT_URL . 'assets/avatars/avatar-placeholder.svg') . '"><div class="fct-avatar-copy" role="status" aria-live="polite" aria-atomic="true"><p class="fct-avatar-status">' . esc_html__('头像，随邮箱而来。', 'fieldnotes-comment-tools') . '</p><p class="fct-avatar-note">' . esc_html__('填写邮箱后，会自动匹配 Gravatar。', 'fieldnotes-comment-tools') . '</p></div><div class="fct-avatar-actions" hidden><button class="fct-avatar-use-local" type="button" hidden>' . esc_html__('选择备用头像', 'fieldnotes-comment-tools') . '</button><button class="fct-avatar-retry" type="button" hidden>' . esc_html__('重新匹配', 'fieldnotes-comment-tools') . '</button></div></div>';
    $markup .= '<fieldset class="fct-avatar-picker" hidden><legend>' . esc_html__('没有 Gravatar 时使用的头像', 'fieldnotes-comment-tools') . '</legend><div class="fct-avatar-options">' . $options . '</div></fieldset>';
    // With scripts disabled the server still gives an existing Gravatar priority.
    $markup .= '<noscript><p class="fct-avatar-noscript">' . esc_html__('已有 Gravatar 会自动使用；没有时，可在下方选择头像。', 'fieldnotes-comment-tools') . '</p><fieldset class="fct-avatar-picker"><legend>' . esc_html__('备用头像', 'fieldnotes-comment-tools') . '</legend><div class="fct-avatar-options">' . str_replace('name="fieldnotes_avatar_fallback"', 'name="fieldnotes_avatar_noscript"', $options) . '</div></fieldset></noscript></div>';
    return $markup;
}

function fct_order_comment_fields($fields) {
    $ordered = array();
    foreach (array('author', 'email') as $key) {
        if (isset($fields[$key])) { $ordered[$key] = $fields[$key]; unset($fields[$key]); }
    }
    $ordered['fieldnotes_avatar_profile'] = fct_avatar_picker_markup();
    if (isset($fields['comment'])) { $ordered['comment'] = $fields['comment']; unset($fields['comment']); }
    return $ordered + $fields;
}
add_filter('comment_form_fields', 'fct_order_comment_fields', 30);

function fct_emoji_groups() {
    return array(
        __('最新加入', 'fieldnotes-comment-tools') => array(
            'cracking-face' => array(__('裂纹脸', 'fieldnotes-comment-tools'), 'cracking-face.svg'),
            'thumb-left' => array(__('向左拇指', 'fieldnotes-comment-tools'), 'thumb-left.svg'),
            'thumb-left-light' => array(__('向左拇指 · 浅肤色', 'fieldnotes-comment-tools'), 'thumb-left-light.svg'),
            'thumb-left-medium-light' => array(__('向左拇指 · 中浅肤色', 'fieldnotes-comment-tools'), 'thumb-left-medium-light.svg'),
            'thumb-left-medium' => array(__('向左拇指 · 中等肤色', 'fieldnotes-comment-tools'), 'thumb-left-medium.svg'),
            'thumb-left-medium-dark' => array(__('向左拇指 · 中深肤色', 'fieldnotes-comment-tools'), 'thumb-left-medium-dark.svg'),
            'thumb-left-dark' => array(__('向左拇指 · 深肤色', 'fieldnotes-comment-tools'), 'thumb-left-dark.svg'),
            'thumb-right' => array(__('向右拇指', 'fieldnotes-comment-tools'), 'thumb-right.svg'),
            'thumb-right-light' => array(__('向右拇指 · 浅肤色', 'fieldnotes-comment-tools'), 'thumb-right-light.svg'),
            'thumb-right-medium-light' => array(__('向右拇指 · 中浅肤色', 'fieldnotes-comment-tools'), 'thumb-right-medium-light.svg'),
            'thumb-right-medium' => array(__('向右拇指 · 中等肤色', 'fieldnotes-comment-tools'), 'thumb-right-medium.svg'),
            'thumb-right-medium-dark' => array(__('向右拇指 · 中深肤色', 'fieldnotes-comment-tools'), 'thumb-right-medium-dark.svg'),
            'thumb-right-dark' => array(__('向右拇指 · 深肤色', 'fieldnotes-comment-tools'), 'thumb-right-dark.svg'),
            'monarch-butterfly' => array(__('帝王蝶', 'fieldnotes-comment-tools'), 'monarch-butterfly.svg'),
            'pickle' => array(__('酸黄瓜', 'fieldnotes-comment-tools'), 'pickle.svg'),
            'lighthouse' => array(__('灯塔', 'fieldnotes-comment-tools'), 'lighthouse.svg'),
            'meteor' => array(__('流星', 'fieldnotes-comment-tools'), 'meteor.svg'),
            'eraser' => array(__('橡皮擦', 'fieldnotes-comment-tools'), 'eraser.svg'),
            'net-with-handle' => array(__('带柄抄网', 'fieldnotes-comment-tools'), 'net-with-handle.svg'),
        ),
        __('常用表情', 'fieldnotes-comment-tools') => array(
            'smile' => array(__('微笑', 'fieldnotes-comment-tools'), 'smile.svg'),
            'grin' => array(__('露齿笑', 'fieldnotes-comment-tools'), 'grin.svg'),
            'blush' => array(__('腼腆笑', 'fieldnotes-comment-tools'), 'blush.svg'),
            'heart-eyes' => array(__('爱心眼', 'fieldnotes-comment-tools'), 'heart-eyes.svg'),
            'smiling-hearts' => array(__('满心欢喜', 'fieldnotes-comment-tools'), 'smiling-hearts.svg'),
            'sunglasses' => array(__('墨镜', 'fieldnotes-comment-tools'), 'sunglasses.svg'),
            'thinking' => array(__('思考', 'fieldnotes-comment-tools'), 'thinking.svg'),
            'joy' => array(__('笑出眼泪', 'fieldnotes-comment-tools'), 'joy.svg'),
            'rofl' => array(__('大笑', 'fieldnotes-comment-tools'), 'rofl.svg'),
            'smiling-tear' => array(__('微笑带泪', 'fieldnotes-comment-tools'), 'smiling-tear.svg'),
            'cry' => array(__('大哭', 'fieldnotes-comment-tools'), 'cry.svg'),
            'sweat-smile' => array(__('尴尬一笑', 'fieldnotes-comment-tools'), 'sweat-smile.svg'),
            'surprised' => array(__('惊讶', 'fieldnotes-comment-tools'), 'surprised.svg'),
            'sleeping' => array(__('睡觉', 'fieldnotes-comment-tools'), 'sleeping.svg'),
            'upside-down' => array(__('倒置笑脸', 'fieldnotes-comment-tools'), 'upside-down.svg'),
            'wink' => array(__('眨眼', 'fieldnotes-comment-tools'), 'wink.svg'),
            'yum' => array(__('好吃', 'fieldnotes-comment-tools'), 'yum.svg'),
            'halo' => array(__('天使光环', 'fieldnotes-comment-tools'), 'halo.svg'),
            'hug' => array(__('拥抱', 'fieldnotes-comment-tools'), 'hug.svg'),
            'hand-over-mouth' => array(__('捂嘴笑', 'fieldnotes-comment-tools'), 'hand-over-mouth.svg'),
            'star-struck' => array(__('星星眼', 'fieldnotes-comment-tools'), 'star-struck.svg'),
            'relieved' => array(__('释然', 'fieldnotes-comment-tools'), 'relieved.svg'),
            'neutral' => array(__('平静', 'fieldnotes-comment-tools'), 'neutral.svg'),
            'grimace' => array(__('龇牙', 'fieldnotes-comment-tools'), 'grimace.svg'),
            'sad' => array(__('难过', 'fieldnotes-comment-tools'), 'sad.svg'),
            'angry' => array(__('生气', 'fieldnotes-comment-tools'), 'angry.svg'),
            'thumbs-up' => array(__('点赞', 'fieldnotes-comment-tools'), 'thumbs-up.svg'),
            'thumbs-down' => array(__('点踩', 'fieldnotes-comment-tools'), 'thumbs-down.svg'),
            'clap' => array(__('鼓掌', 'fieldnotes-comment-tools'), 'clap.svg'),
            'pray' => array(__('祈祷', 'fieldnotes-comment-tools'), 'pray.svg'),
            'heart' => array(__('红心', 'fieldnotes-comment-tools'), 'heart.svg'),
            'two-hearts' => array(__('两颗心', 'fieldnotes-comment-tools'), 'two-hearts.svg'),
            'sparkles' => array(__('闪光', 'fieldnotes-comment-tools'), 'sparkles.svg'),
            'party-popper' => array(__('庆祝', 'fieldnotes-comment-tools'), 'party-popper.svg'),
            'cherry-blossom' => array(__('樱花', 'fieldnotes-comment-tools'), 'cherry-blossom.svg'),
            'herb' => array(__('枝叶', 'fieldnotes-comment-tools'), 'herb.svg'),
            'hot-beverage' => array(__('咖啡', 'fieldnotes-comment-tools'), 'hot-beverage.svg'),
            'camera' => array(__('相机', 'fieldnotes-comment-tools'), 'camera.svg'),
            'luggage' => array(__('行李箱', 'fieldnotes-comment-tools'), 'luggage.svg'),
            'airplane' => array(__('飞机', 'fieldnotes-comment-tools'), 'airplane.svg'),
        ),
    );
}

function fct_emoji_catalog() {
    $catalog = array();
    foreach (fct_emoji_groups() as $group) {
        foreach ($group as $slug => $emoji) {
            $catalog[$slug] = $emoji;
        }
    }
    return $catalog;
}

function fct_comment_emoji_picker($field) {
    if (strpos($field, '<textarea') === false) return $field;
    $field = preg_replace('/^(\s*)<p\b/i', '$1<div', $field, 1);
    $field = preg_replace('/<\/p>(\s*)$/i', '</div>$1', $field, 1);
    $picker_id = wp_unique_id('fct-emoji-picker-');
    $toolbar = '<div class="fct-comment-tools" hidden><button class="fct-emoji-toggle" type="button" aria-label="' . esc_attr__('插入图片表情', 'fieldnotes-comment-tools') . '" aria-expanded="false" aria-controls="' . esc_attr($picker_id) . '"><svg viewBox="0 0 24 24" width="19" height="19" fill="none" stroke="currentColor" stroke-width="1.5" aria-hidden="true"><circle cx="12" cy="12" r="8.25"/><path d="M8.5 14.5c1.8 2 5.2 2 7 0" stroke-linecap="round"/><circle cx="9" cy="9.5" r=".8" fill="currentColor" stroke="none"/><circle cx="15" cy="9.5" r=".8" fill="currentColor" stroke="none"/></svg><span>' . esc_html__('表情', 'fieldnotes-comment-tools') . '</span></button><div id="' . esc_attr($picker_id) . '" class="fct-emoji-picker" hidden role="group" aria-label="' . esc_attr__('选择一个表情图片', 'fieldnotes-comment-tools') . '">';
    foreach (fct_emoji_groups() as $group_label => $emojis) {
        $toolbar .= '<section class="fct-emoji-section"><h3 class="fct-emoji-section-title">' . esc_html($group_label) . '</h3><div class="fct-emoji-grid">';
        foreach ($emojis as $slug => $emoji) {
            $src = FCT_URL . 'assets/emoji/' . $emoji[1];
            $toolbar .= '<button type="button" class="fct-emoji" data-fct-emoji="' . esc_attr($slug) . '" data-fct-src="' . esc_url($src) . '" aria-label="' . esc_attr($emoji[0]) . '" title="' . esc_attr($emoji[0]) . '"><img src="' . esc_url($src) . '" alt="" width="24" height="24"></button>';
        }
        $toolbar .= '</div></section>';
    }
    $toolbar .= '</div></div>';
    $editor = '<div id="' . esc_attr($picker_id . '-editor') . '" class="fct-emoji-editor" hidden contenteditable="true" role="textbox" aria-label="' . esc_attr__('评论内容', 'fieldnotes-comment-tools') . '" aria-multiline="true" aria-required="true" data-placeholder="' . esc_attr__('写下此刻的想法…', 'fieldnotes-comment-tools') . '"></div>';
    return preg_replace_callback('/<textarea\b([^>]*)>([\s\S]*?)<\/textarea>/i', function ($matches) use ($toolbar, $editor) {
        return '<div class="fct-composer-shell">' . $editor . '<textarea data-fct-source="true"' . $matches[1] . '>' . $matches[2] . '</textarea>' . $toolbar . '</div>';
    }, $field, 1);
}
add_filter('comment_form_field_comment', 'fct_comment_emoji_picker', 20);

function fct_render_comment_emoji_images($comment_text, $comment = null, $args = array()) {
    $catalog = fct_emoji_catalog();
    return preg_replace_callback('/\[\[fct_emoji:([a-z0-9-]+)\]\]/', function ($matches) use ($catalog) {
        if (!isset($catalog[$matches[1]])) return esc_html($matches[0]);
        $emoji = $catalog[$matches[1]];
        $src = FCT_URL . 'assets/emoji/' . $emoji[1];
        return '<img class="fct-inline-emoji" src="' . esc_url($src) . '" alt="' . esc_attr($emoji[0]) . '" width="24" height="24" loading="lazy">';
    }, $comment_text);
}
add_filter('comment_text', 'fct_render_comment_emoji_images', 20, 3);

function fct_require_guest_name_email($commentdata) {
    if (!empty($commentdata['user_ID'])) return $commentdata;
    if (!isset($_POST['comment'])) return $commentdata;
    $author = isset($_POST['author']) ? sanitize_text_field(wp_unslash($_POST['author'])) : '';
    $email = isset($_POST['email']) ? sanitize_email(wp_unslash($_POST['email'])) : '';
    if ($author === '' || $email === '' || !is_email($email)) {
        wp_die(
            esc_html__('请填写昵称和有效邮箱后再提交评论。', 'fieldnotes-comment-tools'),
            esc_html__('评论信息不完整', 'fieldnotes-comment-tools'),
            array('response' => 400, 'back_link' => true)
        );
    }
    $commentdata['comment_author'] = $author;
    $commentdata['comment_author_email'] = $email;
    return $commentdata;
}
add_filter('preprocess_comment', 'fct_require_guest_name_email', 5);

function fct_save_avatar_choice($comment_id, $approved, $commentdata) {
    if (!empty($commentdata['comment_type']) && $commentdata['comment_type'] !== 'comment') return;
    $avatar = isset($_POST['fieldnotes_avatar']) && is_string($_POST['fieldnotes_avatar']) ? sanitize_key(wp_unslash($_POST['fieldnotes_avatar'])) : 'gravatar';
    $avatars = fct_avatars();
    if (!isset($avatars[$avatar])) $avatar = 'gravatar';
    $fallback = isset($_POST['fieldnotes_avatar_noscript']) ? $_POST['fieldnotes_avatar_noscript'] : (isset($_POST['fieldnotes_avatar_fallback']) ? $_POST['fieldnotes_avatar_fallback'] : $avatar);
    $fallback = is_string($fallback) ? sanitize_key(wp_unslash($fallback)) : '';
    $local = isset($avatars[$fallback]) && $fallback !== 'gravatar';
    $explicit = isset($_POST['fieldnotes_avatar_override']) && $_POST['fieldnotes_avatar_override'] === '1';
    if ($local && $explicit) {
        // This is an explicit backup choice, never an automatic replacement.
        $avatar = $fallback;
    } elseif ($local) {
        $email = isset($commentdata['comment_author_email']) ? $commentdata['comment_author_email'] : '';
        $lookup = fct_gravatar_lookup($email);
        $avatar = $lookup['status'] === 'missing' ? $fallback : 'gravatar';
    } else {
        // The default email-avatar submission does not wait on an external call.
        $avatar = 'gravatar';
    }
    add_comment_meta($comment_id, '_fct_avatar', $avatar, true);
    if (!headers_sent()) {
        setcookie('fieldnotes_comment_avatar', $avatar, time() + YEAR_IN_SECONDS, COOKIEPATH ? COOKIEPATH : '/', COOKIE_DOMAIN, is_ssl(), true);
    }
}
add_action('comment_post', 'fct_save_avatar_choice', 10, 3);

function fct_selected_comment_avatar($avatar, $id_or_email, $size, $default, $alt, $args) {
    if (!($id_or_email instanceof WP_Comment)) return $avatar;
    $selected = get_comment_meta($id_or_email->comment_ID, '_fct_avatar', true);
    // Preserve the native HTML, including CDN/local-avatar filters and srcset.
    // 1.5.3 incorrectly replaced every email avatar with the new relay route.
    if (!$selected || $selected === 'gravatar') return $avatar;
    $avatars = fct_avatars();
    if (!isset($avatars[$selected])) return $avatar;
    $size = max(16, absint($size));
    $url = FCT_URL . 'assets/avatars/' . $avatars[$selected][1];
    $label = $alt ? $alt : $id_or_email->comment_author;
    return '<img alt="' . esc_attr($label) . '" src="' . esc_url($url) . '" class="avatar avatar-' . esc_attr($size) . ' photo fct-comment-avatar" height="' . esc_attr($size) . '" width="' . esc_attr($size) . '" loading="lazy">';
}
add_filter('get_avatar', 'fct_selected_comment_avatar', 20, 6);
