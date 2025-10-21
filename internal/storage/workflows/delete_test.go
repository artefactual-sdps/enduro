package workflows

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jonboulle/clockwork"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.artefactual.dev/tools/ref"
	temporalsdk_activity "go.temporal.io/sdk/activity"
	temporalsdk_testsuite "go.temporal.io/sdk/testsuite"
	"go.uber.org/mock/gomock"
	_ "gocloud.dev/blob/memblob"

	goastorage "github.com/artefactual-sdps/enduro/internal/api/gen/storage"
	"github.com/artefactual-sdps/enduro/internal/datatypes"
	"github.com/artefactual-sdps/enduro/internal/storage"
	"github.com/artefactual-sdps/enduro/internal/storage/activities"
	"github.com/artefactual-sdps/enduro/internal/storage/enums"
	"github.com/artefactual-sdps/enduro/internal/storage/fake"
	"github.com/artefactual-sdps/enduro/internal/storage/types"
)

const (
	workflowDBID        = 1
	deletionRequestDBID = 2
)

type StorageDeleteWorkflowTestSuite struct {
	env        *temporalsdk_testsuite.TestWorkflowEnvironment
	mockCtrl   *gomock.Controller
	storagesvc *fake.MockService

	req        *storage.StorageDeleteWorkflowRequest
	aip        *goastorage.AIP
	reviewTask *datatypes.Task
}

func NewStorageDeleteWorkflowTestSuite(
	t *testing.T,
	req *storage.StorageDeleteWorkflowRequest,
) *StorageDeleteWorkflowTestSuite {
	s := StorageDeleteWorkflowTestSuite{}

	ts := temporalsdk_testsuite.WorkflowTestSuite{}
	s.env = ts.NewTestWorkflowEnvironment()
	s.mockCtrl = gomock.NewController(t)
	s.storagesvc = fake.NewMockService(s.mockCtrl)
	s.req = req
	s.aip = &goastorage.AIP{
		UUID:         s.req.AIPID,
		LocationUUID: ref.New(uuid.New()),
		Status:       enums.AIPStatusStored.String(),
	}
	s.reviewTask = &datatypes.Task{ID: 1}

	s.env.RegisterActivityWithOptions(
		activities.NewDeleteFromAMSSLocationActivity(false, time.Microsecond*1).Execute,
		temporalsdk_activity.RegisterOptions{Name: storage.DeleteFromAMSSLocationActivityName},
	)
	s.env.RegisterActivityWithOptions(
		activities.NewAIPDeletionReportActivity(
			clockwork.NewFakeClock(),
			storage.AIPDeletionConfig{ReportTemplatePath: "../../../assets/Enduro_AIP_deletion_report_v3.tmpl.pdf"},
			s.storagesvc,
		).Execute,
		temporalsdk_activity.RegisterOptions{Name: activities.AIPDeletionReportActivityName},
	)

	return &s
}

// createDeletionRequest mocks the activities called when creating a
// deletion request.
func (s *StorageDeleteWorkflowTestSuite) createDeletionRequest() {
	s.env.OnActivity(
		storage.ReadAIPLocalActivity,
		mock.AnythingOfType("*context.valueCtx"),
		s.storagesvc,
		s.aip.UUID,
	).Return(s.aip, nil)

	s.env.OnActivity(
		storage.UpdateAIPStatusLocalActivity,
		mock.AnythingOfType("*context.valueCtx"),
		s.storagesvc,
		&storage.UpdateAIPStatusLocalActivityParams{
			AIPID:  s.aip.UUID,
			Status: enums.AIPStatusProcessing,
		},
	).Return(nil)

	s.env.OnActivity(
		storage.CreateWorkflowLocalActivity,
		mock.AnythingOfType("*context.valueCtx"),
		s.storagesvc,
		&storage.CreateWorkflowLocalActivityParams{
			AIPID:      s.aip.UUID,
			TemporalID: "default-test-workflow-id",
			Type:       enums.WorkflowTypeDeleteAip,
		},
	).Return(workflowDBID, nil)

	s.reviewTask.Note = fmt.Sprintf(
		"An AIP deletion has been requested by %s. Reason:\n\n%s",
		s.req.UserEmail,
		s.req.Reason,
	)
	s.env.OnActivity(
		storage.CreateTaskLocalActivity,
		mock.AnythingOfType("*context.valueCtx"),
		s.storagesvc,
		&storage.CreateTaskLocalActivityParams{
			WorkflowDBID: workflowDBID,
			Status:       enums.TaskStatusPending,
			Name:         "Review AIP deletion request",
			Note:         fmt.Sprintf("%s\n\nAwaiting user review.", s.reviewTask.Note),
		},
	).Return(s.reviewTask.ID, nil)

	s.env.OnActivity(
		storage.CreateDeletionRequestLocalActivity,
		mock.AnythingOfType("*context.valueCtx"),
		s.storagesvc,
		&storage.CreateDeletionRequestLocalActivityParams{
			Requester:    s.req.UserEmail,
			RequesterIss: s.req.UserIss,
			RequesterSub: s.req.UserSub,
			Reason:       s.req.Reason,
			WorkflowDBID: workflowDBID,
			AIPUUID:      s.req.AIPID,
		},
	).Return(deletionRequestDBID, nil)

	s.env.OnActivity(
		storage.UpdateWorkflowStatusLocalActivity,
		mock.AnythingOfType("*context.valueCtx"),
		s.storagesvc,
		&storage.UpdateWorkflowStatusLocalActivityParams{
			DBID:   workflowDBID,
			Status: enums.WorkflowStatusPending,
		},
	).Return(nil)

	s.env.OnActivity(
		storage.UpdateAIPStatusLocalActivity,
		mock.AnythingOfType("*context.valueCtx"),
		s.storagesvc,
		&storage.UpdateAIPStatusLocalActivityParams{
			AIPID:  s.aip.UUID,
			Status: enums.AIPStatusPending,
		},
	).Return(nil)
}

