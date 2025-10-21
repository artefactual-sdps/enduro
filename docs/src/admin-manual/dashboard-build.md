# Building and serving the dashboard

This page describes how to build and serve the Enduro dashboard. The dashboard
is a [Vue.js] application built with [Vite] that provides the web interface
for managing and monitoring Enduro.

## Requirements

To build the dashboard from source, you need:

- **Node.js** (version 20 or later)
- **npm** (included with [Node.js])

## Environment

The following environment variables are used to configure the dashboard:

```txt
VITE_OIDC_ENABLED
VITE_OIDC_BASE_URL
VITE_OIDC_AUTHORITY
VITE_OIDC_CLIENT_ID
VITE_OIDC_SCOPES
VITE_OIDC_ABAC_ENABLED
VITE_OIDC_ABAC_CLAIM_PATH
VITE_OIDC_ABAC_CLAIM_PATH_SEPARATOR
VITE_OIDC_ABAC_CLAIM_VALUE_PREFIX
VITE_OIDC_ABAC_USE_ROLES
VITE_OIDC_ABAC_ROLES_MAPPING

VITE_INSTITUTION_LOGO
VITE_INSTITUTION_NAME
VITE_INSTITUTION_URL
```

See the [Dashboard configuration](dashboard-config.md) for more information
about these variables. When building from source, these variables will be read
from the process environment or else will fall back to the values in the `.env`
file.

!!! important

    These environment variables are processed at **build time** by Vite.
    The values in the `.env` file are placeholders that can be replaced by
    deployment injection scripts to customize the dashboard for different
    environments without rebuilding.

## Building from source

### Install dependencies

First, navigate to the dashboard directory and install the required packages:

```bash
cd dashboard
npm install
```

### Production build

To build the dashboard for production:

```bash
npm run build
```

The build creates a `dist/` directory with minified JavaScript, CSS, and HTML
files ready for deployment.

## Serving built assets

### Using pre-built releases

Enduro releases include pre-built dashboard assets. When using these pre-built
assets, you must configure environment variables after deployment using the
provided injection script.

#### Environment variable injection

The injection script replaces environment variable placeholders in the built
assets with actual values. This allows using pre-built dashboard assets while
customizing configuration for different environments.

Set the following environment variables before running the script:

**Required path variables:**

- `ENDURO_DASHBOARD_ROOT`: Path to the directory where dashboard files are
  served
- `ENDURO_DASHBOARD_DIST`: Path to the pre-built dashboard distribution files

**Dashboard configuration variables:**

- Any `VITE_*` environment variables you want to configure (see
  [Environment](#environment) section above)

The script will:

1. Copy distribution files to the dashboard root directory
2. Find all environment variables starting with `VITE_`
3. Replace placeholders in JavaScript and HTML files with actual values
4. Remove any unreplaced placeholders

**Injection script:**

The following script is also included in the source code:

- [dashboard/hack/inject-vite-envs.sh][injection_script]

```bash
#!/usr/bin/env bash

TMP_DIR=/tmp/inject_vite_envs
mkdir $TMP_DIR

# Copy original distribution files to dashboard file root
rm -rf $ENDURO_DASHBOARD_ROOT/*
cp -r $ENDURO_DASHBOARD_DIST/* $ENDURO_DASHBOARD_ROOT/

# Get a comma delimited list of env var names starting with "VITE"
VITE_ENVS=$(printenv | awk -F= '$1 ~ /^VITE/ {print $1}' | sed 's/^/\$/g' | paste -sd,);
echo "Vite envs: ${VITE_ENVS}"

# Inject environment variables into distribution files and remove
# placeholders that were not replaced (env. vars. not set).
for file in $ENDURO_DASHBOARD_ROOT/assets/*.js;
do
    echo "Inject VITE environment variables into assets/$(basename $file)"
    envsubst $VITE_ENVS < $file > $TMP_DIR/$(basename $file)
    sed -E -i 's/\$VITE_[A-Z0-9_]+//g' $TMP_DIR/$(basename $file)
    cp $TMP_DIR/$(basename $file) $file
done
echo "Inject VITE environment variables into index.html"
envsubst $VITE_ENVS < $ENDURO_DASHBOARD_ROOT/index.html > $TMP_DIR/index.html
sed -E -i 's/\$VITE_[A-Z0-9_]+//g' $TMP_DIR/index.html
cp $TMP_DIR/index.html $ENDURO_DASHBOARD_ROOT/index.html

rm -rf $TMP_DIR
```

**Example usage:**

```bash
export ENDURO_DASHBOARD_DIST=/opt/enduro/dashboard/dist
export ENDURO_DASHBOARD_ROOT=/usr/share/nginx/html
./inject-vite-envs.sh
```

### Web server configuration

The dashboard is a single-page application (SPA) that connects to the Enduro
API using the `/api` prefix on the same host. A web server proxy configuration
is required to forward API requests to the actual Enduro API server.

#### NGINX example

```nginx
upstream backend {
    least_conn;
    server enduro-api:9000;
}

server {
    listen 80;
    root /usr/share/nginx/html;
    absolute_redirect off;

    # Content Security Policy (CSP):
    # Replace or remove the OIDC provider and institution logo URLs as needed
    add_header Content-Security-Policy "default-src 'self'; script-src-attr 'none'; img-src 'self' data: http://institution.com/logo.png; frame-src 'none'; object-src 'none'; base-uri 'self'; form-action 'self'; connect-src 'self' http://oidcprovider.com; frame-ancestors 'none'" always;

    # When serving over HTTPS:
    # (Append to CSP) ... ; upgrade-insecure-requests

    # Reporting:
    # (Append to CSP) ... ; report-uri https://example.com/csp

    # Reporting (not broadly supported yet):
    # add_header Reporting-Endpoints 'csp="https://example.com/csp"' always;
    # (Append to CSP) ... ; report-to csp

    # WebSocket support for ingest monitoring
    location /api/ingest/monitor {
        proxy_pass http://backend/ingest/monitor;
        proxy_http_version 1.1;
        proxy_set_header Connection "Upgrade";
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Host $http_host;
    }

    # Large file upload support for SIP uploads
    location /api/ingest/sips/upload {
        client_max_body_size 4096M;
        proxy_pass http://backend/ingest/sips/upload;
        proxy_request_buffering off;
        proxy_read_timeout 24h;
        proxy_send_timeout 24h;
    }

    # WebSocket support for storage monitoring
    location /api/storage/monitor {
        proxy_pass http://backend/storage/monitor;
        proxy_http_version 1.1;
        proxy_set_header Connection "Upgrade";
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Host $http_host;
    }

    # General API proxy
    location /api/ {
        proxy_pass http://backend/;
        proxy_redirect / /api/;
    }

    # Custom content
    location /custom/ {
        try_files \$uri =404;
    }

    # Handle client-side routing for SPA
    location / {
        try_files $uri $uri/ /index.html;
    }
}
```

[vue.js]: https://vuejs.org/
[vite]: https://vite.dev/
[node.js]: https://nodejs.org/
[injection_script]: https://github.com/artefactual-sdps/enduro/blob/main/dashboard/hack/inject-vite-envs.sh
