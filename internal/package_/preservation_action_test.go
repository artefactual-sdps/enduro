package package__test

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"go.artefactual.dev/tools/mockutil"
	"gotest.tools/v3/assert"

	"github.com/artefactual-sdps/enduro/internal/datatypes"
	"github.com/artefactual-sdps/enduro/internal/enums"
	persistence_fake "github.com/artefactual-sdps/enduro/internal/persistence/fake"
)

func TestCreatePreservationAction(t *testing.T) {
	t.Parallel()

	workflowID := "processing-workflow-720db1d4-825c-4911-9a20-61c212cf23ff"
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
		pa      datatypes.PreservationAction
		mock    func(*persistence_fake.MockService, datatypes.PreservationAction) *persistence_fake.MockService
		want    datatypes.PreservationAction
		wantErr string
	}
	for _, tt := range []test{
		{
			name: "Creates a preservation action",
			pa: datatypes.PreservationAction{
				WorkflowID: workflowID,
				PackageID:  1,
			},
			want: datatypes.PreservationAction{
				ID:         11,
				WorkflowID: workflowID,
				Type:       enums.PreservationActionTypeUnspecified,
				Status:     enums.PreservationActionStatusUnspecified,
				StartedAt: sql.NullTime{
					Time:  time.Date(2024, 6, 3, 9, 4, 23, 0, time.UTC),
					Valid: true,
				},
				PackageID: 1,
			},
			mock: func(svc *persistence_fake.MockService, pa datatypes.PreservationAction) *persistence_fake.MockService {
				svc.EXPECT().
					CreatePreservationAction(mockutil.Context(), &pa).
					DoAndReturn(
						func(ctx context.Context, pa *datatypes.PreservationAction) error {
							pa.ID = 11
							pa.StartedAt = sql.NullTime{
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
			name: "Creates a preservation action with optional values",
			pa: datatypes.PreservationAction{
				WorkflowID:  workflowID,
				Type:        enums.PreservationActionTypeCreateAip,
				Status:      enums.PreservationActionStatusDone,
				StartedAt:   startedAt,
				CompletedAt: completedAt,
				PackageID:   1,
			},
			want: datatypes.PreservationAction{
				ID:          11,
				WorkflowID:  workflowID,
				Type:        enums.PreservationActionTypeCreateAip,
				Status:      enums.PreservationActionStatusDone,
				StartedAt:   startedAt,
				CompletedAt: completedAt,
				PackageID:   1,
			},
			mock: func(svc *persistence_fake.MockService, pa datatypes.PreservationAction) *persistence_fake.MockService {
				svc.EXPECT().
					CreatePreservationAction(mockutil.Context(), &pa).
					DoAndReturn(
						func(ctx context.Context, pa *datatypes.PreservationAction) error {
							pa.ID = 11
							return nil
						},
					)
				return svc
			},
		},
		{
			name: "Errors when WorkflowID is missing",
			pa: datatypes.PreservationAction{
				PackageID: 1,
			},
			wantErr: "preservation action: create: invalid data error: field \"WorkflowID\" is required",
			mock: func(svc *persistence_fake.MockService, pa datatypes.PreservationAction) *persistence_fake.MockService {
				svc.EXPECT().
					CreatePreservationAction(mockutil.Context(), &pa).
					DoAndReturn(
						func(ctx context.Context, pa *datatypes.PreservationAction) error {
							return errors.New("invalid data error: field \"WorkflowID\" is required")
						},
					)
				return svc
			},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			pkgSvc, perSvc := testSvc(t, nil, 0)
			if tt.mock != nil {
				tt.mock(perSvc, tt.pa)
			}

			pa := tt.pa
			err := pkgSvc.CreatePreservationAction(context.Background(), &pa)

			if tt.wantErr != "" {
				assert.Error(t, err, tt.wantErr)
				return
			}

			assert.NilError(t, err)
			assert.DeepEqual(t, pa, tt.want)
		})
	}
}
