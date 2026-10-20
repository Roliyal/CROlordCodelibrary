<?php
if (function_exists('fieldnotes_boughs_context') && fieldnotes_boughs_context()) {
    get_header();
    $boughs_key = fieldnotes_boughs_context();
    $boughs_term = get_queried_object();
    $boughs_section = fieldnotes_boughs_sections()[$boughs_key];
    include get_template_directory() . '/template-parts/collection.php';
    get_footer();
    return;
}
get_header();
$topic = get_queried_object();
$topic_profile = fieldnotes_category_profile($topic);
global $wp_query;
$topic_total = (int) $wp_query->found_posts;
$topic_page = max(1, (int) get_query_var('paged'));
?>
<main id="main-content" class="page-shell topic-page">
    <header class="topic-heading">
        <div class="topic-title-row">
            <div><h1><?php echo esc_html(fieldnotes_category_label($topic)); ?></h1><p class="topic-description"><?php echo esc_html(fieldnotes_category_summary($topic)); ?></p></div>
            <p class="topic-count"><strong><?php echo esc_html(number_format_i18n($topic_total)); ?></strong><span><?php esc_html_e('篇记录', 'fieldnotes-journal'); ?></span></p>
        </div>
    </header>
    <div class="topic-layout">
        <section class="topic-articles" aria-label="<?php esc_attr_e('本主题的文章', 'fieldnotes-journal'); ?>">
            <?php if (have_posts()) : $topic_index = 0; ?>
                <div class="topic-posts">
                    <?php while (have_posts()) : the_post();
                        get_template_part('template-parts/topic-post', null, array('featured' => $topic_page === 1 && $topic_index === 0));
                        $topic_index++;
                    endwhile; ?>
                </div>
                <?php the_posts_pagination(array('mid_size' => 1, 'prev_text' => __('上一页', 'fieldnotes-journal'), 'next_text' => __('下一页', 'fieldnotes-journal'), 'screen_reader_text' => __('分类文章分页', 'fieldnotes-journal'))); ?>
            <?php else : ?><div class="topic-empty"><h2><?php esc_html_e('故事，正在路上。', 'fieldnotes-journal'); ?></h2><p><?php esc_html_e('这个主题还没有发布记录。先去别处逛逛，或许会遇到另一段喜欢的文字。', 'fieldnotes-journal'); ?></p><a class="editorial-text-link" href="<?php echo esc_url(home_url('/')); ?>"><?php esc_html_e('回到首页', 'fieldnotes-journal'); ?><span aria-hidden="true">↗</span></a></div><?php endif; ?>
        </section>
        <?php get_template_part('template-parts/category-sidebar', null, array('term' => $topic, 'profile' => $topic_profile)); ?>
    </div>
</main>
<?php get_footer(); ?>
