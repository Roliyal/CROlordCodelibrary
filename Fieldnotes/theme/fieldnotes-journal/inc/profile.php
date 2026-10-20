<?php
/** Editable profile copy. Reading existing content never changes stored data. */
if (!defined('ABSPATH')) exit;

function fieldnotes_profile_details($author_id) {
    $author = fieldnotes_article_author($author_id);
    $public_name = fieldnotes_clean_text(get_theme_mod('fieldnotes_about_name', ''));
    if ($public_name !== '') {
        $author['name'] = $public_name;
        $author['initial'] = wp_html_excerpt($public_name, 1, '');
    }
    $heading = trim((string) get_theme_mod('fieldnotes_about_heading', ''));
    $intro = trim((string) get_theme_mod('fieldnotes_about_intro', ''));
    if ($intro === '') $intro = trim((string) get_theme_mod('fieldnotes_profile', ''));
    return array(
        'author' => $author,
        'heading' => $heading !== '' ? $heading : __("认真生活。\n慢慢记录。", 'fieldnotes-journal'),
        'role' => fieldnotes_clean_text(get_theme_mod('fieldnotes_about_role', __('独立开发者 · 日常记录者', 'fieldnotes-journal'))),
        'intro' => $intro !== '' ? $intro : __('写代码，也慢慢走路。把想法做成作品，把日常里微小而清晰的感受，留在文字与影像里。', 'fieldnotes-journal'),
    );
}

function fieldnotes_profile_destinations() {
    $items = array(
        'archive' => array(fieldnotes_section_label('archive'), __('记录判断，留待回看。', 'fieldnotes-journal'), __('分时与 K 线，观察与复盘。', 'fieldnotes-journal')),
        'map' => array(fieldnotes_section_label('map'), __('沿途，有回声。', 'fieldnotes-journal'), __('从地图上的一个地方，走进一段旅途。', 'fieldnotes-journal')),
        'faves' => array(fieldnotes_section_label('faves'), __('好物，长相伴。', 'fieldnotes-journal'), __('生活里，那些安静而可靠的小选择。', 'fieldnotes-journal')),
    );
    $links = array();
    foreach ($items as $slug => $copy) {
        $page = fieldnotes_section_page($slug);
        if (!$page || get_post_status($page) !== 'publish') continue;
        $links[] = array('label' => $copy[0], 'title' => $copy[1], 'description' => $copy[2], 'url' => fieldnotes_section_url($slug));
    }
    return $links;
}

function fieldnotes_profile_customize($customizer) {
    $customizer->add_section('fieldnotes_profile_page', array(
        'title' => __('博主简介页面', 'fieldnotes-journal'),
        'description' => __('正文与横幅图片在「页面 > 博主简介」编辑；头像读取该页面作者的 WordPress 头像。留空的公开名称使用作者的公开显示名称。', 'fieldnotes-journal'),
        'priority' => 36,
    ));
    $fields = array(
        'fieldnotes_about_name' => array(__('公开名称', 'fieldnotes-journal'), '', 'text'),
        'fieldnotes_about_role' => array(__('身份说明', 'fieldnotes-journal'), __('独立开发者 · 日常记录者', 'fieldnotes-journal'), 'text'),
        'fieldnotes_about_heading' => array(__('开场标题（回车换行）', 'fieldnotes-journal'), __("认真生活。\n慢慢记录。", 'fieldnotes-journal'), 'textarea'),
        'fieldnotes_about_intro' => array(__('开场自述', 'fieldnotes-journal'), '', 'textarea'),
    );
    foreach ($fields as $id => $field) {
        $customizer->add_setting($id, array('default' => $field[1], 'sanitize_callback' => $field[2] === 'textarea' ? 'sanitize_textarea_field' : 'sanitize_text_field'));
        $customizer->add_control($id, array('label' => $field[0], 'section' => 'fieldnotes_profile_page', 'type' => $field[2]));
    }
}
add_action('customize_register', 'fieldnotes_profile_customize', 21);
