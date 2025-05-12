package workflow

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"go.artefactual.dev/tools/mockutil"
	temporalsdk_testsuite "go.temporal.io/sdk/testsuite"
	"go.uber.org/mock/gomock"
	"gotest.tools/v3/assert"

	w "github.com/artefactual-sdps/enduro/internal/datatypes"
	"github.com/artefactual-sdps/enduro/internal/enums"
	ingest_fake "github.com/artefactual-sdps/enduro/internal/ingest/fake"
)

func TestCreateWorkflowLocalActivity(t *testing.T) {
	t.Parallel()

	sipUUID := uuid.New()
	startedAt := time.Date(2024, 6, 13, 17, 50, 13, 0, time.UTC)
	completedAt := time.Date(2024, 6, 13, 17, 50, 14, 0, time.UTC)

	type test struct {
		name      string
		params    *createWorkflowLocalActivityParams
		mockCalls func(m *ingest_fake.MockServiceMockRecorder)
		want      uint
		wantErr   string
	}
	for _, tt := range []test{
		{
			name: "Creates a workflow",
			params: &createWorkflowLocalActivityParams{
				TemporalID:  "workflow-id",
				Type:        enums.WorkflowTypeCreateAip,
				Status:      enums.WorkflowStatusDone,
				StartedAt:   startedAt,
				CompletedAt: completedAt,
				SIPUUID:     sipUUID,
			},
			mockCalls: func(m *ingest_fake.MockServiceMockRecorder) {
				m.CreateWorkflow(mockutil.Context(), &w.Workflow{
					TemporalID:  "workflow-id",
					Type:        enums.WorkflowTypeCreateAip,
					Status:      enums.WorkflowStatusDone,
					StartedAt:   sql.NullTime{Time: startedAt, Valid: true},
					CompletedAt: sql.NullTime{Time: completedAt, Valid: true},
					SIPUUID:     sipUUID,
				}).DoAndReturn(func(ctx context.Context, w *w.Workflow) error {
					w.ID = 1
					return nil
				})
			},
			want: 1,
		},
		{
			name: "Does not pass zero dates",
			params: &createWorkflowLocalActivityParams{
				TemporalID: "workflow-id",
				Type:       enums.WorkflowTypeCreateAip,
				Status:     enums.WorkflowStatusDone,
				SIPUUID:    sipUUID,
			},
			mockCalls: func(m *ingest_fake.MockServiceMockRecorder) {
				m.CreateWorkflow(mockutil.Context(), &w.Workflow{
					TemporalID: "workflow-id",
					Type:       enums.WorkflowTypeCreateAip,
					Status:     enums.WorkflowStatusDone,
					SIPUUID:    sipUUID,
				}).DoAndReturn(func(ctx context.Context, w *w.Workflow) error {
					w.ID = 1
					return nil
				})
			},
			want: 1,
		},
		{
			name: "Fails if there is a persistence error",
			params: &createWorkflowLocalActivityParams{
				TemporalID: "workflow-id",
				Type:       enums.WorkflowTypeCreateAip,
				Status:     enums.WorkflowStatusDone,
				SIPUUID:    sipUUID,
			},
			mockCalls: func(m *ingest_fake.MockServiceMockRecorder) {
				m.CreateWorkflow(mockutil.Context(), &w.Workflow{
					TemporalID: "workflow-id",
					Type:       enums.WorkflowTypeCreateAip,
					Status:     enums.WorkflowStatusDone,
					SIPUUID:    sipUUID,
				}).Return(fmt.Errorf("persistence error"))
			},
			wantErr: "persistence error",
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ts := &temporalsdk_testsuite.WorkflowTestSuite{}
			env := ts.NewTestActivityEnvironment()
			svc := ingest_fake.NewMockService(gomock.NewController(t))
			if tt.mockCalls != nil {
				tt.mockCalls(svc.EXPECT())
			}

			enc, err := env.ExecuteLocalActivity(
				createWorkflowLocalActivity,
				svc,
				tt.params,
			)
			if tt.wantErr != "" {
				assert.ErrorContains(t, err, tt.wantErr)
				return
			}
			assert.NilError(t, err)

			var res uint
			_ = enc.Get(&res)
			assert.DeepEqual(t, res, tt.want)
		})
	}
}
