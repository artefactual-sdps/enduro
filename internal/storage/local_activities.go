package storage

import (
	"context"

	"github.com/google/uuid"

	"github.com/artefactual-sdps/enduro/internal/storage/types"
)

type UpdatePackageLocationLocalActivityParams struct {
	AIPID      uuid.UUID
	LocationID uuid.UUID
}

func UpdatePackageLocationLocalActivity(
	ctx context.Context,
	storagesvc Service,
	params *UpdatePackageLocationLocalActivityParams,
) error {
	return storagesvc.UpdatePackageLocationID(ctx, params.AIPID, params.LocationID)
}

type UpdatePackageStatusLocalActivityParams struct {
	AIPID  uuid.UUID
	Status types.PackageStatus
}

func UpdatePackageStatusLocalActivity(
	ctx context.Context,
	storagesvc Service,
	params *UpdatePackageStatusLocalActivityParams,
) error {
	return storagesvc.UpdatePackageStatus(ctx, params.AIPID, params.Status)
}

type DeleteFromLocationLocalActivityParams struct {
	AIPID uuid.UUID
}

func DeleteFromLocationLocalActivity(
	ctx context.Context,
	storagesvc Service,
	params *DeleteFromLocationLocalActivityParams,
) error {
	return storagesvc.Delete(ctx, params.AIPID)
}
