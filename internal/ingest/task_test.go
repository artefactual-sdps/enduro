package ingest_test

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"go.artefactual.dev/tools/mockutil"
	"go.artefactual.dev/tools/ref"
	"gotest.tools/v3/assert"

	"github.com/artefactual-sdps/enduro/internal/datatypes"
	"github.com/artefactual-sdps/enduro/internal/enums"
	"github.com/artefactual-sdps/enduro/internal/persistence"
	persistence_fake "github.com/artefactual-sdps/enduro/internal/persistence/fake"
)

func TestCreateTask(t *testing.T) {
	t.Parallel()

	taskUUID := uuid.New()
	wUUID := uuid.New()

	type test struct {
		name    string
		task    datatypes.Task
		mock    func(*persistence_fake.MockService, datatypes.Task) *persistence_fake.MockService
		want    datatypes.Task
		wantErr string
	}
	for _, tt := range []test{
		{
			name: "Creates a task",
			task: datatypes.Task{
				UUID:         taskUUID,
				Name:         "PT1",
				Status:       enums.TaskStatusInProgress,
				WorkflowUUID: wUUID,
			},
			want: datatypes.Task{
				ID:           1,
				UUID:         taskUUID,
				Name:         "PT1",
				Status:       enums.TaskStatusInProgress,
				WorkflowUUID: wUUID,
			},
			mock: func(svc *persistence_fake.MockService, task datatypes.Task) *persistence_fake.MockService {
				svc.EXPECT().
					CreateTask(mockutil.Context(), &task).
					DoAndReturn(
						func(ctx context.Context, task *datatypes.Task) error {
							task.ID = 1
							return nil
						},
					)
				return svc
			},
		},
		{
			name: "Creates a task with optional values",
			task: datatypes.Task{
				UUID:   taskUUID,
				Name:   "PT2",
				Status: enums.TaskStatusInProgress,
				StartedAt: sql.NullTime{
					Time:  time.Date(2024, 3, 27, 11, 32, 41, 0, time.UTC),
					Valid: true,
				},
				CompletedAt: sql.NullTime{
					Time:  time.Date(2024, 3, 27, 11, 32, 43, 0, time.UTC),
					Valid: true,
				},
				Note:         "PT2 Note",
				WorkflowUUID: wUUID,
			},
			mock: func(svc *persistence_fake.MockService, task datatypes.Task) *persistence_fake.MockService {
				svc.EXPECT().
					CreateTask(mockutil.Context(), &task).
					DoAndReturn(
						func(ctx context.Context, task *datatypes.Task) error {
							task.ID = 2
							return nil
						},
					)
				return svc
			},
			want: datatypes.Task{
				ID:     2,
				UUID:   taskUUID,
				Name:   "PT2",
				Status: enums.TaskStatusInProgress,
				StartedAt: sql.NullTime{
					Time:  time.Date(2024, 3, 27, 11, 32, 41, 0, time.UTC),
					Valid: true,
				},
				CompletedAt: sql.NullTime{
					Time:  time.Date(2024, 3, 27, 11, 32, 43, 0, time.UTC),
					Valid: true,
				},
				Note:         "PT2 Note",
				WorkflowUUID: wUUID,
			},
		},
		{
			name: "Errors creating a task with a missing TaskID",
			task: datatypes.Task{},
			mock: func(svc *persistence_fake.MockService, task datatypes.Task) *persistence_fake.MockService {
				svc.EXPECT().
					CreateTask(mockutil.Context(), &task).
					Return(
						fmt.Errorf("invalid data error: field \"TaskID\" is required"),
					)
				return svc
			},
			wantErr: "task: create: invalid data error: field \"TaskID\" is required",
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ingestsvc, perSvc, _ := testSvc(t, nil, 0)
			if tt.mock != nil {
				tt.mock(perSvc, tt.task)
			}

			task := tt.task
			err := ingestsvc.CreateTask(context.Background(), &task)

			if tt.wantErr != "" {
				assert.Error(t, err, tt.wantErr)
				return
			}

			assert.NilError(t, err)
			assert.DeepEqual(t, task, tt.want)
		})
	}
}

func TestCompleteTask(t *testing.T) {
	t.Parallel()

	completedAt := time.Date(2024, 4, 2, 10, 35, 32, 0, time.UTC)

	type args struct {
		id          int
		status      enums.TaskStatus
		completedAt time.Time
		note        *string
	}
	type test struct {
		name string
		args args
		mock func(
			*persistence_fake.MockService,
			*datatypes.Task,
		) *persistence_fake.MockService
		want    datatypes.Task
		wantErr string
	}
	for _, tt := range []test{
		{
			name: "Updates a task with note",
			args: args{
				id:          1,
				status:      enums.TaskStatusDone,
				completedAt: completedAt,
				note:        ref.New("Reviewed and accepted"),
			},
			mock: func(
				svc *persistence_fake.MockService,
				task *datatypes.Task,
			) *persistence_fake.MockService {
				svc.EXPECT().
					UpdateTask(
						mockutil.Context(),
						1,
						mockutil.Func(
							"should update task",
							func(updater persistence.TaskUpdater) error {
								_, err := updater(&datatypes.Task{})
								return err
							},
						),
					).
					DoAndReturn(
						func(
							ctx context.Context,
							id int,
							updater persistence.TaskUpdater,
						) (*datatypes.Task, error) {
							task, err := updater(task)
							return task, err
						},
					)
				return svc
			},
			want: datatypes.Task{
				ID:     1,
				Status: enums.TaskStatusDone,
				CompletedAt: sql.NullTime{
					Time:  completedAt,
					Valid: true,
				},
				Note: "Reviewed and accepted",
			},
		},
		{
			name: "Updates a task without note",
			args: args{
				id:          1,
				status:      enums.TaskStatusDone,
				completedAt: completedAt,
			},
			mock: func(
				svc *persistence_fake.MockService,
				task *datatypes.Task,
			) *persistence_fake.MockService {
				svc.EXPECT().
					UpdateTask(
						mockutil.Context(),
						1,
						mockutil.Func(
							"should update task",
							func(updater persistence.TaskUpdater) error {
								_, err := updater(&datatypes.Task{})
								return err
							},
						),
					).
					DoAndReturn(
						func(
							ctx context.Context,
							id int,
							updater persistence.TaskUpdater,
						) (*datatypes.Task, error) {
							task, err := updater(task)
							return task, err
						},
					)
				return svc
			},
			want: datatypes.Task{
				ID:     1,
				Status: enums.TaskStatusDone,
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
				task *datatypes.Task,
			) *persistence_fake.MockService {
				svc.EXPECT().
					UpdateTask(
						mockutil.Context(),
						2,
						mockutil.Func(
							"should update task",
							func(updater persistence.TaskUpdater) error {
								_, err := updater(&datatypes.Task{})
								return err
							},
						),
					).
					Return(
						nil,
						errors.New("not found error: db: task not found"),
					)
				return svc
			},
			wantErr: "error updating task: not found error: db: task not found",
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ingestsvc, perSvc, _ := testSvc(t, nil, 0)
			task := datatypes.Task{
				ID: 1,
			}
			if tt.mock != nil {
				tt.mock(perSvc, &task)
			}

			err := ingestsvc.CompleteTask(
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
			assert.DeepEqual(t, task, tt.want)
		})
	}
}
