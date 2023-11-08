package sfa

import (
	"context"
	"errors"
	"fmt"
	"time"

	temporalsdk_activity "go.temporal.io/sdk/activity"
	temporalsdk_temporal "go.temporal.io/sdk/temporal"
	temporalsdk_worker "go.temporal.io/sdk/worker"
	temporalsdk_workflow "go.temporal.io/sdk/workflow"
	"gocloud.dev/blob"

	"github.com/artefactual-sdps/enduro/internal/config"
	"github.com/artefactual-sdps/enduro/internal/sfa/activities"
	"github.com/artefactual-sdps/enduro/internal/storage"
)

func RegisterActivities(ctx context.Context, w temporalsdk_worker.Worker, cfg config.Configuration) error {
	// Set-up failed transfers bucket.
	var ft *blob.Bucket
	{
		fl, err := storage.NewInternalLocation(&cfg.FailedTransfers)
		if err != nil {
			return fmt.Errorf("error setting up failed transfers location: %v", err)
		}
		ft, err = fl.OpenBucket(ctx)
		if err != nil {
			return fmt.Errorf("error getting failed transfers bucket: %v", err)
		}
	}

	// Set-up failed SIPs bucket.
	var fs *blob.Bucket
	{
		fl, err := storage.NewInternalLocation(&cfg.FailedSIPs)
		if err != nil {
			return fmt.Errorf("error setting up failed SIPs location: %v", err)
		}
		fs, err = fl.OpenBucket(ctx)
		if err != nil {
			return fmt.Errorf("error getting failed SIPs bucket: %v", err)
		}
	}

	// Register activities.
	w.RegisterActivityWithOptions(activities.NewExtractPackage().Execute, temporalsdk_activity.RegisterOptions{Name: activities.ExtractPackageName})
	w.RegisterActivityWithOptions(activities.NewCheckSipStructure().Execute, temporalsdk_activity.RegisterOptions{Name: activities.CheckSipStructureName})
	w.RegisterActivityWithOptions(activities.NewAllowedFileFormatsActivity().Execute, temporalsdk_activity.RegisterOptions{Name: activities.AllowedFileFormatsName})
	w.RegisterActivityWithOptions(activities.NewMetadataValidationActivity().Execute, temporalsdk_activity.RegisterOptions{Name: activities.MetadataValidationName})
	w.RegisterActivityWithOptions(activities.NewSipCreationActivity().Execute, temporalsdk_activity.RegisterOptions{Name: activities.SipCreationName})
	w.RegisterActivityWithOptions(activities.NewSendToFailedBuckeActivity(ft, fs).Execute, temporalsdk_activity.RegisterOptions{Name: activities.SendToFailedBucketName})
	w.RegisterActivityWithOptions(activities.NewRemovePaths().Execute, temporalsdk_activity.RegisterOptions{Name: activities.RemovePathsName})

	return nil
}

func ExecuteActivities(ctx temporalsdk_workflow.Context, path string, key string) (p string, e error) {
	var removePaths []string

	defer func() {
		var result activities.RemovePathsResult
		err := temporalsdk_workflow.ExecuteActivity(withLocalActOpts(ctx), activities.RemovePathsName, &activities.RemovePathsParams{
			Paths: removePaths,
		}).Get(ctx, &result)
		e = errors.Join(e, err)
	}()

	e = func() error {
		// Extract package.
		var extractPackageRes activities.ExtractPackageResult
		err := temporalsdk_workflow.ExecuteActivity(withLocalActOpts(ctx), activities.ExtractPackageName, &activities.ExtractPackageParams{
			Path: path,
			Key:  key,
		}).Get(ctx, &extractPackageRes)
		if err != nil {
			return err
		}

		removePaths = append(removePaths, path, extractPackageRes.Path)

		// Validate SIP structure.
		var checkStructureRes activities.CheckSipStructureResult
		err = temporalsdk_workflow.ExecuteActivity(withLocalActOpts(ctx), activities.CheckSipStructureName, &activities.CheckSipStructureParams{
			SipPath: extractPackageRes.Path,
		}).Get(ctx, &checkStructureRes)
		if err != nil {
			return err
		}

		// Check allowed file formats.
		var allowedFileFormats activities.AllowedFileFormatsResult
		err = temporalsdk_workflow.ExecuteActivity(withLocalActOpts(ctx), activities.AllowedFileFormatsName, &activities.AllowedFileFormatsParams{
			SipPath: extractPackageRes.Path,
		}).Get(ctx, &allowedFileFormats)
		if err != nil {
			return err
		}

		// Validate metadata.xsd.
		var metadataValidation activities.MetadataValidationResult
		err = temporalsdk_workflow.ExecuteActivity(withLocalActOpts(ctx), activities.MetadataValidationName, &activities.MetadataValidationParams{
			SipPath: extractPackageRes.Path,
		}).Get(ctx, &metadataValidation)
		if err != nil {
			return err
		}

		// Repackage SFA Sip into a Bag.
		var sipCreation activities.SipCreationResult
		err = temporalsdk_workflow.ExecuteActivity(withLocalActOpts(ctx), activities.SipCreationName, &activities.SipCreationParams{
			SipPath: extractPackageRes.Path,
		}).Get(ctx, &sipCreation)
		if err != nil {
			return err
		}

		// We do this so that the code above only stops when a non-bussines error is found.
		if !allowedFileFormats.Ok {
			return activities.ErrIlegalFileFormat
		}
		if !checkStructureRes.Ok {
			return activities.ErrInvaliSipStructure
		}

		p = sipCreation.NewSipPath

		return nil
	}()
	if e != nil {
		err := temporalsdk_workflow.ExecuteActivity(withUploadActOpts(ctx), activities.SendToFailedBucketName, &activities.SendToFailedBucketParams{
			FailureType: activities.FailureTransfer,
			Path:        path,
			Key:         key,
		}).Get(ctx, nil)
		e = errors.Join(e, err)

		return p, e
	}

	return p, e
}

func SendToFailedSIPs(ctx temporalsdk_workflow.Context, path string, key string) error {
	return temporalsdk_workflow.ExecuteActivity(withUploadActOpts(ctx), activities.SendToFailedBucketName, &activities.SendToFailedBucketParams{
		FailureType: activities.FailureSIP,
		Path:        path,
		Key:         key,
	}).Get(ctx, nil)
}

func withUploadActOpts(ctx temporalsdk_workflow.Context) temporalsdk_workflow.Context {
	return temporalsdk_workflow.WithActivityOptions(ctx, temporalsdk_workflow.ActivityOptions{
		StartToCloseTimeout: time.Hour * 24,
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
}

func withLocalActOpts(ctx temporalsdk_workflow.Context) temporalsdk_workflow.Context {
	return temporalsdk_workflow.WithActivityOptions(ctx, temporalsdk_workflow.ActivityOptions{
		ScheduleToCloseTimeout: 5 * time.Minute,
		RetryPolicy: &temporalsdk_temporal.RetryPolicy{
			MaximumAttempts: 1,
		},
	})
}
