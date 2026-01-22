# ARMS-Log-collect (Production-hardened demo)

Two services with **mutual HTTP + gRPC calls** and **unified JSON structured logs**.

- Java: Spring Boot + Log4j2 (**logs to stdout**)
- Go: net/http + go-kit/log (**logs to rotating file logs/app.log**)

## Ports (defaults)
- java-service HTTP: `8080`
- java-service gRPC: `9090`
- go-service   HTTP: `8081`
- go-service   gRPC: `9091`

## Quick Run (local)
### 1) Start Go
```bash
cd go-service
go mod tidy
go run ./...
```

### 2) Start Java
```bash
cd java-service
./mvnw -q -DskipTests package
./mvnw -q spring-boot:run
```

### 3) Fire traffic (creates lots of structured logs)
```bash
curl -H "X-Trace-Id: demo-trace-001" -X POST "http://localhost:8080/api/order/create?n=50"
curl -H "X-Trace-Id: demo-trace-002" "http://localhost:8080/api/user/get?n=50"
curl -H "X-Trace-Id: demo-trace-003" -X POST "http://localhost:8080/api/inventory/reserve?n=50"
```

- Java logs: stdout (JSON per line)
- Go logs: `go-service/logs/app.log` (JSON per line)

## Validation plan
See `VALIDATION.md`.
