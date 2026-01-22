# Validation Plan (Production checks)

This plan verifies:
1) Unified JSON schema
2) TraceID propagation (HTTP header + gRPC metadata)
3) Single-line log events (no multi-line)
4) Logging not limited to debug (INFO/WARN/ERROR)
5) Category/source/service are present
6) Go logs to file; Java logs to stdout
7) Mutual calls occur over BOTH HTTP and gRPC

---

## A. Build & Run

### A1. Run locally
1. Start Go:
```bash
cd go-service
go mod tidy
go run ./...
```

2. Start Java:
```bash
cd java-service
./mvnw -q -DskipTests package
./mvnw -q spring-boot:run
```

### A2. Smoke traffic
```bash
curl -H "X-Trace-Id: t-http-001" -X POST "http://localhost:8080/api/order/create?n=20"
curl -H "X-Trace-Id: t-http-002" "http://localhost:8080/api/user/get?n=20"
curl -H "X-Trace-Id: t-http-003" -X POST "http://localhost:8080/api/inventory/reserve?n=20"
```

---

## B. Verify output targets

### B1. Java → stdout
You should see JSON lines in Java console.

### B2. Go → file
```bash
tail -n 5 go-service/logs/app.log
```

---

## C. Verify TraceID propagation across services (HTTP + gRPC)

### C1. Pick one request
```bash
TRACE=t-verify-001
curl -H "X-Trace-Id: $TRACE" -X POST "http://localhost:8080/api/order/create?n=5"
```

### C2. Verify in Go file logs
```bash
grep "$TRACE" go-service/logs/app.log | head
```

### C3. Verify in Java stdout
Search console output for `t-verify-001`.

Expected:
- You will find **both** categories:
  - `remote.http.*` (HTTP outbound/inbound)
  - `remote.grpc.*` (gRPC outbound/inbound)
- And both services show the **same traceId**.

---

## D. Verify unified schema (jq examples)

### D1. Go log schema check
```bash
tail -n 200 go-service/logs/app.log | jq -e '
  .timestamp and .level and .service and .source and .category and .traceId and .protocol and .message
' >/dev/null && echo OK
```

### D2. Single-line check (no embedded newlines as record separators)
Each log record is **one JSON object per line**.
```bash
python3 - <<'PY'
import pathlib
p = pathlib.Path("go-service/logs/app.log")
bad = []
for i,line in enumerate(p.read_text(encoding="utf-8").splitlines(), 1):
    if not line.strip(): 
        continue
    if not (line.lstrip().startswith("{") and line.rstrip().endswith("}")):
        bad.append((i, line[:120]))
print("bad_lines:", len(bad))
print(bad[:5])
PY
```

---

## E. Verify WARN/ERROR exist (not only debug)

### E1. Trigger an error
Stop Go gRPC port by stopping Go service (Ctrl+C), then call Java endpoint again:
```bash
curl -H "X-Trace-Id: t-err-001" -X POST "http://localhost:8080/api/order/create?n=3"
```

Expected:
- Java emits `ERROR` log for gRPC call failure (remote service down)
- Still single-line JSON

---

## F. Load & Volume (optional)
Use any load tool, e.g. `hey`:
```bash
hey -n 200 -c 10 -m POST "http://localhost:8080/api/order/create?n=10"
```

Check log file growth + rotation in Go:
- `go-service/logs/app.log`
- rotated files: `app.log.1`, etc (depends on settings)
