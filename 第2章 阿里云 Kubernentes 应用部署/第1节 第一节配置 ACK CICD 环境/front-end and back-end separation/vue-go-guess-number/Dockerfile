# 第一阶段：构建 Vue 静态资源
FROM node:14-alpine AS vue-build
WORKDIR /app
COPY vue-go-guess-number/package*.json ./
RUN npm install --production
COPY vue-go-guess-number/ ./
RUN npm run build

# 第二阶段：将构建好的静态资源复制到 Nginx 镜像中
FROM nginx:1.21-alpine
COPY --from=vue-build /app/dist /usr/share/nginx/html
COPY nginx/default.conf /etc/nginx/conf.d/default.conf
EXPOSE 80
USER nginx
CMD ["nginx", "-g", "daemon off;"]
