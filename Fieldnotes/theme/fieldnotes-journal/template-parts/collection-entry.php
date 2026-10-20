<?php if (!defined('ABSPATH')) exit; $featured = !empty($args['featured']); ?>
<article class="boughs-entry<?php echo $featured ? ' is-featured' : ''; ?>">
    <a class="boughs-entry-link" href="<?php the_permalink(); ?>">
        <figure><?php if (has_post_thumbnail()) : the_post_thumbnail('large'); else : ?><img src="<?php echo esc_url(fieldnotes_demo_image(get_the_ID())); ?>" alt="" width="900" height="600" loading="lazy"><?php endif; ?></figure>
        <div class="boughs-entry-copy"><time datetime="<?php echo esc_attr(get_the_date('c')); ?>"><?php echo esc_html(get_the_date('Y.m.d')); ?></time><h3><?php the_title(); ?></h3><p><?php echo esc_html(post_password_required() ? '这篇文字有口令保护。' : wp_trim_words(get_the_excerpt(), 38, '…')); ?></p><span aria-hidden="true">↗</span></div>
    </a>
</article>
