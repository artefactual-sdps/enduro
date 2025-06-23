package datatypes

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"gotest.tools/v3/assert"

	goaingest "github.com/artefactual-sdps/enduro/internal/api/gen/ingest"
)

func TestGoa(t *testing.T) {
	t.Parallel()

	uid := uuid.New()
	createdAt := time.Date(2025, 6, 23, 14, 57, 12, 0, time.UTC).Truncate(time.Second)

	type test struct {
		name string
		user *User
		want *goaingest.User
	}
	for _, tt := range []test{
		{
			name: "Converts nil User to nil Goa User",
		},
		{
			name: "Converts User to Goa User with all fields",
			user: &User{
				UUID:      uid,
				CreatedAt: createdAt,
				Email:     "nobody@example.com",
				Name:      "Test User",
			},
			want: &goaingest.User{
				UUID:      uid,
				CreatedAt: "2025-06-23T14:57:12Z",
				Email:     "nobody@example.com",
				Name:      "Test User",
			},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := tt.user.Goa()
			assert.DeepEqual(t, got, tt.want)
		})
	}
}
