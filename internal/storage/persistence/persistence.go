package persistence

import (
	"context"

	"github.com/google/uuid"

	goastorage "github.com/artefactual-sdps/enduro/internal/api/gen/storage"
	"github.com/artefactual-sdps/enduro/internal/storage/purpose"
	"github.com/artefactual-sdps/enduro/internal/storage/source"
	"github.com/artefactual-sdps/enduro/internal/storage/status"
)

type Storage interface {
	// Package.
	CreatePackage(ctx context.Context, name string, AIPID uuid.UUID, objectKey uuid.UUID) (*goastorage.StoredStoragePackage, error)
	ListPackages(ctx context.Context) ([]*goastorage.StoredStoragePackage, error)
	ReadPackage(ctx context.Context, AIPID uuid.UUID) (*goastorage.StoredStoragePackage, error)
	UpdatePackageStatus(ctx context.Context, status status.PackageStatus, AIPID uuid.UUID) error
	UpdatePackageLocation(ctx context.Context, location string, aipID uuid.UUID) error

	// Location.
	CreateLocation(ctx context.Context, name string, description *string, source source.LocationSource, purpose purpose.LocationPurpose, uuid uuid.UUID) (*goastorage.StoredLocation, error)
	ListLocations(ctx context.Context) (goastorage.StoredLocationCollection, error)
	ReadLocation(ctx context.Context, uuid uuid.UUID) (*goastorage.StoredLocation, error)
}
