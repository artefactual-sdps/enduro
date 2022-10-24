# syntax = docker/dockerfile:1.4

FROM node:19 AS builder
ARG oidcProviderUrl=http://dex:5556
ARG oidcRedirectUrl=http://localhost:3000/user/signin-callback
ARG oidcAudienceClientId=enduro
ENV VITE_OIDC_AUTHORITY=$oidcProviderUrl
ENV VITE_OIDC_REDIRECT_URI=$oidcRedirectUrl
ENV VITE_OIDC_AUDIENCE_CLIENT_ID=$oidcAudienceClientId
WORKDIR /app
COPY --link package.json package-lock.json ./
RUN --mount=type=cache,target=/app/.npm \
    npm set cache /app/.npm && npm install-clean
COPY --link . .
RUN npm run build
CMD npm run dev

FROM nginx:alpine
WORKDIR /usr/share/nginx/html
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
