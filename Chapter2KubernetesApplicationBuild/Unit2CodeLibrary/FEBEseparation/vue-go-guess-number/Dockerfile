# 使用 Node.js 的 Docker 镜像
FROM node:14-alpine AS vue-build

WORKDIR /app

# 复制 package.json 和 package-lock.json
COPY package*.json ./

# 安装依赖
RUN npm install --force

# 复制所有文件并构建项目
COPY . .
RUN npm run build

# 使用 Nginx 的 Docker 镜像
FROM nginx:1.21-alpine

# 复制构建结果到 Nginx 默认的静态文件目录
COPY --from=vue-build /app/dist /usr/share/nginx/html

EXPOSE 8080

CMD ["nginx", "-g", "daemon off;"]
