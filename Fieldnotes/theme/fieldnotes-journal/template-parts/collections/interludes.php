<?php if (!defined('ABSPATH')) exit; ?>
<section class="boughs-practice-opening">
    <div><p class="boughs-small-label">云原生实践</p><h2>把问题，写成<br><span>可复现的实践。</span></h2></div>
    <p>环境与版本，方案与配置，验证与复盘。<br>把一次尝试的来龙去脉，留给之后的自己。</p>
</section>
<section class="boughs-repositories" aria-label="代码仓库">
    <?php foreach (array(
        array('CROlordCloudNative', '云原生实践的代码入口。', '云原生 · 实践'),
        array('CROlordCodelibrary', '代码与日常工程记录的入口。', '代码 · 工具')
    ) as $repository) : ?>
        <a class="boughs-repository" href="<?php echo esc_url('https://github.com/Roliyal/' . $repository[0]); ?>" target="_blank" rel="noopener noreferrer">
            <span class="boughs-repository-icon" aria-hidden="true"><svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.2"><path d="m8 7-5 5 5 5m8-10 5 5-5 5M14 4l-4 16"/></svg></span>
            <div><span class="boughs-small-label"><?php echo esc_html($repository[2]); ?></span><h2><?php echo esc_html($repository[0]); ?></h2><p><?php echo esc_html($repository[1]); ?></p></div><span class="boughs-repository-arrow" aria-hidden="true">↗</span>
        </a>
    <?php endforeach; ?>
</section>
<div class="boughs-practice-method"><span>从一个问题开始。</span><p>记录环境 <i aria-hidden="true">→</i> 展开实现 <i aria-hidden="true">→</i> 验证结果</p></div>
<?php fieldnotes_boughs_collection_posts($boughs_term, '实践与拙作'); ?>
