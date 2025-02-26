package workflow

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	"go.artefactual.dev/tools/mockutil"
	temporalsdk_testsuite "go.temporal.io/sdk/testsuite"
	"go.uber.org/mock/gomock"
	"gotest.tools/v3/assert"

	"github.com/artefactual-sdps/enduro/internal/datatypes"
	"github.com/artefactual-sdps/enduro/internal/enums"
	ingest_fake "github.com/artefactual-sdps/enduro/internal/ingest/fake"
)

func TestCreatePreservationActionLocalActivity(t *testing.T) {
	t.Parallel()

	startedAt := time.Date(2024, 6, 13, 17, 50, 13, 0, time.UTC)
	completedAt := time.Date(2024, 6, 13, 17, 50, 14, 0, time.UTC)

	type test struct {
		name      string
		params    *createPreservationActionLocalActivityParams
		mockCalls func(m *ingest_fake.MockServiceMockRecorder)
		want      uint
		wantErr   string
	}
	for _, tt := range []test{
		{
			name: "Creates a preservation action",
			params: &createPreservationActionLocalActivityParams{
				WorkflowID:  "workflow-id",
				Type:        enums.PreservationActionTypeCreateAip,
				Status:      enums.PreservationActionStatusDone,
				StartedAt:   startedAt,
				CompletedAt: completedAt,
				SIPID:       1,
			},
			mockCalls: func(m *ingest_fake.MockServiceMockRecorder) {
				m.CreatePreservationAction(mockutil.Context(), &datatypes.PreservationAction{
					WorkflowID:  "workflow-id",
					Type:        enums.PreservationActionTypeCreateAip,
					Status:      enums.PreservationActionStatusDone,
					StartedAt:   sql.NullTime{Time: startedAt, Valid: true},
					CompletedAt: sql.NullTime{Time: completedAt, Valid: true},
					SIPID:       1,
				}).DoAndReturn(func(ctx context.Context, pa *datatypes.PreservationAction) error {
					pa.ID = 1
					return nil
				})
			},
			want: 1,
		},
		{
			name: "Does not pass zero dates",
			params: &createPreservationActionLocalActivityParams{
				WorkflowID: "workflow-id",
				Type:       enums.PreservationActionTypeCreateAip,
				Status:     enums.PreservationActionStatusDone,
				SIPID:      1,
			},
			mockCalls: func(m *ingest_fake.MockServiceMockRecorder) {
				m.CreatePreservationAction(mockutil.Context(), &datatypes.PreservationAction{
					WorkflowID: "workflow-id",
					Type:       enums.PreservationActionTypeCreateAip,
					Status:     enums.PreservationActionStatusDone,
					SIPID:      1,
				}).DoAndReturn(func(ctx context.Context, pa *datatypes.PreservationAction) error {
					pa.ID = 1
					return nil
				})
			},
			want: 1,
		},
		{
			name: "Fails if there is a persistence error",
			params: &createPreservationActionLocalActivityParams{
				WorkflowID: "workflow-id",
				Type:       enums.PreservationActionTypeCreateAip,
				Status:     enums.PreservationActionStatusDone,
				SIPID:      1,
			},
			mockCalls: func(m *ingest_fake.MockServiceMockRecorder) {
				m.CreatePreservationAction(mockutil.Context(), &datatypes.PreservationAction{
					WorkflowID: "workflow-id",
					Type:       enums.PreservationActionTypeCreateAip,
					Status:     enums.PreservationActionStatusDone,
					SIPID:      1,
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
				createPreservationActionLocalActivity,
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
