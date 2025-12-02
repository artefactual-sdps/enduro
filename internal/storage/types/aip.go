package types

import (
	"time"

	"github.com/google/uuid"

	"github.com/artefactual-sdps/enduro/internal/storage/enums"
)

type AIP struct {
	// UUID is the unique identifier of the AIP.
	UUID uuid.UUID

	// Name of the AIP.
	Name string

	// CreatedAt is the timestamp when the AIP was created.
	CreatedAt time.Time

	// ObjectKey is the storage key of the AIP.
	ObjectKey uuid.UUID

	// Status of the AIP.
	Status enums.AIPStatus

	// LocationUUID is the UUID of the location where the AIP is stored.
	LocationUUID *uuid.UUID

	// DeletionReportKey is the object store key for the AIP deletion report.
	DeletionReportKey *string
}
