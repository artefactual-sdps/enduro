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
					Email:     ref.New("nobody@example.com"),
					Name:      ref.New("Test User"),
					JWTIss:    ref.New("https://oidc.example.com"),
					JWTSub:    ref.New("1234567890"),
				},
			},
			want: &datatypes.User{
				UUID:      uid,
				CreatedAt: createdAt,
				Email:     ref.New("nobody@example.com"),
				Name:      ref.New("Test User"),
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
			user := *tt.args.user // Make a local copy of user.

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

func TestReadUser(t *testing.T) {
	t.Parallel()

	userID := uuid.New()
	createdAt := ref.New(time.Now().Truncate(time.Second))

	type params struct {
		id uuid.UUID
	}
	tests := []struct {
		name    string
		args    params
		want    *datatypes.User
		wantErr string
	}{
		{
			name: "Reads a user with all values",
			args: params{id: userID},
			want: &datatypes.User{
				UUID:      userID,
				CreatedAt: createdAt,
				Email:     ref.New("nobody@example.com"),
				Name:      ref.New("Test User"),
				JWTIss:    ref.New("https://oidc.example.com"),
				JWTSub:    ref.New("1234567890"),
			},
		},
		{
			name:    "Errors when id is nil",
			args:    params{},
			wantErr: "invalid data error: field \"id\" is required",
		},
		{
			name:    "Errors when user is not found",
			args:    params{id: uuid.New()},
			wantErr: "not found error: db: user not found",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			c, svc := setUpClient(t, logr.Discard())
			ctx := t.Context()

			c.User.Create().
				SetUUID(userID).
				SetNillableCreatedAt(createdAt).
				SetEmail("nobody@example.com").
				SetName("Test User").
				SetJwtIss("https://oidc.example.com").
				SetJwtSub("1234567890").
				SaveX(ctx)

			got, err := svc.ReadUser(ctx, tt.args.id)
			if tt.wantErr != "" {
				assert.Error(t, err, tt.wantErr)
				return
			}
			assert.NilError(t, err)

			assert.DeepEqual(t, got, tt.want,
				cmpopts.IgnoreUnexported(db.User{}, db.UserEdges{}),
			)
		})
	}
}
