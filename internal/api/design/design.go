/*
Package design is the single source of truth of Enduro's API. It uses the Goa
design language (https://goa.design) which is a Go DSL.

We describe multiple services (package) which map to resources in
REST or service declarations in gRPC. Services define their own methods, errors,
etc...
*/
package design

import (
	. "goa.design/goa/v3/dsl"
	cors "goa.design/plugins/v3/cors/dsl"
)

var OAuth2Auth = OAuth2Security("oauth2", func() {
	// We only validate for now, but a flow definition is required.
	ClientCredentialsFlow("/oauth2/token", "/oauth2/refresh")
	Description("Secures endpoints by requiring a valid OAuth2 access token.")
})

var _ = API("enduro", func() {
	Title("Enduro API")
	Server("enduro", func() {
		Services("package", "storage", "swagger", "upload")
		Host("localhost", func() {
			URI("http://localhost:9000")
		})
	})
	Security(OAuth2Auth)
	HTTP(func() {
		Consumes("application/json")
	})
	cors.Origin("*", func() {
		cors.Methods("GET", "HEAD", "POST", "PUT", "DELETE", "OPTIONS")
		cors.Headers("Authorization", "Content-Type")
	})
})
