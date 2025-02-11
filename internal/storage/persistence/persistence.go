package persistence

import (
	"context"

	"github.com/google/uuid"

	goastorage "github.com/artefactual-sdps/enduro/internal/api/gen/storage"
	"github.com/artefactual-sdps/enduro/internal/storage/types"
)

type Storage interface {
	// AIP.
	CreateAIP(ctx context.Context, pkg *goastorage.Package) (*goastorage.Package, error)
	ListAIPs(ctx context.Context) (goastorage.PackageCollection, error)
	ReadAIP(ctx context.Context, aipID uuid.UUID) (*goastorage.Package, error)
	UpdateAIPStatus(ctx context.Context, aipID uuid.UUID, status types.AIPStatus) error
	UpdateAIPLocationID(ctx context.Context, aipID, locationID uuid.UUID) error

	// Location.
	CreateLocation(
		ctx context.Context,
		location *goastorage.Location,
		config *types.LocationConfig,
	) (*goastorage.Location, error)
	ListLocations(ctx context.Context) (goastorage.LocationCollection, error)
	ReadLocation(ctx context.Context, locationID uuid.UUID) (*goastorage.Location, error)
	LocationAIPs(ctx context.Context, locationID uuid.UUID) (goastorage.PackageCollection, error)
}
