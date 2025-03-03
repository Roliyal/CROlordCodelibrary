# Use the official Golang image as the base image
FROM golang:1.22.4 AS builder

# Set the working directory
WORKDIR /app

# Set Go proxy to Alibaba Cloud Go Proxy for faster module download
ENV GOPROXY=https://mirrors.aliyun.com/goproxy/

# 下载 instgo 工具并适配架构
RUN uname -m && \
    if [ "$(uname -m)" = "x86_64" ]; then \
        wget "http://arms-apm-ap-southeast-1.oss-ap-southeast-1-internal.aliyuncs.com/instgo/instgo-linux-amd64" -O instgo; \
    elif [ "$(uname -m)" = "aarch64" ]; then \
        wget "http://arms-apm-ap-southeast-1.oss-ap-southeast-1-internal.aliyuncs.com/instgo/instgo-linux-arm64" -O instgo; \
    else \
        echo "Unsupported architecture"; exit 1; \
    fi && \
    chmod +x instgo

# 设置 LicenseKey 和 RegionId
RUN ./instgo set --mse  --licenseKey=djqtzchc9t@27546800413b30e --regionId=ap-southeast-1


# Copy the go.mod, go.sum, and .env.production files to the working directory
COPY go.mod go.sum .env ./

# Download the dependencies
RUN go mod download

# Copy the source code to the working directory
COPY . .

# Compile the Go program

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 ./instgo go build -o main-amd64 .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=arm64 ./instgo go build -o main-arm64 .


# Use the official Alpine image as the base image
FROM alpine

# Add a non-root user
RUN addgroup -S appgroup && adduser -S appuser -G appgroup

# Set the working directory to /app
WORKDIR /app

# Copy the compiled Go binary and .env.production files
COPY --from=builder /app/main-amd64 /app/main-amd64
COPY --from=builder /app/main-arm64 /app/main-arm64
COPY --from=builder /app/.env /app/.env
COPY --from=builder /app/start.sh /app/start.sh

USER root

RUN chmod +x /app/start.sh

RUN mkdir -p /app/log && chown -R appuser:appgroup /app/log

RUN chown -R appuser:appgroup /app

USER appuser

# Expose the port, make sure it matches the port used by your application
EXPOSE 8085

# Add a health check
HEALTHCHECK --interval=60s --timeout=5s --start-period=5s --retries=3 CMD ["/app/start.sh", "check"]

# Run the Go program
ENTRYPOINT ["/app/start.sh"]

