<?php get_header(); ?>
<main id="main-content" class="page-shell page-home">
    <?php
    $profile = trim((string) get_theme_mod('fieldnotes_profile'));
    $site_description = trim((string) get_bloginfo('description'));
    $hero_copy = $profile ? $profile : ($site_description ? $site_description : __('关于旅行、日常与持续生长的私人记录。', 'fieldnotes-journal'));
    $archive_url = fieldnotes_boughs_url('cities');
    $about_page = fieldnotes_section_page('about');
    $latest = new WP_Query(array('post_type' => 'post', 'post_status' => 'publish', 'posts_per_page' => 8, 'ignore_sticky_posts' => true, 'orderby' => array('date' => 'DESC', 'ID' => 'DESC')));
    $latest_post = !empty($latest->posts) ? $latest->posts[0] : null;
    ?>
    <section class="home-hero" aria-labelledby="home-title">
        <div class="home-hero-copy">
            <p class="home-kicker"><span><?php esc_html_e('文字、影像，与日常。', 'fieldnotes-journal'); ?></span></p>
            <h1 id="home-title">把日常写下来，<br><em>让远方慢一点。</em></h1>
            <p class="home-hero-lead"><?php echo esc_html($hero_copy); ?></p>
            <div class="home-hero-actions">
                <?php if ($latest_post) : ?>
                    <a class="home-primary-link" href="<?php echo esc_url(get_permalink($latest_post)); ?>"><?php esc_html_e('开始阅读', 'fieldnotes-journal'); ?><svg viewBox="0 0 24 24" width="18" height="18" fill="none" stroke="currentColor" stroke-width="1.5" aria-hidden="true"><path d="M5 12h14m-6-6 6 6-6 6" stroke-linecap="round" stroke-linejoin="round"/></svg></a>
                <?php else : ?>
                    <span class="home-primary-link is-unavailable"><?php esc_html_e('静候新篇', 'fieldnotes-journal'); ?></span>
                <?php endif; ?>
                <?php if ($about_page) : ?><a class="home-secondary-link" href="<?php echo esc_url(get_permalink($about_page)); ?>"><?php echo esc_html(fieldnotes_section_label('about')); ?><svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="1.5" aria-hidden="true"><path d="M6 18 18 6M6 6h12v12" stroke-linecap="round" stroke-linejoin="round"/></svg></a><?php endif; ?>
            </div>
            <p class="home-hero-note"><span class="home-note-dot" aria-hidden="true"></span><?php esc_html_e('慢一点，记录正在发生的生活。', 'fieldnotes-journal'); ?></p>
        </div>
        <figure class="home-hero-visual">
            <img src="<?php echo esc_url(get_template_directory_uri() . '/assets/photos/kyoto-rain.webp'); ?>" alt="雨后的京都街巷与远处的塔影" width="1536" height="1024" fetchpriority="high">
            <figcaption class="home-hero-caption"><span>行记 · 一</span><span><?php esc_html_e('京都 · 雨后', 'fieldnotes-journal'); ?></span></figcaption>
            <span class="home-hero-index" aria-hidden="true">01 <i></i> 04</span>
        </figure>
        <a class="home-scroll-cue" href="#latest-notes"><span><?php esc_html_e('向下探索', 'fieldnotes-journal'); ?></span><i aria-hidden="true"></i></a>
    </section>

    <section class="home-journal" id="latest-notes" aria-labelledby="latest-title">
        <header class="home-section-head">
            <div>
                <h2 id="latest-title"><?php esc_html_e('日常与远方，慢慢显影。', 'fieldnotes-journal'); ?></h2>
            </div>
            <a class="home-all-link" href="<?php echo esc_url($archive_url); ?>"><?php esc_html_e('城市漫步', 'fieldnotes-journal'); ?><svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="1.5" aria-hidden="true"><path d="M6 18 18 6M6 6h12v12" stroke-linecap="round" stroke-linejoin="round"/></svg></a>
        </header>
        <div class="columns">
            <section class="home-post-column">
                <h3 class="screen-reader-text"><?php esc_html_e('最新文章', 'fieldnotes-journal'); ?></h3>
                <?php if ($latest->have_posts()) : ?><div class="post-list" id="fieldnotes-post-list" data-page="1" data-max-pages="<?php echo esc_attr($latest->max_num_pages); ?>"><?php while ($latest->have_posts()) : $latest->the_post(); get_template_part('template-parts/post-card'); endwhile; ?></div>
                    <nav class="pagination" aria-label="<?php esc_attr_e('文章分页', 'fieldnotes-journal'); ?>"><span><?php if ($latest->max_num_pages > 1) : ?><button class="load-more" id="fieldnotes-load-more" type="button"><?php esc_html_e('加载更多文章', 'fieldnotes-journal'); ?><span aria-hidden="true">↓</span></button><?php endif; ?></span><a href="<?php echo esc_url($archive_url); ?>"><?php esc_html_e('继续读城市漫步', 'fieldnotes-journal'); ?><span aria-hidden="true">↗</span></a></nav>
                <?php else : ?><p class="empty-state"><?php esc_html_e('博客已准备就绪。发布第一篇文章后，首页会显示文章列表。', 'fieldnotes-journal'); ?></p><?php endif; wp_reset_postdata(); ?>
            </section>
            <?php get_template_part('template-parts/sidebar'); ?>
        </div>
    </section>
</main>
<?php get_footer(); ?>
