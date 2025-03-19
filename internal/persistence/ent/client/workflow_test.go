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

func TestCreateWorkflow(t *testing.T) {
	t.Parallel()

	workflowID := "processing-workflow-720db1d4-825c-4911-9a20-61c212cf23ff"
	started := sql.NullTime{Time: time.Now(), Valid: true}
	completed := sql.NullTime{Time: started.Time.Add(time.Second), Valid: true}

	type params struct {
		w        *datatypes.Workflow
		setSIPID bool
	}
	tests := []struct {
		name    string
		args    params
		want    *datatypes.Workflow
		wantErr string
	}{
		{
			name: "Saves a new workflow to the database",
			args: params{
				w: &datatypes.Workflow{
					WorkflowID:  workflowID,
					Type:        enums.WorkflowTypeCreateAip,
					Status:      enums.WorkflowStatusDone,
					StartedAt:   started,
					CompletedAt: completed,
				},
				setSIPID: true,
			},
			want: &datatypes.Workflow{
				ID:          1,
				WorkflowID:  workflowID,
				Type:        enums.WorkflowTypeCreateAip,
				Status:      enums.WorkflowStatusDone,
				StartedAt:   started,
				CompletedAt: completed,
				SIPID:       1,
			},
		},
		{
			name: "Required field error for missing WorkflowID",
			args: params{
				w: &datatypes.Workflow{
					Type:   enums.WorkflowTypeCreateAip,
					Status: enums.WorkflowStatusDone,
				},
			},
			wantErr: "invalid data error: field \"WorkflowID\" is required",
		},
		{
			name: "Required field error for missing SIPID",
			args: params{
				w: &datatypes.Workflow{
					WorkflowID: workflowID,
					Type:       enums.WorkflowTypeCreateAip,
					Status:     enums.WorkflowStatusDone,
				},
			},
			wantErr: "invalid data error: field \"SIPID\" is required",
		},
		{
			name: "Foreign key error on an invalid SIPID",
			args: params{
				w: &datatypes.Workflow{
					WorkflowID: workflowID,
					Type:       9,
					Status:     enums.WorkflowStatusDone,
					SIPID:      12345,
				},
			},
			wantErr: "invalid data error: db: constraint failed: FOREIGN KEY constraint failed: create workflow",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			entc, svc := setUpClient(t, logr.Discard())
			ctx := context.Background()
			sip, _ := createSIP(
				entc,
				"Test SIP",
				enums.SIPStatusInProgress,
			)

			w := *tt.args.w // Make a local copy.
			if tt.args.setSIPID {
				w.SIPID = sip.ID
			}

			err := svc.CreateWorkflow(ctx, &w)
			if tt.wantErr != "" {
				assert.Error(t, err, tt.wantErr)
				return
			}
			assert.NilError(t, err)
			assert.DeepEqual(t, &w, tt.want)
		})
	}
}
