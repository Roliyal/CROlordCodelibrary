<?php
/** Boughs: five collections, their layouts and a one-time classification migration. */
if (!defined('ABSPATH')) exit;

function fieldnotes_boughs_sections() {
    return array(
        'cities' => array('name' => '城市漫步', 'description' => '城市的片刻，日常的回声。', 'intro' => '从一条街、一阵风，写到日常里那些微小而清晰的感受。文字留在这里，足迹留在 Wayfarer。', 'topics' => array('心情随笔', '城市观察', '生活片段'), 'image' => 'kyoto-rain.webp'),
        'wordpress' => array('name' => 'WordPress主题与插件', 'description' => '让写作与表达，拥有趁手的工具。', 'intro' => 'Journal 的页面，Echoes 的对话，地图上的足迹与文字后的回应。四个组件，共同构成这个空间。', 'topics' => array('Journal 主题', 'ChatGPT', 'Map', 'Comment Tools'), 'image' => 'studio-notes.webp'),
        'interludes' => array('name' => '间隙拙作', 'description' => '在工作的间隙，把实践慢慢写成作品。', 'intro' => '留给云原生实践、代码与工程记录。从一个具体的问题出发，留下可以复现的方法与过程。', 'topics' => array('云原生实践', '工程手记', '代码仓库'), 'image' => 'coast-dawn.webp'),
        'investing' => array('name' => '投资理财', 'description' => '观察市场，也记录每一次判断。', 'intro' => '把价格的变化与自己的思考放在同一页。看图、写下依据，再在时间里回看。', 'topics' => array('历史 K 线', '观察记录', '判断与复盘'), 'image' => 'studio-notes.webp'),
        'minimalism' => array('name' => '败家极简', 'description' => '生活里，那些安静而可靠的小选择。', 'intro' => '买过、用过，也认真取舍过。留下适合自己的器物，让日常更轻一点。', 'topics' => array('Keepsakes', '长期使用', '选择与取舍'), 'image' => 'studio-notes.webp'),
    );
}

function fieldnotes_boughs_term($slug) {
    if (!isset(fieldnotes_boughs_sections()[$slug])) return null;
    $term = get_term_by('slug', $slug, 'category');
    return $term && !is_wp_error($term) ? $term : null;
}

function fieldnotes_boughs_url($slug) {
    if ($slug === 'minimalism') return fieldnotes_section_url('faves');
    $page = function_exists('fieldnotes_room_page') ? fieldnotes_room_page($slug) : null;
    if ($page) return get_permalink($page);
    return add_query_arg('room', $slug, fieldnotes_section_url('categories'));
}

function fieldnotes_boughs_context() {
    $room = function_exists('fieldnotes_room_context') ? fieldnotes_room_context() : '';
    if ($room) return $room;
    if (!is_category()) return '';
    $term = get_queried_object();
    return $term && isset(fieldnotes_boughs_sections()[$term->slug ?? '']) ? $term->slug : '';
}

function fieldnotes_boughs_assets() {
    $key = fieldnotes_boughs_context();
    if (!$key && !is_page_template('templates/page-categories.php')) return;
    wp_enqueue_style('fieldnotes-collections', get_template_directory_uri() . '/assets/css/collections.css', array('fieldnotes-editorial'), FIELDNOTES_VERSION);
    if ($key === 'investing') {
        wp_enqueue_style('fieldnotes-market', get_template_directory_uri() . '/assets/css/market.css', array('fieldnotes-collections'), FIELDNOTES_VERSION);
        wp_enqueue_script('fieldnotes-market', get_template_directory_uri() . '/assets/js/market.js', array(), FIELDNOTES_VERSION, true);
        wp_localize_script('fieldnotes-market', 'fieldnotesMarket', array('ajaxUrl' => admin_url('admin-ajax.php', 'relative')));
    }
}
add_action('wp_enqueue_scripts', 'fieldnotes_boughs_assets', 30);

function fieldnotes_boughs_ensure_terms() {
    $ids = array();
    foreach (fieldnotes_boughs_sections() as $slug => $section) {
        $term = fieldnotes_boughs_term($slug);
        // Keep old slugs and parent paths intact until every original permalink is recorded.
        $args = array('name' => $section['name'], 'slug' => $slug, 'parent' => $term ? (int) $term->parent : 0, 'description' => $section['description']);
        $result = $term ? wp_update_term($term->term_id, 'category', $args) : wp_insert_term($section['name'], 'category', $args);
        if (is_wp_error($result)) return $result;
        $ids[$slug] = (int) $result['term_id'];
    }
    return $ids;
}

