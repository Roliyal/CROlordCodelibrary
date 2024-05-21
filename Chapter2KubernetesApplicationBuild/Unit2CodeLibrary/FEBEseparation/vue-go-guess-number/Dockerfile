# First stage: Build the Vue static resources
FROM node:14-alpine AS vue-build
WORKDIR /app
COPY package*.json ./
#ENV NODE_ENV=production
RUN npm install --force
COPY . .
RUN npm run build

# Second stage: Setup Nginx to serve the Vue app
FROM nginx:1.21-alpine
COPY --from=vue-build /app/dist /usr/share/nginx/html
EXPOSE 8080
CMD ["nginx", "-g", "daemon off;"]


* AND __tag__:__path__:/api.log
| SELECT
    __time__ - __time__ % 60 AS Time,
    COUNT_IF(message.path = '/client') AS client_count,
    COUNT_IF(message.path = '/client2') AS client2_count,
    COUNT_IF(message.path = '/client2') AS client3_count
GROUP BY
    Time
ORDER BY
    Time
LIMIT 1000;
