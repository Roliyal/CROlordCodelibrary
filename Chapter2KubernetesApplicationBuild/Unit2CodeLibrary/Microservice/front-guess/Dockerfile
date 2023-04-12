FROM node:14-slim AS build-stage

WORKDIR /app

COPY package*.json ./

RUN npm run build

COPY . .

RUN npm run build

FROM nginx:1.21.0-alpine
COPY --from=build-stage /app/dist /usr/share/nginx/html

COPY --from=build-stage /app/nginx.conf /etc/nginx/conf.d/default.conf

EXPOSE 80

CMD ["nginx", "-g", "daemon off;"]
