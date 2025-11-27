package workflow

import (
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	temporalsdk_testsuite "go.temporal.io/sdk/testsuite"
	temporalsdk_workflow "go.temporal.io/sdk/workflow"
	"go.uber.org/mock/gomock"

	"github.com/artefactual-sdps/enduro/internal/config"
	"github.com/artefactual-sdps/enduro/internal/datatypes"
	"github.com/artefactual-sdps/enduro/internal/enums"
	"github.com/artefactual-sdps/enduro/internal/ingest"
	ingest_fake "github.com/artefactual-sdps/enduro/internal/ingest/fake"
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
)

type BatchWorkflowTestSuite struct {
	suite.Suite
	temporalsdk_testsuite.WorkflowTestSuite

	env *temporalsdk_testsuite.TestWorkflowEnvironment

	workflow *BatchWorkflow
}

func processingChildWorkflow(ctx temporalsdk_workflow.Context, req *ingest.ProcessingWorkflowRequest) error {
	return nil
}

func (s *BatchWorkflowTestSuite) SetupWorkflowTest(cfg config.Configuration) {
	s.env = s.NewTestWorkflowEnvironment()
	s.env.SetStartTime(startTime)

	ctrl := gomock.NewController(s.T())
	ingestsvc := ingest_fake.NewMockService(ctrl)
	rng := rand.New(rand.NewSource(1)) // #nosec: G404

	s.env.RegisterWorkflowWithOptions(
		processingChildWorkflow,
		temporalsdk_workflow.RegisterOptions{Name: ingest.ProcessingWorkflowName},
	)

	s.workflow = NewBatchWorkflow(cfg, rng, ingestsvc, nil)
}

func (s *BatchWorkflowTestSuite) AfterTest(suiteName, testName string) {
	s.env.AssertExpectations(s.T())
}

func (s *BatchWorkflowTestSuite) ExecuteAndValidateWorkflow(
	req *ingest.BatchWorkflowRequest,
	shouldError bool,
) {
	s.env.ExecuteWorkflow(s.workflow.Execute, req)

	s.True(s.env.IsWorkflowCompleted())
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
// - Batch status updates (Processing -> Ingested).
// - SIP creation for each key.
// - Child processing workflows started for each SIP.
// - Waiting for all child workflows to complete.
func (s *BatchWorkflowTestSuite) TestBatch() {
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

	// Mock the processing workflow call for the first SIP.
	s.env.OnWorkflow(
		processingChildWorkflow,
		internalCtx,
		&ingest.ProcessingWorkflowRequest{
			SIPUUID:         batchSIP1UUID,
			SIPName:         batchSIP1Key,
			Key:             batchSIP1Key,
			SIPSourceID:     sourceID,
			Type:            enums.WorkflowTypeCreateAip,
			RetentionPeriod: -1 * time.Second,
			// BatchUUID:       batchUUID,
		},
	).Return(nil)

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

	// Mock the processing workflow call for the second SIP.
	s.env.OnWorkflow(
		processingChildWorkflow,
		internalCtx,
		&ingest.ProcessingWorkflowRequest{
			SIPUUID:         batchSIP2UUID,
			SIPName:         batchSIP2Key,
			Key:             batchSIP2Key,
			SIPSourceID:     sourceID,
			Type:            enums.WorkflowTypeCreateAip,
			RetentionPeriod: -1 * time.Second,
			// BatchUUID:       batchUUID,
		},
	).Return(nil)

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
	}, false)
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
	}, true)
}
