package storage

import (
	"context"
)

type UpdatePackageLocationLocalActivityParams struct {
	AIPID    string
	Location string
}

func UpdatePackageLocationLocalActivity(ctx context.Context, storagesvc Service, params *UpdatePackageLocationLocalActivityParams) error {
	return storagesvc.UpdatePackageLocation(ctx, params.Location, params.AIPID)
}

type UpdatePackageStatusLocalActivityParams struct {
	AIPID    string
	Location string
}

func UpdatePackageStatusLocalActivity(ctx context.Context, storagesvc Service, params *UpdatePackageStatusLocalActivityParams) error {
	return storagesvc.UpdatePackageStatus(ctx, StatusStored, params.AIPID)
}

type DeleteFromLocationLocalActivityParams struct {
	ObjectKey string
}

func DeleteFromLocationLocalActivity(ctx context.Context, storagesvc Service, params *DeleteFromLocationLocalActivityParams) error {
	return storagesvc.Bucket().Delete(ctx, params.ObjectKey)
}
