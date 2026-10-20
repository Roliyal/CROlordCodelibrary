<?php
/** Keepsakes products: explicit JD identifiers and a browser-local shopping list. */
if (!defined('ABSPATH')) exit;

function fieldnotes_jd_sku($value) {
    if (!is_string($value)) return '';
    $value = trim($value);
    if (preg_match('/^[1-9][0-9]{4,17}$/D', $value)) return $value;
    $parts = wp_parse_url($value);
    if (!is_array($parts) || !in_array(strtolower($parts['scheme'] ?? ''), array('https', 'http'), true) || !in_array(strtolower($parts['host'] ?? ''), array('item.jd.com', 'item.m.jd.com'), true) || isset($parts['user']) || isset($parts['pass']) || isset($parts['port'])) return '';
    return preg_match('~^/(?:product/)?([1-9][0-9]{4,17})\.html$~D', $parts['path'] ?? '', $m) ? $m[1] : '';
}

function fieldnotes_jd_search_url($keyword) {
    return add_query_arg(array('keyword' => $keyword, 'enc' => 'utf-8'), 'https://search.jd.com/Search');
}

function fieldnotes_keepsake_product($post_id) {
    $sku = fieldnotes_jd_sku((string) get_post_meta($post_id, '_fieldnotes_jd_sku', true));
    $keyword = trim((string) get_post_meta($post_id, '_fieldnotes_jd_keyword', true));
    if ($keyword === '') $keyword = wp_strip_all_tags(get_the_title($post_id));
    return array('id' => (string) $post_id, 'sku' => $sku, 'keyword' => $keyword, 'title' => get_the_title($post_id), 'url' => $sku ? 'https://item.jd.com/' . $sku . '.html' : fieldnotes_jd_search_url($keyword));
}

function fieldnotes_keepsakes_metabox() {
    add_meta_box('fieldnotes-jd-product', 'Keepsakes · 京东选购', 'fieldnotes_keepsakes_metabox_content', 'journal_fave', 'normal');
}
add_action('add_meta_boxes', 'fieldnotes_keepsakes_metabox');

function fieldnotes_keepsakes_metabox_content($post) {
    wp_nonce_field('fieldnotes_jd_product', 'fieldnotes_jd_nonce');
    echo '<p><label for="fn-jd-sku">京东商品编号或商品网址</label><input class="widefat" id="fn-jd-sku" name="fieldnotes_jd_sku" value="' . esc_attr(get_post_meta($post->ID, '_fieldnotes_jd_sku', true)) . '" placeholder="https://item.jd.com/商品编号.html"></p><p>填写具体规格对应的 SKU；页面可直达此商品。未填写时按下面的关键词到京东搜索，不会自动匹配到某个商品。</p><p><label for="fn-jd-keyword">京东搜索关键词</label><input class="widefat" id="fn-jd-keyword" name="fieldnotes_jd_keyword" value="' . esc_attr(get_post_meta($post->ID, '_fieldnotes_jd_keyword', true)) . '" placeholder="留空使用这件物品的标题"></p><p>价格、库存、规格及加购结果由京东页面确认。本站只保存选购清单，不收取付款信息。</p>';
}

function fieldnotes_keepsakes_save($post_id) {
    if (!isset($_POST['fieldnotes_jd_nonce']) || !is_string($_POST['fieldnotes_jd_nonce']) || !wp_verify_nonce(sanitize_text_field(wp_unslash($_POST['fieldnotes_jd_nonce'])), 'fieldnotes_jd_product') || !current_user_can('edit_post', $post_id) || wp_is_post_revision($post_id) || defined('DOING_AUTOSAVE') && DOING_AUTOSAVE) return;
    $raw = isset($_POST['fieldnotes_jd_sku']) && is_string($_POST['fieldnotes_jd_sku']) ? trim(wp_unslash($_POST['fieldnotes_jd_sku'])) : '';
    $sku = fieldnotes_jd_sku($raw);
    // An invalid non-empty edit must not silently erase an existing product link.
    if ($raw === '') delete_post_meta($post_id, '_fieldnotes_jd_sku');
    elseif ($sku !== '') update_post_meta($post_id, '_fieldnotes_jd_sku', $sku);
    $keyword = isset($_POST['fieldnotes_jd_keyword']) && is_string($_POST['fieldnotes_jd_keyword']) ? sanitize_text_field(wp_unslash($_POST['fieldnotes_jd_keyword'])) : '';
    if (strlen($keyword) <= 300) update_post_meta($post_id, '_fieldnotes_jd_keyword', $keyword);
}
add_action('save_post_journal_fave', 'fieldnotes_keepsakes_save');

function fieldnotes_keepsakes_assets() {
    if (fieldnotes_boughs_context() !== 'minimalism') return;
    wp_enqueue_style('fieldnotes-keepsakes', get_template_directory_uri() . '/assets/css/keepsakes.css', array('fieldnotes-collections'), FIELDNOTES_VERSION);
    wp_enqueue_script('fieldnotes-keepsakes', get_template_directory_uri() . '/assets/js/keepsakes.js', array(), FIELDNOTES_VERSION, true);
}
add_action('wp_enqueue_scripts', 'fieldnotes_keepsakes_assets', 40);
