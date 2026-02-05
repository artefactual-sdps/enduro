package workflow

import (
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	temporalsdk_activity "go.temporal.io/sdk/activity"
	temporalsdk_testsuite "go.temporal.io/sdk/testsuite"
	temporalsdk_workflow "go.temporal.io/sdk/workflow"
	"go.uber.org/mock/gomock"

	"github.com/artefactual-sdps/enduro/internal/childwf"
	"github.com/artefactual-sdps/enduro/internal/config"
	"github.com/artefactual-sdps/enduro/internal/datatypes"
	"github.com/artefactual-sdps/enduro/internal/enums"
	"github.com/artefactual-sdps/enduro/internal/ingest"
	ingest_fake "github.com/artefactual-sdps/enduro/internal/ingest/fake"
	"github.com/artefactual-sdps/enduro/internal/workflow/activities"
)

const (
	batchIdentifier = "test-batch"
	batchSIP1Key    = "sip1.zip"
	batchSIP2Key    = "sip2.zip"
)

var (
	sourceID      = uuid.MustParse("6ba7b814-9dad-41d1-80b4-00c04fd430c8")
	batchUUID     = uuid.MustParse("f47ac10b-58cc-4372-a567-0e02b2c3d479")
	batchSIP1UUID = uuid.MustParse("52fdfc07-2182-454f-963f-5f0f9a621d72")
	batchSIP2UUID = uuid.MustParse("9566c74d-1003-4c4d-bbbb-0407d1e2c649")
	batchAIP1UUID = uuid.MustParse("11111111-1111-1111-1111-111111111111")
	batchAIP2UUID = uuid.MustParse("22222222-2222-2222-2222-222222222222")
)

type BatchWorkflowTestSuite struct {
	suite.Suite
	temporalsdk_testsuite.WorkflowTestSuite

	env *temporalsdk_testsuite.TestWorkflowEnvironment

	workflow *BatchWorkflow

	// childRequests holds the requests received by the child processing workflows.
	childRequests []ingest.ProcessingWorkflowRequest

	// childSignals holds the signals received by the child processing workflows.
	childSignals []ingest.BatchSignal
}

func (s *BatchWorkflowTestSuite) SetupWorkflowTest(cfg config.Configuration) {
	s.env = s.NewTestWorkflowEnvironment()
	s.env.SetStartTime(startTime)

	ctrl := gomock.NewController(s.T())
	ingestsvc := ingest_fake.NewMockService(ctrl)
	rng := rand.New(rand.NewSource(1)) // #nosec: G404

	s.env.RegisterWorkflowWithOptions(
		s.processingChildWorkflow,
		temporalsdk_workflow.RegisterOptions{Name: ingest.ProcessingWorkflowName},
	)
	s.env.RegisterActivityWithOptions(
		activities.NewPollSIPStatusesActivity(ingestsvc, time.Microsecond).Execute,
		temporalsdk_activity.RegisterOptions{Name: activities.PollSIPStatusesActivityName},
	)

	s.env.RegisterWorkflowWithOptions(
		postBatchChildWorkflow,
		temporalsdk_workflow.RegisterOptions{Name: "postbatch"},
	)

	s.workflow = NewBatchWorkflow(cfg, rng, ingestsvc, nil)
}

// processingChildWorkflow is a mock implementation of the processing workflow
// used in tests to capture requests and signals.
func (s *BatchWorkflowTestSuite) processingChildWorkflow(
	ctx temporalsdk_workflow.Context, req *ingest.ProcessingWorkflowRequest,
) error {
	s.childRequests = append(s.childRequests, *req)
	var signal ingest.BatchSignal
	_ = temporalsdk_workflow.GetSignalChannel(ctx, ingest.BatchSignalName).Receive(ctx, &signal)
	s.childSignals = append(s.childSignals, signal)

	return nil
}

// postBatchChildWorkflow is a no-op workflow that will be replaced with a
// mock in test.
func postBatchChildWorkflow(
	ctx temporalsdk_workflow.Context,
	params *childwf.PostbatchParams,
) (*childwf.Result, error) {
	return nil, nil
}

