# 第一阶段：构建 Vue 静态资源
FROM node:14-alpine AS vue-build
WORKDIR /app
COPY package*.json ./
RUN npm cache clean --force
RUN npm install  --no-cache --verbose
COPY . ./
RUN npm run build

# 第二阶段：将构建好的静态资源复制到 Nginx 镜像中
FROM docker.io/nginx:1.21-alpine
COPY --from=vue-build /app/dist /usr/share/nginx/html
COPY default.conf /etc/nginx/conf.d/default.conf
RUN cat /etc/nginx/conf.d/default.conf  # 添加这行来显示文件内容
RUN mkdir -p /var/cache/nginx/client_temp /var/run/nginx && \
    chown -R nginx:nginx /var/cache/nginx /var/run/nginx && \
    chmod -R 755 /var/cache/nginx /var/run/nginx
EXPOSE 8080
CMD ["nginx", "-g", "daemon off;"]