func TestStorageDeleteWorkflow(t *testing.T) {
	t.Parallel()

	t.Run("Create and approve deletion request", func(t *testing.T) {
		t.Parallel()

		req := storage.StorageDeleteWorkflowRequest{
			AIPID:     uuid.New(),
			Reason:    "Reason",
			UserEmail: "requester@example.com",
			UserSub:   "subject",
			UserIss:   "issuer",
			TaskQueue: "global",
		}

		signal := storage.DeletionDecisionSignal{
			Status:    enums.DeletionRequestStatusApproved,
			UserEmail: "reviewer@example.com",
			UserIss:   "issuer",
			UserSub:   "subject-2",
		}

		locationInfo := &storage.ReadLocationInfoLocalActivityResult{
			Source: enums.LocationSourceAmss,
			Config: types.LocationConfig{Value: &types.AMSSConfig{
				URL:      "http://127.0.0.1:62081",
				Username: "test",
				APIKey:   "secret",
			}},
		}

		deleteTaskDBID := 3

		s := NewStorageDeleteWorkflowTestSuite(t, &req)
		s.createDeletionRequest()

		s.env.RegisterDelayedCallback(
			func() {
				s.env.SignalWorkflow(storage.DeletionDecisionSignalName, signal)
			},
			0,
		)

		s.env.OnActivity(
			storage.UpdateAIPStatusLocalActivity,
			mock.AnythingOfType("*context.valueCtx"),
			s.storagesvc,
			&storage.UpdateAIPStatusLocalActivityParams{
				AIPID:  s.aip.UUID,
				Status: enums.AIPStatusProcessing,
			},
		).Return(nil)

		s.env.OnActivity(
			storage.UpdateWorkflowStatusLocalActivity,
			mock.AnythingOfType("*context.valueCtx"),
			s.storagesvc,
			&storage.UpdateWorkflowStatusLocalActivityParams{
				DBID:   workflowDBID,
				Status: enums.WorkflowStatusInProgress,
			},
		).Return(nil)

		s.env.OnActivity(
			storage.UpdateDeletionRequestLocalActivity,
			mock.AnythingOfType("*context.valueCtx"),
			s.storagesvc,
			deletionRequestDBID,
			signal,
		).Return(nil)

		s.env.OnActivity(
			storage.CompleteTaskLocalActivity,
			mock.AnythingOfType("*context.valueCtx"),
			s.storagesvc,
			&storage.CompleteTaskLocalActivityParams{
				DBID:   s.reviewTask.ID,
				Status: enums.TaskStatusDone,
				Note:   fmt.Sprintf("%s\n\nAIP deletion request approved by %s.", s.reviewTask.Note, signal.UserEmail),
			},
		).Return(nil)

		s.env.OnActivity(
			storage.CreateTaskLocalActivity,
			mock.AnythingOfType("*context.valueCtx"),
			s.storagesvc,
			&storage.CreateTaskLocalActivityParams{
				WorkflowDBID: workflowDBID,
				Status:       enums.TaskStatusInProgress,
				Name:         "Delete AIP",
				Note:         "Deleting AIP",
			},
		).Return(deleteTaskDBID, nil)

		s.env.OnActivity(
			storage.ReadLocationInfoLocalActivity,
			mock.AnythingOfType("*context.valueCtx"),
			s.storagesvc,
			*s.aip.LocationUUID,
		).Return(locationInfo, nil)

		s.env.OnActivity(
			storage.DeleteFromAMSSLocationActivityName,
			mock.AnythingOfType("*context.timerCtx"),
			&activities.DeleteFromAMSSLocationActivityParams{
				Config:  *locationInfo.Config.Value.(*types.AMSSConfig),
				AIPUUID: s.aip.UUID,
			},
		).Return(&activities.DeleteFromAMSSLocationActivityResult{Deleted: true}, nil)

		s.env.OnActivity(
			storage.CompleteTaskLocalActivity,
			mock.AnythingOfType("*context.valueCtx"),
			s.storagesvc,
			&storage.CompleteTaskLocalActivityParams{
				DBID:   deleteTaskDBID,
				Status: enums.TaskStatusDone,
				Note: fmt.Sprintf(
					"AIP deleted from %s source location",
					strings.ToUpper(locationInfo.Source.String()),
				),
			},
		).Return(nil)

		s.env.OnActivity(
			activities.AIPDeletionReportActivityName,
			mock.AnythingOfType("*context.timerCtx"),
			&activities.AIPDeletionReportActivityParams{
				AIPID:          s.aip.UUID,
				LocationSource: enums.LocationSourceAmss,
			},
		).Return(
			&activities.AIPDeletionReportActivityResult{
				Key: fmt.Sprintf("aip_deletion_report_%s.pdf", s.aip.UUID),
			},
			nil,
		)

		s.env.OnActivity(
			storage.CompleteWorkflowLocalActivity,
			mock.AnythingOfType("*context.valueCtx"),
			s.storagesvc,
			&storage.CompleteWorkflowLocalActivityParams{
				DBID:   workflowDBID,
				Status: enums.WorkflowStatusDone,
			},
		).Return(nil)

		s.env.OnActivity(
			storage.UpdateAIPStatusLocalActivity,
			mock.AnythingOfType("*context.valueCtx"),
			s.storagesvc,
			&storage.UpdateAIPStatusLocalActivityParams{
				AIPID:  s.aip.UUID,
				Status: enums.AIPStatusDeleted,
			},
		).Return(nil)

		s.env.ExecuteWorkflow(
			NewStorageDeleteWorkflow(
				storage.AIPDeletionConfig{
					ReportTemplatePath: "../../../assets/Enduro_AIP_deletion_report_v3.tmpl.pdf",
				},
				s.storagesvc,
			).Execute,
			req,
		)

		require.True(t, s.env.IsWorkflowCompleted())
		err := s.env.GetWorkflowResult(nil)
		require.NoError(t, err)
	})

	t.Run("Create and cancel deletion request", func(t *testing.T) {
		t.Parallel()

		req := storage.StorageDeleteWorkflowRequest{
			AIPID:     uuid.New(),
			Reason:    "Reason",
			UserEmail: "requester@example.com",
			UserSub:   "subject",
			UserIss:   "issuer",
			TaskQueue: "global",
		}

		signal := storage.DeletionDecisionSignal{
			Status:    enums.DeletionRequestStatusCanceled,
			UserEmail: "requester@example.com",
			UserIss:   "issuer",
			UserSub:   "subject",
		}

		s := NewStorageDeleteWorkflowTestSuite(t, &req)
		s.createDeletionRequest()

		s.env.RegisterDelayedCallback(
			func() {
				s.env.SignalWorkflow(storage.DeletionDecisionSignalName, signal)
			},
			0,
		)

		s.env.OnActivity(
			storage.UpdateAIPStatusLocalActivity,
			mock.AnythingOfType("*context.valueCtx"),
			s.storagesvc,
			&storage.UpdateAIPStatusLocalActivityParams{
				AIPID:  s.aip.UUID,
				Status: enums.AIPStatusProcessing,
			},
		).Return(nil)

		s.env.OnActivity(
			storage.UpdateWorkflowStatusLocalActivity,
			mock.AnythingOfType("*context.valueCtx"),
			s.storagesvc,
			&storage.UpdateWorkflowStatusLocalActivityParams{
				DBID:   workflowDBID,
				Status: enums.WorkflowStatusInProgress,
			},
		).Return(nil)

		s.env.OnActivity(
			storage.UpdateDeletionRequestLocalActivity,
			mock.AnythingOfType("*context.valueCtx"),
			s.storagesvc,
			deletionRequestDBID,
			signal,
		).Return(nil)

		s.env.OnActivity(
			storage.CompleteTaskLocalActivity,
			mock.AnythingOfType("*context.valueCtx"),
			s.storagesvc,
			&storage.CompleteTaskLocalActivityParams{
				DBID:   s.reviewTask.ID,
				Status: enums.TaskStatusDone,
				Note:   fmt.Sprintf("%s\n\nAIP deletion request canceled by %s.", s.reviewTask.Note, signal.UserEmail),
			},
		).Return(nil)

		// These activities are from the deferred workflow completion callback.
		s.env.OnActivity(
			storage.CompleteWorkflowLocalActivity,
			mock.AnythingOfType("*context.valueCtx"),
			s.storagesvc,
			&storage.CompleteWorkflowLocalActivityParams{
				DBID:   workflowDBID,
				Status: enums.WorkflowStatusCanceled,
			},
		).Return(nil)

		s.env.OnActivity(
			storage.UpdateAIPStatusLocalActivity,
			mock.AnythingOfType("*context.valueCtx"),
			s.storagesvc,
			&storage.UpdateAIPStatusLocalActivityParams{
				AIPID:  s.aip.UUID,
				Status: enums.AIPStatusStored,
			},
		).Return(nil)

		s.env.ExecuteWorkflow(
			NewStorageDeleteWorkflow(
				storage.AIPDeletionConfig{
					ReportTemplatePath: "../../../assets/Enduro_AIP_deletion_report_v3.tmpl.pdf",
				},
				s.storagesvc,
			).Execute,
			req,
		)

		require.True(t, s.env.IsWorkflowCompleted())
		err := s.env.GetWorkflowResult(nil)
		require.ErrorContains(t, err, "canceled")
	})
}