func (s *BatchWorkflowTestSuite) AfterTest(suiteName, testName string) {
	s.env.AssertExpectations(s.T())
	s.childRequests = nil
	s.childSignals = nil
}

// ExecuteAndValidateWorkflow executes the batch workflow with the given request,
// validates that the expected child requests and signals were sent, and checks
// for errors based on shouldError.
func (s *BatchWorkflowTestSuite) ExecuteAndValidateWorkflow(
	request *ingest.BatchWorkflowRequest,
	childRequests []ingest.ProcessingWorkflowRequest,
	childSignals []ingest.BatchSignal,
	shouldError bool,
) {
	s.env.ExecuteWorkflow(s.workflow.Execute, request)
	s.True(s.env.IsWorkflowCompleted())

	s.Equal(s.childRequests, childRequests)
	s.Equal(s.childSignals, childSignals)

	err := s.env.GetWorkflowError()
	if shouldError {
		s.Error(err)
	} else {
		s.NoError(err)
	}
}

func TestBatchWorkflow(t *testing.T) {
	suite.Run(t, new(BatchWorkflowTestSuite))
}

// TestBatch tests:
// - Update batch status (processing).
// - Create a SIP for each key.
// - Start a child processing workflow for each SIP.
// - Poll SIP statuses until all are validated.
// - Signal SIP workflows to continue processing.
// - Poll SIP statuses until all are ingested.
// - Wait for all child workflows to complete.
// - Update batch status (ingested).
// - Run postbatch child workflow.
func (s *BatchWorkflowTestSuite) TestBatch() {
	cfg := config.Configuration{
		ChildWorkflows: childwf.Configs{
			{
				Type:         enums.ChildWorkflowTypePostbatch,
				Namespace:    "default",
				TaskQueue:    "postbatch",
				WorkflowName: "postbatch",
			},
		},
	}
	s.SetupWorkflowTest(cfg)

	// Mock initial batch status update.
	s.env.OnActivity(
		updateBatchLocalActivity,
		ctx,
		s.workflow.ingestsvc,
		&updateBatchLocalActivityParams{
			UUID:      batchUUID,
			Status:    enums.BatchStatusProcessing,
			StartedAt: startTime,
		},
	).Return(&updateBatchLocalActivityResult{}, nil)

	// Mock SIP creation for the first SIP.
	s.env.OnActivity(
		createSIPLocalActivity,
		ctx,
		s.workflow.ingestsvc,
		&createSIPLocalActivityParams{
			SIP: datatypes.SIP{
				UUID:   batchSIP1UUID,
				Name:   batchSIP1Key,
				Status: enums.SIPStatusQueued,
				Batch: &datatypes.Batch{
					UUID:       batchUUID,
					Identifier: batchIdentifier,
					Status:     enums.BatchStatusProcessing,
					SIPSCount:  2,
					CreatedAt:  startTime,
					StartedAt:  startTime,
				},
			},
		},
	).Return(1, nil)

	// Mock SIP creation for the second SIP.
	s.env.OnActivity(
		createSIPLocalActivity,
		ctx,
		s.workflow.ingestsvc,
		&createSIPLocalActivityParams{
			SIP: datatypes.SIP{
				UUID:   batchSIP2UUID,
				Name:   batchSIP2Key,
				Status: enums.SIPStatusQueued,
				Batch: &datatypes.Batch{
					UUID:       batchUUID,
					Identifier: batchIdentifier,
					Status:     enums.BatchStatusProcessing,
					SIPSCount:  2,
					CreatedAt:  startTime,
					StartedAt:  startTime,
				},
			},
		},
	).Return(2, nil)

	// Mock validated SIP statuses poll.
	s.env.OnActivity(
		activities.PollSIPStatusesActivityName,
		mock.AnythingOfType("*context.timerCtx"),
		&activities.PollSIPStatusesActivityParams{
			BatchUUID:        batchUUID,
			ExpectedSIPCount: 2,
			ExpectedStatus:   enums.SIPStatusValidated,
		},
	).Return(&activities.PollSIPStatusesActivityResult{AllExpectedStatus: true}, nil)

	// Mock ingested SIP statuses poll.
	s.env.OnActivity(
		activities.PollSIPStatusesActivityName,
		mock.AnythingOfType("*context.timerCtx"),
		&activities.PollSIPStatusesActivityParams{
			BatchUUID:        batchUUID,
			ExpectedSIPCount: 2,
			ExpectedStatus:   enums.SIPStatusIngested,
		},
	).Return(&activities.PollSIPStatusesActivityResult{
		AllExpectedStatus: true,
		SIPIDstoAIPIDs: map[uuid.UUID]uuid.UUID{
			batchSIP1UUID: batchAIP1UUID,
			batchSIP2UUID: batchAIP2UUID,
		},
	}, nil)

	// Mock postbatch child workflow.
	s.env.OnWorkflow(
		postBatchChildWorkflow,
		internalCtx,
		&childwf.PostbatchParams{
			Batch: &childwf.PostbatchBatch{
				UUID:      batchUUID,
				SIPSCount: 2,
			},
			SIPs: []*childwf.PostbatchSIP{
				{
					UUID:  batchSIP1UUID,
					Name:  batchSIP1Key,
					AIPID: &batchAIP1UUID,
				},
				{
					UUID:  batchSIP2UUID,
					Name:  batchSIP2Key,
					AIPID: &batchAIP2UUID,
				},
			},
		},
	).Return(
		&childwf.Result{
			Outcome: childwf.OutcomeSuccess,
			Message: "Postbatch workflow executed successfully",
		},
		nil,
	)

	// Mock final batch status update.
	s.env.OnActivity(
		updateBatchLocalActivity,
		ctx,
		s.workflow.ingestsvc,
		&updateBatchLocalActivityParams{
			UUID:        batchUUID,
			Status:      enums.BatchStatusIngested,
			StartedAt:   startTime,
			CompletedAt: startTime,
		},
	).Return(&updateBatchLocalActivityResult{}, nil)

	s.ExecuteAndValidateWorkflow(
		&ingest.BatchWorkflowRequest{
			Batch: datatypes.Batch{
				UUID:       batchUUID,
				Identifier: batchIdentifier,
				Status:     enums.BatchStatusQueued,
				CreatedAt:  startTime,
				SIPSCount:  2,
			},
			SIPSourceID: sourceID,
			Keys:        []string{batchSIP1Key, batchSIP2Key},
		},
		[]ingest.ProcessingWorkflowRequest{
			{
				SIPUUID:         batchSIP1UUID,
				SIPName:         batchSIP1Key,
				Key:             batchSIP1Key,
				SIPSourceID:     sourceID,
				Type:            enums.WorkflowTypeCreateAip,
				RetentionPeriod: -1 * time.Second,
				BatchUUID:       batchUUID,
			},
			{
				SIPUUID:         batchSIP2UUID,
				SIPName:         batchSIP2Key,
				Key:             batchSIP2Key,
				SIPSourceID:     sourceID,
				Type:            enums.WorkflowTypeCreateAip,
				RetentionPeriod: -1 * time.Second,
				BatchUUID:       batchUUID,
			},
		},
		[]ingest.BatchSignal{
			{Continue: true},
			{Continue: true},
		},
		false,
	)
}

