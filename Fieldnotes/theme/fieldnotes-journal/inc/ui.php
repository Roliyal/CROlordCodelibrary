<?php
/** Shared field interactions and an uninterrupted Echoes header. */
if (!defined('ABSPATH')) exit;

function fieldnotes_is_echoes_page() {
    if (!is_singular('page')) return false;
    if (is_page_template('templates/page-chatgpt.php')) return true;
    $page = get_queried_object();
    if (!$page || ($page->post_type ?? '') !== 'page') return false;
    if (fieldnotes_page_section($page->ID) === 'chatgpt') return true;
    return has_shortcode((string) ($page->post_content ?? ''), 'fieldnotes_chatgpt');
}

function fieldnotes_logo_only_header() {
    return is_single() || fieldnotes_echoes_unlocked();
}

/** Presentation only: let the plugin validate its own session and render its own state. */
function fieldnotes_echoes_unlocked() {
    if (!fieldnotes_is_echoes_page() || !shortcode_exists('fieldnotes_chatgpt')) return false;
    static $unlocked = null;
    if ($unlocked === null) {
        // No cookie presence check, password copy, or separate permission decision in the theme.
        $markup = do_shortcode('[fieldnotes_chatgpt]');
        $unlocked = (bool) preg_match('/class=["\'][^"\']*\bfnchat-shell\b[^"\']*["\']/', $markup);
    }
    return $unlocked;
}

function fieldnotes_ui_body_class($classes) {
    $classes[] = 'fieldnotes-ui';
    if (fieldnotes_is_echoes_page()) $classes[] = fieldnotes_echoes_unlocked() ? 'echoes-logo-only' : 'echoes-gate';
    return array_values(array_unique($classes));
}
add_filter('body_class', 'fieldnotes_ui_body_class');

function fieldnotes_field_assets() {
    $dependencies = array('fieldnotes-editorial');
    foreach (array('fieldnotes-collections', 'fieldnotes-market', 'fieldnotes-chatgpt', 'fieldnotes-comment-tools', 'fntm-front') as $handle) {
        if (wp_style_is($handle, 'enqueued')) $dependencies[] = $handle;
    }
    wp_enqueue_style('fieldnotes-fields', get_template_directory_uri() . '/assets/css/fields.css', $dependencies, FIELDNOTES_VERSION);
}
add_action('wp_enqueue_scripts', 'fieldnotes_field_assets', 100);
