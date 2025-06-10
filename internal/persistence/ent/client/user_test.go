package entclient_test

import (
	"testing"
	"time"

	"github.com/go-logr/logr"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/google/uuid"
	"gotest.tools/v3/assert"

	"github.com/artefactual-sdps/enduro/internal/datatypes"
	"github.com/artefactual-sdps/enduro/internal/persistence/ent/db"
)

func TestCreateUser(t *testing.T) {
	t.Parallel()

	uid := uuid.New()
	createdAt := time.Now().Truncate(time.Second)

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
					Email:     "nobody@example.com",
					Name:      "Test User",
					OIDCIss:   "https://oidc.example.com",
					OIDCSub:   "1234567890",
				},
			},
			want: &datatypes.User{
				UUID:      uid,
				CreatedAt: createdAt,
				Email:     "nobody@example.com",
				Name:      "Test User",
				OIDCIss:   "https://oidc.example.com",
				OIDCSub:   "1234567890",
			},
		},
		{
			name: "Creates a new user with required values only",
			args: params{
				user: &datatypes.User{UUID: uid},
			},
			want: &datatypes.User{
				UUID:      uid,
				CreatedAt: time.Now(),
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
	createdAt := time.Now().Truncate(time.Second)

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
				Email:     "nobody@example.com",
				Name:      "Test User",
				OIDCIss:   "https://oidc.example.com",
				OIDCSub:   "1234567890",
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
				SetCreatedAt(createdAt).
				SetEmail("nobody@example.com").
				SetName("Test User").
				SetOidcIss("https://oidc.example.com").
				SetOidcSub("1234567890").
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

func TestReadOIDCUser(t *testing.T) {
	t.Parallel()

	userID := uuid.New()
	createdAt := time.Now().Truncate(time.Second)

	type params struct {
		iss string
		sub string
	}

	tests := []struct {
		name    string
		args    params
		want    *datatypes.User
		wantErr string
	}{
		{
			name: "Reads a user with all values",
			args: params{iss: "https://oidc.example.com", sub: "1234567890"},
			want: &datatypes.User{
				UUID:      userID,
				CreatedAt: createdAt,
				Email:     "nobody@example.com",
				Name:      "Test User",
				OIDCIss:   "https://oidc.example.com",
				OIDCSub:   "1234567890",
			},
		},
		{
			name:    "Errors when iss is empty",
			args:    params{},
			wantErr: "invalid data error: field \"iss\" is required",
		},
		{
			name:    "Errors when sub is empty",
			args:    params{iss: "https://oidc.example.com"},
			wantErr: "invalid data error: field \"sub\" is required",
		},
		{
			name:    "Errors when user is not found",
			args:    params{iss: "https://oidc.example.com", sub: "not-found"},
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
				SetCreatedAt(createdAt).
				SetEmail("nobody@example.com").
				SetName("Test User").
				SetOidcIss("https://oidc.example.com").
				SetOidcSub("1234567890").
				SaveX(ctx)

			got, err := svc.ReadOIDCUser(ctx, tt.args.iss, tt.args.sub)
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
