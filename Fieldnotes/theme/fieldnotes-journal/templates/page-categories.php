<?php
/* Template Name: Boughs */
if (!defined('ABSPATH')) exit;
get_header();
$key = fieldnotes_room_context();
if ($key) { fieldnotes_render_room($key); get_footer(); return; }
$sections = fieldnotes_boughs_sections();
?>
<main id="main-content" class="page-shell boughs-page boughs-index">
    <header class="boughs-index-heading"><h1>Boughs<span aria-hidden="true">.</span></h1><p>生活的几个方向。<br><span>各自生长，也彼此相连。</span></p></header>
    <?php while (have_posts()) : the_post(); if (trim(get_the_content())) : ?><div class="boughs-custom-copy"><?php the_content(); ?></div><?php endif; endwhile; ?>
    <div class="boughs-index-list">
        <?php foreach ($sections as $slug => $section) : ?>
        <a class="boughs-index-item boughs-index-<?php echo esc_attr($slug); ?>" href="<?php echo esc_url(fieldnotes_boughs_url($slug)); ?>">
            <div class="boughs-index-copy"><h2><?php echo esc_html($slug === 'investing' ? 'Markets' : $section['name']); ?></h2><p><?php echo esc_html($section['description']); ?></p><span class="boughs-index-open">慢慢展开 <span aria-hidden="true">↗</span></span></div>
            <figure class="boughs-index-image">
                <?php if ($slug === 'investing') : ?><div class="boughs-market-art" aria-hidden="true"><span>观察。<br>思考。<br><i>留给时间。</i></span></div>
                <?php elseif ($slug === 'wordpress') : ?><div class="boughs-software-art" aria-hidden="true"><span>J<span>.</span></span><small>一个主题 · 三个插件</small></div>
                <?php elseif ($slug === 'minimalism') : ?><img src="<?php echo esc_url(get_template_directory_uri() . '/assets/demo/still-life.svg'); ?>" alt="" width="800" height="500" loading="lazy">
                <?php else : ?><img src="<?php echo esc_url(get_template_directory_uri() . '/assets/photos/' . $section['image']); ?>" alt="" width="800" height="500" loading="<?php echo $slug === 'cities' ? 'eager' : 'lazy'; ?>"><?php endif; ?>
            </figure>
        </a>
        <?php endforeach; ?>
    </div>
</main>
<?php get_footer(); ?>
