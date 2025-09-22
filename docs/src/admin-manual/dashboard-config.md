# Dashboard configuration

This page describes how to configure a custom institutional logo to be displayed
in the page header of the user interface. For additional Enduro configuration,
see:

* [Configuration]

## Institution logo

The Enduro dashboard can display an institutional logo in the page header by
setting the following [Vite environment variables].

### Institution logo configuration values

```bash
VITE_INSTITUTION_LOGO=http://localhost:8080/artefactual-logo.png
```

`VITE_INSTITUTION_LOGO` sets a URL for the institution logo image file. The logo
file can be hosted on a remote server or CDN (such as Cloudflare). A local logo
can be used by placing the logo file in the `dashboard/public/` directory before
building the Dashboard application. The URL for a local logo file will be the
schema and hostname of the Dashboard server plus the name of the logo file. For
instance, in the Tilt development environment the example URL will load the
local Artefactual logo file at `dashboard/public/artefactual-logo.png`.

If the `VITE_INSTITUTION_LOGO` is not set, or set to any empty string, then no
institutional logo will be displayed.

```bash
VITE_INSTITUTION_NAME="Artefactual Systems Inc."
```

`VITE_INSTITUTION_NAME` sets the name of the institution, to be used as [alt]
text for the logo image. For web accessibility `VITE_INSTITUTION_NAME` should
always be set when a institutional logo is configured.

```bash
VITE_INSTITUTION_URL=https://www.artefactual.com
```

`VITE_INSTITUTION_URL` is an optional setting that provides a URL linking to
an institutional web page. If `VITE_INSTITUTION_URL` is set then clicking on the
institutional logo will open a new browser tab and load the given URL.

If no `VITE_INSTITUTION_URL` is set, then the institutional logo will not be
clickable.

## OIDC settings

The following environment variables can be used to configure an OpenID Connect
(OIDC) provider for authentication and access control.

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

Check the [Identity and access control](iac.md) page for more information.

[alt]: https://developer.mozilla.org/en-US/docs/Web/API/HTMLImageElement/alt
[Configuration]: ../admin-manual/configuration.md
[Vite environment variables]: https://vite.dev/guide/env-and-mode
