package client_test

import (
	"database/sql"
	"errors"
	"fmt"
	"slices"
	"testing"
	"time"

	"github.com/go-logr/logr"
	"github.com/google/uuid"
	"gotest.tools/v3/assert"

	"github.com/artefactual-sdps/enduro/internal/datatypes"
	"github.com/artefactual-sdps/enduro/internal/enums"
	"github.com/artefactual-sdps/enduro/internal/persistence"
	"github.com/artefactual-sdps/enduro/internal/persistence/ent/db"
)

func addDBFixtures(t *testing.T, entc *db.Client) {
	t.Helper()

	sip, err := createSIP(t, entc, "S1", enums.SIPStatusProcessing)
	if err != nil {
		t.Errorf("create SIP: %v", err)
	}

	_, err = createWorkflow(t, entc, sip.ID, enums.WorkflowStatusInProgress)
	if err != nil {
		t.Errorf("create workflow: %v", err)
	}
}

func TestCreateTask(t *testing.T) {
	t.Parallel()

	taskUUID := uuid.New()
	started := sql.NullTime{Time: time.Now(), Valid: true}
	completed := sql.NullTime{Time: started.Time.Add(time.Second), Valid: true}

	tests := []struct {
		name    string
		task    *datatypes.Task
		want    *datatypes.Task
		wantErr string
	}{
		{
			name: "Saves a new task in the DB",
			task: &datatypes.Task{
				UUID:         taskUUID,
				Name:         "PT1",
				Status:       enums.TaskStatusInProgress,
				StartedAt:    started,
				CompletedAt:  completed,
				Note:         "PT1 Note",
				WorkflowUUID: wUUID,
			},
			want: &datatypes.Task{
				ID:           1,
				UUID:         taskUUID,
				Name:         "PT1",
				Status:       enums.TaskStatusInProgress,
				StartedAt:    started,
				CompletedAt:  completed,
				Note:         "PT1 Note",
				WorkflowUUID: wUUID,
			},
		},
		{
			name:    "Errors on invalid UUID",
			task:    &datatypes.Task{},
			wantErr: "invalid data error: field \"UUID\" is required",
		},
		{
			name: "Required field error for missing Name",
			task: &datatypes.Task{
				UUID: taskUUID,
			},
			wantErr: "invalid data error: field \"Name\" is required",
		},
		{
			name: "Required field error for missing WorkflowUUID",
			task: &datatypes.Task{
				UUID:   taskUUID,
				Name:   "PT1",
				Status: enums.TaskStatusInProgress,
			},
			wantErr: "invalid data error: field \"WorkflowUUID\" is required",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			entc, svc := setUpClient(t, logr.Discard())
			ctx := t.Context()
			sip, _ := createSIP(
				t,
				entc,
				"Test SIP",
				enums.SIPStatusIngested,
			)
			_, _ = createWorkflow(
				t,
				entc,
				sip.ID,
				enums.WorkflowStatusDone,
			)

			task := tt.task // Make a local copy of pt.

			err := svc.CreateTask(ctx, task)
			if tt.wantErr != "" {
				assert.Error(t, err, tt.wantErr)
				return
			}
			assert.NilError(t, err)

			assert.Equal(t, task.ID, tt.want.ID)
			assert.Equal(t, task.UUID, tt.want.UUID)
			assert.Equal(t, task.Name, tt.want.Name)
			assert.Equal(t, task.Status, tt.want.Status)
			assert.Equal(t, task.StartedAt, tt.want.StartedAt)
			assert.Equal(t, task.CompletedAt, tt.want.CompletedAt)
			assert.Equal(t, task.Note, tt.want.Note)
			assert.Equal(t, task.WorkflowUUID, tt.want.WorkflowUUID)
		})
	}
}

func TestCreateTasks(t *testing.T) {
	t.Parallel()

	started := sql.NullTime{Time: time.Now(), Valid: true}
	completed := sql.NullTime{Time: started.Time.Add(time.Second), Valid: true}

	tests := []struct {
		name    string
		make    func() []*datatypes.Task
		wantIDs []int
		wantErr string
	}{
		{
			name: "Saves multiple tasks",
			make: func() []*datatypes.Task {
				return []*datatypes.Task{
					{
						UUID:         uuid.New(),
						Name:         "Task A",
						Status:       enums.TaskStatusInProgress,
						StartedAt:    started,
						CompletedAt:  completed,
						WorkflowUUID: wUUID,
					},
					{
						UUID:         uuid.New(),
						Name:         "Task B",
						Status:       enums.TaskStatusDone,
						WorkflowUUID: wUUID,
					},
				}
			},
			wantIDs: []int{1, 2},
		},
		{
			name: "Returns validation error",
			make: func() []*datatypes.Task {
				return []*datatypes.Task{
					{
						UUID:         uuid.New(),
						Status:       enums.TaskStatusInProgress,
						WorkflowUUID: wUUID,
					},
				}
			},
			wantErr: "invalid data error: field \"Name\" is required",
		},
		{
			name: "Workflow not found",
			make: func() []*datatypes.Task {
				return []*datatypes.Task{
					{
						UUID:         uuid.New(),
						Name:         "Missing workflow",
						Status:       enums.TaskStatusInProgress,
						WorkflowUUID: uuid.New(),
					},
				}
			},
			wantErr: "not found error",
		},
		{
			name: "Accepts slices.Values iterator",
			make: func() []*datatypes.Task {
				return []*datatypes.Task{
					{
						UUID:         uuid.New(),
						Name:         "Slice Task 1",
						Status:       enums.TaskStatusInProgress,
						WorkflowUUID: wUUID,
					},
					{
						UUID:         uuid.New(),
						Name:         "Slice Task 2",
						Status:       enums.TaskStatusDone,
						WorkflowUUID: wUUID,
					},
				}
			},
			wantIDs: []int{1, 2},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			entc, svc := setUpClient(t, logr.Discard())
			ctx := t.Context()

			sip, _ := createSIP(
				t,
				entc,
				"Test SIP",
				enums.SIPStatusIngested,
			)
			_, _ = createWorkflow(
				t,
				entc,
				sip.ID,
				enums.WorkflowStatusDone,
			)

			tasks := tt.make()
			var seq persistence.TaskSequence
			if tt.name == "Accepts slices.Values iterator" {
				seq = persistence.TaskSequence(slices.Values(tasks))
			} else {
				seq = persistence.TaskSequence(func(yield func(*datatypes.Task) bool) {
					for _, task := range tasks {
						if !yield(task) {
							return
						}
					}
				})
			}

			err := svc.CreateTasks(ctx, seq)
			if tt.wantErr != "" {
				assert.ErrorContains(t, err, tt.wantErr)
				return
			}

			assert.NilError(t, err)
			for i, want := range tt.wantIDs {
				assert.Equal(t, want, tasks[i].ID)
			}
		})
	}
}

