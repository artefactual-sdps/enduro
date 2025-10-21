# syntax = docker/dockerfile:1.4

FROM node:20 AS builder
WORKDIR /app
COPY --link package.json package-lock.json ./
RUN --mount=type=cache,target=/app/.npm \
    npm set cache /app/.npm && npm install-clean
COPY --link . .
RUN npm run build
CMD npm run dev

FROM nginx:alpine
RUN apk add --no-cache bash

ARG VERSION_PATH
ARG VERSION_LONG
ARG VERSION_SHORT
ARG VERSION_GIT_HASH

ENV ENDURO_DASHBOARD_DIST=/dist
WORKDIR $ENDURO_DASHBOARD_DIST
COPY --from=builder --link /app/dist .

COPY <<EOF /etc/nginx/templates/default.conf.template
upstream backend {
    least_conn;
    server \${ENDURO_API_ADDRESS};
}
server {
    listen 80;
    root /usr/share/nginx/html;
    absolute_redirect off;

    # Content Security Policy (CSP):
    # Replace or remove the OIDC provider and add institution logo URL if needed.
    add_header Content-Security-Policy "default-src 'self'; script-src-attr 'none'; img-src 'self' data:; frame-src 'none'; object-src 'none'; base-uri 'self'; form-action 'self'; connect-src 'self' http://keycloak:7470; frame-ancestors 'none'" always;

    # When serving over HTTPS:
    # (Append to CSP) ... ; upgrade-insecure-requests

    # Reporting:
    # (Append to CSP) ... ; report-uri https://example.com/csp

    # Reporting (not broadly supported yet):
    # add_header Reporting-Endpoints 'csp="https://example.com/csp"' always;
    # (Append to CSP) ... ; report-to csp

    location /api/ingest/monitor {
        proxy_pass http://backend/ingest/monitor;
        proxy_http_version 1.1;
        proxy_set_header Connection "Upgrade";
        proxy_set_header Upgrade \$http_upgrade;
        proxy_set_header Host \$http_host;
    }
    location /api/ingest/sips/upload {
        client_max_body_size 4096M;
        proxy_pass http://backend/ingest/sips/upload;
        proxy_request_buffering off;
        proxy_read_timeout 24h;
        proxy_send_timeout 24h;
    }
    location /api/storage/monitor {
        proxy_pass http://backend/storage/monitor;
        proxy_http_version 1.1;
        proxy_set_header Connection "Upgrade";
        proxy_set_header Upgrade \$http_upgrade;
        proxy_set_header Host \$http_host;
    }
    location /api/ {
        proxy_pass http://backend/;
        proxy_redirect / /api/;
    }
    location /custom/ {
        try_files \$uri =404;
    }
    location / {
        try_files \$uri \$uri/ /index.html;
    }
}
EOF

# Run inject-vite-envs.sh on Nginx start
ENV ENDURO_DASHBOARD_ROOT=/usr/share/nginx/html
COPY --link ./hack/inject-vite-envs.sh /docker-entrypoint.d/50-inject-vite-envs.sh
RUN echo "echo Starting enduro-dashboard version=${VERSION_LONG}" > /docker-entrypoint.d/55-stamping.sh
RUN chmod +x /docker-entrypoint.d/50-inject-vite-envs.sh /docker-entrypoint.d/55-stamping.sh