// TestBatchValidationFailed tests:
// - Batch status update (processing).
// - SIP creation for each key.
// - Child processing workflows started for each SIP.
// - Polling SIP statuses until some fail validation.
// - Signaling SIP workflows to stop processing.
// - Waiting for all child workflows to complete.
// - Batch status update (failed).
func (s *BatchWorkflowTestSuite) TestBatchValidationFailed() {
	s.SetupWorkflowTest(config.Configuration{})

	// Mock initial batch status update.
	s.env.OnActivity(
		updateBatchLocalActivity,
		ctx,
		s.workflow.ingestsvc,
		&updateBatchLocalActivityParams{
			UUID:      batchUUID,
			Status:    enums.BatchStatusProcessing,
			StartedAt: startTime,
		},
	).Return(&updateBatchLocalActivityResult{}, nil)

	// Mock SIP creation for the first SIP.
	s.env.OnActivity(
		createSIPLocalActivity,
		ctx,
		s.workflow.ingestsvc,
		&createSIPLocalActivityParams{
			SIP: datatypes.SIP{
				UUID:   batchSIP1UUID,
				Name:   batchSIP1Key,
				Status: enums.SIPStatusQueued,
				Batch: &datatypes.Batch{
					UUID:       batchUUID,
					Identifier: batchIdentifier,
					Status:     enums.BatchStatusProcessing,
					SIPSCount:  2,
					CreatedAt:  startTime,
					StartedAt:  startTime,
				},
			},
		},
	).Return(1, nil)

	// Mock SIP creation for the second SIP.
	s.env.OnActivity(
		createSIPLocalActivity,
		ctx,
		s.workflow.ingestsvc,
		&createSIPLocalActivityParams{
			SIP: datatypes.SIP{
				UUID:   batchSIP2UUID,
				Name:   batchSIP2Key,
				Status: enums.SIPStatusQueued,
				Batch: &datatypes.Batch{
					UUID:       batchUUID,
					Identifier: batchIdentifier,
					Status:     enums.BatchStatusProcessing,
					SIPSCount:  2,
					CreatedAt:  startTime,
					StartedAt:  startTime,
				},
			},
		},
	).Return(2, nil)

	// Mock validated SIP statuses poll.
	s.env.OnActivity(
		activities.PollSIPStatusesActivityName,
		mock.AnythingOfType("*context.timerCtx"),
		&activities.PollSIPStatusesActivityParams{
			BatchUUID:        batchUUID,
			ExpectedSIPCount: 2,
			ExpectedStatus:   enums.SIPStatusValidated,
		},
	).Return(&activities.PollSIPStatusesActivityResult{AllExpectedStatus: false}, nil)

	// Mock final batch status update.
	s.env.OnActivity(
		updateBatchLocalActivity,
		ctx,
		s.workflow.ingestsvc,
		&updateBatchLocalActivityParams{
			UUID:        batchUUID,
			Status:      enums.BatchStatusFailed,
			StartedAt:   startTime,
			CompletedAt: startTime,
		},
	).Return(&updateBatchLocalActivityResult{}, nil)

	s.ExecuteAndValidateWorkflow(
		&ingest.BatchWorkflowRequest{
			Batch: datatypes.Batch{
				UUID:       batchUUID,
				Identifier: batchIdentifier,
				Status:     enums.BatchStatusQueued,
				CreatedAt:  startTime,
				SIPSCount:  2,
			},
			SIPSourceID: sourceID,
			Keys:        []string{batchSIP1Key, batchSIP2Key},
		},
		[]ingest.ProcessingWorkflowRequest{
			{
				SIPUUID:         batchSIP1UUID,
				SIPName:         batchSIP1Key,
				Key:             batchSIP1Key,
				SIPSourceID:     sourceID,
				Type:            enums.WorkflowTypeCreateAip,
				RetentionPeriod: -1 * time.Second,
				BatchUUID:       batchUUID,
			},
			{
				SIPUUID:         batchSIP2UUID,
				SIPName:         batchSIP2Key,
				Key:             batchSIP2Key,
				SIPSourceID:     sourceID,
				Type:            enums.WorkflowTypeCreateAip,
				RetentionPeriod: -1 * time.Second,
				BatchUUID:       batchUUID,
			},
		},
		[]ingest.BatchSignal{
			{Continue: false},
			{Continue: false},
		},
		true,
	)
}

