# Build stage
FROM node:14-slim AS build-stage

WORKDIR /app

# Copy package.json and package-lock.json
COPY package*.json ./

# Install dependencies
RUN npm ci

# Copy the rest of the project files
COPY . .

# Build the project
RUN npm run build || (cat /app/npm-debug.log && exit 1)

# Production stage
FROM nginx:1.21.0-alpine

# Copy built files from build stage to nginx html folder
COPY --from=build-stage /app/dist /usr/share/nginx/html

# Copy nginx configuration
COPY --from=build-stage /app/nginx.conf /etc/nginx/conf.d/default.conf

# Expose port 80
EXPOSE 80

# Start nginx
CMD ["nginx", "-g", "daemon off;"]
