package storage

import (
	"context"
)

type UpdatePackageLocationLocalActivityParams struct {
	AIPID    string
	Location string
}

type UpdatePackageStatusLocalActivityParams struct {
	AIPID    string
	Location string
}

func UpdatePackageLocationLocalActivity(ctx context.Context, storagesvc Service, params *UpdatePackageLocationLocalActivityParams) error {
	return storagesvc.UpdatePackageLocation(ctx, params.Location, params.AIPID)
}

func UpdatePackageStatusLocalActivity(ctx context.Context, storagesvc Service, params *UpdatePackageStatusLocalActivityParams) error {
	return storagesvc.UpdatePackageStatus(ctx, StatusStored, params.AIPID)
}
