<?php
/* Template Name: Journal profile */
get_header();
?>
<main id="main-content" class="page-shell profile-page">
    <?php while (have_posts()) : the_post(); ?>
        <?php if (post_password_required()) : ?>
            <section class="profile-locked"><h1><?php the_title(); ?></h1><?php the_content(); ?></section>
        <?php else :
            $profile = fieldnotes_profile_details(get_post_field('post_author', get_the_ID()));
            $profile_author = $profile['author'];
        ?>
            <header class="profile-opening">
                <p class="profile-eyebrow"><?php echo esc_html(get_the_title()); ?></p>
                <div class="profile-opening-grid">
                    <h1><?php echo nl2br(esc_html($profile['heading'])); ?></h1>
                    <div class="profile-introduction">
                        <div class="profile-identity">
                            <?php echo fieldnotes_author_portrait($profile_author, 64); ?>
                            <div><p class="profile-name"><?php echo esc_html($profile_author['name']); ?></p><?php if ($profile['role'] !== '') : ?><p class="profile-role"><?php echo esc_html($profile['role']); ?></p><?php endif; ?></div>
                        </div>
                        <p class="profile-lead"><?php echo nl2br(esc_html($profile['intro'])); ?></p>
                    </div>
                </div>
            </header>
            <figure class="profile-figure">
                <div class="profile-image"><?php if (has_post_thumbnail()) : the_post_thumbnail('full', array('loading' => 'eager', 'fetchpriority' => 'high', 'sizes' => '(max-width: 680px) calc(100vw - 40px), (max-width: 1280px) calc(100vw - 112px), 1168px')); else : ?><img src="<?php echo esc_url(get_template_directory_uri() . '/assets/photos/studio-notes.webp'); ?>" alt="<?php esc_attr_e('窗边的写作桌面，笔记、相机与自然光影', 'fieldnotes-journal'); ?>" width="1400" height="1000" loading="eager" fetchpriority="high"><?php endif; ?></div>
                <figcaption><span><?php esc_html_e('为平常的一天，留一个镜头。', 'fieldnotes-journal'); ?></span></figcaption>
            </figure>
            <section class="profile-story" aria-labelledby="profile-story-title">
                <div class="profile-section-heading"><p class="profile-eyebrow"><?php esc_html_e('关于我', 'fieldnotes-journal'); ?></p><h2 id="profile-story-title"><?php esc_html_e('愿意为喜欢的事，', 'fieldnotes-journal'); ?><br><?php esc_html_e('多花一点时间。', 'fieldnotes-journal'); ?></h2></div>
                <div class="profile-body-copy">
                    <?php if (trim(get_the_content())) : the_content(); wp_link_pages(); else :
                        $profile_bio = fieldnotes_clean_text(get_the_author_meta('description', get_post_field('post_author', get_the_ID())));
                    ?>
                        <p><?php echo esc_html($profile_bio ?: __('喜欢把复杂的东西理清楚，也愿意为一束光、一段路多停留一会儿。工作之外，走走看看，读一点书，把注意力从屏幕移回生活。', 'fieldnotes-journal')); ?></p>
                        <p><?php esc_html_e('这个空间，收着旅途的片段、使用过的物件，以及尚未成形的想法。写得不必快，答案也不必完整。过一阵子再回来，能认出那时的自己，就很好。', 'fieldnotes-journal'); ?></p>
                    <?php endif; ?>
                </div>
            </section>
            <section class="profile-practice" aria-labelledby="profile-practice-title">
                <div class="profile-section-heading"><h2 id="profile-practice-title"><?php esc_html_e('把注意力，留给值得的事。', 'fieldnotes-journal'); ?></h2></div>
                <div class="profile-practice-grid">
                    <article><p class="profile-practice-index">其一</p><h3><?php esc_html_e('造物', 'fieldnotes-journal'); ?></h3><p><?php esc_html_e('让一个想法有形，让一件常用的工具更顺手。认真做完，也给下一次尝试留一点余地。', 'fieldnotes-journal'); ?></p></article>
                    <article><p class="profile-practice-index">其二</p><h3><?php esc_html_e('观物', 'fieldnotes-journal'); ?></h3><p><?php esc_html_e('在城市里慢慢走，用文字与摄影留住细节。许多喜欢的瞬间，都发生在计划之外。', 'fieldnotes-journal'); ?></p></article>
                    <article><p class="profile-practice-index">其三</p><h3><?php esc_html_e('惜物', 'fieldnotes-journal'); ?></h3><p><?php esc_html_e('留下适合自己的物件、习惯和关系。安静、可靠，也经得起时间。', 'fieldnotes-journal'); ?></p></article>
                </div>
            </section>
            <?php $profile_links = fieldnotes_profile_destinations(); if ($profile_links) : ?>
            <nav class="profile-explore" aria-label="<?php esc_attr_e('继续浏览这个空间', 'fieldnotes-journal'); ?>">
                <h2><?php esc_html_e('在这里，继续走走。', 'fieldnotes-journal'); ?></h2>
                <div class="profile-explore-grid profile-explore-count-<?php echo esc_attr(count($profile_links)); ?>">
                    <?php foreach ($profile_links as $profile_link) : ?><a href="<?php echo esc_url($profile_link['url']); ?>"><span class="profile-explore-label"><?php echo esc_html($profile_link['label']); ?></span><h3><?php echo esc_html($profile_link['title']); ?></h3><p><?php echo esc_html($profile_link['description']); ?></p><svg viewBox="0 0 24 24" width="22" height="22" fill="none" stroke="currentColor" stroke-width="1.4" aria-hidden="true"><path d="M5 12h14m-6-6 6 6-6 6" stroke-linecap="round" stroke-linejoin="round"/></svg></a><?php endforeach; ?>
                </div>
            </nav>
            <?php endif; ?>
        <?php endif; ?>
    <?php endwhile; ?>
</main>
<?php get_footer(); ?>
