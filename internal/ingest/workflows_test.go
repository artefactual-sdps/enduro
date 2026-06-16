package ingest

import (
	"testing"

	"gotest.tools/v3/assert"

	"github.com/artefactual-sdps/enduro/internal/api/auth"
	"github.com/artefactual-sdps/enduro/pkg/childwf"
)

func TestChildWorkflowUserFromClaims(t *testing.T) {
	t.Parallel()

	for _, tt := range []struct {
		name   string
		claims *auth.Claims
		want   *childwf.User
	}{
		{
			name: "Returns nil without claims",
		},
		{
			name:   "Returns nil without email",
			claims: &auth.Claims{Iss: "issuer", Sub: "subject"},
		},
		{
			name: "Returns email only",
			claims: &auth.Claims{
				Email: "nobody@example.com",
				Name:  "Test User",
				Iss:   "issuer",
				Sub:   "subject",
			},
			want: &childwf.User{Email: "nobody@example.com"},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assert.DeepEqual(t, childWorkflowUserFromClaims(tt.claims), tt.want)
		})
	}
}
