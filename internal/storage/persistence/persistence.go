package persistence

import (
	"context"

	"github.com/google/uuid"

	goastorage "github.com/artefactual-sdps/enduro/internal/api/gen/storage"
	"github.com/artefactual-sdps/enduro/internal/storage/enums"
	"github.com/artefactual-sdps/enduro/internal/storage/types"
)

type Storage interface {
	// AIP.
	CreateAIP(ctx context.Context, aip *goastorage.AIP) (*goastorage.AIP, error)
	ListAIPs(ctx context.Context) (goastorage.AIPCollection, error)
	ReadAIP(ctx context.Context, aipID uuid.UUID) (*goastorage.AIP, error)
	UpdateAIPStatus(ctx context.Context, aipID uuid.UUID, status enums.AIPStatus) error
	UpdateAIPLocationID(ctx context.Context, aipID, locationID uuid.UUID) error

	// Location.
	CreateLocation(
		ctx context.Context,
		location *goastorage.Location,
		config *types.LocationConfig,
	) (*goastorage.Location, error)
	ListLocations(ctx context.Context) (goastorage.LocationCollection, error)
	ReadLocation(ctx context.Context, locationID uuid.UUID) (*goastorage.Location, error)
	LocationAIPs(ctx context.Context, locationID uuid.UUID) (goastorage.AIPCollection, error)
}
