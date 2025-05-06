#!/bin/sh

# 检查当前架构，并根据架构启动对应的程序
ARCH=$(uname -m)

if [ "$ARCH" = "x86_64" ]; then
    echo "Running AMD64 architecture binary..."
    /app/main-amd64
elif [ "$ARCH" = "aarch64" ]; then
    echo "Running ARM64 architecture binary..."
    /app/main-arm64
else
    echo "Unsupported architecture: $ARCH"
    exit 1
fi
