# First stage: Build the Vue static resources
FROM node:14-alpine AS vue-build
WORKDIR /app
COPY package*.json ./
#ENV NODE_ENV=productio
RUN npm cache clean
RUN npm install
COPY . .

# Second stage: Setup Nginx to serve the Vue app
FROM nginx:1.21-alpine
COPY --from=vue-build /app/dist /usr/share/nginx/html
EXPOSE 8080
CMD ["nginx", "-g", "daemon off;"]

