# 使用官方的 Golang 镜像作为基础镜像
FROM golang:1.22.4 AS builder

# 设置工作目录
WORKDIR /app

# 将当前目录内容复制到容器中的 /app 目录
COPY . .

# 安装依赖（如果有的话，Go Modules 会自动处理）
RUN go mod tidy

# 编译 Go 程序
RUN go build -o crolord-mcp-service .

# 创建一个新的轻量级镜像（使用Alpine避免冗余的 Go 环境）
FROM alpine:latest

# 安装需要的依赖库，Go 构建后可能依赖的共享库
RUN apk --no-cache add ca-certificates

# 设置工作目录
WORKDIR /root/

# 从 builder 镜像中复制编译好的二进制文件
COPY --from=builder /app/crolord-mcp-service .

# 暴露容器的 8080 端口
EXPOSE 8080

# 设置容器启动时运行的命令
CMD ["./crolord-mcp-service"]
