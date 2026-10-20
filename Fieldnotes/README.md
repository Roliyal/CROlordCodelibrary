# Fieldnotes

一套 WordPress 主题与三个独立插件。源码、可安装 ZIP 和校验值放在同一处。

| 组件 | 版本 | 下载 | 源码 |
| --- | --- | --- | --- |
| Journal 主题 | 1.9.0 | [安装包](dist/fieldnotes-journal-1.9.0.zip) | [主题源码](theme/fieldnotes-journal) |
| Echoes / ChatGPT | 1.5.0 | [安装包](dist/fieldnotes-chatgpt-1.5.0.zip) | [插件源码](plugins/fieldnotes-chatgpt) |
| Wayfarer / Travel Map | 1.4.0 | [安装包](dist/fieldnotes-travel-map-1.4.0.zip) | [插件源码](plugins/fieldnotes-travel-map) |
| Comment Tools | 1.6.0 | [安装包](dist/fieldnotes-comment-tools-1.6.0.zip) | [插件源码](plugins/fieldnotes-comment-tools) |

点击安装包进入文件页面后使用 Download raw file 下载。不要用整个代码库的 Code → Download ZIP 代替单个 WordPress 安装包。SHA-256 在 [manifest.json](manifest.json) 中。

## 安装与升级

WordPress 6.2+ / PHP 7.4+。在「外观 → 主题 → 上传主题」安装 Journal；三个插件在「插件 → 安装插件 → 上传插件」分别安装。已安装对应版本插件的站点只需替换主题，进入一次 WordPress 后台并清理缓存。现有文章、地图、评论、页面 ID 及插件配置保留。

## Journal 1.9.0

- Echoes 与 Boughs 统一系统字体、标题比例及中性色。Echoes 口令前显示导航与搜索，验证后只留 Logo。
- Markets 复用原 Gleanings 页面：中文名称/简拼/代码检索、明确的候选选择、分时与 K 线、2/5/10/30 秒自动尝试刷新（默认 5 秒）、标签页内临时自选、观察草稿。报价使用行情源的秒级时间戳；请求频率不保证行情无延迟。页面隐藏、非交易时段暂停，连接失败退避。时段判断不包含节假日历。
- Keepsakes 支持京东搜索与具体 SKU、批量多选、临时选购清单及购物车入口。编辑收藏中的「Keepsakes · 京东选购」设置商品编号或网址。清单不等同于京东购物车：在京东逐项确认规格并加购，再到购物车结算；一键批量写入购物车需要京东官方授权，本项目未包含该授权。
- WordPress 组件页面提供本目录内四个独立 ZIP 和源码入口。

完整说明见各组件 README。OpenAI Key、Mapbox Token、访问口令均由站点后台配置，本仓库不含站点配置或密钥。

## 验证范围

检查 PHP 7.4、CSS、JavaScript 语法；用真实 ChatGPT 1.5.0 类验证口令/会话与页头关系；使用捕获的真实行情响应验证解析、候选检索、秒级轮询、缓存/失败处理；验证选购清单、URL 约束、键盘交互与 320–1440px 样式规则。测试使用隔离的 WordPress API、DOM 与静态排版，不代替实际 WordPress 站点的部署验收。未在任何京东账号中执行加购或购买。

## 许可证

遵循各组件的 GPL-2.0-or-later 许可证及随附第三方声明。
