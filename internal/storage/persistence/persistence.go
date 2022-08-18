package persistence

import (
	"context"

	"github.com/google/uuid"

	goastorage "github.com/artefactual-sdps/enduro/internal/api/gen/storage"
	"github.com/artefactual-sdps/enduro/internal/storage/types"
)

type Storage interface {
	// Package.
	CreatePackage(ctx context.Context, name string, AIPID uuid.UUID, objectKey uuid.UUID) (*goastorage.StoredStoragePackage, error)
	ListPackages(ctx context.Context) ([]*goastorage.StoredStoragePackage, error)
	ReadPackage(ctx context.Context, AIPID uuid.UUID) (*goastorage.StoredStoragePackage, error)
	UpdatePackageStatus(ctx context.Context, status types.PackageStatus, AIPID uuid.UUID) error
	UpdatePackageLocation(ctx context.Context, location string, aipID uuid.UUID) error

	// Location.
	CreateLocation(ctx context.Context, name string, description *string, source types.LocationSource, purpose types.LocationPurpose, uuid uuid.UUID, config *types.LocationConfig) (*goastorage.StoredLocation, error)
	ListLocations(ctx context.Context) (goastorage.StoredLocationCollection, error)
	ReadLocation(ctx context.Context, uuid uuid.UUID) (*goastorage.StoredLocation, error)
}
