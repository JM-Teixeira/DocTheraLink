#!/bin/sh
# Docker entrypoint script for DocTheraLink Frontend
# Handles environment variable substitution in runtime configuration

set -e

# Default values
REACT_CARE_API_URL=${REACT_CARE_API_URL:-"http://localhost:9000"}

# Create runtime configuration file
cat > /usr/share/nginx/html/config.js << EOF
window.ENV = {
  REACT_CARE_API_URL: "${REACT_CARE_API_URL}",
  REACT_CUSTOM_DESCRIPTION: "${REACT_CUSTOM_DESCRIPTION:-}",
  REACT_HEADER_LOGO: "${REACT_HEADER_LOGO:-}",
  REACT_MAIN_LOGO: "${REACT_MAIN_LOGO:-}",
  REACT_GMAPS_API_KEY: "${REACT_GMAPS_API_KEY:-}",
  REACT_GOV_DATA_API_KEY: "${REACT_GOV_DATA_API_KEY:-}",
  REACT_RECAPTCHA_SITE_KEY: "${REACT_RECAPTCHA_SITE_KEY:-}",
  REACT_PLAUSIBLE_SITE_DOMAIN: "${REACT_PLAUSIBLE_SITE_DOMAIN:-}",
  REACT_PLAUSIBLE_SERVER_URL: "${REACT_PLAUSIBLE_SERVER_URL:-}",
  REACT_SENTRY_DSN: "${REACT_SENTRY_DSN:-}",
  REACT_SENTRY_ENVIRONMENT: "${REACT_SENTRY_ENVIRONMENT:-}"
};
EOF

# Substitute environment variables in nginx config
envsubst '${REACT_CARE_API_URL}' < /etc/nginx/nginx.conf > /tmp/nginx.conf
mv /tmp/nginx.conf /etc/nginx/nginx.conf

# Execute the main command
exec "$@"