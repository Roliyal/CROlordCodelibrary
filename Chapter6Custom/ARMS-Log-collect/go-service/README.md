# go-service

- HTTP :8081
- gRPC :9091
- Logs to `logs/app.log` (rotating)

Env:
- GO_HTTP_PORT (default 8081)
- GO_GRPC_PORT (default 9091)
- JAVA_HTTP_BASE_URL (default http://localhost:8080)
- JAVA_GRPC_ADDR (default localhost:9090)
- GO_LOG_PATH (default logs/app.log)
- GO_LOG_MAX_SIZE_MB (default 50)
- GO_LOG_MAX_BACKUPS (default 5)
- GO_LOG_MAX_AGE_DAYS (default 7)
- APP_ENV, APP_VERSION
