package storage

import (
	"context"

	"github.com/google/uuid"

	"github.com/artefactual-sdps/enduro/internal/storage/types"
)

type UpdateAIPLocationLocalActivityParams struct {
	AIPID      uuid.UUID
	LocationID uuid.UUID
}

func UpdateAIPLocationLocalActivity(
	ctx context.Context,
	storagesvc Service,
	params *UpdateAIPLocationLocalActivityParams,
) error {
	return storagesvc.UpdateAipLocationID(ctx, params.AIPID, params.LocationID)
}

type UpdateAIPStatusLocalActivityParams struct {
	AIPID  uuid.UUID
	Status types.AIPStatus
}

func UpdateAIPStatusLocalActivity(
	ctx context.Context,
	storagesvc Service,
	params *UpdateAIPStatusLocalActivityParams,
) error {
	return storagesvc.UpdateAipStatus(ctx, params.AIPID, params.Status)
}

type DeleteFromLocationLocalActivityParams struct {
	AIPID uuid.UUID
}

func DeleteFromLocationLocalActivity(
	ctx context.Context,
	storagesvc Service,
	params *DeleteFromLocationLocalActivityParams,
) error {
	return storagesvc.DeleteAip(ctx, params.AIPID)
}
