package entclient_test

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"slices"
	"testing"
	"time"

	"github.com/go-logr/logr"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/google/uuid"
	"gotest.tools/v3/assert"

	"github.com/artefactual-sdps/enduro/internal/datatypes"
	"github.com/artefactual-sdps/enduro/internal/enums"
	"github.com/artefactual-sdps/enduro/internal/persistence"
	"github.com/artefactual-sdps/enduro/internal/persistence/ent/db"
)

func addDBFixtures(
	t *testing.T,
	entc *db.Client,
) (*db.PreservationAction, *db.PreservationAction) {
	t.Helper()

	pkg, err := createPackage(entc, "P1", enums.PackageStatusInProgress)
	if err != nil {
		t.Errorf("create package: %v", err)
	}

	pa, err := createPreservationAction(entc, pkg.ID, enums.PreservationActionStatusInProgress)
	if err != nil {
		t.Errorf("create preservation action: %v", err)
	}

	pa2, err := createPreservationAction(entc, pkg.ID, enums.PreservationActionStatusDone)
	if err != nil {
		t.Errorf("create preservation action 2: %v", err)
	}

	return pa, pa2
}

func TestCreatePreservationTask(t *testing.T) {
	t.Parallel()

	taskID := "ef0193bf-a622-4a8b-b860-cda605a426b5"
	started := sql.NullTime{Time: time.Now(), Valid: true}
	completed := sql.NullTime{Time: started.Time.Add(time.Second), Valid: true}

	type params struct {
		pt                       *datatypes.PreservationTask
		zeroPreservationActionID bool
	}
	tests := []struct {
		name    string
		args    params
		want    *datatypes.PreservationTask
		wantErr string
	}{
		{
			name: "Saves a new preservation task in the DB",
			args: params{
				pt: &datatypes.PreservationTask{
					TaskID:      taskID,
					Name:        "PT1",
					Status:      enums.PreservationTaskStatusInProgress,
					StartedAt:   started,
					CompletedAt: completed,
					Note:        "PT1 Note",
				},
			},
			want: &datatypes.PreservationTask{
				ID:          1,
				TaskID:      taskID,
				Name:        "PT1",
				Status:      enums.PreservationTaskStatusInProgress,
				StartedAt:   started,
				CompletedAt: completed,
				Note:        "PT1 Note",
			},
		},
		{
			name: "Errors on invalid TaskID",
			args: params{
				pt: &datatypes.PreservationTask{
					TaskID: "123456",
				},
			},
			wantErr: "invalid data error: parse error: field \"TaskID\": invalid UUID length: 6",
		},
		{
			name: "Required field error for missing Name",
			args: params{
				pt: &datatypes.PreservationTask{
					TaskID: "ef0193bf-a622-4a8b-b860-cda605a426b5",
				},
			},
			wantErr: "invalid data error: field \"Name\" is required",
		},
		{
			name: "Required field error for missing PreservationActionID",
			args: params{
				pt: &datatypes.PreservationTask{
					TaskID: taskID,
					Name:   "PT1",
					Status: enums.PreservationTaskStatusInProgress,
				},
				zeroPreservationActionID: true,
			},
			wantErr: "invalid data error: field \"PreservationActionID\" is required",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			entc, svc := setUpClient(t, logr.Discard())
			ctx := context.Background()
			pkg, _ := createPackage(
				entc,
				"Test package",
				enums.PackageStatusDone,
			)
			pa, _ := createPreservationAction(
				entc,
				pkg.ID,
				enums.PreservationActionStatusDone,
			)

			pt := *tt.args.pt // Make a local copy of pt.

			if !tt.args.zeroPreservationActionID {
				pt.PreservationActionID = pa.ID
			}

			err := svc.CreatePreservationTask(ctx, &pt)
			if tt.wantErr != "" {
				assert.Error(t, err, tt.wantErr)
				return
			}
			assert.NilError(t, err)

			assert.Equal(t, pt.ID, tt.want.ID)
			assert.Equal(t, pt.TaskID, tt.want.TaskID)
			assert.Equal(t, pt.Name, tt.want.Name)
			assert.Equal(t, pt.Status, tt.want.Status)
			assert.Equal(t, pt.StartedAt, tt.want.StartedAt)
			assert.Equal(t, pt.CompletedAt, tt.want.CompletedAt)
			assert.Equal(t, pt.Note, tt.want.Note)
			assert.Equal(t, pt.PreservationActionID, pa.ID)
		})
	}
}

