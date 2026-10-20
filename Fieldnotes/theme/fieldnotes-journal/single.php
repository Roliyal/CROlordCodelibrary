<?php get_header(); ?>
<main id="main-content" class="page-shell reading-page">
    <?php while (have_posts()) : the_post();
        $reading_locked = post_password_required();
        $reading_author = fieldnotes_article_author(get_post_field('post_author', get_the_ID()));
        $reading_minutes = fieldnotes_reading_minutes();
    ?>
    <article <?php post_class('entry reading-entry'); ?> itemscope itemtype="https://schema.org/BlogPosting">
        <header class="entry-header reading-header">
            <div class="reading-kicker">
                <div class="entry-categories" aria-label="<?php esc_attr_e('文章分类', 'fieldnotes-journal'); ?>">
                    <?php $entry_categories = get_the_category(); if ($entry_categories) : foreach ($entry_categories as $entry_category) : ?>
                        <a class="reading-category" href="<?php echo esc_url(get_category_link($entry_category->term_id)); ?>"><?php echo esc_html(fieldnotes_category_label($entry_category)); ?></a>
                    <?php endforeach; endif; ?>
                </div>
            </div>
            <h1 class="entry-title" itemprop="headline"><?php the_title(); ?></h1>
            <?php if (!$reading_locked && has_excerpt()) : ?><p class="reading-lead"><?php echo esc_html(fieldnotes_clean_text(get_the_excerpt())); ?></p><?php endif; ?>
            <div class="reading-byline">
                <span class="screen-reader-text" itemprop="author" itemscope itemtype="https://schema.org/Person"><span itemprop="name"><?php echo esc_html($reading_author['name']); ?></span></span>
                <time itemprop="datePublished" datetime="<?php echo esc_attr(get_the_date('c')); ?>"><?php echo esc_html(get_the_date('Y年n月j日')); ?></time>
                <?php if ($reading_minutes) : ?><span class="reading-duration"><?php printf(esc_html__('约 %s 分钟阅读', 'fieldnotes-journal'), esc_html(number_format_i18n($reading_minutes))); ?></span><?php endif; ?>
            </div>
        </header>
        <?php if (!$reading_locked) : ?>
        <div class="entry-featured reading-cover">
            <?php if (has_post_thumbnail()) : the_post_thumbnail('large', array('itemprop' => 'image', 'loading' => 'eager', 'fetchpriority' => 'high', 'sizes' => '(max-width: 680px) calc(100vw - 40px), (max-width: 900px) calc(100vw - 80px), 820px')); else : ?>
                <img src="<?php echo esc_url(fieldnotes_demo_image(get_the_ID())); ?>" alt="" itemprop="image" loading="eager" fetchpriority="high" width="1400" height="1000">
            <?php endif; ?>
        </div>
        <?php endif; ?>
        <div class="reading-story">
                <div class="entry-content" itemprop="articleBody"><?php the_content(); wp_link_pages(); ?></div>
                <?php if (!$reading_locked) :
                    $reading_previous = get_previous_post(true, '', 'category');
                    $reading_next = get_next_post(true, '', 'category');
                    if ($reading_previous || $reading_next) : ?>
                    <nav class="reading-next" aria-label="<?php esc_attr_e('继续阅读同主题文章', 'fieldnotes-journal'); ?>">
                        <h2><?php esc_html_e('接着读。', 'fieldnotes-journal'); ?></h2>
                        <div class="reading-next-grid">
                            <?php foreach (array($reading_previous, $reading_next) as $reading_neighbour) : if (!$reading_neighbour) continue; ?>
                                <a href="<?php echo esc_url(get_permalink($reading_neighbour)); ?>"><span><?php echo esc_html(get_the_title($reading_neighbour)); ?></span><svg viewBox="0 0 24 24" width="20" height="20" fill="none" stroke="currentColor" stroke-width="1.4" aria-hidden="true"><path d="M5 12h14m-6-6 6 6-6 6" stroke-linecap="round" stroke-linejoin="round"/></svg></a>
                            <?php endforeach; ?>
                        </div>
                    </nav>
                    <?php endif;
                endif; ?>
        </div>
    </article>
    <?php if (comments_open() || get_comments_number()) comments_template(); endwhile; ?>
</main>
<?php get_footer(); ?>
