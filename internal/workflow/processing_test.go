package workflow

import (
	"testing"
	"time"

	"github.com/go-logr/logr"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	temporalsdk_activity "go.temporal.io/sdk/activity"
	temporalsdk_testsuite "go.temporal.io/sdk/testsuite"
	temporalsdk_worker "go.temporal.io/sdk/worker"

	"github.com/artefactual-sdps/enduro/internal/a3m"
	"github.com/artefactual-sdps/enduro/internal/package_"
	packagefake "github.com/artefactual-sdps/enduro/internal/package_/fake"
	sdps_activities "github.com/artefactual-sdps/enduro/internal/sdps/activities"
	"github.com/artefactual-sdps/enduro/internal/validation"
	watcherfake "github.com/artefactual-sdps/enduro/internal/watcher/fake"
	"github.com/artefactual-sdps/enduro/internal/workflow/activities"
)

type ProcessingWorkflowTestSuite struct {
	suite.Suite
	temporalsdk_testsuite.WorkflowTestSuite

	env *temporalsdk_testsuite.TestWorkflowEnvironment

	// Each test registers the workflow with a different name to avoid dups.
	workflow *ProcessingWorkflow
}

func (s *ProcessingWorkflowTestSuite) SetupTest() {
	s.env = s.NewTestWorkflowEnvironment()
	s.env.SetWorkerOptions(temporalsdk_worker.Options{EnableSessionWorker: true})

	ctrl := gomock.NewController(s.T())
	logger := logr.Discard()
	pkgsvc := packagefake.NewMockService(ctrl)
	wsvc := watcherfake.NewMockService(ctrl)

	s.env.RegisterActivityWithOptions(activities.NewDownloadActivity(wsvc).Execute, temporalsdk_activity.RegisterOptions{Name: activities.DownloadActivityName})
	s.env.RegisterActivityWithOptions(activities.NewBundleActivity(wsvc).Execute, temporalsdk_activity.RegisterOptions{Name: activities.BundleActivityName})
	s.env.RegisterActivityWithOptions(a3m.NewCreateAIPActivity(logger, &a3m.Config{}, pkgsvc).Execute, temporalsdk_activity.RegisterOptions{Name: a3m.CreateAIPActivityName})
	s.env.RegisterActivityWithOptions(activities.NewUploadActivity(nil).Execute, temporalsdk_activity.RegisterOptions{Name: activities.UploadActivityName})
	s.env.RegisterActivityWithOptions(activities.NewMoveToPermanentStorageActivity(nil).Execute, temporalsdk_activity.RegisterOptions{Name: activities.MoveToPermanentStorageActivityName})
	s.env.RegisterActivityWithOptions(activities.NewPollMoveToPermanentStorageActivity(nil).Execute, temporalsdk_activity.RegisterOptions{Name: activities.PollMoveToPermanentStorageActivityName})
	s.env.RegisterActivityWithOptions(sdps_activities.NewValidatePackageActivity().Execute, temporalsdk_activity.RegisterOptions{Name: sdps_activities.ValidatePackageActivityName})
	s.env.RegisterActivityWithOptions(activities.NewValidateTransferActivity().Execute, temporalsdk_activity.RegisterOptions{Name: activities.ValidateTransferActivityName})
	s.env.RegisterActivityWithOptions(sdps_activities.NewIndexActivity(logger, nil).Execute, temporalsdk_activity.RegisterOptions{Name: sdps_activities.IndexActivityName})
	s.env.RegisterActivityWithOptions(activities.NewRejectPackageActivity(nil).Execute, temporalsdk_activity.RegisterOptions{Name: activities.RejectPackageActivityName})

	s.workflow = NewProcessingWorkflow(logger, pkgsvc, wsvc)
}

func (s *ProcessingWorkflowTestSuite) AfterTest(suiteName, testName string) {
	s.env.AssertExpectations(s.T())
}

func TestProcessingWorkflow(t *testing.T) {
	suite.Run(t, new(ProcessingWorkflowTestSuite))
}

