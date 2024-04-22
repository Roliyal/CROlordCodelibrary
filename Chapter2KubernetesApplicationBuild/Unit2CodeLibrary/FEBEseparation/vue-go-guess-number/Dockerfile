# First stage: Build the Vue static resources
FROM registry.aliyuncs.com/aliyun-node/14-alpine AS vue-build
WORKDIR /app
COPY package*.json ./
ENV NODE_ENV=production
RUN npm cache clean --force
RUN npm install --no-cache --verbose
COPY . ./
RUN npm  run build

FROM docker.io/nginx:1.21-alpine
COPY --from=vue-build /app/dist /usr/share/nginx/html
RUN rm /etc/nginx/conf.d/default.conf
RUN mkdir -p /var/cache/nginx/client_temp /var/run/nginx && \
    chown -R nginx:nginx /var/cache/nginx /var/run/nginx && \
    chmod -R 755 /var/cache/nginx /var/run/nginx

EXPOSE 8080

CMD ["nginx", "-g", "daemon off;"]
