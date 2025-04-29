# 使用官方 Golang 镜像作为基础镜像
FROM golang:1.22.4 AS builder

# 设置工作目录
WORKDIR /app


# 设置 Go 代理为阿里云镜像
ENV GOPROXY=https://mirrors.aliyun.com/goproxy/

# 下载 instgo 工具并适配架构
RUN uname -m && \
    if [ "$(uname -m)" = "x86_64" ]; then \
        wget "http://arms-apm-ap-southeast-1.oss-ap-southeast-1.aliyuncs.com/instgo/instgo-linux-amd64" -O instgo; \
    elif [ "$(uname -m)" = "aarch64" ]; then \
        wget "http://arms-apm-ap-southeast-1.oss-ap-southeast-1.aliyuncs.com/instgo/instgo-linux-arm64" -O instgo; \
    else \
        echo "Unsupported architecture"; exit 1; \
    fi && \
    chmod +x instgo

# 设置 LicenseKey 和 RegionId
RUN ls -a
RUN  /app/instgo version
RUN /app/instgo set --mse  --licenseKey=djqtzchc9t@b929339d9ac7fb0 --regionId=ap-southeast-1 --agentVersion=1.6.1

#RUN ./instgo set -mse  --licenseKey=djqtzchc9t@b929339d9ac7fb0 --regionId=ap-southeast-1
#RUN ./instgo set --mse djqtzchc9t@b929339d9ac7fb0 ap-southeast-1

# 复制 go.mod, go.sum 文件到工作目录
COPY go.mod go.sum .env ./

RUN go mod download
# 复制源代码到工作目录
COPY . .

# 编译 AMD64 架构的二进制文件
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 ./instgo go build -o main-amd64 .

# 编译 ARM64 架构的二进制文件
RUN CGO_ENABLED=0 GOOS=linux GOARCH=arm64 ./instgo go build -o main-arm64 .

# 使用 Alpine 镜像作为基础镜像
FROM alpine

# 安装 curl（和 ca-certificates，以支持 HTTPS）
RUN apk update && apk add --no-cache curl ca-certificates

# 添加非 root 用户
RUN addgroup -S appgroup && adduser -S appuser -G appgroup

# 设置工作目录为 /app
WORKDIR /app

# 复制编译好的 Go 二进制文件
COPY --from=builder /app/main-amd64 /app/main-amd64
COPY --from=builder /app/main-arm64 /app/main-arm64
COPY --from=builder /app/.env /app/.env
COPY --from=builder /app/start.sh /app/start.sh

# 临时切换到 root 用户来修改权限
USER root

# 赋予执行权限
RUN chmod +x /app/start.sh

# 创建日志目录并修改权限
RUN mkdir -p /app/log && chown -R appuser:appgroup /app/log

# 更改工作目录的拥有者为非 root 用户
RUN chown -R appuser:appgroup /app

# 切换回非 root 用户
USER appuser

# 暴露端口，确保与应用程序使用的端口一致
EXPOSE 8084

# 添加健康检查
HEALTHCHECK --interval=60s --timeout=5s --start-period=5s --retries=3 CMD ["/app/start.sh", "check"]

# 默认运行启动脚本
ENTRYPOINT ["/bin/sh","-c","exec /app/start.sh"]
