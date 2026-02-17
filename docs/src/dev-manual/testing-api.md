# Testing the API

## Available APIs in the development environment

The development environment exposes two APIs from the Kubernetes cluster:

- **Port 9000** uses the configuration from `enduro.toml`.
- **Port 9002** serves the API without authentication.

The latter is used to remove the need for an access token during development,
like running `make upload-sample-transfer`, while the first one can be used to
test the [identity and access control configuration](../admin-manual/iac.md).

## Getting an access token

To access the Enduro API at port 9000 when authentication is enabled in the
configuration, an access token needs to be sent in the requests. To get one
from the default Keycloak instance in the environment, run the following Make
command from the root of the repository:

```sh
make auth
```

This Make rule allows you to set three parameters to obtain an access token
from a different provider. For example:

```sh
make auth HOST=http://example.com CLIENT=enduro SCOPES=openid,email,profile
```

To use the client credentials flow, set `CLIENT_SECRET`:

```sh
make auth CLIENT=enduro-s2s CLIENT_SECRET=uSh7f2r4j2U5wA9d7mJ3xP6nQ8cT1vL0
```

After authentication, the script will output the token payload for inspection
and its encoded value for API authentication.

## Making requests to the API

Use the access token on each request to the API. For example:

```sh
curl -H "Authorization: Bearer <token>" http://localhost:9000/ingest/sips
```

Replace `<token>` with the access token value obtained with `make auth`.
