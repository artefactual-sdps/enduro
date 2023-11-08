package sfa

import (
	"github.com/stretchr/testify/mock"
	temporalsdk_activity "go.temporal.io/sdk/activity"
	temporalsdk_testsuite "go.temporal.io/sdk/testsuite"
	"gocloud.dev/blob/memblob"

	"github.com/artefactual-sdps/enduro/internal/sfa/activities"
)

func RegisterWorkflowTestActivities(env *temporalsdk_testsuite.TestWorkflowEnvironment) {
	env.RegisterActivityWithOptions(activities.NewExtractPackage().Execute, temporalsdk_activity.RegisterOptions{Name: activities.ExtractPackageName})
	env.RegisterActivityWithOptions(activities.NewCheckSipStructure().Execute, temporalsdk_activity.RegisterOptions{Name: activities.CheckSipStructureName})
	env.RegisterActivityWithOptions(activities.NewAllowedFileFormatsActivity().Execute, temporalsdk_activity.RegisterOptions{Name: activities.AllowedFileFormatsName})
	env.RegisterActivityWithOptions(activities.NewMetadataValidationActivity().Execute, temporalsdk_activity.RegisterOptions{Name: activities.MetadataValidationName})
	env.RegisterActivityWithOptions(activities.NewSipCreationActivity().Execute, temporalsdk_activity.RegisterOptions{Name: activities.SipCreationName})
	env.RegisterActivityWithOptions(activities.NewSendToFailedBuckeActivity(memblob.OpenBucket(nil), memblob.OpenBucket(nil)).Execute, temporalsdk_activity.RegisterOptions{Name: activities.SendToFailedBucketName})
	env.RegisterActivityWithOptions(activities.NewRemovePaths().Execute, temporalsdk_activity.RegisterOptions{Name: activities.RemovePathsName})
}

func AddWorkflowTestExpectations(env *temporalsdk_testsuite.TestWorkflowEnvironment) {
	sessionCtx := mock.AnythingOfType("*context.timerCtx")
	env.OnActivity(activities.ExtractPackageName, sessionCtx, &activities.ExtractPackageParams{Key: "transfer.tgz"}).Return(&activities.ExtractPackageResult{}, nil).Once()
	env.OnActivity(activities.CheckSipStructureName, sessionCtx, &activities.CheckSipStructureParams{}).Return(&activities.CheckSipStructureResult{Ok: true}, nil).Once()
	env.OnActivity(activities.AllowedFileFormatsName, sessionCtx, &activities.AllowedFileFormatsParams{}).Return(&activities.AllowedFileFormatsResult{Ok: true}, nil).Once()
	env.OnActivity(activities.MetadataValidationName, sessionCtx, &activities.MetadataValidationParams{}).Return(&activities.MetadataValidationResult{}, nil).Once()
	env.OnActivity(activities.SipCreationName, sessionCtx, &activities.SipCreationParams{}).Return(&activities.SipCreationResult{}, nil).Once()
	env.OnActivity(activities.SendToFailedBucketName, sessionCtx, &activities.SendToFailedBucketParams{}).Return(&activities.SendToFailedBucketResult{}, nil).Maybe()
	env.OnActivity(activities.RemovePathsName, sessionCtx, &activities.RemovePathsParams{Paths: []string{"", ""}}).Return(&activities.RemovePathsResult{}, nil).Once()
}
