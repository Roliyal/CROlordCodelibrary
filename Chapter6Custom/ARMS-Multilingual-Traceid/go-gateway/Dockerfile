# ---------- build ----------
FROM golang:1.22-alpine AS build
WORKDIR /src

COPY go-gateway/go.mod go-gateway/go.sum ./
RUN go mod download

COPY go-gateway/ ./

# 编译
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/go-gateway ./main.go

# ---------- runtime ----------
FROM alpine:3.20
RUN apk add --no-cache ca-certificates
WORKDIR /app

COPY --from=build /out/go-gateway /app/go-gateway

COPY .env /app/.env

ENV DOTENV_PATH=/app/.env

EXPOSE 8080
ENTRYPOINT ["/app/go-gateway"]
