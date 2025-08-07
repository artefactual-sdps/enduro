package entclient

import (
	"context"
	"errors"

	"github.com/google/uuid"

	"github.com/artefactual-sdps/enduro/internal/datatypes"
	"github.com/artefactual-sdps/enduro/internal/entfilter"
	"github.com/artefactual-sdps/enduro/internal/persistence"
	"github.com/artefactual-sdps/enduro/internal/persistence/ent/db"
	"github.com/artefactual-sdps/enduro/internal/persistence/ent/db/sip"
	"github.com/artefactual-sdps/enduro/internal/persistence/ent/db/user"
)

// CreateUser creates and persists a new user.
func (c *client) CreateUser(ctx context.Context, u *datatypes.User) error {
	_, err := createUser(ctx, c.ent, u)
	return err
}

// createUser creates and persists a new user.
// It allows passing an Ent Client that binds to a running transaction.
func createUser(ctx context.Context, entc *db.Client, u *datatypes.User) (int, error) {
	// UUID is a required field.
	if u.UUID == uuid.Nil {
		return 0, newRequiredFieldError("UUID")
	}

	q := entc.User.Create().
		SetUUID(u.UUID).
		SetEmail(u.Email).
		SetName(u.Name).
		SetOidcIss(u.OIDCIss).
		SetOidcSub(u.OIDCSub)

	// Optionally set CreatedAt if it is not zero.
	if !u.CreatedAt.IsZero() {
		q.SetCreatedAt(u.CreatedAt)
	}

	// Save the User.
	dbu, err := q.Save(ctx)
	if err != nil {
		return 0, newDBErrorWithDetails(err, "create user")
	}

	// Update User with DB data, to get generated values (e.g. ID).
	*u = *convertUser(dbu)

	return dbu.ID, nil
}

// ReadUser retrieves a user by UUID.
func (c *client) ReadUser(ctx context.Context, id uuid.UUID) (*datatypes.User, error) {
	// Validate required fields.
	if id == uuid.Nil {
		return nil, newRequiredFieldError("id")
	}

	// Query the user by UUID.
	dbu, err := c.ent.User.Query().Where(user.UUID(id)).Only(ctx)
	if err != nil {
		return nil, newDBError(err)
	}

	return convertUser(dbu), nil
}

// ReadOIDCUser retrieves a user by OIDC issuer and subject.
func (c *client) ReadOIDCUser(ctx context.Context, iss, sub string) (*datatypes.User, error) {
	_, u, err := readOIDCUser(ctx, c.ent, iss, sub)
	return u, err
}

// readOIDCUser retrieves a user by OIDC issuer and subject.
// It allows passing an Ent Client that binds to a running transaction.
func readOIDCUser(ctx context.Context, entc *db.Client, iss, sub string) (int, *datatypes.User, error) {
	// Validate required fields.
	if iss == "" {
		return 0, nil, newRequiredFieldError("iss")
	}
	if sub == "" {
		return 0, nil, newRequiredFieldError("sub")
	}

	// Query the user by iss and sub.
	q := entc.User.Query()
	q.Where(user.And(user.OidcIss(iss), user.OidcSub(sub)))

	dbu, err := q.Only(ctx)
	if err != nil {
		return 0, nil, newDBError(err)
	}

	return dbu.ID, convertUser(dbu), nil
}

// findOrCreateUser finds or creates a user by OIDC issuer and subject.
// It allows passing an Ent Client that binds to a running transaction.
func findOrCreateUser(ctx context.Context, entc *db.Client, u *datatypes.User) (int, error) {
	// Try to find the user by OIDC Iss and Sub.
	uID, readUser, err := readOIDCUser(ctx, entc, u.OIDCIss, u.OIDCSub)
	if err != nil {
		if !errors.Is(err, persistence.ErrNotFound) {
			return 0, err
		}

		// The user does not exist, so create a new one.
		uID, err = createUser(ctx, entc, u)
		if err != nil {
			return 0, err
		}
	} else {
		*u = *readUser
	}

	return uID, nil
}

func (c *client) ListUsers(ctx context.Context, f *persistence.UserFilter) (
	[]*datatypes.User, *persistence.Page, error,
) {
	if f == nil {
		f = &persistence.UserFilter{}
	}

	page, whole := filterUsers(c.ent.User.Query(), f)
	r, err := page.All(ctx)
	if err != nil {
		return nil, nil, newDBError(err)
	}

	// Convert to datatypes.User slice.
	users := make([]*datatypes.User, len(r))
	for i, dbu := range r {
		users[i] = convertUser(dbu)
	}

	total, err := whole.Count(ctx)
	if err != nil {
		return nil, nil, newDBError(err)
	}

	pp := &persistence.Page{
		Limit:  f.Limit,
		Offset: f.Offset,
		Total:  total,
	}

	return users, pp, nil
}

func sortableFields() entfilter.SortableFields {
	return entfilter.SortableFields{
		user.FieldID:        {Name: "ID", Default: true},
		user.FieldEmail:     {Name: "Email"},
		user.FieldName:      {Name: "Name"},
		user.FieldCreatedAt: {Name: "CreatedAt"},
	}
}

// filterUsers applies the User filter f to query q and return a paginated amd
// unpaginated query.
func filterUsers(q *db.UserQuery, f *persistence.UserFilter) (page, whole *db.UserQuery) {
	qf := entfilter.NewFilter(q, sortableFields())

	qf.AddDateRange(sip.FieldCreatedAt, f.CreatedAt)
	qf.Contains(user.FieldName, f.Name)
	qf.Contains(user.FieldEmail, f.Email)

	// Update the pager values with the actual values set on the query.
	// E.g. calling `qf.Page(0,0)` will set the query limit equal to the default
	// page size.
	f.Limit = qf.Limit
	f.Offset = qf.Offset

	return qf.Apply()
}
