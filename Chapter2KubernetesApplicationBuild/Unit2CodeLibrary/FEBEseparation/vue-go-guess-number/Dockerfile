# First stage: Build the Vue static resources
FROM node:14-alpine AS vue-build
WORKDIR /app
COPY package*.json ./

# Set production environment variable
ENV NODE_ENV=production

# Clean npm cache and remove old node_modules
RUN npm cache clean --force
RUN rm -rf node_modules

# Install dependencies
RUN npm install --no-cache --verbose

# Copy the rest of your app's source code from your host to your image filesystem.
COPY . ./

# Build the project for production
# Using npx to ensure that vue-cli-service is correctly executed
RUN npx vue-cli-service build

# Second stage: Copy the built static resources into an Nginx image
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
