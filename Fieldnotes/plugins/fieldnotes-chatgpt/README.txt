Fieldnotes ChatGPT Connector / Echoes 1.5.0
WordPress 6.2+ · PHP 7.4+ · 适配 Fieldnotes Journal 1.6.8

安装与升级
1. 后台 → 插件 → 安装插件 → 上传 fieldnotes-chatgpt-1.5.0.zip。
2. 选择替换现有插件。保持目录名 fieldnotes-chatgpt，不要并列安装两个版本。
3. 原 API Key、口令哈希、对话模型和页面关联会保留。清理站点/CDN 缓存后刷新 Echoes。
4. 后台 → 设置 → ChatGPT 接入：保存附件、音频与 MCP 设置。
5. 页面仍使用 [fieldnotes_chatgpt] 短代码；已有页面不需要重建。重复短代码会只显示一套对话框。

界面
白色纸面、黑灰配色、系统字体、克制的 Echoes 字标。Apple 设备使用系统字体与 PingFang SC；其他系统使用其本机回退字体，不分发 Apple 字体文件。
底部一处输入区，支持多行、中文输入法 Enter 保护、Shift+Enter 换行、拖放、粘贴图片。
附件可在发送前预览；音频可试听和查看转写；视频可在本地播放。
回复支持安全的基础 Markdown：标题、粗体、代码、引用、列表、链接与表格；模型文字不作为 HTML 执行。
支持复制回复与导出本次对话为 Markdown。手机端会调整工作区和输入区。
本版没有 SSE 流式输出；等待过程显示真实请求状态，不使用模拟打字效果。

附件能力
图片：JPG/JPEG、PNG、WebP、GIF。作为 input_image 发送，取决于所选模型的视觉能力。GIF 以 API 支持的静态图片处理为准。
文档：PDF、TXT、Markdown、CSV、TSV、JSON、Word、PowerPoint、Excel。作为 input_file 发送。
PDF 可包含页面图像；普通 Office 文档主要提取文字，内嵌图表不保证完整读取。表格输入也不等于完整的计算环境。
音频：MP3、MPGA、MPEG、M4A、WAV、音频 WebM。先调用 /v1/audio/transcriptions 转成文字，再将转写交给对话模型。不是声音/音乐分析。
视频：浏览器可解码的 MP4、MOV、视频 WebM，原文件不超过 100 MB。在浏览器取四张等间隔关键帧，合成一张带时间的拼图上传；原视频不上载。只分析这些画面，不解析声音，也不保证捕捉所有动作和细节。浏览器不支持的编码会明确报错。
每条消息最多 4 份附件，单份默认 10 MB，可设 1–20 MB；还受 PHP、Web 服务器上传限制约束。
本段对话累计不超过 12 份附件、40 MB；对话最多 24 轮。达到限制后开启新对话。
选择文件后即上传到 OpenAI，音频即开始转写；文档/图片和转写均可能产生 API 用量。

附件与上下文的保存
PHP 上传临时文件直接转发，不移动到 WordPress 公开媒体库，不产生公开下载 URL。
OpenAI 文件使用 purpose=user_data，并明确设置创建后一小时过期。音频转写不创建长期 Files 条目。
本站以 WordPress Transients 临时缓存会话上下文、附件引用与转写，最长一小时，并绑定当前口令访问凭据。API 响应使用 store=false，通过服务器保存的完整输出延续多轮与 MCP。
这不是“完全不落盘”：WordPress 未配置对象缓存时，Transients 可以落在数据库中；到期后停止使用，物理删除遵循 WordPress 清理机制和站点备份策略。OpenAI 自身的数据政策独立适用。
刷新后界面不恢复历史；刷新遗留的上下文/文件按到期策略失效。
“新对话”会清理旧上下文并请求删除旧文件，同时移除未发送附件，保留文字草稿。“锁定”也会先尝试清理。
删除请求只做短时尝试；网络失败会说明远端文件未能立即删除，文件仍按一小时过期策略兜底。不可把此机制解释成即时擦除第三方所有记录。

