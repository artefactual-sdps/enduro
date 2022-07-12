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

type saveLocationMovePreservationActionLocalActivityParams struct {
	PackageID   uint
	Location    string
	WorkflowID  string
	Status      package_.PreservationTaskStatus
	StartedAt   time.Time
	CompletedAt time.Time
}

func saveLocationMovePreservationActionLocalActivity(ctx context.Context, pkgsvc package_.Service, params *saveLocationMovePreservationActionLocalActivityParams) error {
	paID, err := createPreservationActionLocalActivity(ctx, pkgsvc, &createPreservationActionLocalActivityParams{
		Name:        "Move package", // XXX: move to a translatable constant?
		WorkflowID:  params.WorkflowID,
		StartedAt:   params.StartedAt,
		CompletedAt: params.CompletedAt,
		PackageID:   params.PackageID,
	})
	if err != nil {
		return err
	}

	pt := package_.PreservationTask{
		TaskID:               uuid.NewString(),
		Name:                 fmt.Sprintf("Moved to %s", params.Location),
		Status:               params.Status,
		PreservationActionID: paID,
	}
	pt.StartedAt.Time = params.StartedAt
	pt.CompletedAt.Time = params.CompletedAt

	return pkgsvc.CreatePreservationTask(ctx, &pt)
}

type createPreservationActionLocalActivityParams struct {
	Name        string
	WorkflowID  string
	StartedAt   time.Time
	CompletedAt time.Time
	PackageID   uint
}

func createPreservationActionLocalActivity(ctx context.Context, pkgsvc package_.Service, params *createPreservationActionLocalActivityParams) (uint, error) {
	pa := package_.PreservationAction{
		Name:       params.Name,
		WorkflowID: params.WorkflowID,
		PackageID:  params.PackageID,
	}
	pa.StartedAt.Time = params.StartedAt
	pa.CompletedAt.Time = params.CompletedAt

	if err := pkgsvc.CreatePreservationAction(ctx, &pa); err != nil {
		return 0, err
	}

	return pa.ID, nil
}

func completePreservationActionLocalActivity(ctx context.Context, pkgsvc package_.Service, paID uint, completedAt time.Time) error {
	return pkgsvc.CompletePreservationAction(ctx, paID, completedAt)
}
