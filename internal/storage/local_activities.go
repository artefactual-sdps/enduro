package storage

import (
	"context"

	"github.com/google/uuid"

	"github.com/artefactual-sdps/enduro/internal/storage/types"
)

type UpdatePackageLocationLocalActivityParams struct {
	AIPID      string
	LocationID uuid.UUID
}

func UpdatePackageLocationLocalActivity(ctx context.Context, storagesvc Service, params *UpdatePackageLocationLocalActivityParams) error {
	return storagesvc.UpdatePackageLocationID(ctx, params.LocationID, params.AIPID)
}

type UpdatePackageStatusLocalActivityParams struct {
	AIPID  string
	Status types.PackageStatus
}

func UpdatePackageStatusLocalActivity(ctx context.Context, storagesvc Service, params *UpdatePackageStatusLocalActivityParams) error {
	return storagesvc.UpdatePackageStatus(ctx, params.Status, params.AIPID)
}

type DeleteFromLocationLocalActivityParams struct {
	AIPID string
}

func DeleteFromLocationLocalActivity(ctx context.Context, storagesvc Service, params *DeleteFromLocationLocalActivityParams) error {
	return storagesvc.Delete(ctx, params.AIPID)
}
