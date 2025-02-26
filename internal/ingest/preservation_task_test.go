package ingest_test

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"testing"
	"time"

	"go.artefactual.dev/tools/mockutil"
	"go.artefactual.dev/tools/ref"
	"gotest.tools/v3/assert"

	"github.com/artefactual-sdps/enduro/internal/datatypes"
	"github.com/artefactual-sdps/enduro/internal/enums"
	"github.com/artefactual-sdps/enduro/internal/persistence"
	persistence_fake "github.com/artefactual-sdps/enduro/internal/persistence/fake"
)

func TestCreatePreservationTask(t *testing.T) {
	t.Parallel()

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
			name: "Errors creating a preservation task with a missing TaskID",
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

			ingestsvc, perSvc := testSvc(t, nil, 0)
			if tt.mock != nil {
				tt.mock(perSvc, tt.pt)
			}

			pt := tt.pt
			err := ingestsvc.CreatePreservationTask(context.Background(), &pt)

			if tt.wantErr != "" {
				assert.Error(t, err, tt.wantErr)
				return
			}

			assert.NilError(t, err)
			assert.DeepEqual(t, pt, tt.want)
		})
	}
}

func TestCompletePreservationTask(t *testing.T) {
	t.Parallel()

	completedAt := time.Date(2024, 4, 2, 10, 35, 32, 0, time.UTC)

	type args struct {
		id          int
		status      enums.PreservationTaskStatus
		completedAt time.Time
		note        *string
	}
	type test struct {
		name string
		args args
		mock func(
			*persistence_fake.MockService,
			*datatypes.PreservationTask,
		) *persistence_fake.MockService
		want    datatypes.PreservationTask
		wantErr string
	}
	for _, tt := range []test{
		{
			name: "Updates a preservation task with note",
			args: args{
				id:          1,
				status:      enums.PreservationTaskStatusDone,
				completedAt: completedAt,
				note:        ref.New("Reviewed and accepted"),
			},
			mock: func(
				svc *persistence_fake.MockService,
				pt *datatypes.PreservationTask,
			) *persistence_fake.MockService {
				svc.EXPECT().
					UpdatePreservationTask(
						mockutil.Context(),
						1,
						mockutil.Func(
							"should update preservation task",
							func(updater persistence.PresTaskUpdater) error {
								_, err := updater(&datatypes.PreservationTask{})
								return err
							},
						),
					).
					DoAndReturn(
						func(
							ctx context.Context,
							id int,
							updater persistence.PresTaskUpdater,
						) (*datatypes.PreservationTask, error) {
							pt, err := updater(pt)
							return pt, err
						},
					)
				return svc
			},
			want: datatypes.PreservationTask{
				ID:     1,
				Status: enums.PreservationTaskStatusDone,
				CompletedAt: sql.NullTime{
					Time:  completedAt,
					Valid: true,
				},
				Note: "Reviewed and accepted",
			},
		},
		{
			name: "Updates a preservation task without note",
			args: args{
				id:          1,
				status:      enums.PreservationTaskStatusDone,
				completedAt: completedAt,
			},
			mock: func(
				svc *persistence_fake.MockService,
				pt *datatypes.PreservationTask,
			) *persistence_fake.MockService {
				svc.EXPECT().
					UpdatePreservationTask(
						mockutil.Context(),
						1,
						mockutil.Func(
							"should update preservation task",
							func(updater persistence.PresTaskUpdater) error {
								_, err := updater(&datatypes.PreservationTask{})
								return err
							},
						),
					).
					DoAndReturn(
						func(
							ctx context.Context,
							id int,
							updater persistence.PresTaskUpdater,
						) (*datatypes.PreservationTask, error) {
							pt, err := updater(pt)
							return pt, err
						},
					)
				return svc
			},
			want: datatypes.PreservationTask{
				ID:     1,
				Status: enums.PreservationTaskStatusDone,
				CompletedAt: sql.NullTime{
					Time:  completedAt,
					Valid: true,
				},
			},
		},
		{
			name: "Errors on persistence error",
			args: args{id: 2},
			mock: func(
				svc *persistence_fake.MockService,
				pt *datatypes.PreservationTask,
			) *persistence_fake.MockService {
				svc.EXPECT().
					UpdatePreservationTask(
						mockutil.Context(),
						2,
						mockutil.Func(
							"should update preservation task",
							func(updater persistence.PresTaskUpdater) error {
								_, err := updater(&datatypes.PreservationTask{})
								return err
							},
						),
					).
					Return(
						nil,
						errors.New("not found error: db: preservation_task not found"),
					)
				return svc
			},
			wantErr: "error updating preservation task: not found error: db: preservation_task not found",
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ingestsvc, perSvc := testSvc(t, nil, 0)
			pt := datatypes.PreservationTask{
				ID: 1,
			}
			if tt.mock != nil {
				tt.mock(perSvc, &pt)
			}

			err := ingestsvc.CompletePreservationTask(
				context.Background(),
				tt.args.id,
				tt.args.status,
				tt.args.completedAt,
				tt.args.note,
			)

			if tt.wantErr != "" {
				assert.Error(t, err, tt.wantErr)
				return
			}

			assert.NilError(t, err)
			assert.DeepEqual(t, pt, tt.want)
		})
	}
}
