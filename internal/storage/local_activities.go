package storage

import (
	"context"

	"github.com/artefactual-sdps/enduro/internal/storage/status"
)

type UpdatePackageLocationLocalActivityParams struct {
	AIPID    string
	Location string
}

func UpdatePackageLocationLocalActivity(ctx context.Context, storagesvc Service, params *UpdatePackageLocationLocalActivityParams) error {
	return storagesvc.UpdatePackageLocation(ctx, params.Location, params.AIPID)
}

type UpdatePackageStatusLocalActivityParams struct {
	AIPID  string
	Status status.PackageStatus
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
