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

# Run inject-vite-envs.sh on Nginx start
COPY --link ./hack/inject-vite-envs.sh /docker-entrypoint.d/50-inject-vite-envs.sh
RUN chmod +x /docker-entrypoint.d/50-inject-vite-envs.sh

ENV ENDURO_DASHBOARD_ROOT=/usr/share/nginx/html
WORKDIR ${ENDURO_DASHBOARD_ROOT}
RUN rm -rf ./*
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
        proxy_set_header Host \$host;
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
