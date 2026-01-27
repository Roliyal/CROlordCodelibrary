#!/usr/bin/env bash
set -euo pipefail

# Generate Go protobuf + gRPC stubs from proto/bridge/v1/bridge.proto
#
# Install tools (pin versions to match go.mod):
#   go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.36.11
#   go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.6.0
#
# Install protoc (macOS):
#   brew install protobuf && brew link protobuf

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT_DIR"

rm -f pb/*.go 2>/dev/null || true
mkdir -p pb

protoc -I proto \
  proto/bridge/v1/bridge.proto \
  --go_out=. --go_opt=module=armslogcollect/go-service \
  --go-grpc_out=. --go-grpc_opt=module=armslogcollect/go-service

echo "[ok] generated go-service/pb/*.go"
