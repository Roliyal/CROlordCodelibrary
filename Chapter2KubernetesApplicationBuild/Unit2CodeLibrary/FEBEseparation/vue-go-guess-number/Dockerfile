# 第一阶段：构建 Vue 静态资源
FROM node:14-alpine AS vue-build
WORKDIR /app
COPY package*.json ./
RUN npm install  --no-cache --verbose --registry=https://registry.npm.taobao.org
COPY . ./
RUN npm run build

# 第二阶段：将构建好的静态资源复制到 Nginx 镜像中
FROM nginx:1.21-alpine
COPY --from=vue-build /app/dist /usr/share/nginx/html
COPY default.conf /etc/nginx/conf.d/default.conf
RUN mkdir -p /var/cache/nginx/client_temp /var/run/nginx && \
    chown -R nginx:nginx /var/cache/nginx /var/run/nginx && \
    chmod -R 755 /var/cache/nginx /var/run/nginx
EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]
