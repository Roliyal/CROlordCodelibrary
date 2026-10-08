## Production image that is still "kubectl exec"-friendly (has /bin/sh)
## - Fixes CrashLoopBackOff (ExitCode=0/Completed) by copying proto sources into the runtime image.
## - Supports multi-arch via TARGETOS/TARGETARCH.

FROM golang:1.24 AS build

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download

COPY . .

ARG TARGETOS=linux
ARG TARGETARCH=amd64
RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} \
    go build -trimpath -ldflags="-s -w" -o /out/go-service .

# Runtime (shell included for exec/debug; lightweight)
FROM alpine:3.20

RUN apk add --no-cache ca-certificates \
    && addgroup -S -g 10001 app \
    && adduser -S -u 10001 app -G app \
    && mkdir -p /app/logs /var/log/arms \
    && chown -R app:app /app /var/log/arms

WORKDIR /app
COPY --from=build /out/go-service /app/go-service

# IMPORTANT: runtime loads proto from filesystem (grpc/bridge_desc.go)
COPY proto /app/proto

EXPOSE 8081 9091

USER app
ENTRYPOINT ["/app/go-service"]
