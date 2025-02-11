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
		pa       *datatypes.PreservationAction
		setSIPID bool
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
					Type:        enums.PreservationActionTypeCreateAip,
					Status:      enums.PreservationActionStatusDone,
					StartedAt:   started,
					CompletedAt: completed,
				},
				setSIPID: true,
			},
			want: &datatypes.PreservationAction{
				ID:          1,
				WorkflowID:  workflowID,
				Type:        enums.PreservationActionTypeCreateAip,
				Status:      enums.PreservationActionStatusDone,
				StartedAt:   started,
				CompletedAt: completed,
				SIPID:       1,
			},
		},
		{
			name: "Required field error for missing WorkflowID",
			args: params{
				pa: &datatypes.PreservationAction{
					Type:   enums.PreservationActionTypeCreateAip,
					Status: enums.PreservationActionStatusDone,
				},
			},
			wantErr: "invalid data error: field \"WorkflowID\" is required",
		},
		{
			name: "Required field error for missing SIPID",
			args: params{
				pa: &datatypes.PreservationAction{
					WorkflowID: workflowID,
					Type:       enums.PreservationActionTypeCreateAip,
					Status:     enums.PreservationActionStatusDone,
				},
			},
			wantErr: "invalid data error: field \"SIPID\" is required",
		},
		{
			name: "Foreign key error on an invalid SIPID",
			args: params{
				pa: &datatypes.PreservationAction{
					WorkflowID: workflowID,
					Type:       9,
					Status:     enums.PreservationActionStatusDone,
					SIPID:      12345,
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
			sip, _ := createSIP(
				entc,
				"Test SIP",
				enums.SIPStatusInProgress,
			)

			pa := *tt.args.pa // Make a local copy.
			if tt.args.setSIPID {
				pa.SIPID = sip.ID
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
