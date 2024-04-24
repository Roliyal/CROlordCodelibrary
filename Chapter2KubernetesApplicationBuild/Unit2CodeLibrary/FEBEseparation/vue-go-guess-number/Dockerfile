# First stage: Build the Vue static resources
FROM node:14-alpine AS vue-build
WORKDIR /app
COPY package*.json ./
ENV NODE_ENV=production
RUN npm install --no-cache --verbose
COPY . .
RUN npm run build

# List contents of the /app/dist directory
RUN ls /app/dist

# Second stage: Setup Nginx to serve the Vue app
FROM nginx:1.21-alpine
COPY --from=vue-build /app/dist /usr/share/nginx/html
RUN rm /etc/nginx/conf.d/default.conf
RUN mkdir -p /var/cache/nginx/client_temp /var/run/nginx && \
    chown -R nginx:nginx /var/cache/nginx /var/run/nginx && \
    chmod -R 755 /var/cache/nginx /var/run/nginx

EXPOSE 8080
CMD ["nginx", "-g", "daemon off;"]