func TestUpdateTask(t *testing.T) {
	t.Parallel()

	taskID := uuid.MustParse("c5f7c35a-d5a6-4e00-b4da-b036ce5b40bc")
	taskID2 := uuid.MustParse("c04d0191-d7ce-46dd-beff-92d6830082ff")

	started := sql.NullTime{
		Time:  time.Date(2024, 3, 31, 10, 11, 12, 0, time.UTC),
		Valid: true,
	}
	started2 := sql.NullTime{
		Time:  time.Date(2024, 4, 1, 17, 5, 49, 0, time.UTC),
		Valid: true,
	}

	completed := sql.NullTime{Time: started.Time.Add(time.Second), Valid: true}
	completed2 := sql.NullTime{Time: started2.Time.Add(time.Second), Valid: true}

	type params struct {
		task    *datatypes.Task
		updater persistence.TaskUpdater
	}
	tests := []struct {
		name    string
		args    params
		want    *datatypes.Task
		wantErr string
	}{
		{
			name: "Updates all task columns",
			args: params{
				task: &datatypes.Task{
					UUID:         taskID,
					Name:         "Task 1",
					Status:       enums.TaskStatusInProgress,
					StartedAt:    started,
					CompletedAt:  completed,
					Note:         "Task1 Note",
					WorkflowUUID: wUUID,
				},
				updater: func(p *datatypes.Task) (*datatypes.Task, error) {
					p.ID = 100 // No-op, can't update ID.
					p.Name = "Task1 Update"
					p.UUID = taskID2 // No-op, can't update UUID.
					p.Status = enums.TaskStatusDone
					p.StartedAt = started2
					p.CompletedAt = completed2
					p.Note = "Task1 Note updated"
					p.WorkflowUUID = uuid.New() // No-op, can't update WorkflowUUID.
					return p, nil
				},
			},
			want: &datatypes.Task{
				UUID:         taskID,
				Name:         "Task1 Update",
				Status:       enums.TaskStatusDone,
				StartedAt:    started2,
				CompletedAt:  completed2,
				Note:         "Task1 Note updated",
				WorkflowUUID: wUUID,
			},
		},
		{
			name: "Updates selected task columns",
			args: params{
				task: &datatypes.Task{
					Name:         "Task 1",
					UUID:         taskID,
					Status:       enums.TaskStatusInProgress,
					StartedAt:    started,
					WorkflowUUID: wUUID,
				},
				updater: func(p *datatypes.Task) (*datatypes.Task, error) {
					p.Status = enums.TaskStatusDone
					p.CompletedAt = completed
					p.Note = "Task1 Note updated"
					return p, nil
				},
			},
			want: &datatypes.Task{
				UUID:         taskID,
				Name:         "Task 1",
				Status:       enums.TaskStatusDone,
				StartedAt:    started,
				CompletedAt:  completed,
				Note:         "Task1 Note updated",
				WorkflowUUID: wUUID,
			},
		},
		{
			name: "Errors when target task isn't found",
			args: params{
				updater: func(p *datatypes.Task) (*datatypes.Task, error) {
					return nil, errors.New("Bad input")
				},
			},
			wantErr: "not found error: db: task not found",
		},
		{
			name: "Errors when the updater fails",
			args: params{
				task: &datatypes.Task{
					Name:         "Task 1",
					UUID:         taskID,
					Status:       enums.TaskStatusInProgress,
					WorkflowUUID: wUUID,
				},
				updater: func(p *datatypes.Task) (*datatypes.Task, error) {
					return nil, fmt.Errorf("Bad input")
				},
			},
			wantErr: "invalid data error: updater error: Bad input",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctx := t.Context()
			entc, svc := setUpClient(t, logr.Discard())
			addDBFixtures(t, entc)

			updater := tt.args.updater
			var id int
			if tt.args.task != nil {
				task := *tt.args.task // Make a local copy of pt.

				// Create task to be updated.
				err := svc.CreateTask(ctx, &task)
				if err != nil {
					t.Errorf("create task: %v", err)
				}
				id = task.ID
			}

			task, err := svc.UpdateTask(ctx, id, updater)
			if tt.wantErr != "" {
				assert.Error(t, err, tt.wantErr)
				return
			}

			tt.want.ID = id
			assert.DeepEqual(t, task, tt.want)
		})
	}
}