// TestBatchIngestFailedContinue tests:
// - Batch status update (processing).
// - SIP creation for each key.
// - Child processing workflows started for each SIP.
// - Polling SIP statuses until all are validated.
// - Signaling SIP workflows to continue processing.
// - Polling SIP statuses until some fail to ingest.
// - Waiting for batch decision signal (continue).
// - Waiting for all child workflows to complete.
// - Batch status update (ingested).
func (s *BatchWorkflowTestSuite) TestBatchIngestFailedContinue() {
	s.SetupWorkflowTest(config.Configuration{})

	// Mock initial batch status update.
	s.env.OnActivity(
		updateBatchLocalActivity,
		ctx,
		s.workflow.ingestsvc,
		&updateBatchLocalActivityParams{
			UUID:      batchUUID,
			Status:    enums.BatchStatusProcessing,
			StartedAt: startTime,
		},
	).Return(&updateBatchLocalActivityResult{}, nil)

	// Mock SIP creation for the first SIP.
	s.env.OnActivity(
		createSIPLocalActivity,
		ctx,
		s.workflow.ingestsvc,
		&createSIPLocalActivityParams{
			SIP: datatypes.SIP{
				UUID:   batchSIP1UUID,
				Name:   batchSIP1Key,
				Status: enums.SIPStatusQueued,
				Batch: &datatypes.Batch{
					UUID:       batchUUID,
					Identifier: batchIdentifier,
					Status:     enums.BatchStatusProcessing,
					SIPSCount:  2,
					CreatedAt:  startTime,
					StartedAt:  startTime,
				},
			},
		},
	).Return(1, nil)

	// Mock SIP creation for the second SIP.
	s.env.OnActivity(
		createSIPLocalActivity,
		ctx,
		s.workflow.ingestsvc,
		&createSIPLocalActivityParams{
			SIP: datatypes.SIP{
				UUID:   batchSIP2UUID,
				Name:   batchSIP2Key,
				Status: enums.SIPStatusQueued,
				Batch: &datatypes.Batch{
					UUID:       batchUUID,
					Identifier: batchIdentifier,
					Status:     enums.BatchStatusProcessing,
					SIPSCount:  2,
					CreatedAt:  startTime,
					StartedAt:  startTime,
				},
			},
		},
	).Return(2, nil)

	// Mock validated SIP statuses poll.
	s.env.OnActivity(
		activities.PollSIPStatusesActivityName,
		mock.AnythingOfType("*context.timerCtx"),
		&activities.PollSIPStatusesActivityParams{
			BatchUUID:        batchUUID,
			ExpectedSIPCount: 2,
			ExpectedStatus:   enums.SIPStatusValidated,
		},
	).Return(&activities.PollSIPStatusesActivityResult{AllExpectedStatus: true}, nil)

	// Mock ingested SIP statuses poll.
	s.env.OnActivity(
		activities.PollSIPStatusesActivityName,
		mock.AnythingOfType("*context.timerCtx"),
		&activities.PollSIPStatusesActivityParams{
			BatchUUID:        batchUUID,
			ExpectedSIPCount: 2,
			ExpectedStatus:   enums.SIPStatusIngested,
		},
	).Return(&activities.PollSIPStatusesActivityResult{AllExpectedStatus: false}, nil)

	// Mock batch status update to pending.
	s.env.OnActivity(
		updateBatchLocalActivity,
		ctx,
		s.workflow.ingestsvc,
		&updateBatchLocalActivityParams{
			UUID:      batchUUID,
			Status:    enums.BatchStatusPending,
			StartedAt: startTime,
		},
	).Return(&updateBatchLocalActivityResult{}, nil)

	// Simulate user decision to continue with partial ingest.
	s.env.RegisterDelayedCallback(func() {
		s.env.SignalWorkflow(ingest.BatchDecisionSignalName, ingest.BatchDecisionSignal{Continue: true})
	}, 0)

	// Mock batch status update to processing.
	s.env.OnActivity(
		updateBatchLocalActivity,
		ctx,
		s.workflow.ingestsvc,
		&updateBatchLocalActivityParams{
			UUID:      batchUUID,
			Status:    enums.BatchStatusProcessing,
			StartedAt: startTime,
		},
	).Return(&updateBatchLocalActivityResult{}, nil)

	// Mock final batch status update.
	s.env.OnActivity(
		updateBatchLocalActivity,
		ctx,
		s.workflow.ingestsvc,
		&updateBatchLocalActivityParams{
			UUID:        batchUUID,
			Status:      enums.BatchStatusIngested,
			StartedAt:   startTime,
			CompletedAt: startTime,
		},
	).Return(&updateBatchLocalActivityResult{}, nil)

	s.ExecuteAndValidateWorkflow(
		&ingest.BatchWorkflowRequest{
			Batch: datatypes.Batch{
				UUID:       batchUUID,
				Identifier: batchIdentifier,
				Status:     enums.BatchStatusQueued,
				CreatedAt:  startTime,
				SIPSCount:  2,
			},
			SIPSourceID: sourceID,
			Keys:        []string{batchSIP1Key, batchSIP2Key},
		},
		[]ingest.ProcessingWorkflowRequest{
			{
				SIPUUID:         batchSIP1UUID,
				SIPName:         batchSIP1Key,
				Key:             batchSIP1Key,
				SIPSourceID:     sourceID,
				Type:            enums.WorkflowTypeCreateAip,
				RetentionPeriod: -1 * time.Second,
				BatchUUID:       batchUUID,
			},
			{
				SIPUUID:         batchSIP2UUID,
				SIPName:         batchSIP2Key,
				Key:             batchSIP2Key,
				SIPSourceID:     sourceID,
				Type:            enums.WorkflowTypeCreateAip,
				RetentionPeriod: -1 * time.Second,
				BatchUUID:       batchUUID,
			},
		},
		[]ingest.BatchSignal{
			{Continue: true},
			{Continue: true},
		},
		false,
	)
}

