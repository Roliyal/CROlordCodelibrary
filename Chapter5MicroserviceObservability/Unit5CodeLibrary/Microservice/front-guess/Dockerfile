# Build stage 
FROM node:18 AS build-stage

WORKDIR /app

# Install dependencies
COPY package*.json ./
#RUN npm install --production && npm cache clean --force
RUN npm ci  # 构建时需要 devDeps --production

#RUN npm install axios

COPY . .
RUN npm run build

# Production stage
FROM nginx:1.23.0-alpine

# Copy built files and nginx configuration
COPY --from=build-stage /app/dist /usr/share/nginx/html
COPY --from=build-stage /app/nginx.conf /etc/nginx/conf.d/default.conf


# Expose port 80
EXPOSE 80

# Start Nginx
CMD ["nginx", "-g", "daemon off;"]
