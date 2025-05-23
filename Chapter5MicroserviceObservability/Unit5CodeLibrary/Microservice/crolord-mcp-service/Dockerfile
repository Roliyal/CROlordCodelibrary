# Use the official Golang image as the base image
FROM golang:1.22.4 AS builder

# Set the working directory
WORKDIR /app

# Copy the go.mod, go.sum, and .env.production files to the working directory
COPY * ./

# Download the dependencies
RUN go mod download

# Copy the source code to the working directory
COPY . .

# Compile the Go program
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o main .

# Use the official Alpine image as the base image
FROM alpine

# Add a non-root user
RUN addgroup -S appgroup && adduser -S appuser -G appgroup

# Set the working directory to /app
WORKDIR /app

# Copy the compiled Go binary and .env.production files
COPY --from=builder /app/main /app/main

# Create the log directory and change its ownership to the non-root user
RUN mkdir -p /app/log && chown -R appuser:appgroup /app/log

# Change the ownership of the working directory to the non-root user
RUN chown -R appuser:appgroup /app

# Switch to the non-root user
USER appuser

# Expose the port, make sure it matches the port used by your application
EXPOSE 8080


# Run the Go program
ENTRYPOINT ["/app/main"]
