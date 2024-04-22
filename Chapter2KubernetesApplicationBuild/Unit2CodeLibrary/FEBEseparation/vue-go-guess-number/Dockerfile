# 第一阶段：构建 Vue 静态资源
FROM node:14-alpine AS vue-build
WORKDIR /app
COPY package*.json ./
RUN npm cache clean --force
RUN npm install --no-cache --verbose
COPY . ./
RUN npm run build

# 第二阶段：将构建好的静态资源复制到 Nginx 镜像中
FROM docker.io/nginx:1.21-alpine
COPY --from=vue-build /app/dist /usr/share/nginx/html

# Ensure the Nginx config directory is clean to avoid conflicts with ConfigMap volumes
RUN rm /etc/nginx/conf.d/default.conf
RUN mkdir -p /var/cache/nginx/client_temp /var/run/nginx && \
    chown -R nginx:nginx /var/cache/nginx /var/run/nginx && \
    chmod -R 755 /var/cache/nginx /var/run/nginx

# Expose the port Nginx is listening on
EXPOSE 8080

# Use the default CMD from the nginx image, which runs Nginx in the foreground
CMD ["nginx", "-g", "daemon off;"]
