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
		SetNillableCreatedAt(u.CreatedAt).
		SetNillableEmail(u.Email).
		SetNillableName(u.Name).
		SetNillableJwtIss(u.JWTIss).
		SetNillableJwtSub(u.JWTSub)

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

// ReadUserJWT retrieves a user by JWT issuer and subject.
func (c *client) ReadUserJWT(ctx context.Context, iss, sub string) (*datatypes.User, error) {
	// Validate required fields.
	if iss == "" {
		return nil, newRequiredFieldError("iss")
	}
	if sub == "" {
		return nil, newRequiredFieldError("sub")
	}

	// Query the user by iss and sub.
	q := c.ent.User.Query()
	q.Where(user.And(user.JwtIss(iss), user.JwtSub(sub)))

	dbu, err := q.Only(ctx)
	if err != nil {
		return nil, newDBError(err)
	}

	return convertUser(dbu), nil
}