function fieldnotes_boughs_normalize_path($url) {
    $path = wp_parse_url($url, PHP_URL_PATH);
    return '/' . trim(rawurldecode(is_string($path) ? $path : ''), '/') . '/';
}

function fieldnotes_boughs_start() {
    if (!current_user_can('manage_categories') || (function_exists('wp_doing_ajax') && wp_doing_ajax())) return;
    if (get_option('fieldnotes_boughs_170_done')) return;
    fieldnotes_boughs_migrate_batch();
}
add_action('admin_init', 'fieldnotes_boughs_start', 30);

function fieldnotes_boughs_migrate_batch() {
    global $wpdb;
    if (get_option('fieldnotes_boughs_170_done')) return;
    $lock = get_option('fieldnotes_boughs_170_lock');
    if (is_array($lock) && ($lock['until'] ?? 0) < time()) delete_option('fieldnotes_boughs_170_lock');
    if (!add_option('fieldnotes_boughs_170_lock', array('until' => time() + 60), '', false)) return;
    try {
        $state = get_option('fieldnotes_boughs_170_state');
        if (!is_array($state)) {
            $old_terms = get_terms(array('taxonomy' => 'category', 'hide_empty' => false));
            if (is_wp_error($old_terms)) return;
            $backup = array('created_at' => gmdate('c'), 'default_category' => (int) get_option('default_category'), 'terms' => array());
            foreach ($old_terms as $term) $backup['terms'][] = array('id' => (int) $term->term_id, 'name' => $term->name, 'slug' => $term->slug, 'description' => $term->description, 'parent' => (int) $term->parent, 'url' => get_category_link($term->term_id));
            // Persist the complete term snapshot before renaming or removing a term.
            if (!add_option('fieldnotes_boughs_170_backup', $backup, '', false) && !get_option('fieldnotes_boughs_170_backup')) return;
            $ids = fieldnotes_boughs_ensure_terms();
            if (is_wp_error($ids)) { update_option('fieldnotes_boughs_170_error', $ids->get_error_message(), false); return; }
            $cutoff = (int) $wpdb->get_var("SELECT MAX(ID) FROM {$wpdb->posts} WHERE post_type = 'post'");
            $total = (int) $wpdb->get_var("SELECT COUNT(*) FROM {$wpdb->posts} WHERE post_type = 'post'");
            $state = array('last' => 0, 'cutoff' => $cutoff, 'total' => $total, 'moved' => 0, 'ids' => $ids);
            if (!update_option('fieldnotes_boughs_170_state', $state, false)) return;
            update_option('default_category', $ids['cities']);
        }
        $term_backup = get_option('fieldnotes_boughs_170_backup', array());
        $old_slugs = array();
        foreach ($term_backup['terms'] ?? array() as $old_term) $old_slugs[(int) $old_term['id']] = $old_term['slug'];
        $posts = $wpdb->get_col($wpdb->prepare("SELECT ID FROM {$wpdb->posts} WHERE post_type = 'post' AND ID > %d AND ID <= %d ORDER BY ID ASC LIMIT 100", $state['last'], $state['cutoff']));
        foreach ($posts as $post_id) {
            $post_id = (int) $post_id;
            if (!get_post($post_id)) { $state['last'] = $post_id; update_option('fieldnotes_boughs_170_state', $state, false); continue; }
            if (!metadata_exists('post', $post_id, '_fn_boughs_170_categories')) {
                $original = wp_get_post_categories($post_id);
                if (is_wp_error($original) || !add_post_meta($post_id, '_fn_boughs_170_categories', array_map('intval', $original), true)) {
                    update_option('fieldnotes_boughs_170_error', '原分类记录未能保存；整理已暂停，可在数据库恢复后重试。', false); return;
                }
                $old_url = get_permalink($post_id);
                if (!wp_parse_url($old_url, PHP_URL_QUERY)) add_post_meta($post_id, '_fn_boughs_170_path', fieldnotes_boughs_normalize_path($old_url), true);
                $old_categories = get_post_meta($post_id, '_fn_boughs_170_categories', true);
                add_post_meta($post_id, '_fn_boughs_170_image', basename(fieldnotes_demo_image($post_id, $old_slugs[$old_categories[0] ?? 0] ?? '')), true);
            }
            $result = wp_set_post_categories($post_id, array($state['ids']['cities']), false);
            if (is_wp_error($result)) { update_option('fieldnotes_boughs_170_error', $result->get_error_message(), false); return; }
            $state['last'] = $post_id;
            $state['moved']++;
            update_option('fieldnotes_boughs_170_state', $state, false);
        }
        if (count($posts) === 100) {
            if (!wp_next_scheduled('fieldnotes_boughs_170_batch')) wp_schedule_single_event(time() + 15, 'fieldnotes_boughs_170_batch');
            return;
        }
        $terms = get_terms(array('taxonomy' => 'category', 'hide_empty' => false));
        if (is_wp_error($terms)) return;
        if ((int) get_option('default_category') !== (int) $state['ids']['cities']) {
            update_option('default_category', $state['ids']['cities']);
            if ((int) get_option('default_category') !== (int) $state['ids']['cities']) { update_option('fieldnotes_boughs_170_error', '默认分类未能保存，旧分类尚未移除。', false); return; }
        }
        $backup = get_option('fieldnotes_boughs_170_backup', array());
        $old_ids = array_map('intval', array_column($backup['terms'] ?? array(), 'id'));
        foreach ($state['ids'] as $id) {
            $result = wp_update_term($id, 'category', array('parent' => 0));
            if (is_wp_error($result)) { update_option('fieldnotes_boughs_170_error', '栏目层级未能整理，旧分类尚未移除。', false); return; }
        }
        // Only category terms are removed. Posts, favourites, map points and comments are never deleted.
        foreach ($terms as $term) {
            if (in_array((int) $term->term_id, $state['ids'], true)) continue;
            if (!in_array((int) $term->term_id, $old_ids, true)) continue;
            $result = wp_delete_term($term->term_id, 'category', array('default' => $state['ids']['cities']));
            if (is_wp_error($result) || !$result) { update_option('fieldnotes_boughs_170_error', '旧分类未能完全整理，稍后会继续。', false); return; }
        }
        update_option('fieldnotes_boughs_170_done', array('at' => gmdate('c'), 'moved' => $state['moved']), false);
        delete_option('fieldnotes_boughs_170_error');
        wp_clear_scheduled_hook('fieldnotes_boughs_170_batch');
    } finally { delete_option('fieldnotes_boughs_170_lock'); }
}
add_action('fieldnotes_boughs_170_batch', 'fieldnotes_boughs_migrate_batch');

