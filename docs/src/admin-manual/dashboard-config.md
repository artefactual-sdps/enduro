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
VITE_INSTITUTION_LOGO_LIGHT=http://localhost:8080/institution-logo-light.png
VITE_INSTITUTION_LOGO_DARK=http://localhost:8080/institution-logo-dark.png
VITE_INSTITUTION_LOGO=http://localhost:8080/institution-logo.png
```

`VITE_INSTITUTION_LOGO_LIGHT` and `VITE_INSTITUTION_LOGO_DARK` set the logo URL
for the light and dark themes. For each theme, an unset or empty theme-specific
setting falls back to `VITE_INSTITUTION_LOGO`. This keeps existing single-logo
configurations working. If neither the active theme's setting nor
`VITE_INSTITUTION_LOGO` is configured, no institutional logo is displayed.

Logo files can be hosted on a remote server or CDN. A local logo can be placed
in `dashboard/public/` before building the dashboard; its URL uses the dashboard
server's scheme and host plus the file name.

!!! note

    If a configured logo URL is hosted on a different origin, update the
    dashboard Content Security Policy (CSP) `img-src` directive to allow that
    origin. Otherwise, the browser will block the logo. See the
    [nginx configuration example].

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

## Custom home page HTML

The Enduro dashboard can display custom HTML content on the home page by
setting the following environment variable:

```bash
VITE_CUSTOM_HOME_URL=/custom/home.html
```

`VITE_CUSTOM_HOME_URL` sets a URL for a custom HTML file to be displayed on the
home page. If it's not set, or set to an empty string, then the default home
page content will be displayed.

The custom HTML content is sanitized using [DOMPurify] before being rendered.
If the custom HTML file cannot be loaded, an error message will be displayed
and the default home page content will be shown instead. Due to Content
Security Policy restrictions, inline CSS styles are not permitted by default.
Use Bootstrap CSS classes instead, or modify the CSP header. Local files can be
served from the `/custom/` directory (see the [nginx configuration example]).

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
[Configuration]: configuration.md
[DOMPurify]: https://github.com/cure53/DOMPurify
[nginx configuration example]: dashboard-build.md#nginx-example
[Vite environment variables]: https://vite.dev/guide/env-and-mode
