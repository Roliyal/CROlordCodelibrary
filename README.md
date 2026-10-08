# CROlord Code Library

阿里云容器、微服务治理和可观测性实践示例。各目录是独立场景，不能从根目录统一构建。

| 目录 | 用途 | 入口与状态 |
| --- | --- | --- |
| [Chapter1Toolchain](Chapter1Toolchain/) | 工具链、Jenkins 配置 | 按子目录使用；未进行端到端验证 |
| [Chapter2KubernetesApplicationBuild](Chapter2KubernetesApplicationBuild/) | 前后端分离及猜数字微服务部署 | Unit2CodeLibrary；需准备数据库、Nacos 等依赖 |
| [Chapter4MicroserviceGovern](Chapter4MicroserviceGovern/README.md) | 治理、灰度及上下线实践 | 参见场景差异与验证说明 |
| [Chapter5MicroserviceObservability](Chapter5MicroserviceObservability/) | 微服务观测、Prometheus 及 MCP 示例 | Unit5CodeLibrary；未进行端到端验证 |
| [Chapter6Custom/ARMS-Log-collect](Chapter6Custom/ARMS-Log-collect/README.md) | Java/Go HTTP、gRPC 与结构化日志 | README + VALIDATION.md；需 Go 1.24、JDK 17、Maven |
| [Chapter6Custom/ARMS-Multilingual-Traceid](Chapter6Custom/ARMS-Multilingual-Traceid/README.md) | RUM 与四种语言的链路透传 | 先复制 .env.example，再按语言启动 |
| [Chapter6Custom/ARMS-SourceMap-Demo](Chapter6Custom/ARMS-SourceMap-Demo/) | 前端 SourceMap 示例 | frontend、backend；未进行端到端验证 |

仓库目前没有第三、第七章对应目录。示例中的镜像、域名、命名空间和云资源需按自己的环境配置。

## 维护约定

- 运行日志、Nacos 缓存、Java target、前端依赖和本地配置不提交。历史提交仍可能保留旧产物，本次清理不改写 Git 历史。
- 教学版本差异先记录再抽取，公共修复需对照第二、四、五章检查。
- 提交信息说明具体模块和行为变化；验证说明区分静态检查、构建和真实环境验收。
