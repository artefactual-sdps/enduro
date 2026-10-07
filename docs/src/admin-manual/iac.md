# Identity and access control

Enduro optionally supports external OpenID Connect (OIDC) compatible providers
for authentication and access control. Users can authenticate against the
external provider from the dashboard to receive an access token that will be
sent to the API on each request.

Enduro uses Attribute Based Access Control (ABAC) to determine the actions and
resources to which an authenticated user has access. It looks for a configurable
claim in the access token to know the attributes assigned to the user in their
external provider.

This section explains the identity provider requirements and the authentication
configuration for Enduro's API, dashboard, and ingest storage client.

## Identity provider requirements

Enduro supports OIDC for user authentication. Dashboard login and
authentication between services use different OAuth 2.0 flows and can use
different identity providers.

### Dashboard login

The dashboard uses the **Authorization Code flow with Proof Key for Code
Exchange (PKCE)**, using the `S256` challenge method. It is a public browser
application, also called a single-page application (SPA), and does not use a
client secret.

The identity provider's dashboard registration must allow this flow and include
the sign-in and post-logout redirect URIs described in
[Dashboard configuration](#dashboard-configuration).

The dashboard requests access tokens for the Enduro API. Depending on the
provider, this may require API scopes or additional authorization request
parameters beyond the default `openid email profile` scopes.

### Authentication between services

When ingest-to-storage authentication is enabled, ingest uses the **Client
Credentials flow** to obtain access tokens for the storage API without user
interaction. This requires a confidential client with a client ID and client
secret. A separate client registration is recommended so that service
credentials and permissions can be managed independently of dashboard login.

The service client can use the same identity provider as the dashboard or a
different provider. At least one of the API's configured OIDC verifiers must
accept its tokens. When ABAC is enabled for that verifier, the tokens must also
contain the attributes required by the storage endpoints used by ingest.

The provider, credentials, and any required scopes or audience parameter are
configured under `[ingest.storage.oidc]`, as described in
[Ingest storage client configuration](#ingest-storage-client-configuration).

### Access-token requirements

The API validates signed JWT access tokens, including their signature, issuer,
audience, and expiry. The issuer must match a configured OIDC provider, and the
`aud` claim must include the corresponding API `clientID` value.

User access tokens must include a stable `sub` claim uniquely identifying the
user within the issuer. Enduro uses the combination of `iss` and `sub` to
identify users when recording ingest actions and handling deletion requests.
Requesting, reviewing, or canceling an AIP deletion also requires at least one
nonempty `email`, `preferred_username`, or `name` claim, checked in that order,
for the display name. These identity claims are also required in service tokens
used for automatic AIP deletion, such as during batch cancellation.

When attribute-based access control (ABAC) is enabled, the configured
permissions claim must contain an array ofstrings in the access token. These
values can be Enduro attributes or roles mapped to attributes through
`rolesMapping`. The dashboard and API must use matching access control settings
for user tokens. See [API configuration](#api-configuration) and [Required
attributes](#required-attributes) for the configuration and supported
permissions.

By default, the API also requires `email_verified: true` in the access token.
The `skipEmailVerifiedCheck` setting disables this requirement for an OIDC
verifier, for example when a provider does not supply the claim or when service
identities do not have an email address. Claims used by the API must be present
in the access token; providing them only in an ID token or UserInfo response is
insufficient.

## API configuration

The following example configures `enduro` for dashboard users and `enduro-s2s`
for service authentication. ABAC is enabled for dashboard users only:

```toml
[api]
# TCP address for the server to listen on, in the form "host:port".
listen = "0.0.0.0:9000"
# Allowed CORS origin URL.
corsOrigin = "https://enduro.example.com"

[api.auth]
# Verify the access token submitted with each request.
enabled = true

# User access tokens obtained by the dashboard.
[[api.auth.oidc]]
providerURL = "https://idp.example.com"
clientID = "enduro"
skipEmailVerifiedCheck = false

[api.auth.oidc.abac]
enabled = true
claimPath = "roles"
claimPathSeparator = ""
claimValuePrefix = ""
useRoles = true
rolesMapping = '{"admin": ["*"]}'

# Service access tokens obtained by ingest.
[[api.auth.oidc]]
providerURL = "https://idp.example.com"
clientID = "enduro-s2s"
# Skip the email verification check for service tokens.
skipEmailVerifiedCheck = true

[api.auth.ticket.redis]
# Redis URI to store tickets used for browser download handoffs.
address = "redis://redis:6379"
# Prefix used as part of the ticket keys in Redis.
prefix = "enduro"
```

The [API authentication reference](configuration.md#oidc-authentication-providers-configuration)
describes all verifier and ABAC settings. The
[required attributes](#required-attributes) determine the permissions to map
for additional user roles or service operations.

## Ingest storage client configuration

Use the following section to configure ingest as an authenticated client of the
storage API using the OAuth 2.0 Client Credentials flow. Tokens generated by this
OIDC provider must be verified by at least one of the providers from the full
API OIDC configuration.

```toml
[ingest.storage]
# Storage API host:port for ingest client requests.
address = "enduro-api:9000"
# Default destination location for permanent storage workflows.
defaultPermanentLocationId = "f2cc963f-c14d-4eaa-b950-bd207189a1f1"

[ingest.storage.oidc]
# Enable service to service OIDC authentication for storage API requests.
enabled = true
# OIDC provider URL used for token endpoint discovery.
providerURL = "https://idp.example.com"
# Optional token endpoint URL. If set, discovery is skipped.
tokenURL = ""
# OIDC client credentials used for client_credentials token requests.
clientID = "enduro-s2s"
clientSecret = "replace-me"
# Optional scopes requested during token retrieval.
scopes = ""
# Optional audience to include in token endpoint params.
audience = ""
# Refresh token before expiry to avoid edge races.
tokenExpiryLeeway = "30s"
# Retry settings for transient token endpoint failures.
retryMaxAttempts = 3
retryInitialInterval = "500ms"
retryMaxInterval = "2s"
retryBackoffCoefficient = 2.0
```

## Dashboard configuration

The dashboard's [OIDC settings](dashboard-config.md#oidc-settings) configure
authentication and access control. This section describes how those settings
relate to the identity provider and API configuration.

!!! important

    The `VITE_OIDC_AUTHORITY` environment variable is included in the Content
    Security Policy's `connect-src` directive. If this URL contains a path,
    ensure it ends with a slash (`/`) to allow connections to the necessary
    OIDC endpoints for authentication.

`VITE_OIDC_AUTHORITY` identifies the provider used for dashboard login and must
correspond to one of the API's configured OIDC providers. `VITE_OIDC_CLIENT_ID`
identifies the public dashboard client. The API's corresponding `clientID`
setting must be included in the `aud` claim of the resulting access token and
can differ from `VITE_OIDC_CLIENT_ID`.

`VITE_OIDC_BASE_URL` will be used to generate the signin and signout callback
URLs, to set them in the OIDC provider for this client, they will be:

- Signin: `VITE_OIDC_BASE_URL` + `/user/signin-callback`
- Signout: `VITE_OIDC_BASE_URL` + `/user/signout-callback`

The Authorization Code flow with PKCE requests the `openid email profile`
scopes by default. `VITE_OIDC_SCOPES` replaces that list with a space-separated
list of scopes. When additional API scopes are required, the configured list
must also retain `openid` and any required profile scopes.

`VITE_OIDC_EXTRA_QUERY_PARAMS` can be set to specify further query string
parameters required by the provider in the authorization request. The expected
format is comma-separated `key=value` pairs, for example
`audience=api-audience,key=value`.

The ABAC variables work in the same way as the API's ABAC settings. In this
example, they match the user token verifier:

```sh
VITE_OIDC_ENABLED=true
VITE_OIDC_BASE_URL=https://enduro.example.com
VITE_OIDC_AUTHORITY=https://idp.example.com
VITE_OIDC_CLIENT_ID=enduro
VITE_OIDC_ABAC_ENABLED=true
VITE_OIDC_ABAC_CLAIM_PATH=roles
VITE_OIDC_ABAC_USE_ROLES=true
VITE_OIDC_ABAC_ROLES_MAPPING='{"admin": ["*"]}'
```

See [Building and serving the dashboard](dashboard-build.md#environment) for
applying these settings at build time, or
[Environment variable injection](dashboard-build.md#environment-variable-injection)
for configuring pre-built assets.

## Required attributes

The following table shows the attributes required for each API endpoint. The
attributes allow a wildcard hierarchical declaration. For example,
`ingest:sips:*` will give access to endpoints requiring `ingest:sips:list`,
`ingest:sips:read`, etc. The `*` attribute will provide full access to the API.

The ingest and storage monitor streams are SSE endpoints authenticated with the
same bearer token used by the rest of the API. User claims are checked before
sending events to the stream.

Similarly, to be able to stream a SIP, AIP or AIP deletion report download from
the browser, the `GET` endpoints require a cookie obtained from the `POST`
endpoints.

| Method | Endpoint                              | Attributes                       |
| ------ | ------------------------------------- | -------------------------------- |
| GET    | /about                                | `-`                              |
| POST   | /ingest/batches                       | `ingest:batches:create`          |
| GET    | /ingest/batches                       | `ingest:batches:list`            |
| GET    | /ingest/batches/{uuid}                | `ingest:batches:read`            |
| POST   | /ingest/batches/{uuid}/review         | `ingest:batches:review`          |
| GET    | /ingest/monitor                       | `-`                              |
| GET    | /ingest/sip-sources/{uuid}/objects    | `ingest:sipsources:objects:list` |
| GET    | /ingest/sips                          | `ingest:sips:list`               |
| POST   | /ingest/sips                          | `ingest:sips:create`             |
| POST   | /ingest/sips/upload                   | `ingest:sips:upload`             |
| GET    | /ingest/sips/{uuid}                   | `ingest:sips:read`               |
| POST   | /ingest/sips/{uuid}/confirm           | `ingest:sips:review`             |
| GET    | /ingest/sips/{uuid}/decision          | `ingest:sips:decision`           |
| POST   | /ingest/sips/{uuid}/decision          | `ingest:sips:decision`           |
| GET    | /ingest/sips/{uuid}/download          | `-`                              |
| POST   | /ingest/sips/{uuid}/download          | `ingest:sips:download`           |
| POST   | /ingest/sips/{uuid}/reject            | `ingest:sips:review`             |
| GET    | /ingest/sips/{uuid}/workflows         | `ingest:sips:workflows:list`     |
| GET    | /ingest/users                         | `ingest:users:list`              |
| GET    | /storage/aips                         | `storage:aips:list`              |
| POST   | /storage/aips                         | `storage:aips:create`            |
| GET    | /storage/aips/{uuid}                  | `storage:aips:read`              |
| POST   | /storage/aips/{uuid}/deletion-auto    | `storage:aips:deletion:auto`     |
| POST   | /storage/aips/{uuid}/deletion-cancel  | `storage:aips:deletion:request`  |
| GET    | /storage/aips/{uuid}/deletion-report  | `-`                              |
| POST   | /storage/aips/{uuid}/deletion-report  | `storage:aips:deletion:report`   |
| POST   | /storage/aips/{uuid}/deletion-request | `storage:aips:deletion:request`  |
| POST   | /storage/aips/{uuid}/deletion-review  | `storage:aips:deletion:review`   |
| GET    | /storage/aips/{uuid}/download         | `-`                              |
| POST   | /storage/aips/{uuid}/download         | `storage:aips:download`          |
| POST   | /storage/aips/{uuid}/reject           | `storage:aips:review`            |
| GET    | /storage/aips/{uuid}/store            | `storage:aips:move`              |
| POST   | /storage/aips/{uuid}/store            | `storage:aips:move`              |
| GET    | /storage/aips/{uuid}/workflows        | `storage:aips:workflows:list`    |
| GET    | /storage/locations                    | `storage:locations:list`         |
| POST   | /storage/locations                    | `storage:locations:create`       |
| GET    | /storage/locations/{uuid}             | `storage:locations:read`         |
| GET    | /storage/locations/{uuid}/aips        | `storage:locations:aips:list`    |
| GET    | /storage/monitor                      | `-`                              |
