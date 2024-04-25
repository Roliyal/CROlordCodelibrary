# First stage: Build the Vue static resources
FROM node:14-alpine AS vue-build

WORKDIR /app

# Improved layer caching by only copying package files and installing dependencies before copying entire project
COPY package*.json ./

# Uncomment the following line if you are using a `.npmrc` file for custom registry configuration
# COPY .npmrc ./

#ENV NODE_ENV=production
RUN npm install

# Now copy the rest of your app's source code from your host to your image filesystem.
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
