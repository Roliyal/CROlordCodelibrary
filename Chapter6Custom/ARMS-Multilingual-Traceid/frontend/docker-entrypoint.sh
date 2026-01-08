#!/bin/sh
set -eu

cat >/usr/share/nginx/html/env.js <<EOF
window.__ENV__ = {
  VITE_ARMS_RUM_PID: "${VITE_ARMS_RUM_PID:-}",
  VITE_ARMS_RUM_ENDPOINT: "${VITE_ARMS_RUM_ENDPOINT:-}",
  VITE_APP_VERSION: "${VITE_APP_VERSION:-}"
};
EOF

exec nginx -g 'daemon off;'



