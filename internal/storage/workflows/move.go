package workflows

import (
	"time"

	temporalsdk_temporal "go.temporal.io/sdk/temporal"
	temporalsdk_workflow "go.temporal.io/sdk/workflow"

	"github.com/artefactual-sdps/enduro/internal/storage"
	"github.com/artefactual-sdps/enduro/internal/storage/types"
)

type StorageMoveWorkflow struct {
	storagesvc storage.Service
}

func NewStorageMoveWorkflow(storagesvc storage.Service) *StorageMoveWorkflow {
	return &StorageMoveWorkflow{
		storagesvc: storagesvc,
	}
}

func (w *StorageMoveWorkflow) Execute(ctx temporalsdk_workflow.Context, req storage.StorageMoveWorkflowRequest) error {
	// Set package status to moving.
	{
		if err := w.updatePackageStatus(ctx, types.StatusMoving, req.AIPID); err != nil {
			return err
		}
	}

	// Copy package from its current bucket to a new permanent location bucket
	{
		activityOpts := temporalsdk_workflow.WithActivityOptions(ctx, temporalsdk_workflow.ActivityOptions{
			StartToCloseTimeout: time.Hour * 2,
			RetryPolicy: &temporalsdk_temporal.RetryPolicy{
				InitialInterval:    time.Second,
				BackoffCoefficient: 2,
				MaximumInterval:    time.Minute * 10,
				MaximumAttempts:    5,
				NonRetryableErrorTypes: []string{
					"TemporalTimeout:StartToClose",
				},
			},
		})
		err := temporalsdk_workflow.ExecuteActivity(activityOpts, storage.CopyToPermanentLocationActivityName, &storage.CopyToPermanentLocationActivityParams{
			AIPID:      req.AIPID,
			LocationID: req.LocationID,
		}).Get(activityOpts, nil)
		if err != nil {
			return err
		}
	}

	// Delete package from its current bucket
	{

		activityOpts := temporalsdk_workflow.WithLocalActivityOptions(ctx, temporalsdk_workflow.LocalActivityOptions{
			ScheduleToCloseTimeout: 5 * time.Second,
			RetryPolicy: &temporalsdk_temporal.RetryPolicy{
				InitialInterval:    time.Second,
				BackoffCoefficient: 2,
				MaximumInterval:    time.Minute,
				MaximumAttempts:    3,
			},
		})
		err := temporalsdk_workflow.ExecuteLocalActivity(activityOpts, storage.DeleteFromLocationLocalActivity, w.storagesvc, &storage.DeleteFromLocationLocalActivityParams{
			AIPID: req.AIPID,
		}).Get(activityOpts, nil)
		if err != nil {
			return err
		}
	}

	// Update package location
	{
		activityOpts := temporalsdk_workflow.WithLocalActivityOptions(ctx, temporalsdk_workflow.LocalActivityOptions{
			ScheduleToCloseTimeout: 5 * time.Second,
			RetryPolicy: &temporalsdk_temporal.RetryPolicy{
				InitialInterval:    time.Second,
				BackoffCoefficient: 2,
				MaximumInterval:    time.Minute,
				MaximumAttempts:    3,
			},
		})
		err := temporalsdk_workflow.ExecuteLocalActivity(activityOpts, storage.UpdatePackageLocationLocalActivity, w.storagesvc, &storage.UpdatePackageLocationLocalActivityParams{
			AIPID:      req.AIPID,
			LocationID: req.LocationID,
		}).Get(activityOpts, nil)
		if err != nil {
			return err
		}
	}

	// Set package status to stored.
	{
		if err := w.updatePackageStatus(ctx, types.StatusStored, req.AIPID); err != nil {
			return err
		}
	}

	return nil
}

func (w *StorageMoveWorkflow) updatePackageStatus(ctx temporalsdk_workflow.Context, st types.PackageStatus, AIPID string) error {
	activityOpts := temporalsdk_workflow.WithLocalActivityOptions(ctx, temporalsdk_workflow.LocalActivityOptions{
		ScheduleToCloseTimeout: 5 * time.Second,
		RetryPolicy: &temporalsdk_temporal.RetryPolicy{
			InitialInterval:    time.Second,
			BackoffCoefficient: 2,
			MaximumInterval:    time.Minute,
			MaximumAttempts:    3,
		},
	})

	params := &storage.UpdatePackageStatusLocalActivityParams{
		AIPID:  AIPID,
		Status: st,
	}

	return temporalsdk_workflow.ExecuteLocalActivity(activityOpts, storage.UpdatePackageStatusLocalActivity, w.storagesvc, params).Get(activityOpts, nil)
}
