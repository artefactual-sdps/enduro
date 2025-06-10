package entclient

import (
	"context"

	"github.com/google/uuid"

	"github.com/artefactual-sdps/enduro/internal/datatypes"
	"github.com/artefactual-sdps/enduro/internal/persistence/ent/db/user"
)

// CreateUser creates and persists a new user.
func (c *client) CreateUser(ctx context.Context, u *datatypes.User) error {
	// UUID is a required field.
	if u.UUID == uuid.Nil {
		return newRequiredFieldError("UUID")
	}

	q := c.ent.User.Create().
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
		return newDBErrorWithDetails(err, "create user")
	}

	// Update User with DB data, to get generated values (e.g. ID).
	*u = *convertUser(dbu)

	return nil
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
	// Validate required fields.
	if iss == "" {
		return nil, newRequiredFieldError("iss")
	}
	if sub == "" {
		return nil, newRequiredFieldError("sub")
	}

	// Query the user by iss and sub.
	q := c.ent.User.Query()
	q.Where(user.And(user.OidcIss(iss), user.OidcSub(sub)))

	dbu, err := q.Only(ctx)
	if err != nil {
		return nil, newDBError(err)
	}

	return convertUser(dbu), nil
}
