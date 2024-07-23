# 使用 Node.js 的 Docker 镜像
FROM node:latest AS vue-build

WORKDIR /app

COPY package*.json ./
RUN npm cache clean --force
RUN npm install --force

COPY . .
RUN npm run build

FROM nginx:1.21-alpine

COPY --from=vue-build /app/dist/ /usr/share/nginx/html
# 复制自定义的 Nginx 配置到 Nginx 配置目录
COPY default.conf /etc/nginx/conf.d/default.conf

EXPOSE 8080

CMD ["nginx", "-g", "daemon off;"]
