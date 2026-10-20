<?php if (!defined('ABSPATH')) exit; ?>
<section class="fnchat-admin-section">
    <h2>附件与媒体</h2>
    <p>附件使用 PHP 临时目录转发，不进入公开媒体库。选取文件后即上传至 OpenAI；音频会先转写，转写也产生 API 用量。</p>
    <table class="form-table" role="presentation">
        <tr><th scope="row">附件上传</th><td><label><input type="checkbox" name="<?php echo esc_attr(self::OPTION); ?>[uploads_enabled]" value="1" <?php checked(!empty($settings['uploads_enabled'])); ?>> 允许图片、文档与视频关键帧</label></td></tr>
        <tr><th scope="row"><label for="fnchat-upload-size">单个附件上限</label></th><td><input id="fnchat-upload-size" type="number" min="1" max="20" name="<?php echo esc_attr(self::OPTION); ?>[max_upload_mb]" value="<?php echo esc_attr($settings['max_upload_mb']); ?>"> MB<p class="description">每条消息最多 4 份附件；本段对话累计不超过 12 份、40 MB。服务器上传上限也会生效。视频在浏览器生成四帧拼图，原视频不上载、不分析声音。</p></td></tr>
        <tr><th scope="row">音频转写</th><td><label><input type="checkbox" name="<?php echo esc_attr(self::OPTION); ?>[audio_enabled]" value="1" <?php checked(!empty($settings['audio_enabled'])); ?>> 允许上传音频并转成文字</label></td></tr>
        <tr><th scope="row"><label for="fnchat-transcription">转写模型</label></th><td><input id="fnchat-transcription" class="regular-text" name="<?php echo esc_attr(self::OPTION); ?>[transcription_model]" value="<?php echo esc_attr($settings['transcription_model']); ?>" spellcheck="false"><p class="description">默认 gpt-transcribe。需为 API Key 开通 Audio Transcriptions；对话模型仍使用上方的模型 ID。</p></td></tr>
    </table>
</section>
<section class="fnchat-admin-section">
    <h2>MCP 连接</h2>
    <p>填写远程 MCP 的公网 HTTPS 地址（Streamable HTTP 或 HTTP/SSE）。由 OpenAI 连接服务；本地 stdio 进程不能直接填写。启用后，访客仍需在对话中选取工具，并逐项确认调用。</p>
    <p>令牌仅在服务器保存与转发。留空保留原令牌；更换地址或标识后需重新填写。填写原始访问令牌，不包含 “Bearer ” 前缀。OAuth 登录及自动刷新不在此版本内。</p>
    <?php
    $rows = array_values((array) $settings['mcp_servers']);
    while (count($rows) < 4) $rows[] = array('enabled' => 0, 'label' => '', 'title' => '', 'url' => '', 'token' => '', 'allowed_tools' => array());
    foreach (array_slice($rows, 0, 4) as $index => $server) :
        $prefix = self::OPTION . '[mcp_servers][' . $index . ']';
    ?>
    <details class="fnchat-server-setting" <?php echo !empty($server['url']) || $index === 0 ? 'open' : ''; ?>>
        <summary><?php echo esc_html($server['title'] ?: ('连接 ' . ($index + 1))); ?></summary>
        <div class="fnchat-server-fields">
            <label class="fnchat-server-enabled"><input type="checkbox" name="<?php echo esc_attr($prefix); ?>[enabled]" value="1" <?php checked(!empty($server['enabled'])); ?>> 启用这个连接</label>
            <label>英文标识<input name="<?php echo esc_attr($prefix); ?>[label]" value="<?php echo esc_attr($server['label']); ?>" maxlength="40" placeholder="my_tools" spellcheck="false"></label>
            <label>显示名称<input name="<?php echo esc_attr($prefix); ?>[title]" value="<?php echo esc_attr($server['title']); ?>" maxlength="80" placeholder="我的资料"></label>
            <label class="fnchat-server-wide">服务地址<input type="url" name="<?php echo esc_attr($prefix); ?>[url]" value="<?php echo esc_attr($server['url']); ?>" placeholder="https://your-server.example/mcp" spellcheck="false"></label>
            <label class="fnchat-server-wide">访问令牌<?php echo !empty($server['token']) ? '（已保存）' : '（如需要）'; ?><input type="password" name="<?php echo esc_attr($prefix); ?>[token]" value="" autocomplete="new-password" spellcheck="false"></label>
            <label class="fnchat-server-wide">允许的工具名称<input name="<?php echo esc_attr($prefix); ?>[allowed_tools]" value="<?php echo esc_attr(implode(', ', (array) $server['allowed_tools'])); ?>" placeholder="留空可发现所有工具；用英文逗号分隔限制范围" spellcheck="false"></label>
            <label class="fnchat-server-wide"><input type="checkbox" name="<?php echo esc_attr($prefix); ?>[clear_token]" value="1"> 清除保存的令牌</label>
        </div>
    </details>
    <?php endforeach; ?>
    <p class="description">上下文在本站临时缓存最多一小时；OpenAI Responses 使用 store=false。上传的 OpenAI 文件设置一小时过期。新对话会请求删除旧文件；网络失败时以文件过期策略兜底。选择的 MCP 服务会收到确认调用中的参数。</p>
</section>