function fieldnotes_boughs_admin_notice() {
    if (!current_user_can('manage_categories')) return;
    $error = get_option('fieldnotes_boughs_170_error');
    $state = get_option('fieldnotes_boughs_170_state');
    if ($error) echo '<div class="notice notice-error"><p>Boughs：' . esc_html($error) . '</p></div>';
    elseif (is_array($state) && !get_option('fieldnotes_boughs_170_done')) echo '<div class="notice notice-info"><p>Boughs 正在整理现有文章：' . esc_html($state['moved']) . ' / ' . esc_html($state['total']) . '。文章与评论会保留。</p></div>';
}
add_action('admin_notices', 'fieldnotes_boughs_admin_notice');

function fieldnotes_boughs_tools_menu() {
    add_theme_page('Boughs 栏目整理', 'Boughs 栏目整理', 'manage_categories', 'fieldnotes-boughs', 'fieldnotes_boughs_tools_page');
}
add_action('admin_menu', 'fieldnotes_boughs_tools_menu');

function fieldnotes_boughs_tools_page() {
    if (!current_user_can('manage_categories')) return;
    $state = get_option('fieldnotes_boughs_170_state', array());
    $done = get_option('fieldnotes_boughs_170_done');
    echo '<div class="wrap"><h1>Boughs 栏目整理</h1><p>五个栏目：城市漫步、WordPress主题与插件、间隙拙作、投资理财、败家极简。</p>';
    echo '<p>' . esc_html($done ? '现有文章整理完成，共 ' . (int) $done['moved'] . ' 篇；后续新文章按你选择的分类保存。' : '正在整理：' . (int) ($state['moved'] ?? 0) . ' / ' . (int) ($state['total'] ?? 0)) . '</p>';
    echo '<p>本次仅重新分配现有普通文章的分类。文章正文、发布时间、作者、评论、收藏及地图地点均保留。原分类保存在本站，可下载记录。</p>';
    $url = wp_nonce_url(admin_url('admin-post.php?action=fieldnotes_boughs_backup'), 'fieldnotes_boughs_backup');
    echo '<p><a class="button" href="' . esc_url($url) . '">下载原分类记录</a></p><p>新页面使用固定的五个栏目。今后在编辑器选择对应分类；未选择分类的文章默认进入城市漫步。</p></div>';
}

