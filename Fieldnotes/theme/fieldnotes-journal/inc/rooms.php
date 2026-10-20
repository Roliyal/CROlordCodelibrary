<?php
/** Independent destinations; categories are kept only as editorial storage. */
if (!defined('ABSPATH')) exit;

function fieldnotes_room_page($key) {
    if ($key === 'investing') return fieldnotes_section_page('archive');
    if ($key === 'minimalism') return fieldnotes_section_page('faves');
    if (!in_array($key, array('cities', 'wordpress', 'interludes'), true)) return null;
    $pages = get_pages(array('post_status' => 'publish', 'meta_key' => '_fieldnotes_room', 'meta_value' => $key, 'number' => 1, 'sort_column' => 'ID', 'sort_order' => 'ASC'));
    return $pages ? $pages[0] : null;
}

function fieldnotes_room_context() {
    if (!is_singular('page')) return '';
    if (is_page_template('templates/page-archive.php')) return 'investing';
    if (is_page_template('templates/page-faves.php')) return 'minimalism';
    if (is_page_template('templates/page-collection.php')) {
        $key = get_post_meta(get_queried_object_id(), '_fieldnotes_room', true);
        return in_array($key, array('cities', 'wordpress', 'interludes'), true) ? $key : '';
    }
    if (is_page_template('templates/page-categories.php') && isset($_GET['room']) && is_string($_GET['room']) && isset(fieldnotes_boughs_sections()[$_GET['room']])) return $_GET['room'];
    return '';
}

function fieldnotes_room_title($key) {
    return $key === 'minimalism' ? 'Keepsakes' : ($key === 'investing' ? 'Markets' : (fieldnotes_boughs_sections()[$key]['name'] ?? ''));
}

function fieldnotes_render_room($key) {
    if (!isset(fieldnotes_boughs_sections()[$key])) return;
    $boughs_key = $key;
    $boughs_section = fieldnotes_boughs_sections()[$key];
    $boughs_term = fieldnotes_boughs_term($key);
    include get_template_directory() . '/template-parts/collection.php';
}

function fieldnotes_rooms_install() {
    if (!current_user_can('edit_pages') || wp_doing_ajax() || get_option('fieldnotes_rooms_180_done')) return;
    $until = (int) get_option('fieldnotes_rooms_180_lock');
    if ($until && $until < time()) delete_option('fieldnotes_rooms_180_lock');
    if (!add_option('fieldnotes_rooms_180_lock', time() + 60, '', false)) return;
    try {
    $registry = array('cities' => array('cities', '城市漫步', 'templates/page-collection.php'), 'wordpress' => array('wordpress', 'WordPress主题与插件', 'templates/page-collection.php'), 'interludes' => array('interludes', '间隙拙作', 'templates/page-collection.php'), 'investing' => array('archive', '投资理财', 'templates/page-archive.php'), 'minimalism' => array('faves', 'Keepsakes', 'templates/page-faves.php'));
    foreach ($registry as $key => $definition) {
        if (fieldnotes_room_page($key)) continue;
        $id = wp_insert_post(array('post_type' => 'page', 'post_status' => 'publish', 'post_title' => $definition[1], 'post_name' => $definition[0], 'post_content' => '', 'meta_input' => array('_wp_page_template' => $definition[2], '_fieldnotes_room' => $key)), true);
        if (is_wp_error($id) || !$id) return;
    }
    update_option('fieldnotes_rooms_180_done', 1, false);
    } finally { delete_option('fieldnotes_rooms_180_lock'); }
}
add_action('admin_init', 'fieldnotes_rooms_install', 45);

function fieldnotes_room_document_title($parts) {
    $key = fieldnotes_room_context();
    if ($key) $parts['title'] = fieldnotes_room_title($key);
    return $parts;
}
add_filter('document_title_parts', 'fieldnotes_room_document_title', 25);

function fieldnotes_room_assets() {
    if (fieldnotes_is_echoes_page()) {
        $deps = array('fieldnotes-refinement');
        if (wp_style_is('fieldnotes-chatgpt', 'enqueued')) $deps[] = 'fieldnotes-chatgpt';
        wp_enqueue_style('fieldnotes-echoes-header', get_template_directory_uri() . '/assets/css/echoes.css', $deps, FIELDNOTES_VERSION);
    }
}
add_action('wp_enqueue_scripts', 'fieldnotes_room_assets', 40);
