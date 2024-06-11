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
	"goa.design/goa/v3/expr"
	cors "goa.design/plugins/v3/cors/dsl"
	_ "goa.design/plugins/v3/otel"
)

var JWTAuth = JWTSecurity("jwt", func() {
	Description("Secures endpoint by requiring a valid JWT token.")
})

var _ = API("enduro", func() {
	Title("Enduro API")
	Randomizer(expr.NewDeterministicRandomizer())
	Server("enduro", func() {
		Services("package", "storage", "swagger", "upload")
		Host("localhost", func() {
			URI("http://localhost:9000")
		})
	})
	Security(JWTAuth)
	HTTP(func() {
		Consumes("application/json")
	})
	cors.Origin("$ENDURO_API_CORS_ORIGIN", func() {
		cors.Methods("GET", "HEAD", "POST", "PUT", "DELETE", "OPTIONS")
		cors.Headers("Authorization", "Content-Type")
	})
})
