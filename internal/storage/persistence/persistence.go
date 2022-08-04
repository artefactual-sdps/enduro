package persistence

import (
	"context"

	"github.com/google/uuid"

	goastorage "github.com/artefactual-sdps/enduro/internal/api/gen/storage"
	"github.com/artefactual-sdps/enduro/internal/storage/status"
)

type Storage interface {
	CreatePackage(ctx context.Context, name string, AIPID uuid.UUID, objectKey uuid.UUID) (*goastorage.StoredStoragePackage, error)
	ListPackages(ctx context.Context) ([]*goastorage.StoredStoragePackage, error)
	ReadPackage(ctx context.Context, AIPID uuid.UUID) (*goastorage.StoredStoragePackage, error)
	UpdatePackageStatus(ctx context.Context, status status.PackageStatus, AIPID uuid.UUID) error
	UpdatePackageLocation(ctx context.Context, location string, aipID uuid.UUID) error
}
