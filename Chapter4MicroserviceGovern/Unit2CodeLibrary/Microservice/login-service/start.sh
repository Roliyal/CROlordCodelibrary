#!/bin/sh

set -e

# 1) 选择可执行文件
ARCH=$(uname -m)
case "$ARCH" in
    x86_64)  BIN="/app/main-amd64" ;;
    aarch64) BIN="/app/main-arm64" ;;
    *) echo "Unsupported architecture: $ARCH"; exit 1 ;;
esac

echo "Launching $BIN ..."
# 2) 用 exec 替换当前 shell，使二进制成为 PID 1
exec "$BIN"
