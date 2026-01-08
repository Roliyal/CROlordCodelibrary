# ---------- build ----------
FROM node:latest AS build
WORKDIR /app

COPY package.json package-lock.json* ./
RUN npm i

COPY . .
RUN npm run build

# ---------- runtime ----------
FROM nginx:latest

COPY --from=build /app/dist /usr/share/nginx/html
COPY docker-entrypoint.sh /docker-entrypoint.sh
RUN chmod +x /docker-entrypoint.sh

EXPOSE 80
ENTRYPOINT ["/docker-entrypoint.sh"]