func TestCreatePreservationTasks(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		tasks   func(entc *db.Client) []*datatypes.PreservationTask
		wantErr string
	}{
		{
			name: "Creates multiple preservation tasks",
			tasks: func(entc *db.Client) []*datatypes.PreservationTask {
				pa1, _ := addDBFixtures(t, entc)
				return []*datatypes.PreservationTask{
					{
						TaskID:               "20b8b39c-3642-4138-9dab-b266bfca1e87",
						Name:                 "Task 1",
						Status:               enums.PreservationTaskStatusDone,
						StartedAt:            sql.NullTime{Time: time.Now(), Valid: true},
						CompletedAt:          sql.NullTime{Time: time.Now().Add(time.Second), Valid: true},
						Note:                 "Note",
						PreservationActionID: pa1.ID,
					},
					{
						TaskID:               "9cad7233-58b2-46c4-a2e3-20f892859648",
						Name:                 "Task 2",
						Status:               enums.PreservationTaskStatusDone,
						StartedAt:            sql.NullTime{Time: time.Now(), Valid: true},
						CompletedAt:          sql.NullTime{Time: time.Now().Add(time.Second), Valid: true},
						Note:                 "Note",
						PreservationActionID: pa1.ID,
					},
				}
			},
		},
		{
			name: "Creates multiple preservation tasks exceeding batch size",
			tasks: func(entc *db.Client) []*datatypes.PreservationTask {
				pa1, _ := addDBFixtures(t, entc)
				pts := make([]*datatypes.PreservationTask, 0, 300) // Three batches.
				for range cap(pts) {
					pts = append(pts, &datatypes.PreservationTask{
						TaskID:               uuid.NewString(),
						Name:                 "Task",
						Status:               enums.PreservationTaskStatusDone,
						StartedAt:            sql.NullTime{Time: time.Now(), Valid: true},
						CompletedAt:          sql.NullTime{Time: time.Now().Add(time.Second), Valid: true},
						Note:                 "Note",
						PreservationActionID: pa1.ID,
					})
				}
				return pts
			},
		},
		{
			name: "Errors on invalid TaskID",
			tasks: func(entc *db.Client) []*datatypes.PreservationTask {
				return []*datatypes.PreservationTask{
					{
						TaskID: "123456",
					},
				}
			},
			wantErr: "invalid data error: parse error: field \"TaskID\": invalid UUID length: 6",
		},
		{
			name: "Required field error for missing Name",
			tasks: func(entc *db.Client) []*datatypes.PreservationTask {
				return []*datatypes.PreservationTask{
					{
						TaskID: "ef0193bf-a622-4a8b-b860-cda605a426b5",
					},
				}
			},
			wantErr: "invalid data error: field \"Name\" is required",
		},
		{
			name: "Required field error for missing PreservationActionID",
			tasks: func(entc *db.Client) []*datatypes.PreservationTask {
				return []*datatypes.PreservationTask{
					{
						TaskID: "20b8b39c-3642-4138-9dab-b266bfca1e87",
						Name:   "PT1",
						Status: enums.PreservationTaskStatusInProgress,
					},
				}
			},
			wantErr: "invalid data error: field \"PreservationActionID\" is required",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			entc, svc := setUpClient(t, logr.Discard())

			tasks := tt.tasks(entc)
			ret, err := svc.CreatePreservationTasks(ctx, slices.Values(tasks))

			if tt.wantErr != "" {
				assert.Error(t, err, tt.wantErr)
				return
			}
			assert.NilError(t, err)

			assert.Equal(t, entc.PreservationTask.Query().CountX(ctx), len(tasks))
			assert.DeepEqual(t, tasks, ret,
				// The ID is assigned after creation, so we need to exclude it from the comparison.
				cmpopts.IgnoreFields(datatypes.PreservationTask{}, "ID"),
			)
		})
	}
}

