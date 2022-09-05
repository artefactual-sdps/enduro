package persistence

import (
	"context"

	"github.com/google/uuid"

	goastorage "github.com/artefactual-sdps/enduro/internal/api/gen/storage"
	"github.com/artefactual-sdps/enduro/internal/storage/types"
)

type Storage interface {
	// Package.
	CreatePackage(ctx context.Context, pkg *goastorage.Package) (*goastorage.Package, error)
	ListPackages(ctx context.Context) ([]*goastorage.Package, error)
	ReadPackage(ctx context.Context, aipID uuid.UUID) (*goastorage.Package, error)
	UpdatePackageStatus(ctx context.Context, aipID uuid.UUID, status types.PackageStatus) error
	UpdatePackageLocationID(ctx context.Context, aipID, locationID uuid.UUID) error

	// Location.
	CreateLocation(ctx context.Context, location *goastorage.Location, config *types.LocationConfig) (*goastorage.Location, error)
	ListLocations(ctx context.Context) (goastorage.LocationCollection, error)
	ReadLocation(ctx context.Context, locationID uuid.UUID) (*goastorage.Location, error)
	LocationPackages(ctx context.Context, locationID uuid.UUID) (goastorage.PackageCollection, error)
}
