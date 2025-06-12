package entclient

import (
	"context"

	"github.com/artefactual-sdps/enduro/internal/datatypes"
	"github.com/artefactual-sdps/enduro/internal/persistence/ent/db/user"
	"github.com/google/uuid"
)

// CreateUser creates and persists a new user.
func (c *client) CreateUser(ctx context.Context, u *datatypes.User) error {
	// Validate required fields.
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
