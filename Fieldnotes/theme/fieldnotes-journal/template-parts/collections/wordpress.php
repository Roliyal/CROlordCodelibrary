<?php
if (!defined('ABSPATH')) exit;
$components = array(
    array('Journal', '主题', FIELDNOTES_VERSION, '让内容，成为页面的主角。', '首页、文章与五个栏目，在同一套克制的排版里展开。为文字、影像和长期记录保留空间。', 'fieldnotes-journal-' . FIELDNOTES_VERSION . '.zip', array('系统字体与阅读排版', '五个 Boughs 栏目', '首页与文章页面')),
    array('Echoes', 'ChatGPT', '1.5.0', '给思绪，一处回响。', '口令保护的私人对话空间。文字、图片、文档与音频可以一起进入思考，也可以连接远程 MCP 工具。', 'fieldnotes-chatgpt-1.5.0.zip', array('附件与音频转写', '视频四帧视觉摘要', '远程 MCP 与逐项确认')),
    array('Wayfarer', 'Map', '1.4.0', '足迹不止于一个坐标。', '用地图串联走过的城市与当时的文字。地点、年份和文章相互连接，让记忆可以沿路重返。', 'fieldnotes-travel-map-1.4.0.zip', array('地点与文章关联', '国家与年份筛选', '本地底图回退')),
    array('Comment Tools', '评论插件', '1.6.0', '让回应，也有自己的表情。', '从昵称、邮箱到头像与图片表情，把评论写成自然的交流，延续整页的阅读节奏。', 'fieldnotes-comment-tools-1.6.0.zip', array('邮箱头像与自选头像', '图片表情编辑', '评论与审核工具')),
);
?>
<div class="boughs-software-grid">
    <?php foreach ($components as $index => $component) : ?>
        <article class="boughs-software boughs-software-<?php echo esc_attr((string) $index); ?>">
            <div class="boughs-software-mark" aria-hidden="true"><?php echo esc_html(array('J.', 'E.', 'W.', 'C.')[$index]); ?></div>
            <p class="boughs-software-meta"><?php echo esc_html($component[1]); ?><span><?php echo esc_html($component[2]); ?></span></p>
            <h2><?php echo esc_html($component[0]); ?></h2><h3><?php echo esc_html($component[3]); ?></h3><p class="boughs-software-copy"><?php echo esc_html($component[4]); ?></p>
            <details class="boughs-software-details"><summary>功能与使用</summary><ul><?php foreach ($component[6] as $feature) : ?><li><?php echo esc_html($feature); ?></li><?php endforeach; ?></ul><p><?php echo $index === 0 ? '通过外观中的主题设置调整页面；文章在 WordPress 编辑器维护。' : ($index === 1 ? '安装后进入「设置 → ChatGPT 接入」，填写服务器 API Key、口令及需要的 MCP 连接。' : ($index === 2 ? '在「旅行地点」维护地图数据，在文章编辑器关联地点。' : '在插件设置中管理评论头像、表情与审核规则。')); ?></p></details>
            <div class="boughs-software-actions"><a class="boughs-download" href="<?php echo esc_url('https://github.com/Roliyal/CROlordCodelibrary/raw/refs/heads/main/Fieldnotes/dist/' . $component[5]); ?>" download>下载 ZIP <span aria-hidden="true">↓</span></a><a class="boughs-source" href="<?php echo esc_url('https://github.com/Roliyal/CROlordCodelibrary/tree/main/Fieldnotes/' . ($index === 0 ? 'theme/fieldnotes-journal' : 'plugins/' . array('', 'fieldnotes-chatgpt', 'fieldnotes-travel-map', 'fieldnotes-comment-tools')[$index])); ?>" target="_blank" rel="noopener noreferrer">GitHub 源码 ↗</a></div><p class="boughs-software-caption"><?php echo $index === 0 ? '在 WordPress 外观中上传主题。' : '在 WordPress 插件中上传并启用。'; ?></p>
        </article>
    <?php endforeach; ?>
</div>
<?php if ($boughs_term && (int) $boughs_term->count > 0) fieldnotes_boughs_collection_posts($boughs_term, '设计与实现手记'); ?>