function fieldnotes_boughs_download_backup() {
    if (!current_user_can('manage_categories')) wp_die('无权访问。', '', array('response' => 403));
    check_admin_referer('fieldnotes_boughs_backup');
    $backup = get_option('fieldnotes_boughs_170_backup', array());
    $backup['posts'] = array();
    $ids = get_posts(array('post_type' => 'post', 'post_status' => array_keys(get_post_stati()), 'numberposts' => -1, 'fields' => 'ids', 'meta_key' => '_fn_boughs_170_categories'));
    foreach ($ids as $id) $backup['posts'][] = array('id' => (int) $id, 'categories' => get_post_meta($id, '_fn_boughs_170_categories', true), 'old_path' => get_post_meta($id, '_fn_boughs_170_path', true));
    nocache_headers();
    header('Content-Type: application/json; charset=utf-8');
    header('Content-Disposition: attachment; filename="boughs-original-categories.json"');
    echo wp_json_encode($backup, JSON_PRETTY_PRINT | JSON_UNESCAPED_UNICODE);
    exit;
}
add_action('admin_post_fieldnotes_boughs_backup', 'fieldnotes_boughs_download_backup');

function fieldnotes_boughs_redirect_legacy() {
    if (!get_option('fieldnotes_boughs_170_done')) return;
    if (is_category()) {
        $term = get_queried_object();
        if ($term && isset(fieldnotes_boughs_sections()[$term->slug])) {
            $url = fieldnotes_boughs_url($term->slug);
            foreach (array('kind', 'stock', 'period', 'adjust', 'notes_page', 'faves_page') as $arg) if (isset($_GET[$arg]) && is_string($_GET[$arg])) $url = add_query_arg($arg, sanitize_text_field(wp_unslash($_GET[$arg])), $url);
            wp_safe_redirect($url, 302); exit;
        }
    }
    if (!is_404()) return;
    $path = fieldnotes_boughs_normalize_path(wp_unslash($_SERVER['REQUEST_URI'] ?? '/'));
    $post = get_posts(array('post_type' => 'post', 'post_status' => 'publish', 'posts_per_page' => 1, 'meta_key' => '_fn_boughs_170_path', 'meta_value' => $path));
    if ($post) { wp_safe_redirect(get_permalink($post[0]), 301); exit; }
    $backup = get_option('fieldnotes_boughs_170_backup', array());
    foreach ((array) ($backup['terms'] ?? array()) as $term) {
        if (!is_string($term['url']) || fieldnotes_boughs_normalize_path($term['url']) !== $path) continue;
        $target = in_array($term['slug'], array('wordpress-themes', 'wordpress-plugins', 'wordpress'), true) ? 'wordpress' : (in_array($term['slug'], array('investing', 'minimalism', 'interludes'), true) ? $term['slug'] : 'cities');
        wp_safe_redirect(fieldnotes_boughs_url($target), 301); exit;
    }
}
add_action('template_redirect', 'fieldnotes_boughs_redirect_legacy', 5);

function fieldnotes_boughs_collection_posts($term, $title = '这里的文字') {
    if (!$term) { echo '<p class="boughs-empty">第一篇文字，留给下一次认真记录。</p>'; return; }
    $page = isset($_GET['notes_page']) && is_scalar($_GET['notes_page']) ? max(1, absint($_GET['notes_page'])) : 1;
    $query = new WP_Query(array('post_type' => 'post', 'post_status' => 'publish', 'cat' => (int) $term->term_id, 'posts_per_page' => 8, 'paged' => $page, 'ignore_sticky_posts' => true, 'orderby' => array('date' => 'DESC', 'ID' => 'DESC')));
    echo '<section class="boughs-entries"><div class="boughs-section-heading"><h2>' . esc_html($title) . '</h2><span>' . esc_html(number_format_i18n($query->found_posts)) . ' 篇记录</span></div>';
    if ($query->have_posts()) {
        echo '<div class="boughs-entry-list">';
        $entry_index = 0;
        while ($query->have_posts()) { $query->the_post(); get_template_part('template-parts/collection-entry', null, array('featured' => $term->slug === 'cities' && $page === 1 && $entry_index === 0)); $entry_index++; }
        echo '</div>';
        echo wp_kses_post(paginate_links(array('base' => add_query_arg('notes_page', '%#%', fieldnotes_boughs_url($term->slug)), 'format' => '', 'current' => $page, 'total' => $query->max_num_pages, 'prev_text' => '上一页', 'next_text' => '下一页')));
    } else echo '<p class="boughs-empty">第一篇文字，留给下一次认真记录。</p>';
    wp_reset_postdata();
    echo '</section>';
}
