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

func TestCreatePreservationAction(t *testing.T) {
	t.Parallel()

	workflowID := "processing-workflow-720db1d4-825c-4911-9a20-61c212cf23ff"
	started := sql.NullTime{Time: time.Now(), Valid: true}
	completed := sql.NullTime{Time: started.Time.Add(time.Second), Valid: true}

	type params struct {
		pa           *datatypes.PreservationAction
		setPackageID bool
	}
	tests := []struct {
		name    string
		args    params
		want    *datatypes.PreservationAction
		wantErr string
	}{
		{
			name: "Saves a new preservation action to the database",
			args: params{
				pa: &datatypes.PreservationAction{
					WorkflowID:  workflowID,
					Type:        enums.PreservationActionTypeCreateAIP,
					Status:      enums.PreservationActionStatusDone,
					StartedAt:   started,
					CompletedAt: completed,
				},
				setPackageID: true,
			},
			want: &datatypes.PreservationAction{
				ID:          1,
				WorkflowID:  workflowID,
				Type:        enums.PreservationActionTypeCreateAIP,
				Status:      enums.PreservationActionStatusDone,
				StartedAt:   started,
				CompletedAt: completed,
				PackageID:   1,
			},
		},
		{
			name: "Required field error for missing WorkflowID",
			args: params{
				pa: &datatypes.PreservationAction{
					Type:   enums.PreservationActionTypeCreateAIP,
					Status: enums.PreservationActionStatusDone,
				},
			},
			wantErr: "invalid data error: field \"WorkflowID\" is required",
		},
		{
			name: "Required field error for missing PackageID",
			args: params{
				pa: &datatypes.PreservationAction{
					WorkflowID: workflowID,
					Type:       enums.PreservationActionTypeCreateAIP,
					Status:     enums.PreservationActionStatusDone,
				},
			},
			wantErr: "invalid data error: field \"PackageID\" is required",
		},
		{
			name: "Foreign key error on an invalid PackageID",
			args: params{
				pa: &datatypes.PreservationAction{
					WorkflowID: workflowID,
					Type:       9,
					Status:     enums.PreservationActionStatusDone,
					PackageID:  12345,
				},
			},
			wantErr: "invalid data error: db: constraint failed: FOREIGN KEY constraint failed: create preservation action",
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

			pa := *tt.args.pa // Make a local copy.
			if tt.args.setPackageID {
				pa.PackageID = uint(pkg.ID)
			}

			err := svc.CreatePreservationAction(ctx, &pa)
			if tt.wantErr != "" {
				assert.Error(t, err, tt.wantErr)
				return
			}
			assert.NilError(t, err)
			assert.DeepEqual(t, &pa, tt.want)
		})
	}
}
