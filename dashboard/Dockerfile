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
    location /api/package/monitor {
        proxy_pass http://backend/package/monitor;
        proxy_http_version 1.1;
        proxy_set_header Connection "Upgrade";
        proxy_set_header Upgrade \$http_upgrade;
        proxy_set_header Host \$http_host;
    }
    location /api/ {
        proxy_pass http://backend/;
        proxy_redirect / /api/;
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
