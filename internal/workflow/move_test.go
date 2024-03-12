package workflow

import (
	"errors"
	"testing"

	"github.com/go-logr/logr"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	temporalsdk_activity "go.temporal.io/sdk/activity"
	temporalsdk_testsuite "go.temporal.io/sdk/testsuite"
	temporalsdk_worker "go.temporal.io/sdk/worker"
	"go.uber.org/mock/gomock"

	"github.com/artefactual-sdps/enduro/internal/enums"
	"github.com/artefactual-sdps/enduro/internal/package_"
	packagefake "github.com/artefactual-sdps/enduro/internal/package_/fake"
	"github.com/artefactual-sdps/enduro/internal/temporal"
	"github.com/artefactual-sdps/enduro/internal/workflow/activities"
)

type MoveWorkflowTestSuite struct {
	suite.Suite
	temporalsdk_testsuite.WorkflowTestSuite

	env *temporalsdk_testsuite.TestWorkflowEnvironment

	// Each test registers the workflow with a different name to avoid dups.
	workflow *MoveWorkflow
}

func (s *MoveWorkflowTestSuite) SetupTest() {
	s.env = s.NewTestWorkflowEnvironment()
	s.env.SetWorkerOptions(temporalsdk_worker.Options{EnableSessionWorker: true})

	ctrl := gomock.NewController(s.T())
	logger := logr.Discard()
	pkgsvc := packagefake.NewMockService(ctrl)

	s.env.RegisterActivityWithOptions(
		activities.NewMoveToPermanentStorageActivity(nil).Execute,
		temporalsdk_activity.RegisterOptions{Name: activities.MoveToPermanentStorageActivityName},
	)
	s.env.RegisterActivityWithOptions(
		activities.NewPollMoveToPermanentStorageActivity(nil).Execute,
		temporalsdk_activity.RegisterOptions{Name: activities.PollMoveToPermanentStorageActivityName},
	)

	s.workflow = NewMoveWorkflow(logger, pkgsvc)
}

func (s *MoveWorkflowTestSuite) AfterTest(suiteName, testName string) {
	s.env.AssertExpectations(s.T())
}

func TestMoveWorkflow(t *testing.T) {
	suite.Run(t, new(MoveWorkflowTestSuite))
}

func (s *MoveWorkflowTestSuite) TestSuccessfulMove() {
	pkgID := uint(1)
	AIPID := uuid.NewString()
	locationID := uuid.MustParse("51328c02-2b63-47be-958e-e8088aa1a61f")

	// Package is set to in progress status.
	s.env.OnActivity(setStatusLocalActivity, mock.Anything, mock.Anything, pkgID, enums.PackageStatusInProgress).Return(nil, nil)

	// Move operation succeeds.
	s.env.OnActivity(
		activities.MoveToPermanentStorageActivityName,
		mock.Anything,
		&activities.MoveToPermanentStorageActivityParams{
			AIPID:      AIPID,
			LocationID: locationID,
		},
	).Return(nil, nil)

	// Polling of move operation succeeds.
	s.env.OnActivity(
		activities.PollMoveToPermanentStorageActivityName,
		mock.Anything,
		&activities.PollMoveToPermanentStorageActivityParams{
			AIPID: AIPID,
		},
	).Return(nil, nil)

	// Package is set back to done status.
	s.env.OnActivity(setStatusLocalActivity, mock.Anything, mock.Anything, pkgID, enums.PackageStatusDone).Return(nil, nil)

	// Package location is set.
	s.env.OnActivity(setLocationIDLocalActivity, mock.Anything, mock.Anything, pkgID, locationID).Return(nil, nil)

	// Preservation action is created with successful status.
	s.env.OnActivity(
		saveLocationMovePreservationActionLocalActivity,
		mock.Anything,
		mock.Anything,
		mock.Anything,
	).Return(nil, nil)

	s.env.ExecuteWorkflow(
		s.workflow.Execute,
		&package_.MoveWorkflowRequest{
			ID:         pkgID,
			AIPID:      AIPID,
			LocationID: locationID,
			TaskQueue:  temporal.GlobalTaskQueue,
		},
	)

	s.True(s.env.IsWorkflowCompleted())
	s.NoError(s.env.GetWorkflowResult(nil))
}

func (s *MoveWorkflowTestSuite) TestFailedMove() {
	pkgID := uint(1)
	AIPID := uuid.NewString()
	locationID := uuid.MustParse("51328c02-2b63-47be-958e-e8088aa1a61f")

	// Package is set to in progress status.
	s.env.OnActivity(setStatusLocalActivity, mock.Anything, mock.Anything, pkgID, enums.PackageStatusInProgress).Return(nil, nil)

	// Move operation fails.
	s.env.OnActivity(
		activities.MoveToPermanentStorageActivityName,
		mock.Anything,
		&activities.MoveToPermanentStorageActivityParams{
			AIPID:      AIPID,
			LocationID: locationID,
		},
	).Return(nil, errors.New("error moving package"))

	// Package is set back to done status.
	s.env.OnActivity(setStatusLocalActivity, mock.Anything, mock.Anything, pkgID, enums.PackageStatusDone).Return(nil, nil)

	// Preservation action is created with failed status.
	s.env.OnActivity(
		saveLocationMovePreservationActionLocalActivity,
		mock.Anything,
		mock.Anything,
		mock.Anything,
	).Return(nil, nil)

	s.env.ExecuteWorkflow(
		s.workflow.Execute,
		&package_.MoveWorkflowRequest{
			ID:         pkgID,
			AIPID:      AIPID,
			LocationID: locationID,
		},
	)

	s.True(s.env.IsWorkflowCompleted())
	s.NoError(s.env.GetWorkflowResult(nil))
}
