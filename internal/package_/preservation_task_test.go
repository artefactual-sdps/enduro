package package__test

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	"go.artefactual.dev/tools/mockutil"
	"gotest.tools/v3/assert"

	"github.com/artefactual-sdps/enduro/internal/datatypes"
	"github.com/artefactual-sdps/enduro/internal/enums"
	persistence_fake "github.com/artefactual-sdps/enduro/internal/persistence/fake"
)

func TestCreatePreservationTask(t *testing.T) {
	taskID := "a499e8fc-7309-4e26-b39d-d8ab68466c27"

	type test struct {
		name    string
		pt      datatypes.PreservationTask
		mock    func(*persistence_fake.MockService, datatypes.PreservationTask) *persistence_fake.MockService
		want    datatypes.PreservationTask
		wantErr string
	}
	for _, tt := range []test{
		{
			name: "Creates a preservation task",
			pt: datatypes.PreservationTask{
				TaskID:               taskID,
				Name:                 "PT1",
				Status:               enums.PreservationTaskStatusInProgress,
				PreservationActionID: 11,
			},
			want: datatypes.PreservationTask{
				ID:                   1,
				TaskID:               taskID,
				Name:                 "PT1",
				Status:               enums.PreservationTaskStatusInProgress,
				PreservationActionID: 11,
			},
			mock: func(svc *persistence_fake.MockService, pt datatypes.PreservationTask) *persistence_fake.MockService {
				svc.EXPECT().
					CreatePreservationTask(mockutil.Context(), &pt).
					DoAndReturn(
						func(ctx context.Context, pt *datatypes.PreservationTask) error {
							pt.ID = 1
							return nil
						},
					)
				return svc
			},
		},
		{
			name: "Creates a preservation task with optional values",
			pt: datatypes.PreservationTask{
				TaskID: taskID,
				Name:   "PT2",
				Status: enums.PreservationTaskStatusInProgress,
				StartedAt: sql.NullTime{
					Time:  time.Date(2024, 3, 27, 11, 32, 41, 0, time.UTC),
					Valid: true,
				},
				CompletedAt: sql.NullTime{
					Time:  time.Date(2024, 3, 27, 11, 32, 43, 0, time.UTC),
					Valid: true,
				},
				Note:                 "PT2 Note",
				PreservationActionID: 12,
			},
			mock: func(svc *persistence_fake.MockService, pt datatypes.PreservationTask) *persistence_fake.MockService {
				svc.EXPECT().
					CreatePreservationTask(mockutil.Context(), &pt).
					DoAndReturn(
						func(ctx context.Context, pt *datatypes.PreservationTask) error {
							pt.ID = 2
							return nil
						},
					)
				return svc
			},
			want: datatypes.PreservationTask{
				ID:     2,
				TaskID: taskID,
				Name:   "PT2",
				Status: enums.PreservationTaskStatusInProgress,
				StartedAt: sql.NullTime{
					Time:  time.Date(2024, 3, 27, 11, 32, 41, 0, time.UTC),
					Valid: true,
				},
				CompletedAt: sql.NullTime{
					Time:  time.Date(2024, 3, 27, 11, 32, 43, 0, time.UTC),
					Valid: true,
				},
				Note:                 "PT2 Note",
				PreservationActionID: 12,
			},
		},
		{
			name: "Errors creating a package with a missing TaskID",
			pt:   datatypes.PreservationTask{},
			mock: func(svc *persistence_fake.MockService, pt datatypes.PreservationTask) *persistence_fake.MockService {
				svc.EXPECT().
					CreatePreservationTask(mockutil.Context(), &pt).
					Return(
						fmt.Errorf("invalid data error: field \"TaskID\" is required"),
					)
				return svc
			},
			wantErr: "preservation task: create: invalid data error: field \"TaskID\" is required",
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			pkgSvc, perSvc := testSvc(t)
			if tt.mock != nil {
				tt.mock(perSvc, tt.pt)
			}

			pt := tt.pt
			err := pkgSvc.CreatePreservationTask(context.Background(), &pt)

			if tt.wantErr != "" {
				assert.Error(t, err, tt.wantErr)
				return
			}

			assert.NilError(t, err)
			assert.DeepEqual(t, pt, tt.want)
		})
	}
}
