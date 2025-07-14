package entclient_test

import (
	"database/sql"
	"testing"
	"time"

	"github.com/go-logr/logr"
	"github.com/google/uuid"
	"gotest.tools/v3/assert"

	"github.com/artefactual-sdps/enduro/internal/datatypes"
	"github.com/artefactual-sdps/enduro/internal/enums"
)

func TestCreateWorkflow(t *testing.T) {
	t.Parallel()

	workflowUUID := uuid.New()
	temporalID := "processing-workflow-720db1d4-825c-4911-9a20-61c212cf23ff"
	started := sql.NullTime{Time: time.Now(), Valid: true}
	completed := sql.NullTime{Time: started.Time.Add(time.Second), Valid: true}

	tests := []struct {
		name    string
		args    *datatypes.Workflow
		want    *datatypes.Workflow
		wantErr string
	}{
		{
			name: "Saves a new workflow to the database",
			args: &datatypes.Workflow{
				UUID:        workflowUUID,
				TemporalID:  temporalID,
				Type:        enums.WorkflowTypeCreateAip,
				Status:      enums.WorkflowStatusDone,
				StartedAt:   started,
				CompletedAt: completed,
				SIPUUID:     sipUUID,
			},
			want: &datatypes.Workflow{
				ID:          1,
				UUID:        workflowUUID,
				TemporalID:  temporalID,
				Type:        enums.WorkflowTypeCreateAip,
				Status:      enums.WorkflowStatusDone,
				StartedAt:   started,
				CompletedAt: completed,
				SIPUUID:     sipUUID,
			},
		},
		{
			name: "Required field error for missing UUID",
			args: &datatypes.Workflow{
				Type:   enums.WorkflowTypeCreateAip,
				Status: enums.WorkflowStatusDone,
			},
			wantErr: "invalid data error: field \"UUID\" is required",
		},
		{
			name: "Required field error for missing TemporalID",
			args: &datatypes.Workflow{
				UUID:   workflowUUID,
				Type:   enums.WorkflowTypeCreateAip,
				Status: enums.WorkflowStatusDone,
			},
			wantErr: "invalid data error: field \"TemporalID\" is required",
		},
		{
			name: "Required field error for missing SIPUUID",
			args: &datatypes.Workflow{
				UUID:       workflowUUID,
				TemporalID: temporalID,
				Type:       enums.WorkflowTypeCreateAip,
				Status:     enums.WorkflowStatusDone,
			},
			wantErr: "invalid data error: field \"SIPUUID\" is required",
		},
		{
			name: "Invalid Type field error",
			args: &datatypes.Workflow{
				UUID:       workflowUUID,
				SIPUUID:    sipUUID,
				TemporalID: temporalID,
				Type:       "invalid",
			},
			wantErr: "invalid data error: field \"Type\" is invalid \"invalid\"",
		},
		{
			name: "Not found SIP error",
			args: &datatypes.Workflow{
				UUID:       workflowUUID,
				TemporalID: temporalID,
				Type:       enums.WorkflowTypeCreateAip,
				Status:     enums.WorkflowStatusDone,
				SIPUUID:    uuid.New(),
			},
			wantErr: "not found error: db: sip not found: create workflow",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			entc, svc := setUpClient(t, logr.Discard())
			ctx := t.Context()
			_, _ = createSIP(t, entc, "Test SIP", enums.SIPStatusProcessing)

			err := svc.CreateWorkflow(ctx, tt.args)
			if tt.wantErr != "" {
				assert.Error(t, err, tt.wantErr)
				return
			}
			assert.NilError(t, err)
			assert.DeepEqual(t, tt.args, tt.want)
		})
	}
}
