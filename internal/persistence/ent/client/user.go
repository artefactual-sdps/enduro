package entclient

import (
	"context"

	"github.com/artefactual-sdps/enduro/internal/datatypes"
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
