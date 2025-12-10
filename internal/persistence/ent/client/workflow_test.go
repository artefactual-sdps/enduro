package client_test

import (
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/go-logr/logr"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/google/uuid"
	"gotest.tools/v3/assert"

	"github.com/artefactual-sdps/enduro/internal/datatypes"
	"github.com/artefactual-sdps/enduro/internal/enums"
	"github.com/artefactual-sdps/enduro/internal/persistence"
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

func TestUpdateWorkflow(t *testing.T) {
	t.Parallel()

	workflowUUID := uuid.New()
	started := sql.NullTime{Time: time.Now(), Valid: true}
	completed := sql.NullTime{Time: started.Time.Add(time.Second), Valid: true}

	type params struct {
		workflow *datatypes.Workflow
		updater  persistence.WorkflowUpdater
	}
	tests := []struct {
		name    string
		args    params
		want    *datatypes.Workflow
		wantErr string
	}{
		{
			name: "Updates status and completion",
			args: params{
				workflow: &datatypes.Workflow{
					UUID:       workflowUUID,
					TemporalID: "processing-workflow-1",
					Type:       enums.WorkflowTypeCreateAip,
					Status:     enums.WorkflowStatusQueued,
					StartedAt:  started,
					SIPUUID:    sipUUID,
				},
				updater: func(w *datatypes.Workflow) (*datatypes.Workflow, error) {
					w.Status = enums.WorkflowStatusDone
					w.CompletedAt = completed
					return w, nil
				},
			},
			want: &datatypes.Workflow{
				ID:          1,
				UUID:        workflowUUID,
				TemporalID:  "processing-workflow-1",
				Type:        enums.WorkflowTypeCreateAip,
				Status:      enums.WorkflowStatusDone,
				StartedAt:   started,
				CompletedAt: completed,
				SIPUUID:     sipUUID,
			},
		},
		{
			name: "Propagates updater error",
			args: params{
				workflow: &datatypes.Workflow{
					UUID:       workflowUUID,
					TemporalID: "processing-workflow-2",
					Type:       enums.WorkflowTypeCreateAip,
					Status:     enums.WorkflowStatusQueued,
					SIPUUID:    sipUUID,
				},
				updater: func(w *datatypes.Workflow) (*datatypes.Workflow, error) {
					return nil, errors.New("boom")
				},
			},
			wantErr: "invalid data error: updater error: boom",
		},
		{
			name: "Returns not found when workflow missing",
			args: params{
				updater: func(w *datatypes.Workflow) (*datatypes.Workflow, error) {
					return w, nil
				},
			},
			wantErr: "not found error: db: workflow not found",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctx := t.Context()
			entc, svc := setUpClient(t, logr.Discard())
			sip, _ := createSIP(t, entc, "Test SIP", enums.SIPStatusProcessing)

			var id int
			if tt.args.workflow != nil {
				wf, err := entc.Workflow.Create().
					SetUUID(tt.args.workflow.UUID).
					SetTemporalID(tt.args.workflow.TemporalID).
					SetType(tt.args.workflow.Type).
					SetStatus(int8(tt.args.workflow.Status)). // #nosec G115 -- constrained value.
					SetStartedAt(tt.args.workflow.StartedAt.Time).
					SetSipID(sip.ID).
					Save(ctx)
				assert.NilError(t, err)
				id = wf.ID
			}

			workflow, err := svc.UpdateWorkflow(ctx, id, tt.args.updater)
			if tt.wantErr != "" {
				assert.Error(t, err, tt.wantErr)
				return
			}

			tt.want.ID = id
			assert.DeepEqual(t, workflow, tt.want, cmpopts.EquateEmpty())
		})
	}
}

