FROM node:14-alpine AS vue-build

WORKDIR /app

COPY package*.json ./

RUN npm ci

COPY . .

RUN npm run build

# Second stage: Setup Nginx to serve the Vue app
FROM nginx:1.21-alpine

# Copy static assets from builder stage
COPY --from=vue-build /app/dist /usr/share/nginx/html

# Optional: Copy your nginx configuration file if you have custom configurations
# COPY path/to/your/nginx.conf /etc/nginx/nginx.conf

EXPOSE 8080
CMD ["nginx", "-g", "daemon off;"]
