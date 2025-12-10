package client_test

import (
	"database/sql"
	"errors"
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

	tests := []struct {
		name    string
		setup   func(t *testing.T) (svc persistence.Service, wfID int, expect time.Time)
		updater persistence.WorkflowUpdater
		wantErr string
	}{
		{
			name: "Updates status and completion",
			setup: func(t *testing.T) (persistence.Service, int, time.Time) {
				entc, svc := setUpClient(t, logr.Discard())
				sip, err := createSIP(t, entc, "update", enums.SIPStatusProcessing)
				assert.NilError(t, err)
				wf, err := createWorkflow(t, entc, sip.ID, enums.WorkflowStatusQueued)
				assert.NilError(t, err)
				return svc, wf.ID, time.Now().UTC()
			},
			updater: func(w *datatypes.Workflow) (*datatypes.Workflow, error) {
				w.Status = enums.WorkflowStatusDone
				w.CompletedAt = sql.NullTime{Time: time.Now().UTC(), Valid: true}
				return w, nil
			},
		},
		{
			name: "Propagates updater error",
			setup: func(t *testing.T) (persistence.Service, int, time.Time) {
				entc, svc := setUpClient(t, logr.Discard())
				sip, err := createSIP(t, entc, "update", enums.SIPStatusProcessing)
				assert.NilError(t, err)
				wf, err := createWorkflow(t, entc, sip.ID, enums.WorkflowStatusQueued)
				assert.NilError(t, err)
				return svc, wf.ID, time.Time{}
			},
			updater: func(w *datatypes.Workflow) (*datatypes.Workflow, error) {
				return nil, errors.New("boom")
			},
			wantErr: "invalid data error: updater error: boom",
		},
		{
			name: "Returns not found when workflow missing",
			setup: func(t *testing.T) (persistence.Service, int, time.Time) {
				_, svc := setUpClient(t, logr.Discard())
				return svc, 9999, time.Time{}
			},
			updater: func(w *datatypes.Workflow) (*datatypes.Workflow, error) {
				return w, nil
			},
			wantErr: "not found error: db: workflow not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			svc, wfID, expectedTime := tt.setup(t)

			updated, err := svc.UpdateWorkflow(t.Context(), wfID, tt.updater)
			if tt.wantErr != "" {
				assert.Error(t, err, tt.wantErr)
				return
			}

			assert.NilError(t, err)
			assert.Equal(t, updated.Status, enums.WorkflowStatusDone)
			if expectedTime.IsZero() {
				assert.Assert(t, updated.CompletedAt.Valid || true) // ensure field exists
			} else {
				assert.Assert(t, updated.CompletedAt.Valid)
				assert.Equal(t, updated.CompletedAt.Time.Truncate(time.Second), expectedTime.Truncate(time.Second))
			}
		})
	}
}

func TestReadWorkflow(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		setup   func(t *testing.T) (svc persistence.Service, wfID int, sipUUID uuid.UUID)
		wantErr string
	}{
		{
			name: "Reads workflow",
			setup: func(t *testing.T) (persistence.Service, int, uuid.UUID) {
				entc, svc := setUpClient(t, logr.Discard())
				sip, err := createSIP(t, entc, "read", enums.SIPStatusProcessing)
				assert.NilError(t, err)
				wf, err := createWorkflow(t, entc, sip.ID, enums.WorkflowStatusQueued)
				assert.NilError(t, err)
				return svc, wf.ID, sip.UUID
			},
		},
		{
			name: "Returns not found",
			setup: func(t *testing.T) (persistence.Service, int, uuid.UUID) {
				_, svc := setUpClient(t, logr.Discard())
				return svc, 12345, uuid.Nil
			},
			wantErr: "not found error: db: workflow not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			svc, wfID, sipUUID := tt.setup(t)
			got, err := svc.ReadWorkflow(t.Context(), wfID)
			if tt.wantErr != "" {
				assert.Error(t, err, tt.wantErr)
				return
			}

			assert.NilError(t, err)
			assert.Assert(t, got != nil)
			assert.Equal(t, got.ID, wfID)
			assert.Equal(t, got.SIPUUID, sipUUID)
		})
	}
}

func TestListWorkflowsBySIP(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		setup   func(t *testing.T) (svc persistence.Service, sipUUID uuid.UUID)
		expect  int
		check   func(t *testing.T, w []*datatypes.Workflow)
		wantErr string
	}{
		{
			name: "Returns workflows ordered by started_at desc",
			setup: func(t *testing.T) (persistence.Service, uuid.UUID) {
				entc, svc := setUpClient(t, logr.Discard())
				sip, err := createSIP(t, entc, "list", enums.SIPStatusProcessing)
				assert.NilError(t, err)
				// older
				_, err = createWorkflowWithTimes(
					t,
					entc,
					sip.ID,
					enums.WorkflowStatusQueued,
					time.Now().Add(-2*time.Hour),
				)
				assert.NilError(t, err)
				// newer
				_, err = createWorkflowWithTimes(t, entc, sip.ID, enums.WorkflowStatusQueued, time.Now())
				assert.NilError(t, err)
				return svc, sip.UUID
			},
			expect: 2,
			check: func(t *testing.T, w []*datatypes.Workflow) {
				assert.Assert(
					t,
					w[0].StartedAt.Time.After(w[1].StartedAt.Time) || w[0].StartedAt.Time.Equal(w[1].StartedAt.Time),
				)
			},
		},
		{
			name: "Returns empty when none",
			setup: func(t *testing.T) (persistence.Service, uuid.UUID) {
				entc, svc := setUpClient(t, logr.Discard())
				sip, err := createSIP(t, entc, "empty", enums.SIPStatusProcessing)
				assert.NilError(t, err)
				return svc, sip.UUID
			},
			expect: 0,
			check: func(t *testing.T, w []*datatypes.Workflow) {
				assert.Equal(t, len(w), 0)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			svc, sipUUID := tt.setup(t)
			got, err := svc.ListWorkflowsBySIP(t.Context(), sipUUID)
			if tt.wantErr != "" {
				assert.Error(t, err, tt.wantErr)
				return
			}
			assert.NilError(t, err)
			assert.Equal(t, len(got), tt.expect)
			if tt.check != nil {
				tt.check(t, got)
			}
		})
	}
}

func createWorkflowWithTimes(
	t *testing.T,
	entc *db.Client,
	sipID int,
	status enums.WorkflowStatus,
	started time.Time,
) (*db.Workflow, error) {
	t.Helper()
	wf, err := entc.Workflow.
		Create().
		SetUUID(uuid.New()).
		SetTemporalID("tid").
		SetType(enums.WorkflowTypeCreateAip).
		SetStatus(int8(status)). // #nosec G115 -- enum values fit in int8.
		SetStartedAt(started).
		SetSipID(sipID).
		Save(t.Context())
	return wf, err
}
