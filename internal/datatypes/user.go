package datatypes

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	// UUID is a unique identifier for the user that can be used as a public
	// identifier if no better options (e.g. name, email) are available.
	UUID uuid.UUID `db:"uuid"`

	// CreatedAt is an optional field. If the value is nil when creating a new
	// user, CreatedAt will be set to the current time.
	CreatedAt *time.Time `db:"created_at"`

	// The Name and Email fields are optional as we are not sure if they will be
	// provided by all OIDC providers. Because they are nice human-readable
	// identifiers they should be used for display purposes when available.
	Email *string `db:"email"`
	Name  *string `db:"name"`

	// The JWT Iss and Sub fields are optional SECRET values that uniquely
	// identify the user in the OIDC provider. They are optional values to allow
	// for users that are not authenticated via OIDC (e.g. system users).
	JWTIss *string `db:"jwt_iss"`
	JWTSub *string `db:"jwt_sub"`
}
