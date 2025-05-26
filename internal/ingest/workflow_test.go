package ingest_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"go.artefactual.dev/tools/mockutil"
	"gotest.tools/v3/assert"

	"github.com/artefactual-sdps/enduro/internal/datatypes"
	"github.com/artefactual-sdps/enduro/internal/enums"
	persistence_fake "github.com/artefactual-sdps/enduro/internal/persistence/fake"
)

func TestCreateWorkflow(t *testing.T) {
	t.Parallel()

	sipUUID := uuid.New()
	temporalID := "processing-workflow-720db1d4-825c-4911-9a20-61c212cf23ff"
	startedAt := sql.NullTime{
		Time:  time.Date(2024, 6, 3, 8, 51, 35, 0, time.UTC),
		Valid: true,
	}
	completedAt := sql.NullTime{
		Time:  time.Date(2024, 6, 3, 8, 52, 18, 0, time.UTC),
		Valid: true,
	}

	type test struct {
		name    string
		w       datatypes.Workflow
		mock    func(*persistence_fake.MockService, datatypes.Workflow) *persistence_fake.MockService
		want    datatypes.Workflow
		wantErr string
	}
	for _, tt := range []test{
		{
			name: "Creates a workflow",
			w: datatypes.Workflow{
				TemporalID: temporalID,
				SIPUUID:    sipUUID,
				Type:       enums.WorkflowTypeCreateAip,
			},
			want: datatypes.Workflow{
				ID:         11,
				TemporalID: temporalID,
				Type:       enums.WorkflowTypeCreateAip,
				Status:     enums.WorkflowStatusUnspecified,
				StartedAt: sql.NullTime{
					Time:  time.Date(2024, 6, 3, 9, 4, 23, 0, time.UTC),
					Valid: true,
				},
				SIPUUID: sipUUID,
			},
			mock: func(svc *persistence_fake.MockService, w datatypes.Workflow) *persistence_fake.MockService {
				svc.EXPECT().
					CreateWorkflow(mockutil.Context(), &w).
					DoAndReturn(
						func(ctx context.Context, w *datatypes.Workflow) error {
							w.ID = 11
							w.StartedAt = sql.NullTime{
								Time:  time.Date(2024, 6, 3, 9, 4, 23, 0, time.UTC),
								Valid: true,
							}
							return nil
						},
					)
				return svc
			},
		},
		{
			name: "Creates a workflow with optional values",
			w: datatypes.Workflow{
				TemporalID:  temporalID,
				Type:        enums.WorkflowTypeCreateAip,
				Status:      enums.WorkflowStatusDone,
				StartedAt:   startedAt,
				CompletedAt: completedAt,
				SIPUUID:     sipUUID,
			},
			want: datatypes.Workflow{
				ID:          11,
				TemporalID:  temporalID,
				Type:        enums.WorkflowTypeCreateAip,
				Status:      enums.WorkflowStatusDone,
				StartedAt:   startedAt,
				CompletedAt: completedAt,
				SIPUUID:     sipUUID,
			},
			mock: func(svc *persistence_fake.MockService, w datatypes.Workflow) *persistence_fake.MockService {
				svc.EXPECT().
					CreateWorkflow(mockutil.Context(), &w).
					DoAndReturn(
						func(ctx context.Context, w *datatypes.Workflow) error {
							w.ID = 11
							return nil
						},
					)
				return svc
			},
		},
		{
			name: "Errors when TemporalID is missing",
			w: datatypes.Workflow{
				SIPUUID: sipUUID,
			},
			wantErr: "workflow: create: invalid data error: field \"TemporalID\" is required",
			mock: func(svc *persistence_fake.MockService, w datatypes.Workflow) *persistence_fake.MockService {
				svc.EXPECT().
					CreateWorkflow(mockutil.Context(), &w).
					DoAndReturn(
						func(ctx context.Context, w *datatypes.Workflow) error {
							return errors.New("invalid data error: field \"TemporalID\" is required")
						},
					)
				return svc
			},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ingestsvc, perSvc, _ := testSvc(t, nil, 0)
			if tt.mock != nil {
				tt.mock(perSvc, tt.w)
			}

			w := tt.w
			err := ingestsvc.CreateWorkflow(context.Background(), &w)

			if tt.wantErr != "" {
				assert.Error(t, err, tt.wantErr)
				return
			}

			assert.NilError(t, err)
			assert.DeepEqual(t, w, tt.want)
		})
	}
}
