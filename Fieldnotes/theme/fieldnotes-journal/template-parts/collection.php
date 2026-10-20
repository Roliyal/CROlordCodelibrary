<?php if (!defined('ABSPATH')) exit; ?>
<main id="main-content" class="page-shell boughs-page boughs-room boughs-<?php echo esc_attr($boughs_key); ?>">
    <header class="boughs-heading">
        <div><h1><?php if ($boughs_key === 'wordpress') : ?>WordPress<span class="boughs-heading-detail">主题与插件</span><?php else : ?><?php echo esc_html(fieldnotes_room_title($boughs_key)); ?><?php endif; ?></h1><?php if ($boughs_key === 'minimalism') : ?><p class="boughs-room-subtitle">生活里的长期之选。</p><?php endif; ?></div>
        <p><?php echo esc_html($boughs_section['intro']); ?></p>
    </header>
    <?php include get_template_directory() . '/template-parts/collections/' . $boughs_key . '.php'; ?>
</main>
