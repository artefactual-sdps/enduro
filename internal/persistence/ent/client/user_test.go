package entclient_test

import (
	"testing"
	"time"

	"github.com/go-logr/logr"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/google/uuid"
	"go.artefactual.dev/tools/ref"
	"gotest.tools/v3/assert"

	"github.com/artefactual-sdps/enduro/internal/datatypes"
	"github.com/artefactual-sdps/enduro/internal/persistence"
	"github.com/artefactual-sdps/enduro/internal/persistence/ent/db"
	"github.com/artefactual-sdps/enduro/internal/timerange"
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

func TestListUsers(t *testing.T) {
	t.Parallel()

	userUUID := uuid.New()
	userUUID2 := uuid.New()
	createdAt := time.Date(2025, 6, 23, 13, 29, 55, 0, time.UTC)
	createdAt2 := time.Date(2025, 6, 23, 13, 32, 22, 0, time.UTC)

	tests := []struct {
		name    string
		filter  *persistence.UserFilter
		want    []*datatypes.User
		wantPg  *persistence.Page
		wantErr string
	}{
		{
			name: "Lists all users",
			want: []*datatypes.User{
				{
					UUID:      userUUID,
					CreatedAt: createdAt,
					Email:     "nobody@example.com",
					Name:      "Nobody Example",
					OIDCIss:   "https://oidc.example.com",
					OIDCSub:   "1234567890",
				},
				{
					UUID:      userUUID2,
					CreatedAt: createdAt2,
					Email:     "test@example.com",
					Name:      "Test User",
					OIDCIss:   "https://oidc.example.com",
					OIDCSub:   "0987654321",
				},
			},
			wantPg: &persistence.Page{
				Limit:  20,
				Offset: 0,
				Total:  2,
			},
		},
		{
			name: "Lists users filtered by CreatedAt",
			filter: &persistence.UserFilter{
				CreatedAt: &timerange.Range{
					Start: time.Date(2025, 6, 23, 13, 20, 0, 0, time.UTC),
					End:   time.Date(2025, 6, 23, 13, 30, 0, 0, time.UTC),
				},
			},
			want: []*datatypes.User{
				{
					UUID:      userUUID,
					CreatedAt: createdAt,
					Email:     "nobody@example.com",
					Name:      "Nobody Example",
					OIDCIss:   "https://oidc.example.com",
					OIDCSub:   "1234567890",
				},
			},
			wantPg: &persistence.Page{
				Limit:  20,
				Offset: 0,
				Total:  1,
			},
		},
		{
			name: "Lists users filtered by Email",
			filter: &persistence.UserFilter{
				Email: ref.New("nobody@example.com"),
			},
			want: []*datatypes.User{
				{
					UUID:      userUUID,
					CreatedAt: createdAt,
					Email:     "nobody@example.com",
					Name:      "Nobody Example",
					OIDCIss:   "https://oidc.example.com",
					OIDCSub:   "1234567890",
				},
			},
			wantPg: &persistence.Page{
				Limit:  20,
				Offset: 0,
				Total:  1,
			},
		},
		{
			name: "Lists users filtered by Name",
			filter: &persistence.UserFilter{
				Name: ref.New("test"),
			},
			want: []*datatypes.User{
				{
					UUID:      userUUID2,
					CreatedAt: createdAt2,
					Email:     "test@example.com",
					Name:      "Test User",
					OIDCIss:   "https://oidc.example.com",
					OIDCSub:   "0987654321",
				},
			},
			wantPg: &persistence.Page{
				Limit:  20,
				Offset: 0,
				Total:  1,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			entc, svc := setUpClient(t, logr.Discard())
			ctx := t.Context()

			entc.User.Create().
				SetUUID(userUUID).
				SetCreatedAt(createdAt).
				SetEmail("nobody@example.com").
				SetName("Nobody Example").
				SetOidcIss("https://oidc.example.com").
				SetOidcSub("1234567890").
				SaveX(ctx)

			entc.User.Create().
				SetUUID(userUUID2).
				SetCreatedAt(createdAt2).
				SetEmail("test@example.com").
				SetName("Test User").
				SetOidcIss("https://oidc.example.com").
				SetOidcSub("0987654321").
				SaveX(ctx)

			got, pg, err := svc.ListUsers(ctx, tt.filter)
			if tt.wantErr != "" {
				assert.Error(t, err, tt.wantErr)
				return
			}
			assert.NilError(t, err)

			assert.DeepEqual(t, got, tt.want,
				cmpopts.IgnoreUnexported(db.User{}, db.UserEdges{}),
			)
			assert.DeepEqual(t, pg, tt.wantPg)
		})
	}
}
