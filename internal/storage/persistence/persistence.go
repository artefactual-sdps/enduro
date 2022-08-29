package persistence

import (
	"context"

	"github.com/google/uuid"

	goastorage "github.com/artefactual-sdps/enduro/internal/api/gen/storage"
	"github.com/artefactual-sdps/enduro/internal/storage/types"
)

type Storage interface {
	// Package.
	CreatePackage(ctx context.Context, pkg *goastorage.StoragePackage) (*goastorage.StoredStoragePackage, error)
	ListPackages(ctx context.Context) ([]*goastorage.StoredStoragePackage, error)
	ReadPackage(ctx context.Context, AIPID uuid.UUID) (*goastorage.StoredStoragePackage, error)
	UpdatePackageStatus(ctx context.Context, status types.PackageStatus, AIPID uuid.UUID) error
	UpdatePackageLocationID(ctx context.Context, locationID uuid.UUID, aipID uuid.UUID) error

	// Location.
	CreateLocation(ctx context.Context, location *goastorage.Location, config *types.LocationConfig) (*goastorage.StoredLocation, error)
	ListLocations(ctx context.Context) (goastorage.StoredLocationCollection, error)
	ReadLocation(ctx context.Context, locationID uuid.UUID) (*goastorage.StoredLocation, error)
	LocationPackages(ctx context.Context, locationID uuid.UUID) (goastorage.StoredStoragePackageCollection, error)
}
