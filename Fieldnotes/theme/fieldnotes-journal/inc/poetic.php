<?php
/** Public-facing section names. Stored page titles, menu data and URLs stay intact. */
if (!defined('ABSPATH')) exit;

function fieldnotes_section_labels() {
    return array('about' => __('About', 'fieldnotes-journal'), 'archive' => __('Markets', 'fieldnotes-journal'), 'map' => __('Wayfarer', 'fieldnotes-journal'), 'categories' => __('Boughs', 'fieldnotes-journal'), 'faves' => __('Keepsakes', 'fieldnotes-journal'), 'chatgpt' => __('Echoes', 'fieldnotes-journal'));
}

function fieldnotes_section_label($slug) {
    $labels = fieldnotes_section_labels();
    return $labels[$slug] ?? '';
}

/** Resolve the actual published page, including a page whose slug was changed. */
function fieldnotes_section_page($section) {
    if (!isset(fieldnotes_section_labels()[$section])) return null;
    $page = get_page_by_path($section, OBJECT, 'page');
    if ($page && get_post_status($page) === 'publish') return $page;
    $pages = get_pages(array(
        'post_status' => 'publish',
        'meta_key' => '_wp_page_template',
        'meta_value' => 'templates/page-' . $section . '.php',
        'sort_column' => 'menu_order,ID',
        'sort_order' => 'ASC',
        'number' => 1,
    ));
    return !empty($pages) ? $pages[0] : null;
}

function fieldnotes_section_url($section) {
    $page = fieldnotes_section_page($section);
    $url = $page ? get_permalink($page) : home_url('/' . $section . '/');
    // These two bare page URLs received permanent redirects in 1.7.
    // A stable view marker avoids the browser's old redirect cache and loops.
    if (in_array($section, array('map', 'faves'), true)) $url = add_query_arg('view', $section === 'map' ? 'wayfarer' : 'keepsakes', $url);
    return $url;
}

function fieldnotes_public_copy_enabled() {
    return !is_admin() && !(defined('REST_REQUEST') && REST_REQUEST);
}

function fieldnotes_page_section($post_id) {
    $page = get_post($post_id);
    if (!$page || $page->post_type !== 'page') return '';
    $labels = fieldnotes_section_labels();
    if (isset($labels[$page->post_name])) return $page->post_name;
    $templates = array('templates/page-about.php' => 'about', 'templates/page-archive.php' => 'archive', 'templates/page-map.php' => 'map', 'templates/page-categories.php' => 'categories', 'templates/page-faves.php' => 'faves', 'templates/page-chatgpt.php' => 'chatgpt');
    $template = get_page_template_slug($page->ID);
    return $templates[$template] ?? '';
}

function fieldnotes_poetic_page_title($title, $post_id = 0) {
    if (!fieldnotes_public_copy_enabled() || !$post_id) return $title;
    $section = fieldnotes_page_section($post_id);
    return $section !== '' ? fieldnotes_section_label($section) : $title;
}
add_filter('the_title', 'fieldnotes_poetic_page_title', 20, 2);

function fieldnotes_poetic_document_title($parts) {
    if (!fieldnotes_public_copy_enabled() || !is_singular('page') || is_front_page()) return $parts;
    $section = fieldnotes_page_section(get_queried_object_id());
    if ($section !== '') $parts['title'] = fieldnotes_section_label($section);
    return $parts;
}
add_filter('document_title_parts', 'fieldnotes_poetic_document_title', 20);

function fieldnotes_poetic_menu($items, $args) {
    if (!fieldnotes_public_copy_enabled() || empty($args->theme_location) || $args->theme_location !== 'primary') return $items;
    $hints = array('about' => __('博主简介', 'fieldnotes-journal'), 'archive' => __('Markets', 'fieldnotes-journal'), 'map' => __('旅行地图', 'fieldnotes-journal'), 'categories' => __('分类浏览', 'fieldnotes-journal'), 'faves' => __('日常选择', 'fieldnotes-journal'), 'chatgpt' => __('ChatGPT 对话', 'fieldnotes-journal'));
    $page_urls = null;
    foreach ($items as $item) {
        $section = isset($item->object, $item->object_id) && $item->object === 'page' ? fieldnotes_page_section($item->object_id) : '';
        if ($section === '' && isset($item->url)) {
            if ($page_urls === null) {
                $page_urls = array();
                foreach (fieldnotes_section_labels() as $slug => $label) {
                    $page = fieldnotes_section_page($slug);
                    if ($page) $page_urls[untrailingslashit(get_permalink($page))] = $slug;
                }
            }
            $section = $page_urls[untrailingslashit($item->url)] ?? '';
        }
        if ($section === '') continue;
        $item->title = fieldnotes_section_label($section);
        if (in_array($section, array('map', 'faves'), true)) $item->url = fieldnotes_section_url($section);
        if (empty($item->attr_title)) $item->attr_title = $hints[$section];
    }
    return $items;
}
add_filter('wp_nav_menu_objects', 'fieldnotes_poetic_menu', 20, 2);

function fieldnotes_poetic_map_text($translation, $text, $domain) {
    if (fieldnotes_public_copy_enabled() && $domain === 'fieldnotes-travel-map' && $text === '走过的地方。' && function_exists('fieldnotes_boughs_context') && fieldnotes_boughs_context() === 'cities') return '地图里的足迹。';
    if (fieldnotes_public_copy_enabled() && $domain === 'fieldnotes-travel-map' && $text === '走过的地方。') return fieldnotes_section_label('map');
    return $translation;
}
add_filter('gettext', 'fieldnotes_poetic_map_text', 20, 3);

function fieldnotes_poetic_map_output($output, $tag) {
    if (fieldnotes_public_copy_enabled() && $tag === 'fieldnotes_travel_map') {
        $output = str_replace('<span class="fntm-map-kicker">FIELDNOTES / ATLAS</span>', '', $output);
    }
    return $output;
}
add_filter('do_shortcode_tag', 'fieldnotes_poetic_map_output', 20, 2);
