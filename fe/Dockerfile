# Klyniva frontend image
# Build: docker build -t klyniva/frontend ./fe

# ---- Build stage ----------------------------------------------------------
FROM node:24-alpine AS builder
WORKDIR /app

# Install all dependencies (Vite, TypeScript and Tailwind are devDependencies
# and are needed for the build). --ignore-scripts skips the husky git hook.
COPY package.json package-lock.json ./
RUN npm ci --ignore-scripts --no-audit --no-fund

COPY . .

# The real API URL is injected at container start through /config.js
# (see docker-entrypoint.sh); Vite only needs a value at build time.
ENV REACT_CARE_API_URL=http://placeholder
# generate = git-ignored sources (pluginMap.ts, supportedBrowsers.ts)
RUN npm run build:meta && npm run generate && npm run build:react

# ---- Runtime stage --------------------------------------------------------
FROM nginx:1.29-alpine

COPY nginx.conf /etc/nginx/nginx.conf.template
COPY --from=builder /app/build /usr/share/nginx/html
COPY docker-entrypoint.sh /docker-entrypoint-klyniva.sh
RUN chmod +x /docker-entrypoint-klyniva.sh

# Where nginx proxies /api/ (compose service name by default)
ENV BACKEND_UPSTREAM=backend:9000
EXPOSE 8080

# Run as the image's unprivileged nginx user (uid 101); see nginx.conf header.
USER 101

HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget -qO- http://127.0.0.1:8080/health >/dev/null || exit 1

ENTRYPOINT ["/docker-entrypoint-klyniva.sh"]
CMD ["nginx", "-c", "/tmp/nginx.conf", "-g", "daemon off;"]
