# Build stage 
FROM node:14-slim AS build-stage

WORKDIR /app

# Install dependencies
COPY package*.json ./
RUN npm install --production && npm cache clean --force
RUN npm install axios

# Copy source files
COPY . .

# Build the application
RUN npm run build

# Production stage
FROM nginx:1.21.0-alpine

# Copy built files and nginx configuration
COPY --from=build-stage /app/dist /usr/share/nginx/html
COPY --from=build-stage /app/nginx.conf /etc/nginx/conf.d/default.conf


# Expose port 80
EXPOSE 80

# Start Nginx
CMD ["nginx", "-g", "daemon off;"]
