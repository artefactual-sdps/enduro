package workflow

import (
	"fmt"
	"math/rand"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	temporalsdk_testsuite "go.temporal.io/sdk/testsuite"
	"go.uber.org/mock/gomock"

	"github.com/artefactual-sdps/enduro/internal/config"
	"github.com/artefactual-sdps/enduro/internal/datatypes"
	"github.com/artefactual-sdps/enduro/internal/enums"
	"github.com/artefactual-sdps/enduro/internal/ingest"
	ingest_fake "github.com/artefactual-sdps/enduro/internal/ingest/fake"
)

var (
	batchUUID = uuid.MustParse("f47ac10b-58cc-4372-a567-0e02b2c3d479")
	sourceID  = uuid.MustParse("6ba7b814-9dad-41d1-80b4-00c04fd430c8")
)

type BatchWorkflowTestSuite struct {
	suite.Suite
	temporalsdk_testsuite.WorkflowTestSuite

	env *temporalsdk_testsuite.TestWorkflowEnvironment

	workflow *BatchWorkflow
}

func (s *BatchWorkflowTestSuite) SetupWorkflowTest(cfg config.Configuration) {
	s.env = s.NewTestWorkflowEnvironment()
	s.env.SetStartTime(startTime)

	ctrl := gomock.NewController(s.T())
	ingestsvc := ingest_fake.NewMockService(ctrl)
	rng := rand.New(rand.NewSource(1)) // #nosec: G404

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
func (s *BatchWorkflowTestSuite) TestBatch() {
	s.SetupWorkflowTest(config.Configuration{})

	s.env.OnActivity(
		updateBatchLocalActivity,
		mock.AnythingOfType("*context.valueCtx"),
		s.workflow.ingestsvc,
		&updateBatchLocalActivityParams{
			UUID:      batchUUID,
			Status:    enums.BatchStatusProcessing,
			StartedAt: startTime,
		},
	).Return(&updateBatchLocalActivityResult{}, nil)

	s.env.OnActivity(
		updateBatchLocalActivity,
		mock.AnythingOfType("*context.valueCtx"),
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
			Identifier: "test-batch",
			Status:     enums.BatchStatusQueued,
			CreatedAt:  startTime,
			SIPSCount:  2,
		},
		SIPSourceID: sourceID,
		Keys:        []string{"sip1.zip", "sip2.zip"},
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
		mock.AnythingOfType("*context.valueCtx"),
		s.workflow.ingestsvc,
		mock.AnythingOfType("*workflow.updateBatchLocalActivityParams"),
	).Return(nil, fmt.Errorf("update error")).Times(6)

	s.ExecuteAndValidateWorkflow(&ingest.BatchWorkflowRequest{
		Batch: datatypes.Batch{
			UUID:       batchUUID,
			Identifier: "test-batch",
			Status:     enums.BatchStatusQueued,
			CreatedAt:  startTime,
			SIPSCount:  2,
		},
		SIPSourceID: sourceID,
		Keys:        []string{"sip1.zip", "sip2.zip"},
	}, true)
}