func TestReadWorkflow(t *testing.T) {
	t.Parallel()

	workflowUUID := uuid.New()
	started := sql.NullTime{Time: time.Now(), Valid: true}

	for _, tt := range []struct {
		name    string
		want    *datatypes.Workflow
		wantErr string
	}{
		{
			name: "Reads a workflow",
			want: &datatypes.Workflow{
				ID:         1,
				UUID:       workflowUUID,
				TemporalID: "processing-workflow-1",
				Type:       enums.WorkflowTypeCreateAip,
				Status:     enums.WorkflowStatusQueued,
				StartedAt:  started,
				SIPUUID:    sipUUID,
			},
		},
		{
			name:    "Returns not found for missing workflow",
			wantErr: "not found error: db: workflow not found",
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctx := t.Context()
			entc, svc := setUpClient(t, logr.Discard())
			sip, _ := createSIP(t, entc, "Test SIP", enums.SIPStatusProcessing)

			var id int
			if tt.want != nil {
				wf, err := entc.Workflow.Create().
					SetUUID(tt.want.UUID).
					SetTemporalID(tt.want.TemporalID).
					SetType(tt.want.Type).
					SetStatus(int8(tt.want.Status)). // #nosec G115 -- constrained value.
					SetStartedAt(tt.want.StartedAt.Time).
					SetSipID(sip.ID).
					Save(ctx)
				assert.NilError(t, err)
				id = wf.ID
			}

			workflow, err := svc.ReadWorkflow(ctx, id)
			if tt.wantErr != "" {
				assert.Error(t, err, tt.wantErr)
				return
			}

			tt.want.ID = id
			assert.DeepEqual(t, workflow, tt.want, cmpopts.EquateEmpty())
		})
	}
}

func TestListWorkflowsBySIP(t *testing.T) {
	t.Parallel()

	workflowUUID1 := uuid.New()
	workflowUUID2 := uuid.New()
	started := sql.NullTime{Time: time.Date(2024, 9, 25, 9, 31, 11, 0, time.UTC), Valid: true}
	started2 := sql.NullTime{Time: time.Date(2024, 9, 25, 10, 3, 42, 0, time.UTC), Valid: true}

	for _, tt := range []struct {
		name      string
		workflows []*datatypes.Workflow
		want      []*datatypes.Workflow
		wantErr   string
	}{
		{
			name: "Returns workflows ordered by started_at desc",
			workflows: []*datatypes.Workflow{
				{
					UUID:       workflowUUID1,
					TemporalID: "processing-workflow-1",
					Type:       enums.WorkflowTypeCreateAip,
					Status:     enums.WorkflowStatusQueued,
					StartedAt:  started,
					SIPUUID:    sipUUID,
				},
				{
					UUID:       workflowUUID2,
					TemporalID: "processing-workflow-2",
					Type:       enums.WorkflowTypeCreateAip,
					Status:     enums.WorkflowStatusInProgress,
					StartedAt:  started2,
					SIPUUID:    sipUUID,
				},
			},
			want: []*datatypes.Workflow{
				{
					ID:         2, // Newer first.
					UUID:       workflowUUID2,
					TemporalID: "processing-workflow-2",
					Type:       enums.WorkflowTypeCreateAip,
					Status:     enums.WorkflowStatusInProgress,
					StartedAt:  started2,
					SIPUUID:    sipUUID,
				},
				{
					ID:         1, // Older second.
					UUID:       workflowUUID1,
					TemporalID: "processing-workflow-1",
					Type:       enums.WorkflowTypeCreateAip,
					Status:     enums.WorkflowStatusQueued,
					StartedAt:  started,
					SIPUUID:    sipUUID,
				},
			},
		},
		{
			name: "Returns empty when none",
			want: []*datatypes.Workflow{},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctx := t.Context()
			entc, svc := setUpClient(t, logr.Discard())
			sip, _ := createSIP(t, entc, "Test SIP", enums.SIPStatusProcessing)

			for _, wf := range tt.workflows {
				_, err := entc.Workflow.Create().
					SetUUID(wf.UUID).
					SetTemporalID(wf.TemporalID).
					SetType(wf.Type).
					SetStatus(int8(wf.Status)). // #nosec G115 -- constrained value.
					SetStartedAt(wf.StartedAt.Time).
					SetSipID(sip.ID).
					Save(ctx)
				assert.NilError(t, err)
			}

			got, err := svc.ListWorkflowsBySIP(ctx, sipUUID)
			if tt.wantErr != "" {
				assert.Error(t, err, tt.wantErr)
				return
			}
			assert.NilError(t, err)
			assert.DeepEqual(t, got, tt.want, cmpopts.EquateEmpty())
		})
	}
}
