#!/bin/sh
# Docker entrypoint for the Klyniva frontend.
# Writes the runtime configuration (/config.js) and renders nginx.conf so one
# image works in every environment.

set -e

# Public URL the browser uses to reach the API. Behind the bundled nginx proxy
# this is the frontend's own origin (nginx forwards /api/ to the backend).
REACT_CARE_API_URL=${REACT_CARE_API_URL:-"http://localhost:4000"}

# Served at /config.js by nginx (see "location = /config.js" in nginx.conf)
cat > /tmp/config.js << EOF
window.ENV = {
  REACT_CARE_API_URL: "${REACT_CARE_API_URL}",
  REACT_MAIN_LOGO: "${REACT_MAIN_LOGO:-}",
  REACT_GMAPS_API_KEY: "${REACT_GMAPS_API_KEY:-}",
  REACT_RECAPTCHA_SITE_KEY: "${REACT_RECAPTCHA_SITE_KEY:-}"
};
EOF

# Only substitute our variable; nginx's own $variables must stay untouched.
export BACKEND_UPSTREAM=${BACKEND_UPSTREAM:-"backend:9000"}
envsubst '${BACKEND_UPSTREAM}' < /etc/nginx/nginx.conf.template > /tmp/nginx.conf

exec "$@"
