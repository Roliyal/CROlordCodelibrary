<?php
/* Template Name: Travel map */
get_header(); ?>
<main id="main-content" class="page-shell">
    <?php while (have_posts()) : the_post(); ?>
        <div class="map-page-heading">
            <span class="map-page-emblem" aria-hidden="true"><svg viewBox="0 0 72 72" role="presentation"><circle cx="36" cy="36" r="31" fill="none" stroke="currentColor" stroke-width="1.5"/><path d="M12 45h48M18 44V31l11-9 12 10v12M41 44V27l7-6 9 7v16M26 44V35h9v9M46 34h7v10M15 53c7-4 13-4 20 0s13 4 22 0" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"/><path d="M22 29h4m18 2h5" stroke="currentColor" stroke-width="1.8" stroke-linecap="round"/></svg></span>
            <h1 class="page-title"><?php the_title(); ?></h1>
        </div>
        <p class="page-description"><?php esc_html_e('山海有相逢，行路有回声。', 'fieldnotes-journal'); ?></p>
        <?php if (trim(get_the_content())) : ?><div class="page-description"><?php the_content(); ?></div><?php endif; ?>
    <?php endwhile; ?>
    <?php if (shortcode_exists('fieldnotes_travel_map')) : ?>
        <?php echo do_shortcode('[fieldnotes_travel_map]'); ?>
    <?php elseif (shortcode_exists('marker_pro_map')) : ?>
        <?php echo do_shortcode('[marker_pro_map]'); ?>
    <?php else : ?>
        <div class="map-canvas map-plugin-required"><?php esc_html_e('请安装并启用 Fieldnotes Travel Map 插件以显示交互式旅行地图。', 'fieldnotes-journal'); ?></div>
    <?php endif; ?>
</main>
<?php get_footer(); ?>
