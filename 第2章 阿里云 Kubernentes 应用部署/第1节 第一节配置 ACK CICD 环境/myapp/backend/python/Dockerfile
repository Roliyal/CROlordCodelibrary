# 设置基础镜像为 python:latest
FROM python:latest

# 创建一个工作目录
WORKDIR /app

# 将所有文件复制到工作目录中
COPY . .

# 安装依赖
RUN pip install -r requirements.txt

# 启动应用
CMD ["python", "app.py"]
