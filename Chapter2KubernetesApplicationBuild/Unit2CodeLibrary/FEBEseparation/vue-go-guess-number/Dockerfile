# 使用 Node.js 的 Docker 镜像
FROM node:14-alpine AS vue-build

WORKDIR /app

COPY package*.json ./

RUN npm install --force

COPY . .
RUN npm run build

FROM nginx:1.21-alpine

COPY --from=vue-build /app/dist /usr/share/nginx/html

EXPOSE 8080

CMD ["nginx", "-g", "daemon off;"]
