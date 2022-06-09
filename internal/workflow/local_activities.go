package workflow

import (
	"context"
	"time"

	"github.com/go-logr/logr"
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

//nolint:deadcode,unused
func setStatusLocalActivity(ctx context.Context, pkgsvc package_.Service, pkgID uint, status package_.Status) error {
	return pkgsvc.SetStatus(ctx, pkgID, status)
}
