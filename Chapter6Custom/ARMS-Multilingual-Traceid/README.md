# ARMS Multilingual TraceId Demo（RUM + OpenTelemetry + OTLP）

该仓库用于演示 **前端 ARMS RUM（W3C TraceContext 注入）** 与 **多语言后端（Go / Python / Java / C++）OpenTelemetry Trace** 的端到端链路打通，并通过 **OTLP/HTTP** 上报到阿里云 ARMS 服务。

---

## 目录结构

```text
.
├── frontend      # Vite + Vue3 + ARMS RUM（自动注入 traceparent）
├── go-gateway    # Go 入口服务（otelhttp server/client + OTLP exporter）
├── python-svc    # Flask + OTel instrumentation + OTLP exporter
├── java-svc      # JDK HttpServer + OTel SDK 手动初始化 + 透传
└── cpp-svc       # httplib + opentelemetry-cpp + OTLP/HTTP exporter
```

---

## 调用链与端口

调用链：

`Browser(frontend) → /api/hello(go-gateway) → /py/work(python) → /java/work(java) → /cpp/work(cpp)`

默认端口（与 `.env` 一致）：


| 模块       |              端口 | 入口                             |
| ---------- | ----------------: | -------------------------------- |
| go-gateway |              8080 | `GET /api/hello`、`GET /healthz` |
| python-svc |              8081 | `GET /py/work`、`GET /healthz`   |
| java-svc   |              8082 | `GET /java/work`、`GET /healthz` |
| cpp-svc    |              8083 | `GET /cpp/work`、`GET /healthz`  |
| frontend   | 5173（Vite 默认） | 浏览器访问                       |

---

## 环境配置

在本示例根目录执行 `cp .env.example .env`，填写自己的 ARMS OTLP/HTTP traces 地址和前端 RUM 参数。默认地址指向本机 Collector，不会向已有云实例发送数据。`.env` 不纳入版本控制。

本地从各服务目录启动时，后端读取 `../.env`，Vite 读取示例根目录配置。不要在服务子目录另建 `.env`，避免配置覆盖。Java 当前使用 JDK HttpServer 和手动初始化的 OpenTelemetry SDK；使用 `OTEL_EXPORTER_OTLP_ENDPOINT`，不读取旧的 `JAVA_OTEL_*` 配置。

容器构建不包含 `.env`。运行每个后端容器时，将配置文件只读挂载到 `/app/.env`；容器间的 `PY_URL`、`JAVA_URL`、`CPP_URL` 必须使用可达的服务地址，不能使用指向本容器的 `127.0.0.1`。前端镜像使用 Dockerfile 中的 `VITE_*` build args。

依赖：Go >= 1.22、JDK 17 + Maven 3.9、Python 3.11、Node.js 20；C++ 使用 CMake >= 3.20 和 vcpkg（Dockerfile 固定为 2024.12.16），需安装清单中的依赖并传入 vcpkg toolchain。

# 启动与运行

## 启动顺序（建议）

1. cpp-svc
2. java-svc
3. python-svc
4. go-gateway
5. frontend

---

## 1）启动 cpp-svc（C++）

### 构建

```bash
cd cpp-svc
cmake -S . -B build -DCMAKE_TOOLCHAIN_FILE="$VCPKG_ROOT/scripts/buildsystems/vcpkg.cmake"
cmake --build build -j
```

### 运行

```bash
./build/cpp-svc
```

验证：

```bash
curl -s http://127.0.0.1:8083/healthz
curl -s http://127.0.0.1:8083/cpp/work | head
```

---

## 2）启动 java-svc（Java）

### 构建

```bash
cd java-svc
mvn -q -DskipTests package
```

### 运行

在 `java-svc` 目录构建后执行（入口为 `com.example.JavaSvc`）：

```bash
java -jar target/app.jar
```

验证：

```bash
curl -s http://127.0.0.1:8082/healthz
curl -s http://127.0.0.1:8082/java/work | head
```

---

## 3）启动 python-svc（Python）

### 安装依赖

```bash
cd python-svc
python3 -m venv .venv
source .venv/bin/activate
pip install -r requirements.txt
```

### 运行

```bash
python app.py
```

验证：

```bash
curl -s http://127.0.0.1:8081/healthz
curl -s http://127.0.0.1:8081/py/work | head
```

---

## 4）启动 go-gateway（Go）

### 运行

```bash
cd go-gateway
go run .
```

验证：

```bash
curl -s http://127.0.0.1:8080/healthz
curl -s http://127.0.0.1:8080/api/hello | head
```

---

## 5）启动 frontend（Vite + Vue3）

### 安装依赖

```bash
cd frontend
npm i
```

### 本地运行

```bash
npm run dev
```

访问：

* [http://127.0.0.1:5173](http://127.0.0.1:5173)

说明：

* `vite.config.js` 已将 `/api` 代理到 `http://127.0.0.1:8080`

### 打包

```bash
npm run build
```

产物：

* `frontend/dist`

### 预览

```bash
npm run preview
```

---

# 如何使用（演示操作）

## 1）端到端链路演示

1. 浏览器打开前端页面
2. 点击：`Call /api/hello (Go → Py → Java → C++)`
3. 页面输出 JSON，包含：

   * `trace_id`、`span_id`（Go 入口 span）
   * `python_response`（包含 python/java/cpp 的 trace 信息）

## 2）验证 traceparent 注入（前端）

浏览器 DevTools → Network → 选中 `/api/hello` 请求 → Request Headers：

* 存在 `traceparent`

## 3）验证后端透传（trace_id 一致）

* `go-gateway` 返回的 `trace_id` 与 python/java/cpp 返回的 `trace_id` 应一致（同一条 Trace）
* Java 返回中的 `cpp.traceparent_in` 应包含同一 Trace ID；`cpp.trace_id` 应与 Java 的 `trace_id` 一致。

---

# 打包命令汇总


| 模块       | 安装                              | 运行               | 打包                         |
| ---------- | --------------------------------- | ------------------ | ---------------------------- |
| frontend   | `npm i`                           | `npm run dev`      | `npm run build`              |
| go-gateway | -                                 | `go run .`         | `go build -o go-gateway`     |
| python-svc | `pip install -r requirements.txt` | `python app.py`    | （按部署方式：容器/zip）     |
| java-svc   | -                                 | `java -jar target/app.jar` | `mvn -DskipTests package` |
| cpp-svc    | -                                 | `./build/cpp-svc`  | `cmake --build build`        |

---

# 常见问题

## 1）/api/hello 调用失败

* 检查下游服务是否启动：8081/8082/8083
* 检查 `.env` 中 `PY_URL / JAVA_URL / CPP_URL` 是否为本机可达地址

## 2）ARMS 上无 Trace 数据

* 检查 `OTEL_EXPORTER_OTLP_ENDPOINT` 是否可达
* Java 直接使用 `OTEL_EXPORTER_OTLP_ENDPOINT` 创建 HTTP exporter；当前实现没有配置额外鉴权 header 的入口。
