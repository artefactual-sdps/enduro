package entclient_test

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/go-logr/logr"
	"gotest.tools/v3/assert"

	"github.com/artefactual-sdps/enduro/internal/datatypes"
	"github.com/artefactual-sdps/enduro/internal/enums"
)

func TestCreatePreservationTask(t *testing.T) {
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
				enums.PackageStatusInProgress,
			)
			pa, _ := createPreservationAction(
				entc,
				pkg.ID,
				enums.PreservationActionStatusInProgress,
			)

			pt := *tt.args.pt // Make a local copy of pt.

			if !tt.args.zeroPreservationActionID {
				pt.PreservationActionID = uint(pa.ID)
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
			assert.Equal(t, pt.PreservationActionID, uint(pa.ID))
		})
	}
}
