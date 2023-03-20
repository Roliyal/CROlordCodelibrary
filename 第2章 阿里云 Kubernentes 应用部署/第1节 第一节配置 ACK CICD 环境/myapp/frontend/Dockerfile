# 设置基础镜像为 node:latest
FROM node:latest

# 创建一个工作目录
WORKDIR /app

# 将所有文件复制到工作目录中
COPY . .

# 安装依赖
RUN npm install

# 构建应用
RUN npm run build

# 启动应用
CMD ["npm", "run", "serve"]
