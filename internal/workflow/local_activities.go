package workflow

import (
	"context"
	"fmt"
	"time"

	"github.com/go-logr/logr"
	"github.com/google/uuid"
	temporalsdk_activity "go.temporal.io/sdk/activity"

	"github.com/artefactual-labs/enduro/internal/package_"
)

type createPackageLocalActivityParams struct {
	Key    string
	Status package_.Status
}

func createPackageLocalActivity(ctx context.Context, logger logr.Logger, pkgsvc package_.Service, params *createPackageLocalActivityParams) (uint, error) {
	info := temporalsdk_activity.GetInfo(ctx)

	col := &package_.Package{
		Name:       params.Key,
		WorkflowID: info.WorkflowExecution.ID,
		RunID:      info.WorkflowExecution.RunID,
		Status:     params.Status,
	}

	if err := pkgsvc.Create(ctx, col); err != nil {
		logger.Error(err, "Error creating package")
		return 0, err
	}

	return col.ID, nil
}

type updatePackageLocalActivityParams struct {
	PackageID uint
	Key       string
	SIPID     string
	StoredAt  time.Time
	Status    package_.Status
}

func updatePackageLocalActivity(ctx context.Context, logger logr.Logger, pkgsvc package_.Service, params *updatePackageLocalActivityParams) error {
	info := temporalsdk_activity.GetInfo(ctx)

	err := pkgsvc.UpdateWorkflowStatus(
		ctx,
		params.PackageID,
		params.Key,
		info.WorkflowExecution.ID,
		info.WorkflowExecution.RunID,
		params.SIPID,
		params.Status,
		params.StoredAt,
	)
	if err != nil {
		logger.Error(err, "Error updating package")
		return err
	}

	return nil
}

func setStatusInProgressLocalActivity(ctx context.Context, pkgsvc package_.Service, pkgID uint, startedAt time.Time) error {
	return pkgsvc.SetStatusInProgress(ctx, pkgID, startedAt)
}

func setStatusLocalActivity(ctx context.Context, pkgsvc package_.Service, pkgID uint, status package_.Status) error {
	return pkgsvc.SetStatus(ctx, pkgID, status)
}

func setLocationLocalActivity(ctx context.Context, pkgsvc package_.Service, pkgID uint, location string) error {
	return pkgsvc.SetLocation(ctx, pkgID, location)
}

func saveLocationMovePreservationActionLocalActivity(ctx context.Context, pkgsvc package_.Service, pkgID uint, location string, status package_.PreservationActionStatus, startedAt, completedAt time.Time) error {
	pa := package_.PreservationAction{
		ActionID:  uuid.NewString(),
		Name:      fmt.Sprintf("Moved to %s", location),
		Status:    status,
		PackageID: pkgID,
	}
	pa.StartedAt.Time = startedAt
	pa.CompletedAt.Time = completedAt
	return pkgsvc.CreatePreservationAction(ctx, &pa)
}
