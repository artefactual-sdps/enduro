package workflow

import (
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	temporalsdk_activity "go.temporal.io/sdk/activity"
	temporalsdk_testsuite "go.temporal.io/sdk/testsuite"
	temporalsdk_worker "go.temporal.io/sdk/worker"
	"go.uber.org/mock/gomock"

	"github.com/artefactual-sdps/enduro/internal/enums"
	"github.com/artefactual-sdps/enduro/internal/ingest"
	ingest_fake "github.com/artefactual-sdps/enduro/internal/ingest/fake"
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
	ingestsvc := ingest_fake.NewMockService(ctrl)

	s.env.RegisterActivityWithOptions(
		activities.NewMoveToPermanentStorageActivity(nil).Execute,
		temporalsdk_activity.RegisterOptions{Name: activities.MoveToPermanentStorageActivityName},
	)
	s.env.RegisterActivityWithOptions(
		activities.NewPollMoveToPermanentStorageActivity(nil).Execute,
		temporalsdk_activity.RegisterOptions{Name: activities.PollMoveToPermanentStorageActivityName},
	)

	s.workflow = NewMoveWorkflow(ingestsvc)
}

func (s *MoveWorkflowTestSuite) AfterTest(suiteName, testName string) {
	s.env.AssertExpectations(s.T())
}

func TestMoveWorkflow(t *testing.T) {
	suite.Run(t, new(MoveWorkflowTestSuite))
}

func (s *MoveWorkflowTestSuite) TestSuccessfulMove() {
	sipID := 1
	AIPID := uuid.NewString()
	locationID := uuid.MustParse("51328c02-2b63-47be-958e-e8088aa1a61f")

	// SIP is set to in progress status.
	s.env.OnActivity(setStatusLocalActivity, mock.Anything, mock.Anything, sipID, enums.SIPStatusInProgress).
		Return(nil, nil)

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

	// SIP is set back to done status.
	s.env.OnActivity(setStatusLocalActivity, mock.Anything, mock.Anything, sipID, enums.SIPStatusDone).
		Return(nil, nil)

	// SIP location is set.
	s.env.OnActivity(setLocationIDLocalActivity, mock.Anything, mock.Anything, sipID, locationID).Return(nil, nil)

	// Workflow is created with successful status.
	s.env.OnActivity(
		saveLocationMoveWorkflowLocalActivity,
		mock.Anything,
		mock.Anything,
		mock.Anything,
	).Return(nil, nil)

	s.env.ExecuteWorkflow(
		s.workflow.Execute,
		&ingest.MoveWorkflowRequest{
			ID:         sipID,
			AIPID:      AIPID,
			LocationID: locationID,
			TaskQueue:  temporal.GlobalTaskQueue,
		},
	)

	s.True(s.env.IsWorkflowCompleted())
	s.NoError(s.env.GetWorkflowResult(nil))
}

func (s *MoveWorkflowTestSuite) TestFailedMove() {
	sipID := 1
	AIPID := uuid.NewString()
	locationID := uuid.MustParse("51328c02-2b63-47be-958e-e8088aa1a61f")

	// SIP is set to in progress status.
	s.env.OnActivity(setStatusLocalActivity, mock.Anything, mock.Anything, sipID, enums.SIPStatusInProgress).
		Return(nil, nil)

	// Move operation fails.
	s.env.OnActivity(
		activities.MoveToPermanentStorageActivityName,
		mock.Anything,
		&activities.MoveToPermanentStorageActivityParams{
			AIPID:      AIPID,
			LocationID: locationID,
		},
	).Return(nil, errors.New("error moving AIP"))

	// SIP is set back to done status.
	s.env.OnActivity(setStatusLocalActivity, mock.Anything, mock.Anything, sipID, enums.SIPStatusDone).
		Return(nil, nil)

	// Workflow is created with failed status.
	s.env.OnActivity(
		saveLocationMoveWorkflowLocalActivity,
		mock.Anything,
		mock.Anything,
		mock.Anything,
	).Return(nil, nil)

	s.env.ExecuteWorkflow(
		s.workflow.Execute,
		&ingest.MoveWorkflowRequest{
			ID:         sipID,
			AIPID:      AIPID,
			LocationID: locationID,
		},
	)

	s.True(s.env.IsWorkflowCompleted())
	s.NoError(s.env.GetWorkflowResult(nil))
}
