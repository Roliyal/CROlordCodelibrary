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

# 暴露容器的 8080 端口
EXPOSE 8080

# 设置容器启动时运行的命令
CMD ["./mcp-service"]