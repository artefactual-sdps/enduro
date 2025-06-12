package entclient_test

import (
	"testing"
	"time"

	"github.com/artefactual-sdps/enduro/internal/datatypes"
	"github.com/artefactual-sdps/enduro/internal/persistence/ent/db"
	"github.com/go-logr/logr"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/google/uuid"
	"go.artefactual.dev/tools/ref"
	"gotest.tools/v3/assert"
)

func TestCreateUser(t *testing.T) {
	t.Parallel()

	uid := uuid.New()
	createdAt := ref.New(time.Now().Truncate(time.Second))

	type params struct {
		user *datatypes.User
	}
	tests := []struct {
		name    string
		args    params
		want    *datatypes.User
		wantErr string
	}{
		{
			name: "Creates a new user with all optional values",
			args: params{
				user: &datatypes.User{
					UUID:      uid,
					CreatedAt: createdAt,
					Name:      ref.New("Test User"),
					Email:     ref.New("nobody@example.com"),
					JWTIss:    ref.New("https://oidc.example.com"),
					JWTSub:    ref.New("1234567890"),
				},
			},
			want: &datatypes.User{
				UUID:      uid,
				CreatedAt: createdAt,
				Name:      ref.New("Test User"),
				Email:     ref.New("nobody@example.com"),
				JWTIss:    ref.New("https://oidc.example.com"),
				JWTSub:    ref.New("1234567890"),
			},
		},
		{
			name: "Creates a new user with required values only",
			args: params{
				user: &datatypes.User{UUID: uid},
			},
			want: &datatypes.User{
				UUID:      uid,
				CreatedAt: ref.New(time.Now()),
			},
		},
		{
			name: "Errors when UUID is missing",
			args: params{
				user: &datatypes.User{},
			},
			wantErr: "invalid data error: field \"UUID\" is required",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			_, svc := setUpClient(t, logr.Discard())
			ctx := t.Context()
			user := *tt.args.user // Make a local copy of sip.

			err := svc.CreateUser(ctx, &user)
			if tt.wantErr != "" {
				assert.Error(t, err, tt.wantErr)
				return
			}
			assert.NilError(t, err)

			assert.DeepEqual(t, &user, tt.want,
				cmpopts.EquateApproxTime(time.Millisecond*100),
				cmpopts.IgnoreUnexported(db.User{}, db.UserEdges{}),
			)
		})
	}
}
