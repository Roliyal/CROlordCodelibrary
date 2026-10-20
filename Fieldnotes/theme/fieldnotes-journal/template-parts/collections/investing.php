<?php
if (!defined('ABSPATH')) exit;
$symbol_input = isset($_GET['stock']) && is_string($_GET['stock']) ? wp_unslash($_GET['stock']) : '';
$symbol = $symbol_input !== '' ? fieldnotes_market_symbol($symbol_input) : null;
if (is_wp_error($symbol)) $symbol = null;
$notes_page = isset($_GET['notes_page']) && is_scalar($_GET['notes_page']) ? max(1, absint($_GET['notes_page'])) : 1;
$notes = fieldnotes_market_notes_data($symbol, $notes_page);
?>
<section class="fn-market<?php echo $symbol ? ' has-stock' : ''; ?>" data-symbol="<?php echo esc_attr($symbol ? $symbol['canonical'] : ''); ?>" data-notes-page="<?php echo esc_attr($notes_page); ?>" aria-label="股票分时、历史 K 线与观察记录">
    <form class="fn-market-search" method="get" action="<?php echo esc_url(fieldnotes_boughs_url('investing')); ?>">
        <?php $market_url = wp_parse_url(fieldnotes_boughs_url('investing')); parse_str($market_url['query'] ?? '', $market_query); foreach ($market_query as $key => $value) : if (is_scalar($value) && !in_array($key, array('stock', 'period', 'adjust', 'notes_page'), true)) : ?><input type="hidden" name="<?php echo esc_attr($key); ?>" value="<?php echo esc_attr($value); ?>"><?php endif; endforeach; ?>
        <label for="fn-market-code">从一只股票开始。</label>
        <div class="fn-market-search-wrap"><div class="fn-market-search-field"><input id="fn-market-code" name="stock" value="<?php echo esc_attr($symbol ? $symbol['canonical'] : ''); ?>" placeholder="股票名称或代码，如 贵州茅台" maxlength="32" autocomplete="off" spellcheck="false" role="combobox" aria-autocomplete="list" aria-expanded="false" aria-controls="fn-market-matches" aria-describedby="fn-market-search-hint"><button type="submit">展开行情 <span aria-hidden="true">↗</span></button></div><div class="fn-market-matches" id="fn-market-matches" role="listbox" aria-label="匹配的股票" hidden></div></div>
        <p id="fn-market-search-hint" role="status">沪深北 A 股 · 中文名称、简拼或六位代码</p>
    </form>
    <section class="fn-market-watchlist" aria-label="临时自选" hidden><div class="fn-market-watchlist-heading"><span>临时自选</span><small>仅保留在此标签页</small><button type="button" class="fn-market-watch-clear">清空</button></div><div class="fn-market-watch-items"></div></section>
    <div class="fn-market-panel">
        <header class="fn-market-panel-heading">
            <div><h2 class="fn-market-stock-title"><?php echo $symbol ? esc_html($symbol['canonical']) : '价格的变化，留给图表。'; ?></h2><p class="fn-market-stock-subtitle">判断的来由，留在文字里。</p></div>
            <div class="fn-market-periods" role="group" aria-label="行情视图"><button type="button" data-period="minute" aria-pressed="true">分时</button><button type="button" data-period="day" aria-pressed="false">日</button><button type="button" data-period="week" aria-pressed="false">周</button><button type="button" data-period="month" aria-pressed="false">月</button></div>
        </header>
        <div class="fn-market-live" hidden><div><strong class="fn-market-live-price"></strong><span class="fn-market-live-change"></span></div><p class="fn-market-live-time"></p></div>
        <div class="fn-market-auto"><label for="fn-market-interval">自动刷新<select id="fn-market-interval"><option value="2">每 2 秒</option><option value="5" selected>每 5 秒</option><option value="10">每 10 秒</option><option value="30">每 30 秒</option><option value="0">暂停</option></select></label><span class="fn-market-auto-state" role="status">选择股票后开始</span></div>
        <div class="fn-market-toolbar">
            <div class="fn-market-ma-legend"><span class="ma5">MA5</span><span class="ma10">MA10</span><span class="ma20">MA20</span></div>
            <div><label class="screen-reader-text" for="fn-market-adjust">复权方式</label><select id="fn-market-adjust"><option value="qfq">前复权</option><option value="none">不复权</option></select><button class="fn-market-zoom-out" type="button" aria-label="显示更多 K 线">−</button><button class="fn-market-zoom-in" type="button" aria-label="放大 K 线">＋</button><button class="fn-market-zoom-reset" type="button">重置</button><button class="fn-market-refresh" type="button" aria-label="刷新当前股票行情">刷新</button></div>
        </div>
        <div class="fn-market-chart" tabindex="0" role="group" aria-label="股票 K 线图，使用左右方向键查看数据">
            <svg class="fn-market-svg" viewBox="0 0 1000 500" role="img" aria-label="等待查询股票历史行情"></svg>
            <div class="fn-market-chart-empty"><span aria-hidden="true">／</span><p>输入名称或代码，展开一段价格记录。</p></div>
            <div class="fn-market-tooltip" hidden></div>
        </div>
        <p class="fn-market-status" role="status" aria-live="polite"></p>
        <div class="fn-market-provenance"><span class="fn-market-source">历史行情 · 可能存在延迟</span><button class="fn-market-retry" type="button" hidden>重新读取</button></div>
        <details class="fn-market-data-table" hidden><summary>查看近期数据</summary><div><table><thead><tr><th>日期</th><th>开盘</th><th>收盘</th><th>最高</th><th>最低</th><th>成交量（手）</th></tr></thead><tbody></tbody></table></div></details>
        <noscript><p class="fn-market-status">启用 JavaScript 后可显示交互 K 线。下方文章仍可直接阅读。</p></noscript>
    </div>
    <section class="fn-market-observations" aria-labelledby="fn-market-notes-title">
        <div class="boughs-section-heading"><div><h2 id="fn-market-notes-title">观察与复盘</h2><p class="fn-market-notes-caption"><?php echo $symbol ? esc_html($symbol['canonical'] . ' 的记录') : '市场的变化，和自己的思考。'; ?></p></div><?php if (current_user_can('edit_posts')) : ?><a class="boughs-text-link" href="#fn-market-write">写一段观察 ↗</a><?php endif; ?></div>
        <dl class="fn-market-note-stats"><div><dt>全部观察</dt><dd data-stat="all"><?php echo esc_html($notes['all_total']); ?></dd></div><div data-match-stat><dt class="fn-market-match-label"><?php echo $symbol ? '这只股票的记录' : '已发表记录'; ?></dt><dd data-stat="matched"><?php echo esc_html($notes['total']); ?></dd></div><div><dt>本页最近一笔</dt><dd data-stat="latest"><?php echo esc_html($notes['latest'] ?: '待续'); ?></dd></div></dl>
        <div class="fn-market-note-list" aria-live="polite">
            <?php foreach ($notes['posts'] as $note) : ?><article class="fn-market-note"><a class="fn-market-note-link" href="<?php echo esc_url($note['url']); ?>"><time datetime="<?php echo esc_attr($note['iso_date']); ?>"><?php echo esc_html($note['date']); ?></time><div><h3><?php echo esc_html($note['title']); ?></h3><p><?php echo esc_html($note['excerpt']); ?></p></div><span aria-hidden="true">↗</span></a></article><?php endforeach; ?>
            <?php if (!$notes['posts']) : ?><div class="fn-market-empty"><span aria-hidden="true">＋</span><h3>下一次判断，从这里留下。</h3><p>写下看见的变化，保留当时的依据。回看时，才有迹可循。</p></div><?php endif; ?>
        </div>
        <nav class="fn-market-note-pagination" aria-label="观察记录分页"><?php for ($page = max(1, $notes['page'] - 2); $page <= min($notes['pages'], $notes['page'] + 2); $page++) : ?><a data-notes-page="<?php echo esc_attr($page); ?>" href="<?php echo esc_url(add_query_arg(array('stock' => $symbol ? $symbol['canonical'] : '', 'notes_page' => $page), fieldnotes_boughs_url('investing'))); ?>"<?php echo $page === $notes['page'] ? ' aria-current="page"' : ''; ?>><?php echo esc_html($page); ?></a><?php endfor; ?></nav>
    </section>
    <?php if (current_user_can('edit_posts')) : ?>
    <details class="fn-market-write" id="fn-market-write">
        <summary>写一段观察 <span aria-hidden="true">＋</span></summary>
        <form action="<?php echo esc_url(admin_url('admin-post.php')); ?>" method="post">
            <input type="hidden" name="action" value="fieldnotes_market_note"><?php wp_nonce_field('fieldnotes_market_note'); ?>
            <div class="fn-market-write-top"><label>股票代码<input name="stock" value="<?php echo esc_attr($symbol ? $symbol['canonical'] : ''); ?>" placeholder="例如 600519" maxlength="16" required></label><label>记录标题<input name="note_title" placeholder="为这次观察，写一个标题" maxlength="160" required></label></div>
            <label>观察与依据<textarea name="note_content" rows="7" maxlength="24000" placeholder="看到了什么？判断的依据是什么？哪些变化值得下一次回看？" required></textarea></label>
            <div class="fn-market-write-bottom"><p>保存为 WordPress 草稿，进入编辑器继续完善与发布。</p><button type="submit">保存草稿，继续编辑 ↗</button></div>
        </form>
    </details>
    <?php endif; ?>
</section>
