# ---------- build ----------
FROM node:20-alpine AS build
WORKDIR /app

ARG VITE_ARMS_RUM_PID
ARG VITE_ARMS_RUM_ENDPOINT
ARG VITE_APP_VERSION

ENV VITE_ARMS_RUM_PID=${VITE_ARMS_RUM_PID}
ENV VITE_ARMS_RUM_ENDPOINT=${VITE_ARMS_RUM_ENDPOINT}
ENV VITE_APP_VERSION=${VITE_APP_VERSION}

COPY package.json package-lock.json* ./
RUN npm ci

COPY . .
RUN npm run build

# ---------- runtime ----------
FROM nginx:1.27-alpine

COPY --from=build /app/dist /usr/share/nginx/html

COPY nginx.conf /etc/nginx/conf.d/default.conf

EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]
