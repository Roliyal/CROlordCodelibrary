# 设置基础镜像为 golang:latest
FROM golang:latest

# 创建一个工作目录
WORKDIR /app

# 将所有文件复制到工作目录中
COPY . .

# 构建应用
RUN go build -o main .

# 启动应用
CMD ["./main"]
