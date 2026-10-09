# 微服务治理示例

| 目录 | 场景 | 与其他章节的关系 |
| --- | --- | --- |
| Unit1CodeLibrary/Govern/Microservice | 猜数字前端、登录、游戏、排行榜及 Nacos 接入 | 与第二章 Microservice 存在相同源码文件，公共修复需同步检查 |
| Unit2CodeLibrary/FEBEseparation | 前后端分离、base/gray 配置 | 保留灰度教学差异 |
| Unit2CodeLibrary/Microservice | K8s、Nacos 生命周期及灰度响应 | deployment.yaml 为当前清单；back_deployment.yaml、demo-deployment_back_.yaml 为历史对照，不要批量 apply |

第二章作为基础部署示例，第四章表达治理差异，第五章增加观测场景。各目录当前仍为独立副本，尚未抽成共享模块；合并前需核对 Nacos、数据库、接口与灰度行为。

## 使用顺序

1. 准备数据库、Nacos、镜像仓库和命名空间，检查各服务的 go.mod、Dockerfile、环境变量及部署清单。
2. 分别构建 login-service、game-service、scoreboard-service 和 front-guess，替换部署镜像及环境地址。
3. 逐个应用明确选定的 deployment.yaml；不要递归应用包含历史备份的目录。
4. 验证注册、登录、猜数字和排行榜，再验证灰度及下线。

## game-service 注册与下线验证

Unit2 的服务实际监听 8084；K8s Service、探针和生命周期 SERVICE_PORT 均应为 8084。

- 启动后查询 Nacos，确认 game-service 实例地址为 Pod 可达地址、端口为 8084，并请求 `/health`。
- 通过前端完成登录和猜数字，确认请求能到达注册的实例。
- 滚动更新时观察旧 Pod 对应的 **8084** 实例摘流、注销，新 Pod 就绪后加入；旧的错误端口实例需先核对归属再清理。

以上为真实环境验收步骤，不表示已在你的集群执行。前端子目录 README 的 npm 命令仅负责前端构建，完整示例还依赖上述后端与云资源。