func (s *ProcessingWorkflowTestSuite) TestPackageConfirmation() {
	pkgID := uint(1)
	locationID := uuid.MustParse("51328c02-2b63-47be-958e-e8088aa1a61f")
	watcherName := "watcher"
	retentionPeriod := 1 * time.Second

	// Signal handler that mimics package confirmation
	s.env.RegisterDelayedCallback(
		func() {
			s.env.SignalWorkflow(
				package_.ReviewPerformedSignalName,
				package_.ReviewPerformedSignal{Accepted: true, LocationID: &locationID},
			)
		},
		0,
	)

	// Activity mocks/assertions sequence
	s.env.OnActivity(createPackageLocalActivity, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(pkgID, nil)
	s.env.OnActivity(setStatusInProgressLocalActivity, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
	s.env.OnActivity(createPreservationActionLocalActivity, mock.Anything, mock.Anything, mock.Anything).Return(uint(0), nil)
	s.env.OnActivity(activities.DownloadActivityName, mock.Anything, mock.Anything, mock.Anything).Return("", nil)
	s.env.OnActivity(activities.BundleActivityName, mock.Anything, mock.Anything).Return(&activities.BundleActivityResult{FullPath: "/tmp/aip", FullPathBeforeStrip: "/tmp/aip"}, nil)
	s.env.OnActivity(sdps_activities.ValidatePackageActivityName, mock.Anything, mock.Anything).Return(nil)
	s.env.OnActivity(activities.ValidateTransferActivityName, mock.Anything, mock.Anything).Return(nil)
	s.env.OnActivity(a3m.CreateAIPActivityName, mock.Anything, mock.Anything).Return(nil, nil)
	s.env.OnActivity(updatePackageLocalActivity, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
	s.env.OnActivity(createPreservationTaskLocalActivity, mock.Anything, mock.Anything, mock.Anything).Return(uint(0), nil)
	s.env.OnActivity(activities.UploadActivityName, mock.Anything, mock.Anything).Return(nil)
	s.env.OnActivity(setStatusLocalActivity, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
	s.env.OnActivity(setPreservationActonStatusLocalActivity, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
	s.env.OnActivity(completePreservationTaskLocalActivity, mock.Anything, mock.Anything, mock.Anything).Return(nil)
	s.env.OnActivity(activities.MoveToPermanentStorageActivityName, mock.Anything, mock.Anything).Return(nil)
	s.env.OnActivity(activities.PollMoveToPermanentStorageActivityName, mock.Anything, mock.Anything).Return(nil)
	s.env.OnActivity(setLocationIDLocalActivity, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
	s.env.OnActivity(sdps_activities.IndexActivityName, mock.Anything, mock.Anything).Return(nil)
	// TODO: CleanUpActivityName
	// TODO: DeleteOriginalActivityName
	s.env.OnActivity(updatePackageLocalActivity, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
	s.env.OnActivity(completePreservationActionLocalActivity, mock.Anything, mock.Anything, mock.Anything).Return(nil)

	s.env.ExecuteWorkflow(
		s.workflow.Execute,
		&package_.ProcessingWorkflowRequest{
			ValidationConfig: validation.Config{ChecksumsCheckEnabled: true},
			WatcherName:      watcherName,
			RetentionPeriod:  &retentionPeriod,
		},
	)

	s.True(s.env.IsWorkflowCompleted())
	s.NoError(s.env.GetWorkflowResult(nil))
}

func (s *ProcessingWorkflowTestSuite) TestPackageRejection() {
	pkgID := uint(1)
	watcherName := "watcher"
	retentionPeriod := 1 * time.Second

	// Signal handler that mimics package rejection
	s.env.RegisterDelayedCallback(
		func() {
			s.env.SignalWorkflow(
				package_.ReviewPerformedSignalName,
				package_.ReviewPerformedSignal{Accepted: false},
			)
		},
		0,
	)

	// Activity mocks/assertions sequence
	s.env.OnActivity(createPackageLocalActivity, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(pkgID, nil)
	s.env.OnActivity(setStatusInProgressLocalActivity, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
	s.env.OnActivity(createPreservationActionLocalActivity, mock.Anything, mock.Anything, mock.Anything).Return(uint(0), nil)
	s.env.OnActivity(activities.DownloadActivityName, mock.Anything, mock.Anything, mock.Anything).Return("", nil)
	s.env.OnActivity(activities.BundleActivityName, mock.Anything, mock.Anything).Return(&activities.BundleActivityResult{FullPath: "/tmp/aip", FullPathBeforeStrip: "/tmp/aip"}, nil)
	s.env.OnActivity(sdps_activities.ValidatePackageActivityName, mock.Anything, mock.Anything).Return(nil)
	s.env.OnActivity(activities.ValidateTransferActivityName, mock.Anything, mock.Anything).Return(nil)
	s.env.OnActivity(a3m.CreateAIPActivityName, mock.Anything, mock.Anything).Return(nil, nil)
	s.env.OnActivity(updatePackageLocalActivity, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
	s.env.OnActivity(completePreservationTaskLocalActivity, mock.Anything, mock.Anything, mock.Anything).Return(nil)
	s.env.OnActivity(activities.UploadActivityName, mock.Anything, mock.Anything).Return(nil)
	s.env.OnActivity(setStatusLocalActivity, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
	s.env.OnActivity(setPreservationActonStatusLocalActivity, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
	s.env.OnActivity(createPreservationTaskLocalActivity, mock.Anything, mock.Anything, mock.Anything).Return(uint(0), nil)
	s.env.OnActivity(activities.RejectPackageActivityName, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
	// TODO: CleanUpActivityName
	// TODO: DeleteOriginalActivityName
	s.env.OnActivity(updatePackageLocalActivity, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
	s.env.OnActivity(completePreservationActionLocalActivity, mock.Anything, mock.Anything, mock.Anything).Return(nil)

	s.env.ExecuteWorkflow(
		s.workflow.Execute,
		&package_.ProcessingWorkflowRequest{
			ValidationConfig: validation.Config{ChecksumsCheckEnabled: true},
			WatcherName:      watcherName,
			RetentionPeriod:  &retentionPeriod,
		},
	)

	s.True(s.env.IsWorkflowCompleted())
	s.NoError(s.env.GetWorkflowResult(nil))
}
