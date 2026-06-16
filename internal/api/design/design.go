/*
Package design is the single source of truth of Enduro's API. It uses the Goa
design language (https://goa.design) which is a Go DSL.

We describe multiple services which map to resources in REST or service declarations
in gRPC. Services define their own methods, errors, etc...
*/
package design

import (
	"encoding/json"

	. "goa.design/goa/v3/dsl" //nolint:staticcheck
	"goa.design/goa/v3/expr"
	cors "goa.design/plugins/v3/cors/dsl"

	"github.com/artefactual-sdps/enduro/internal/api/auth"
)

var BearerAuth = BearerSecurity("bearer", func() {
	Description("Secures endpoint by requiring a valid bearer token.")
	Scope(auth.IngestBatchesCreateAttr)
	Scope(auth.IngestBatchesListAttr)
	Scope(auth.IngestBatchesReadAttr)
	Scope(auth.IngestBatchesReviewAttr)
	Scope(auth.IngestSIPSCreateAttr)
	Scope(auth.IngestSIPSDecisionAttr)
	Scope(auth.IngestSIPSDownloadAttr)
	Scope(auth.IngestSIPSListAttr)
	Scope(auth.IngestSIPSReadAttr)
	Scope(auth.IngestSIPSReviewAttr)
	Scope(auth.IngestSIPSUploadAttr)
	Scope(auth.IngestSIPSWorkflowsListAttr)
	Scope(auth.IngestSIPSourcesObjectsListAttr)
	Scope(auth.IngestUsersListAttr)
	Scope(auth.StorageAIPSCreateAttr)
	Scope(auth.StorageAIPSDeletionAutoAttr)
	Scope(auth.StorageAIPSDeletionReportAttr)
	Scope(auth.StorageAIPSDeletionRequestAttr)
	Scope(auth.StorageAIPSDeletionReviewAttr)
	Scope(auth.StorageAIPSDownloadAttr)
	Scope(auth.StorageAIPSListAttr)
	Scope(auth.StorageAIPSMoveAttr)
	Scope(auth.StorageAIPSReadAttr)
	Scope(auth.StorageAIPSReviewAttr)
	Scope(auth.StorageAIPSWorkflowsListAttr)
	Scope(auth.StorageLocationsAIPSListAttr)
	Scope(auth.StorageLocationsCreateAttr)
	Scope(auth.StorageLocationsListAttr)
	Scope(auth.StorageLocationsReadAttr)
})

func BearerAuthScopes(scopes ...string) {
	Security(BearerAuth, func() {
		for _, scope := range scopes {
			Scope(scope)
		}
	})

	requiredScopes := append([]string{}, scopes...)
	data, err := json.Marshal(requiredScopes)
	if err != nil {
		panic(err)
	}
	Meta("openapi:extension:x-required-scopes", string(data))
}

var _ = API("enduro", func() {
	Title("Enduro API")
	Randomizer(expr.NewDeterministicRandomizer())
	Server("enduro", func() {
		Services("about", "ingest", "storage")
		Host("localhost", func() {
			URI("http://localhost:9000")
		})
	})
	Security(BearerAuth)
	HTTP(func() {
		Consumes("application/json")
	})
	cors.Origin("$ENDURO_API_CORS_ORIGIN", func() {
		cors.Methods("GET", "HEAD", "POST", "PUT", "DELETE", "OPTIONS")
		cors.Headers("Authorization", "Content-Type")
	})
})
