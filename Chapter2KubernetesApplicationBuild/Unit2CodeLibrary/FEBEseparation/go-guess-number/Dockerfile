# 使用官方的 Go 1.20 镜像进行构建
FROM golang:1.20  AS builder

WORKDIR /app

# 复制 go.mod 和 go.sum 文件，并下载依赖项
COPY go.mod go.sum ./
RUN go mod download

# 复制所有源代码
COPY . .

# 构建 Go 程序
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o main .

# 使用官方的 Alpine 镜像作为运行环境
FROM alpine

# 添加非 root 用户
RUN addgroup -S appgroup && adduser -S appuser -G appgroup

WORKDIR /app

# 从构建阶段复制构建好的 Go 二进制文件
COPY --from=builder /app/main /app/main

RUN mkdir -p /app/log && chown -R appuser:appgroup /app/log
RUN chown -R appuser:appgroup /app

USER appuser

EXPOSE 8081

HEALTHCHECK --interval=60s --timeout=5s --start-period=5s --retries=3 CMD ["/app/main", "check"]

ENTRYPOINT ["/app/main"]