// TestBatchIngestFailedCancel tests:
// - Batch status update (processing).
// - SIP creation for each key.
// - Child processing workflows started for each SIP.
// - Polling SIP statuses until all are validated.
// - Signaling SIP workflows to continue processing.
// - Polling SIP statuses until some fail to ingest.
// - Waiting for batch decision signal (cancel).
// - Waiting for all child workflows to complete.
// - Batch status update (canceled).
func (s *BatchWorkflowTestSuite) TestBatchIngestFailedCancel() {
	s.SetupWorkflowTest(config.Configuration{})

	// Mock initial batch status update.
	s.env.OnActivity(
		updateBatchLocalActivity,
		ctx,
		s.workflow.ingestsvc,
		&updateBatchLocalActivityParams{
			UUID:      batchUUID,
			Status:    enums.BatchStatusProcessing,
			StartedAt: startTime,
		},
	).Return(&updateBatchLocalActivityResult{}, nil)

	// Mock SIP creation for the first SIP.
	s.env.OnActivity(
		createSIPLocalActivity,
		ctx,
		s.workflow.ingestsvc,
		&createSIPLocalActivityParams{
			SIP: datatypes.SIP{
				UUID:   batchSIP1UUID,
				Name:   batchSIP1Key,
				Status: enums.SIPStatusQueued,
				Batch: &datatypes.Batch{
					UUID:       batchUUID,
					Identifier: batchIdentifier,
					Status:     enums.BatchStatusProcessing,
					SIPSCount:  2,
					CreatedAt:  startTime,
					StartedAt:  startTime,
				},
			},
		},
	).Return(1, nil)

	// Mock SIP creation for the second SIP.
	s.env.OnActivity(
		createSIPLocalActivity,
		ctx,
		s.workflow.ingestsvc,
		&createSIPLocalActivityParams{
			SIP: datatypes.SIP{
				UUID:   batchSIP2UUID,
				Name:   batchSIP2Key,
				Status: enums.SIPStatusQueued,
				Batch: &datatypes.Batch{
					UUID:       batchUUID,
					Identifier: batchIdentifier,
					Status:     enums.BatchStatusProcessing,
					SIPSCount:  2,
					CreatedAt:  startTime,
					StartedAt:  startTime,
				},
			},
		},
	).Return(2, nil)

	// Mock validated SIP statuses poll.
	s.env.OnActivity(
		activities.PollSIPStatusesActivityName,
		mock.AnythingOfType("*context.timerCtx"),
		&activities.PollSIPStatusesActivityParams{
			BatchUUID:        batchUUID,
			ExpectedSIPCount: 2,
			ExpectedStatus:   enums.SIPStatusValidated,
		},
	).Return(&activities.PollSIPStatusesActivityResult{AllExpectedStatus: true}, nil)

	// Mock ingested SIP statuses poll.
	s.env.OnActivity(
		activities.PollSIPStatusesActivityName,
		mock.AnythingOfType("*context.timerCtx"),
		&activities.PollSIPStatusesActivityParams{
			BatchUUID:        batchUUID,
			ExpectedSIPCount: 2,
			ExpectedStatus:   enums.SIPStatusIngested,
		},
	).Return(&activities.PollSIPStatusesActivityResult{AllExpectedStatus: false}, nil)

	// Mock batch status update to pending.
	s.env.OnActivity(
		updateBatchLocalActivity,
		ctx,
		s.workflow.ingestsvc,
		&updateBatchLocalActivityParams{
			UUID:      batchUUID,
			Status:    enums.BatchStatusPending,
			StartedAt: startTime,
		},
	).Return(&updateBatchLocalActivityResult{}, nil)

	// Simulate user decision to not continue with partial ingest.
	s.env.RegisterDelayedCallback(func() {
		s.env.SignalWorkflow(ingest.BatchDecisionSignalName, ingest.BatchDecisionSignal{Continue: false})
	}, 0)

	// Mock final batch status update.
	s.env.OnActivity(
		updateBatchLocalActivity,
		ctx,
		s.workflow.ingestsvc,
		&updateBatchLocalActivityParams{
			UUID:        batchUUID,
			Status:      enums.BatchStatusCanceled,
			StartedAt:   startTime,
			CompletedAt: startTime,
		},
	).Return(&updateBatchLocalActivityResult{}, nil)

	s.ExecuteAndValidateWorkflow(
		&ingest.BatchWorkflowRequest{
			Batch: datatypes.Batch{
				UUID:       batchUUID,
				Identifier: batchIdentifier,
				Status:     enums.BatchStatusQueued,
				CreatedAt:  startTime,
				SIPSCount:  2,
			},
			SIPSourceID: sourceID,
			Keys:        []string{batchSIP1Key, batchSIP2Key},
		},
		[]ingest.ProcessingWorkflowRequest{
			{
				SIPUUID:         batchSIP1UUID,
				SIPName:         batchSIP1Key,
				Key:             batchSIP1Key,
				SIPSourceID:     sourceID,
				Type:            enums.WorkflowTypeCreateAip,
				RetentionPeriod: -1 * time.Second,
				BatchUUID:       batchUUID,
			},
			{
				SIPUUID:         batchSIP2UUID,
				SIPName:         batchSIP2Key,
				Key:             batchSIP2Key,
				SIPSourceID:     sourceID,
				Type:            enums.WorkflowTypeCreateAip,
				RetentionPeriod: -1 * time.Second,
				BatchUUID:       batchUUID,
			},
		},
		[]ingest.BatchSignal{
			{Continue: true},
			{Continue: true},
		},
		true,
	)
}

// TestBatchError tests:
// - Batch status updates failures.
func (s *BatchWorkflowTestSuite) TestBatchError() {
	s.SetupWorkflowTest(config.Configuration{})

	// Mock both activity calls with flexible parameter matching to
	// handle retries (3 + 3) and dynamic completion times.
	s.env.OnActivity(
		updateBatchLocalActivity,
		ctx,
		s.workflow.ingestsvc,
		mock.AnythingOfType("*workflow.updateBatchLocalActivityParams"),
	).Return(nil, fmt.Errorf("update error")).Times(6)

	s.ExecuteAndValidateWorkflow(&ingest.BatchWorkflowRequest{
		Batch: datatypes.Batch{
			UUID:       batchUUID,
			Identifier: batchIdentifier,
			Status:     enums.BatchStatusQueued,
			CreatedAt:  startTime,
			SIPSCount:  2,
		},
		SIPSourceID: sourceID,
		Keys:        []string{batchSIP1Key, batchSIP2Key},
	}, nil, nil, true)
}
