<?php if (!defined('ABSPATH')) exit; ?>
<div class="boughs-city-intro"><p>这里收存文字，地图另有一页。</p><a class="boughs-text-link" href="<?php echo esc_url(fieldnotes_section_url('map')); ?>">去 Wayfarer 看足迹 <span aria-hidden="true">↗</span></a></div>
<?php fieldnotes_boughs_collection_posts($boughs_term, '日常的片刻'); ?>
