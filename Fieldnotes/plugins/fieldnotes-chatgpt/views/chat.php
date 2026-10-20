<?php
if (!defined('ABSPATH')) exit;
$icon = function ($name, $size = 20) {
    $paths = array(
        'arrow' => '<path d="M12 19V5m-6 6 6-6 6 6"/>',
        'plus' => '<path d="M12 5v14M5 12h14"/>',
        'lock' => '<rect x="5" y="10" width="14" height="11" rx="3"/><path d="M8 10V7a4 4 0 0 1 8 0v3"/>',
        'new' => '<path d="M12 4H6a2 2 0 0 0-2 2v12a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2v-6M16 4l4 4M10 14l1-5 8-8 4 4-8 8-5 1Z"/>',
        'tools' => '<path d="m12 3 8 4.5v9L12 21l-8-4.5v-9L12 3ZM4 7.5l8 4.5 8-4.5M12 12v9M8 5.3l8 4.5"/>',
        'eye' => '<path d="M2.5 12s3.5-6 9.5-6 9.5 6 9.5 6-3.5 6-9.5 6-9.5-6-9.5-6Z"/><circle cx="12" cy="12" r="2.5"/>',
        'chevron' => '<path d="m9 5 7 7-7 7"/>',
        'close' => '<path d="m6 6 12 12M6 18 18 6"/>',
    );
    echo '<svg viewBox="0 0 24 24" width="' . (int) $size . '" height="' . (int) $size . '" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">' . ($paths[$name] ?? '') . '</svg>';
};
?>
<section class="fnchat" data-ready="<?php echo $api_ready ? '1' : '0'; ?>" data-state="<?php echo $authorized && $this->transport_allowed() ? 'conversation' : 'gate'; ?>" aria-label="Echoes 私人对话">
    <?php if (!$authorized || !$this->transport_allowed()) : ?>
        <div class="fnchat-gate">
            <div class="fnchat-gate-intro">
                <span class="fnchat-signature">Echoes<span aria-hidden="true">.</span></span>
                <h2>为思绪，<br><span>留一处回响。</span></h2>
                <p class="fnchat-gate-copy">文字、影像，和未完成的念头。<br>在这里，慢慢找到彼此的联系。</p>
                <div class="fnchat-gate-caption"><span aria-hidden="true">01</span><p>一方私人的思考空间</p></div>
            </div>
            <div class="fnchat-gate-entry">
                <span class="fnchat-gate-symbol" aria-hidden="true"><?php $icon('lock', 24); ?></span>
                <?php if (!$this->transport_allowed()) : ?>
                    <p class="fnchat-notice">HTTP 测试已关闭。请使用 HTTPS，或在后台开启 HTTP 测试。</p>
                <?php elseif (!$password_ready || !$page_id) : ?>
                    <p class="fnchat-notice">这处空间尚未准备好。
                    <?php if (current_user_can('manage_options')) : ?><a href="<?php echo esc_url(admin_url('options-general.php?page=fieldnotes-chatgpt')); ?>">完成接入设置</a><?php endif; ?></p>
                <?php else : ?>
                    <form class="fnchat-unlock" novalidate>
                        <h3>欢迎回来。</h3>
                        <label for="fnchat-password">用口令，打开这处空间</label>
                        <div class="fnchat-password-field">
                            <input id="fnchat-password" name="password" type="password" autocomplete="current-password" placeholder="访问口令" maxlength="512" required aria-describedby="fnchat-unlock-status">
                            <button class="fnchat-show-password" type="button" aria-label="显示口令" aria-pressed="false" data-show="显示口令" data-hide="隐藏口令"><?php $icon('eye'); ?></button>
                        </div>
                        <button class="fnchat-enter" type="submit"><span>进入对话</span><?php $icon('chevron', 18); ?></button>
                        <p id="fnchat-unlock-status" class="fnchat-form-status" role="status" aria-live="polite"></p>
                    </form>
                <?php endif; ?>
                <p class="fnchat-gate-footnote">留白，也是对话的一部分。</p>
            </div>
        </div>
    <?php else : ?>
        <div class="fnchat-shell">
            <header class="fnchat-toolbar">
                <div class="fnchat-product"><span class="fnchat-signature">Echoes<span aria-hidden="true">.</span></span><small><?php echo esc_html($settings['model']); ?></small></div>
                <div class="fnchat-tools">
                    <button class="fnchat-export fnchat-icon-button" type="button" aria-label="导出本次对话" title="导出对话"><svg viewBox="0 0 24 24" width="19" height="19" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true"><path d="M12 3v12m-4-4 4 4 4-4M5 15v5h14v-5"/></svg></button>
                    <button class="fnchat-clear fnchat-icon-button" type="button" aria-label="开启新对话" title="新对话"><?php $icon('new', 19); ?><span>新对话</span></button>
                    <button class="fnchat-lock fnchat-icon-button" type="button" aria-label="锁定这处空间" title="锁定"><?php $icon('lock', 18); ?></button>
                </div>
            </header>
            <?php if (!$api_ready) : ?>
                <p class="fnchat-notice">尚未配置 OpenAI API Key，暂时无法发送消息。<?php if (current_user_can('manage_options')) : ?><a href="<?php echo esc_url(admin_url('options-general.php?page=fieldnotes-chatgpt')); ?>">配置 API Key</a><?php endif; ?></p>
            <?php endif; ?>
            <div class="fnchat-log" role="log" aria-label="对话记录" aria-live="polite" aria-relevant="additions" tabindex="0">
                <div class="fnchat-welcome">
                    <div class="fnchat-orbit" aria-hidden="true"><svg viewBox="0 0 180 110" fill="none"><ellipse cx="90" cy="55" rx="66" ry="26" transform="rotate(-24 90 55)" stroke="#d2d2d7" stroke-width=".8"/><ellipse cx="90" cy="55" rx="43" ry="43" stroke="#b7b7bc" stroke-width=".8"/><path d="M90 12v86M47 55h86" stroke="#e5e5e8" stroke-width=".8"/><circle cx="127" cy="32" r="4" fill="#1d1d1f"/></svg></div>
                    <p class="fnchat-eyebrow">一方思考的留白</p>
                    <h2>把想法，<span>放在这里。</span></h2>
                    <p class="fnchat-welcome-copy">一段文字，一张照片，或一个尚未成形的念头。</p>
                    <div class="fnchat-prompts">
                        <button type="button" data-prompt="请帮我把一个模糊的想法，整理成清晰的思路。"><strong>整理思绪</strong><span>让零散的念头渐渐清晰</span><?php $icon('chevron', 14); ?></button>
                        <button type="button" data-prompt="请帮我润色一段文字，保持自然、克制、有温度。"><strong>打磨文字</strong><span>为想说的话，找到表达</span><?php $icon('chevron', 14); ?></button>
                        <button type="button" data-prompt="请阅读我附上的资料，概括重点，再提出值得深入思考的问题。"><strong>读一份资料</strong><span>从细节里，看见新的联系</span><?php $icon('chevron', 14); ?></button>
                    </div>
                </div>
            </div>
            <div class="fnchat-composer-area">
                <div class="fnchat-approval-area" hidden aria-label="工具调用确认"></div>
                <div class="fnchat-connect-panel" id="fnchat-connections" hidden>
                    <div class="fnchat-connect-heading"><strong>连接工具</strong><button class="fnchat-close-connections fnchat-icon-button" type="button" aria-label="关闭工具列表"><?php $icon('close', 16); ?></button></div>
                    <p>选取本次对话可使用的服务；调用前会请你确认。</p>
                    <div class="fnchat-connect-list"></div>
                </div>
                <form class="fnchat-compose" enctype="multipart/form-data">
                    <div class="fnchat-attachments" aria-label="待发送附件" hidden></div>
                    <label class="screen-reader-text" for="fnchat-message">输入消息</label>
                    <textarea id="fnchat-message" name="message" rows="1" maxlength="5000" placeholder="写下想法，或放入一份资料…" aria-describedby="fnchat-compose-status fnchat-privacy" <?php disabled(!$api_ready); ?>></textarea>
                    <div class="fnchat-compose-bottom">
                        <div class="fnchat-compose-actions">
                            <button class="fnchat-attach fnchat-icon-button" type="button" aria-label="添加附件" title="添加图片、文档、音频或视频" <?php disabled(!$api_ready || empty($settings['uploads_enabled'])); ?>><?php $icon('plus', 21); ?></button>
                            <button class="fnchat-connections-toggle" type="button" aria-expanded="false" aria-controls="fnchat-connections"><?php $icon('tools', 16); ?><span>工具</span><small class="fnchat-tool-count" hidden>0</small></button>
                        </div>
                        <span class="fnchat-key-hint">Enter 发送</span>
                        <button class="fnchat-send" type="submit" aria-label="发送消息" <?php disabled(!$api_ready); ?>><?php $icon('arrow', 20); ?></button>
                    </div>
                    <input class="fnchat-file-input" type="file" multiple hidden accept=".jpg,.jpeg,.png,.webp,.gif,.pdf,.txt,.md,.csv,.tsv,.json,.doc,.docx,.ppt,.pptx,.xls,.xlsx,.mp3,.mpga,.mpeg,.m4a,.wav,.mp4,.webm,.mov">
                    <div class="fnchat-drop-label" hidden>放在这里，加入对话</div>
                </form>
                <p id="fnchat-compose-status" class="fnchat-form-status" role="status" aria-live="polite"></p>
                <button class="fnchat-reauth" type="button" hidden>重新验证口令</button>
                <p id="fnchat-privacy" class="fnchat-bottom-line">内容与附件交由 OpenAI 处理 · 本站上下文临时保留一小时</p>
            </div>
        </div>
        <div class="fnchat-preview" role="dialog" aria-modal="true" aria-label="附件预览" hidden>
            <div class="fnchat-preview-inner"><header><strong class="fnchat-preview-title"></strong><button class="fnchat-preview-close fnchat-icon-button" type="button" aria-label="关闭预览"><?php $icon('close'); ?></button></header><div class="fnchat-preview-content"></div></div>
        </div>
    <?php endif; ?>
</section>
