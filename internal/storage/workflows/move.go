package workflows

import (
	"time"

	temporalsdk_temporal "go.temporal.io/sdk/temporal"
	temporalsdk_workflow "go.temporal.io/sdk/workflow"

	"github.com/artefactual-labs/enduro/internal/storage"
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
	// Copy package from internal processing bucket to permanent location bucket
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
			AIPID:    req.AIPID,
			Location: req.Location,
		}).Get(activityOpts, nil)
		if err != nil {
			return err
		}
	}

	// TODO: should we delete the package from the internal aips bucket here?

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
			AIPID:    req.AIPID,
			Location: req.Location,
		}).Get(activityOpts, nil)
		if err != nil {
			return err
		}
	}

	// Update package status
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
		err := temporalsdk_workflow.ExecuteLocalActivity(activityOpts, storage.UpdatePackageStatusLocalActivity, w.storagesvc, &storage.UpdatePackageStatusLocalActivityParams{
			AIPID:    req.AIPID,
			Location: req.Location,
		}).Get(activityOpts, nil)
		if err != nil {
			return err
		}
	}

	return nil
}
