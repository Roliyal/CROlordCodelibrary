# Use the official Golang image as the base image
FROM golang:1.22.4 AS builder

# Set the working directory
WORKDIR /app

# Set Go proxy to Alibaba Cloud Go Proxy for faster module download
ENV GOPROXY=https://mirrors.aliyun.com/goproxy/

# 下载 instgo 工具
RUN wget "http://arms-apm-cn-hangzhou.oss-cn-hangzhou.aliyuncs.com/instgo/instgo-linux-amd64" -O instgo && \
    chmod +x instgo

# 设置 LicenseKey 和 RegionId（如果你有 LicenseKey，可以替换 {Y ourLicenseKey} 和 {YourRegionId}）
RUN ./instgo set --mse --licenseKey={YourLicenseKey} --regionId={YourRegionId} --dev=true

# Copy the go.mod, go.sum, and .env.production files to the working directory
COPY go.mod go.sum .env ./

# Download the dependencies
RUN go mod download

# Copy the source code to the working directory
COPY . .

# Compile the Go program
#RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o main .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 ./instgo go build -o main .


# Use the official Alpine image as the base image
FROM alpine

# Add a non-root user
RUN addgroup -S appgroup && adduser -S appuser -G appgroup

# Set the working directory to /app
WORKDIR /app

# Copy the compiled Go binary and .env.production files
COPY --from=builder /app/main /app/main
COPY --from=builder /app/.env /app/.env

# Create the log directory and change its ownership to the non-root user
RUN mkdir -p /app/log && chown -R appuser:appgroup /app/log

# Change the ownership of the working directory to the non-root user
RUN chown -R appuser:appgroup /app

# Switch to the non-root user
USER appuser

# Expose the port, make sure it matches the port used by your application
EXPOSE 8083

# Add a health check
HEALTHCHECK --interval=60s --timeout=5s --start-period=5s --retries=3 CMD ["/app/main", "check"]

# Run the Go program
ENTRYPOINT ["/app/main"]
