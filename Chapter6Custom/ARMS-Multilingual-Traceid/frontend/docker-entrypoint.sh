#!/bin/sh
set -eu

# 1) 写 env.js（RUM 用）
cat >/usr/share/nginx/html/env.js <<EOF
window.__ENV__ = {
  VITE_ARMS_RUM_PID: "${VITE_ARMS_RUM_PID:-}",
  VITE_ARMS_RUM_ENDPOINT: "${VITE_ARMS_RUM_ENDPOINT:-}",
  VITE_APP_VERSION: "${VITE_APP_VERSION:-}"
};
EOF

echo "[entrypoint] wrote /usr/share/nginx/html/env.js"
cat /usr/share/nginx/html/env.js || true

# 2) 生成 nginx conf
: "${API_UPSTREAM:=http://go-gateway:8080}"
export API_UPSTREAM

envsubst '${API_UPSTREAM}' \
  < /etc/nginx/conf.d/default.conf.template \
  > /etc/nginx/conf.d/default.conf

echo "[entrypoint] generated /etc/nginx/conf.d/default.conf"
cat /etc/nginx/conf.d/default.conf || true

exec nginx -g 'daemon off;'
