#!/bin/sh
set -e

if [ "$1" = "check" ]; then
    curl -sf 127.0.0.1:8083/health || exit 1
    exit 0
fi

case "$(uname -m)" in
  x86_64)  BIN="/app/main-amd64" ;;
  aarch64) BIN="/app/main-arm64" ;;
  *) echo "Unsupported arch"; exit 1 ;;
esac

echo "Launching $BIN ..."
exec "$BIN"
