#!/bin/sh
set -eu

cat >/usr/share/nginx/html/env.js <<EOF
window.__ENV__ = {
  VITE_ARMS_RUM_PID: "${VITE_ARMS_RUM_PID:-}",
  VITE_ARMS_RUM_ENDPOINT: "${VITE_ARMS_RUM_ENDPOINT:-}",
  VITE_APP_VERSION: "${VITE_APP_VERSION:-}"
};
EOF

echo "[entrypoint] wrote /usr/share/nginx/html/env.js"
cat /usr/share/nginx/html/env.js || true

: "${API_UPSTREAM:=http://go-gateway:8080}"
export API_UPSTREAM

envsubst '${API_UPSTREAM}' \
  < /etc/nginx/templates/nginx.conf \
  > /etc/nginx/conf.d/nginx.conf

echo "[entrypoint] generated nginx conf:"
cat /etc/nginx/conf.d/nginx.conf || true

exec nginx -g 'daemon off;'
