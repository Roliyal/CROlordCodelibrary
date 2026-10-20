<?php get_header(); ?>
<main id="main-content" class="page-shell">
    <div class="columns">
        <section>
            <h1 class="section-heading"><?php echo is_home() || is_front_page() ? esc_html__('最新文章', 'fieldnotes-journal') : esc_html(fieldnotes_clean_text(get_the_archive_title())); ?></h1>
            <?php if (have_posts()) : ?><div class="post-list"><?php while (have_posts()) : the_post(); get_template_part('template-parts/post-card'); endwhile; ?></div>
                <nav class="pagination" aria-label="<?php esc_attr_e('文章导航', 'fieldnotes-journal'); ?>"><span><?php previous_posts_link(__('← 较新文章', 'fieldnotes-journal')); ?></span><span><?php next_posts_link(__('更早文章 →', 'fieldnotes-journal')); ?></span></nav>
            <?php else : ?><p class="empty-state"><?php esc_html_e('还没有文章。请在 WordPress 后台发布文章。', 'fieldnotes-journal'); ?></p><?php endif; ?>
        </section>
        <?php get_template_part('template-parts/sidebar'); ?>
    </div>
</main>
<?php get_footer(); ?>
