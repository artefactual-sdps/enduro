# Identity and access control

Enduro optionally supports external OpenID Connect (OIDC) compatible providers
for authentication and access control. Users can authenticate against the
external provider from the dashboard to receive an access token that will be
sent to the API on each request.

Enduro uses Attribute Based Access Control (ABAC) to determine the actions and
resources to which an authenticated user has access. It looks for a configurable
claim in the access token to know the attributes assigned to the user in their
external provider.

This section explains how to configure the OIDC provider in Enduro's API and
dashboard.

## API configuration

Below is a self-documented API section from an Enduro configuration file in
TOML format:

```toml
[api]
# TCP address for the server to listen on, in the form "host:port".
listen = "0.0.0.0:9000"
# Enable debug mode.
debug = false
# Allowed CORS origin URL.
corsOrigin = "http://localhost"

[api.auth]
# Enable API authentication. OIDC is the only protocol supported at the
# moment. When enabled the API verifies the access token submitted with
# each request. The API client is responsible for obtaining an access
# token from the provider.
enabled = true

[api.auth.oidc]
# OIDC provider URL. Required when auth. is enabled.
providerURL = "http://keycloak:7470/realms/artefactual"
# OIDC client ID. The client ID must be included in the `aud` claim of
# the access token. Required when auth. is enabled.
clientID = "enduro"
# Do not check if the `email_verified` claim is present and set to `true`.
skipEmailVerifiedCheck = false

[api.auth.oidc.abac]
# Enable Attribute Based Access Control (ABAC). If enabled, the API will
# check a configurable multivalue claim against required attributes based
# on each endpoint configuration.
enabled = true
# Claim path of the Enduro attributes within the access token. If the claim
# path is nested then include all fields separated by `claimPathSeparator`
# (see below). E.g. "attributes.enduro" with `claimPathSeparator = "."`.
# Required when ABAC is enabled.
claimPath = "enduro"
# Separator used to split the claim path fields. The default value of "" will
# try to match the claim path as-is to a top-level field from the access token.
claimPathSeparator = ""
# Add a prefix to filter the values of the configured claim. If the claim
# contains values unrelated to Enduro's ABAC, the values relevant to Enduro
# should be prefixed so they are the only values used for access control.
# For example, a claim with values ["enduro:*", "unrelated"] will be filtered
# to a value of ["*"] when `claimValuePrefix = "enduro:"`. The default "" will
# not filter any value.
claimValuePrefix = ""
# Consider the values obtained from the claim as roles and use the `rolesMapping`
# config below to map them to Enduro attributes.
useRoles = false
# A JSON formatted string specifying a mapping from expected roles to Enduro
# attributes. JSON format:
# {
#   "role1": ["attribute1", "atrribute2"],
#   "role2": ["attribute1", "atrribute2", "attribute3", "atrribute4"]
# }
# Example:
# rolesMapping = '{"admin": ["*"], "operator": ["ingest:sips:list", "ingest:sips:actions:list", "ingest:sips:move", "ingest:sips:read", "ingest:sips:upload"], "readonly": ["ingest:sips:list", "ingest:sips:actions:list", "ingest:sips:read"]}'
rolesMapping = ""

[api.auth.ticket.redis]
# Redis URI to store a ticket used to set a websocket connection.
address = "redis://redis.enduro-sdps:6379"
# Prefix used as part of the ticket keys in Redis.
prefix = "enduro"
```

## Dashboard configuration

The following environment variables can be used to configure the dashboard:

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
```

They must match the ones configured in the API. `VITE_OIDC_AUTHORITY` has to be
the same OIDC provider URL and `VITE_OIDC_CLIENT_ID` needs to be the same or a
trusted client. This client (or the one used in the API configuration, if they
are not the same) must be included in the `aud` claim from the access token.
`VITE_OIDC_BASE_URL` will be used to generate the signin and signout callback
URLs, to set them in the OIDC provider for this client, they will be:

- Signin: `VITE_OIDC_BASE_URL` + `/user/signin-callback`
- Signout: `VITE_OIDC_BASE_URL` + `/user/signout-callback`

The authorization flow will request the `openid email profile` scopes by
default. If needed, `VITE_OIDC_SCOPES` can be used to replace those scopes.

`VITE_OIDC_EXTRA_QUERY_PARAMS` can be set to specify further query string
parameters to be including in the authorization request. E.g, when using Azure
AD a resource parameter is required, or using Auth0 you may need to send an
audience client ID. The expected format is key-value pairs separated by `=`
(`audience=client-id`), if more than one parameter is needed they can be added
separated by comma (`audience=client-id,key=value`).

The ABAC variables will work in the same way as they do in the API, they are
explained in detail in the API configuration comments above.

These environment variables can be set at build time, or they can be replaced in
the final assets. For example, the following script uses `envsubst` to do the
replacement:

```bash
#!/usr/bin/env bash

ENDURO_DASHBOARD_ROOT=/usr/lib/enduro-dashboard
TMP_DIR=/tmp/inject_vite_envs
mkdir $TMP_DIR

# Get a comma delimited list of env var names starting with "VITE"
VITE_ENVS=$(printenv | awk -F= '$1 ~ /^VITE/ {print $1}' | sed 's/^/\$/g' | paste -sd,);
echo "Vite envs: ${VITE_ENVS}"

# Inject environment variables into distribution files
for file in $ENDURO_DASHBOARD_ROOT/assets/*.js;
do
    echo "Inject VITE environment variables into $(basename $file)"
    envsubst $VITE_ENVS < $file > $TMP_DIR/$(basename $file)
    cp $TMP_DIR/$(basename $file) $file
done

rm -rf $TMP_DIR
```

## Required attributes

The following table shows the attributes required for each API endpoint. The
attributes allow a wildcard hierarchical declaration. For example, `ingest:sips:*`
will give access to endpoints requiring `ingest:sips:list`, `ingest:sips:read`, etc.
The `*` attribute will provide full access to the API.

| Method | Endpoint                               | Attributes                    |
| ------ | -------------------------------------- | ----------------------------- |
| GET    | /ingest/sips                           | `ingest:sips:list`            |
| GET    | /ingest/sips/{id}                      | `ingest:sips:read`            |
| POST   | /ingest/sips/{id}/confirm              | `ingest:sips:review`          |
| GET    | /ingest/sips/{id}/move                 | `ingest:sips:move`            |
| POST   | /ingest/sips/{id}/move                 | `ingest:sips:move`            |
| GET    | /ingest/sips/{id}/preservation-actions | `ingest:sips:actions:list`    |
| POST   | /ingest/sips/{id}/reject               | `ingest:sips:review`          |
| POST   | /ingest/sips/upload                    | `ingest:sips:upload`          |
| POST   | /storage/aips                          | `storage:aips:create`         |
| GET    | /storage/aips/{uuid}                   | `storage:aips:read`           |
| GET    | /storage/aips/{uuid}/download          | `storage:aips:download`       |
| POST   | /storage/aips/{uuid}/reject            | `storage:aips:review`         |
| GET    | /storage/aips/{uuid}/store             | `storage:aips:move`           |
| POST   | /storage/aips/{uuid}/store             | `storage:aips:move`           |
| POST   | /storage/aips/{uuid}/submit            | `storage:aips:submit`         |
| POST   | /storage/aips/{uuid}/update            | `storage:aips:submit`         |
| GET    | /storage/locations                     | `storage:locations:list`      |
| POST   | /storage/locations                     | `storage:locations:create`    |
| GET    | /storage/locations/{uuid}              | `storage:locations:read`      |
| GET    | /storage/locations/{uuid}/aips         | `storage:locations:aips:list` |