MCP 接入
支持最多四个管理员配置的远程 MCP 服务，Streamable HTTP 或 HTTP/SSE。
服务需要能被 OpenAI 访问的公网 HTTPS URL；WordPress 本机 localhost、容器内地址或 stdio 命令不能直接填入。已有本地 MCP 需先通过你管理的远程 MCP 网关暴露。
每个连接填写：启用、唯一英文标识、显示名称、HTTPS 地址、访问令牌（如需要）、允许的工具名称（逗号分隔，留空发现全部）。
令牌只在后端保存与转发，前端不会收到。留空保留原令牌；更换 URL/标识后不沿用旧令牌。只支持现成的原始 Bearer/OAuth access token，不提供 OAuth 登录或自动刷新。
没有自动启用的第三方服务。访客在输入框“工具”中选取服务；本次上下文建立后选择固定，开启新对话可更换。
每次 MCP 调用 require_approval=always。界面显示目标服务、工具名和实际参数，逐项允许或拒绝后继续；只点选连接不会自动批准工具执行。
允许的操作范围也受管理员填写的工具名单和 MCP 令牌权限约束。工具列表发现会发生在调用确认之前。
工具确认使用服务器保存的上下文、访问绑定、随机 revision、逐项布尔决策和互斥锁，不能通过前端伪造工具 URL/令牌或重放已经消费的确认。
工具确认提交后遇到网络超时，会将结果视为未知，停止重发。请先到相应服务核对结果，再开启新对话；不会自动重试可能已执行的操作。
工具执行需要 MCP 端本身可用且令牌有效。MCP 的外部服务可能单独计费。

可选测试连接（不默认启用）
英文标识：openai_docs
显示名称：OpenAI 文档
URL：https://developers.openai.com/mcp
访问令牌：空
允许工具：先留空，由接口发现当前工具。
保存后在 Echoes 的“工具”中选中，再询问官方 API 文档。必须先保证 WordPress → OpenAI 的连接正常。

服务器与 API 设置
对话模型保留原配置，默认 gpt-6-luna；音频转写默认 gpt-transcribe，可填当前项目已开通的转写模型。
API Key 优先级：非空 FIELDNOTES_CHATGPT_API_KEY 常量 → 非空 OPENAI_API_KEY 环境变量 → 后台保存值。Key 永远只放服务器。
需要 Responses 的调用权限；上传文件还需要 Files 的创建/删除权限；音频另需 Transcriptions 权限。后台模型检测仅验证 Models Read，不证明生成、文件、音频和 MCP 均已开通。
保留当前 HTTP 测试设置。前台 HTTP 测试与上游 API 是两条连接；上游始终使用 HTTPS 并保持证书校验。
对话请求时限默认 60 秒，可设 15–120 秒。检测请求总时限为 20 秒；WordPress 网络库的连接阶段可能更早超时。
本版将“连接建立超时”和“等待响应超时”区分。约 10 秒的 cURL Connection timed out、没有 HTTP 状态，说明尚未获得 API 响应，不能由此判断 Key 错误。
新增附件/MCP 不会修复服务器到 OpenAI 的网络出口。需要在 WordPress 所在服务器/容器内验证 DNS、出站 443、TLS、代理与实际出口，并核对 OpenAI 官方支持地区。
已有可信 HTTP 代理可按 WordPress 官方 WP_PROXY_HOST / WP_PROXY_PORT 配置；Docker 中 127.0.0.1 指当前容器，不应盲填代理地址。本插件不包含代理服务器或第三方 API 中转。
附件报 413 时核对 Nginx client_max_body_size、PHP upload_max_filesize / post_max_size；post_max_size 要包含 multipart 开销。PHP 与反向代理的执行/读取时限需足以覆盖实际请求。

网络初检（在 WordPress 所在服务器/容器执行；不含 Key）
curl -q -4 -sS -o /dev/null --connect-timeout 10 --max-time 20 -w 'HTTP=%{http_code} DNS=%{time_namelookup}s TCP=%{time_connect}s TLS=%{time_appconnect}s TOTAL=%{time_total}s\n' https://api.openai.com/v1/models
无 Key 时，正常收到 401 说明已取得 API HTTP 响应，不代表你的 Key 已验证。HTTP 000 + cURL 28 则继续检查网络。不要通过关闭证书验证修复连接。

验证记录
本次发布在隔离环境验证 PHP 7.4 语法兼容性、JavaScript/CSS 语法、78 项原有权限/配置回归、74 项新后端校验、47 项前端交互、6 项后台检测交互及 78 项响应式 CSS/DOM 校验。
API 和上传由测试桩模拟，视频使用受控媒体/画布夹具；静态设计校样用于检查桌面和手机排版。没有使用你的 API Key，也没有在你的服务器或真实 MCP 服务完成联网联调。
上线后请先验证：纯文字 → 图片/文档 → 音频 → 已配置的 MCP；确认每一步的 Key 权限与出口正常。

官方接口参考
https://developers.openai.com/api/docs/guides/tools-connectors-mcp
https://developers.openai.com/api/docs/guides/file-inputs
https://developers.openai.com/api/docs/guides/images-vision
https://developers.openai.com/api/docs/guides/speech-to-text
https://developers.openai.com/api/docs/guides/reasoning
https://developer.wordpress.org/reference/classes/wp_http_proxy/

GPL-2.0-or-later. 保留现有 LICENSE。