func TestUpdatePreservationTask(t *testing.T) {
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
		pt      *datatypes.PreservationTask
		updater persistence.PresTaskUpdater
	}
	tests := []struct {
		name    string
		args    params
		want    *datatypes.PreservationTask
		wantErr string
	}{
		{
			name: "Updates all preservation task columns",
			args: params{
				pt: &datatypes.PreservationTask{
					TaskID:      taskID.String(),
					Name:        "PT 1",
					Status:      enums.PreservationTaskStatusInProgress,
					StartedAt:   started,
					CompletedAt: completed,
					Note:        "PT1 Note",
				},
				updater: func(p *datatypes.PreservationTask) (*datatypes.PreservationTask, error) {
					p.ID = 100 // No-op, can't update ID.
					p.Name = "PT1 Update"
					p.TaskID = taskID2.String()
					p.Status = enums.PreservationTaskStatusDone
					p.StartedAt = started2
					p.CompletedAt = completed2
					p.Note = "PT1 Note updated"
					return p, nil
				},
			},
			want: &datatypes.PreservationTask{
				TaskID:      taskID2.String(),
				Name:        "PT1 Update",
				Status:      enums.PreservationTaskStatusDone,
				StartedAt:   started2,
				CompletedAt: completed2,
				Note:        "PT1 Note updated",
			},
		},
		{
			name: "Updates selected preservation task columns",
			args: params{
				pt: &datatypes.PreservationTask{
					Name:      "PT 1",
					TaskID:    taskID.String(),
					Status:    enums.PreservationTaskStatusInProgress,
					StartedAt: started,
				},
				updater: func(p *datatypes.PreservationTask) (*datatypes.PreservationTask, error) {
					p.Status = enums.PreservationTaskStatusDone
					p.CompletedAt = completed
					p.Note = "PT1 Note updated"
					return p, nil
				},
			},
			want: &datatypes.PreservationTask{
				TaskID:      taskID.String(),
				Name:        "PT 1",
				Status:      enums.PreservationTaskStatusDone,
				StartedAt:   started,
				CompletedAt: completed,
				Note:        "PT1 Note updated",
			},
		},
		{
			name: "Errors when target preservation task isn't found",
			args: params{
				updater: func(p *datatypes.PreservationTask) (*datatypes.PreservationTask, error) {
					return nil, errors.New("Bad input")
				},
			},
			wantErr: "not found error: db: preservation_task not found",
		},
		{
			name: "Errors when the updater fails",
			args: params{
				pt: &datatypes.PreservationTask{
					Name:   "PT 1",
					TaskID: taskID.String(),
					Status: enums.PreservationTaskStatusInProgress,
				},
				updater: func(p *datatypes.PreservationTask) (*datatypes.PreservationTask, error) {
					return nil, fmt.Errorf("Bad input")
				},
			},
			wantErr: "invalid data error: updater error: Bad input",
		},
		{
			name: "Errors on an invalid TaskID",
			args: params{
				pt: &datatypes.PreservationTask{
					Name:   "PT 1",
					TaskID: taskID.String(),
					Status: enums.PreservationTaskStatusInProgress,
				},
				updater: func(p *datatypes.PreservationTask) (*datatypes.PreservationTask, error) {
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

			ctx := context.Background()
			entc, svc := setUpClient(t, logr.Discard())
			pa, pa2 := addDBFixtures(t, entc)

			updater := tt.args.updater
			var id int
			if tt.args.pt != nil {
				pt := *tt.args.pt // Make a local copy of pt.
				pt.PreservationActionID = pa.ID

				// Create preservation task to be updated.
				err := svc.CreatePreservationTask(ctx, &pt)
				if err != nil {
					t.Errorf("create preservation task: %v", err)
				}
				id = pt.ID

				// Update PreservationActionID to pa2.ID.
				updater = func(pt *datatypes.PreservationTask) (*datatypes.PreservationTask, error) {
					pt, err := tt.args.updater(pt)
					if err != nil {
						return nil, err
					}
					pt.PreservationActionID = pa2.ID

					return pt, nil
				}
			}

			pt, err := svc.UpdatePreservationTask(ctx, id, updater)
			if tt.wantErr != "" {
				assert.Error(t, err, tt.wantErr)
				return
			}

			tt.want.ID = id
			tt.want.PreservationActionID = pa2.ID
			assert.DeepEqual(t, pt, tt.want)
		})
	}
}
