package entclient_test

import (
	"database/sql"
	"errors"
	"fmt"
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

func addDBFixtures(t *testing.T, entc *db.Client) (*db.Workflow, *db.Workflow) {
	t.Helper()

	sip, err := createSIP(t, entc, "S1", enums.SIPStatusProcessing)
	if err != nil {
		t.Errorf("create SIP: %v", err)
	}

	w, err := createWorkflow(t, entc, sip.ID, enums.WorkflowStatusInProgress)
	if err != nil {
		t.Errorf("create workflow: %v", err)
	}

	pa2, err := createWorkflow(t, entc, sip.ID, enums.WorkflowStatusDone)
	if err != nil {
		t.Errorf("create workflow 2: %v", err)
	}

	return w, pa2
}

func TestCreateTask(t *testing.T) {
	t.Parallel()

	taskID := "ef0193bf-a622-4a8b-b860-cda605a426b5"
	started := sql.NullTime{Time: time.Now(), Valid: true}
	completed := sql.NullTime{Time: started.Time.Add(time.Second), Valid: true}

	type params struct {
		task           *datatypes.Task
		zeroWorkflowID bool
	}
	tests := []struct {
		name    string
		args    params
		want    *datatypes.Task
		wantErr string
	}{
		{
			name: "Saves a new task in the DB",
			args: params{
				task: &datatypes.Task{
					TaskID:      taskID,
					Name:        "PT1",
					Status:      enums.TaskStatusInProgress,
					StartedAt:   started,
					CompletedAt: completed,
					Note:        "PT1 Note",
				},
			},
			want: &datatypes.Task{
				ID:          1,
				TaskID:      taskID,
				Name:        "PT1",
				Status:      enums.TaskStatusInProgress,
				StartedAt:   started,
				CompletedAt: completed,
				Note:        "PT1 Note",
			},
		},
		{
			name: "Errors on invalid TaskID",
			args: params{
				task: &datatypes.Task{
					TaskID: "123456",
				},
			},
			wantErr: "invalid data error: parse error: field \"TaskID\": invalid UUID length: 6",
		},
		{
			name: "Required field error for missing Name",
			args: params{
				task: &datatypes.Task{
					TaskID: "ef0193bf-a622-4a8b-b860-cda605a426b5",
				},
			},
			wantErr: "invalid data error: field \"Name\" is required",
		},
		{
			name: "Required field error for missing WorkflowID",
			args: params{
				task: &datatypes.Task{
					TaskID: taskID,
					Name:   "PT1",
					Status: enums.TaskStatusInProgress,
				},
				zeroWorkflowID: true,
			},
			wantErr: "invalid data error: field \"WorkflowID\" is required",
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
			w, _ := createWorkflow(
				t,
				entc,
				sip.ID,
				enums.WorkflowStatusDone,
			)

			task := *tt.args.task // Make a local copy of pt.

			if !tt.args.zeroWorkflowID {
				task.WorkflowID = w.ID
			}

			err := svc.CreateTask(ctx, &task)
			if tt.wantErr != "" {
				assert.Error(t, err, tt.wantErr)
				return
			}
			assert.NilError(t, err)

			assert.Equal(t, task.ID, tt.want.ID)
			assert.Equal(t, task.TaskID, tt.want.TaskID)
			assert.Equal(t, task.Name, tt.want.Name)
			assert.Equal(t, task.Status, tt.want.Status)
			assert.Equal(t, task.StartedAt, tt.want.StartedAt)
			assert.Equal(t, task.CompletedAt, tt.want.CompletedAt)
			assert.Equal(t, task.Note, tt.want.Note)
			assert.Equal(t, task.WorkflowID, w.ID)
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
					TaskID:      taskID.String(),
					Name:        "Task 1",
					Status:      enums.TaskStatusInProgress,
					StartedAt:   started,
					CompletedAt: completed,
					Note:        "Task1 Note",
				},
				updater: func(p *datatypes.Task) (*datatypes.Task, error) {
					p.ID = 100 // No-op, can't update ID.
					p.Name = "Task1 Update"
					p.TaskID = taskID2.String()
					p.Status = enums.TaskStatusDone
					p.StartedAt = started2
					p.CompletedAt = completed2
					p.Note = "Task1 Note updated"
					return p, nil
				},
			},
			want: &datatypes.Task{
				TaskID:      taskID2.String(),
				Name:        "Task1 Update",
				Status:      enums.TaskStatusDone,
				StartedAt:   started2,
				CompletedAt: completed2,
				Note:        "Task1 Note updated",
			},
		},
		{
			name: "Updates selected task columns",
			args: params{
				task: &datatypes.Task{
					Name:      "Task 1",
					TaskID:    taskID.String(),
					Status:    enums.TaskStatusInProgress,
					StartedAt: started,
				},
				updater: func(p *datatypes.Task) (*datatypes.Task, error) {
					p.Status = enums.TaskStatusDone
					p.CompletedAt = completed
					p.Note = "Task1 Note updated"
					return p, nil
				},
			},
			want: &datatypes.Task{
				TaskID:      taskID.String(),
				Name:        "Task 1",
				Status:      enums.TaskStatusDone,
				StartedAt:   started,
				CompletedAt: completed,
				Note:        "Task1 Note updated",
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
					Name:   "Task 1",
					TaskID: taskID.String(),
					Status: enums.TaskStatusInProgress,
				},
				updater: func(p *datatypes.Task) (*datatypes.Task, error) {
					return nil, fmt.Errorf("Bad input")
				},
			},
			wantErr: "invalid data error: updater error: Bad input",
		},
		{
			name: "Errors on an invalid TaskID",
			args: params{
				task: &datatypes.Task{
					Name:   "Task 1",
					TaskID: taskID.String(),
					Status: enums.TaskStatusInProgress,
				},
				updater: func(p *datatypes.Task) (*datatypes.Task, error) {
					p.TaskID = "123456"
					return p, nil
				},
			},
			wantErr: "invalid data error: parse error: field \"TaskID\": invalid UUID length: 6",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctx := t.Context()
			entc, svc := setUpClient(t, logr.Discard())
			w, w2 := addDBFixtures(t, entc)

			updater := tt.args.updater
			var id int
			if tt.args.task != nil {
				task := *tt.args.task // Make a local copy of pt.
				task.WorkflowID = w.ID

				// Create task to be updated.
				err := svc.CreateTask(ctx, &task)
				if err != nil {
					t.Errorf("create task: %v", err)
				}
				id = task.ID

				// Update WorkflowID to w2.ID.
				updater = func(task *datatypes.Task) (*datatypes.Task, error) {
					task, err := tt.args.updater(task)
					if err != nil {
						return nil, err
					}
					task.WorkflowID = w2.ID

					return task, nil
				}
			}

			task, err := svc.UpdateTask(ctx, id, updater)
			if tt.wantErr != "" {
				assert.Error(t, err, tt.wantErr)
				return
			}

			tt.want.ID = id
			tt.want.WorkflowID = w2.ID
			assert.DeepEqual(t, task, tt.want)
		})
	}
}
