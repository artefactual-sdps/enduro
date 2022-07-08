package workflow

import (
	"errors"
	"testing"

	"github.com/go-logr/logr"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	temporalsdk_activity "go.temporal.io/sdk/activity"
	temporalsdk_testsuite "go.temporal.io/sdk/testsuite"
	temporalsdk_worker "go.temporal.io/sdk/worker"

	"github.com/artefactual-labs/enduro/internal/package_"
	packagefake "github.com/artefactual-labs/enduro/internal/package_/fake"
	"github.com/artefactual-labs/enduro/internal/workflow/activities"
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
	location := "perma-aips-100"

	// Package is set to in progress status.
	s.env.OnActivity(setStatusLocalActivity, mock.Anything, mock.Anything, pkgID, package_.StatusInProgress).Return(nil)

	// Move operation succeeds.
	s.env.OnActivity(
		activities.MoveToPermanentStorageActivityName,
		mock.Anything,
		&activities.MoveToPermanentStorageActivityParams{
			AIPID:    AIPID,
			Location: location,
		},
	).Return(nil)

	// Polling of move operation succeeds.
	s.env.OnActivity(
		activities.PollMoveToPermanentStorageActivityName,
		mock.Anything,
		&activities.PollMoveToPermanentStorageActivityParams{
			AIPID: AIPID,
		},
	).Return(nil)

	// Package is set back to done status.
	s.env.OnActivity(setStatusLocalActivity, mock.Anything, mock.Anything, pkgID, package_.StatusDone).Return(nil)

	// Package location is set.
	s.env.OnActivity(setLocationLocalActivity, mock.Anything, mock.Anything, pkgID, location).Return(nil)

	// Preservation action is created with successful status.
	s.env.OnActivity(
		saveLocationMovePreservationActionLocalActivity,
		mock.Anything,
		mock.Anything,
		pkgID,
		location,
		package_.StatusComplete,
		mock.Anything,
		mock.Anything,
	).Return(nil)

	s.env.ExecuteWorkflow(
		s.workflow.Execute,
		&package_.MoveWorkflowRequest{
			ID:       pkgID,
			AIPID:    AIPID,
			Location: location,
		},
	)

	s.True(s.env.IsWorkflowCompleted())
	s.NoError(s.env.GetWorkflowResult(nil))
}

func (s *MoveWorkflowTestSuite) TestFailedMove() {
	pkgID := uint(1)
	AIPID := uuid.NewString()
	location := "perma-aips-100"

	// Package is set to in progress status.
	s.env.OnActivity(setStatusLocalActivity, mock.Anything, mock.Anything, pkgID, package_.StatusInProgress).Return(nil)

	// Move operation fails.
	s.env.OnActivity(
		activities.MoveToPermanentStorageActivityName,
		mock.Anything,
		&activities.MoveToPermanentStorageActivityParams{
			AIPID:    AIPID,
			Location: location,
		},
	).Return(errors.New("error moving package"))

	// Package is set back to done status.
	s.env.OnActivity(setStatusLocalActivity, mock.Anything, mock.Anything, pkgID, package_.StatusDone).Return(nil)

	// Preservation action is created with failed status.
	s.env.OnActivity(
		saveLocationMovePreservationActionLocalActivity,
		mock.Anything,
		mock.Anything,
		pkgID,
		location,
		package_.StatusFailed,
		mock.Anything,
		mock.Anything,
	).Return(nil)

	s.env.ExecuteWorkflow(
		s.workflow.Execute,
		&package_.MoveWorkflowRequest{
			ID:       pkgID,
			AIPID:    AIPID,
			Location: location,
		},
	)

	s.True(s.env.IsWorkflowCompleted())
	s.NoError(s.env.GetWorkflowResult(nil))
}