package workflow

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"go.artefactual.dev/tools/mockutil"
	temporalsdk_testsuite "go.temporal.io/sdk/testsuite"
	"go.uber.org/mock/gomock"
	"gotest.tools/v3/assert"

	"github.com/artefactual-sdps/enduro/internal/datatypes"
	"github.com/artefactual-sdps/enduro/internal/enums"
	ingest_fake "github.com/artefactual-sdps/enduro/internal/ingest/fake"
	"github.com/artefactual-sdps/enduro/internal/persistence"
)

func TestCreateWorkflowLocalActivity(t *testing.T) {
	t.Parallel()

	sipUUID := uuid.New()
	startedAt := time.Date(2024, 6, 13, 17, 50, 13, 0, time.UTC)
	completedAt := time.Date(2024, 6, 13, 17, 50, 14, 0, time.UTC)

	type test struct {
		name      string
		params    *createWorkflowLocalActivityParams
		mockCalls func(m *ingest_fake.MockServiceMockRecorder)
		want      uint
		wantErr   string
	}
	for _, tt := range []test{
		{
			name: "Creates a workflow",
			params: &createWorkflowLocalActivityParams{
				TemporalID:  "workflow-id",
				Type:        enums.WorkflowTypeCreateAip,
				Status:      enums.WorkflowStatusDone,
				StartedAt:   startedAt,
				CompletedAt: completedAt,
				SIPUUID:     sipUUID,
			},
			mockCalls: func(m *ingest_fake.MockServiceMockRecorder) {
				m.CreateWorkflow(mockutil.Context(), &datatypes.Workflow{
					TemporalID:  "workflow-id",
					Type:        enums.WorkflowTypeCreateAip,
					Status:      enums.WorkflowStatusDone,
					StartedAt:   sql.NullTime{Time: startedAt, Valid: true},
					CompletedAt: sql.NullTime{Time: completedAt, Valid: true},
					SIPUUID:     sipUUID,
				}).DoAndReturn(func(ctx context.Context, w *datatypes.Workflow) error {
					w.ID = 1
					return nil
				})
			},
			want: 1,
		},
		{
			name: "Does not pass zero dates",
			params: &createWorkflowLocalActivityParams{
				TemporalID: "workflow-id",
				Type:       enums.WorkflowTypeCreateAip,
				Status:     enums.WorkflowStatusDone,
				SIPUUID:    sipUUID,
			},
			mockCalls: func(m *ingest_fake.MockServiceMockRecorder) {
				m.CreateWorkflow(mockutil.Context(), &datatypes.Workflow{
					TemporalID: "workflow-id",
					Type:       enums.WorkflowTypeCreateAip,
					Status:     enums.WorkflowStatusDone,
					SIPUUID:    sipUUID,
				}).DoAndReturn(func(ctx context.Context, w *datatypes.Workflow) error {
					w.ID = 1
					return nil
				})
			},
			want: 1,
		},
		{
			name: "Fails if there is a persistence error",
			params: &createWorkflowLocalActivityParams{
				TemporalID: "workflow-id",
				Type:       enums.WorkflowTypeCreateAip,
				Status:     enums.WorkflowStatusDone,
				SIPUUID:    sipUUID,
			},
			mockCalls: func(m *ingest_fake.MockServiceMockRecorder) {
				m.CreateWorkflow(mockutil.Context(), &datatypes.Workflow{
					TemporalID: "workflow-id",
					Type:       enums.WorkflowTypeCreateAip,
					Status:     enums.WorkflowStatusDone,
					SIPUUID:    sipUUID,
				}).Return(fmt.Errorf("persistence error"))
			},
			wantErr: "persistence error",
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ts := &temporalsdk_testsuite.WorkflowTestSuite{}
			env := ts.NewTestActivityEnvironment()
			svc := ingest_fake.NewMockService(gomock.NewController(t))
			if tt.mockCalls != nil {
				tt.mockCalls(svc.EXPECT())
			}

			enc, err := env.ExecuteLocalActivity(
				createWorkflowLocalActivity,
				svc,
				tt.params,
			)
			if tt.wantErr != "" {
				assert.ErrorContains(t, err, tt.wantErr)
				return
			}
			assert.NilError(t, err)

			var res uint
			_ = enc.Get(&res)
			assert.DeepEqual(t, res, tt.want)
		})
	}
}

func TestUpdateSIPLocalActivity(t *testing.T) {
	t.Parallel()

	sipUUID := uuid.New()
	aipUUID := uuid.New()
	name := "Test SIP"
	completedAt := time.Now()

	for _, tt := range []struct {
		name      string
		params    *updateSIPLocalActivityParams
		mockCalls func(context.Context, *ingest_fake.MockService)
		wantErr   string
	}{
		{
			name: "Updates a SIP",
			params: &updateSIPLocalActivityParams{
				UUID:        sipUUID,
				Name:        name,
				Status:      enums.SIPStatusIngested,
				CompletedAt: completedAt,
				AIPUUID:     aipUUID.String(),
			},
			mockCalls: func(ctx context.Context, svc *ingest_fake.MockService) {
				svc.EXPECT().
					UpdateSIP(
						ctx,
						sipUUID,
						mockutil.Func(
							"should update sip",
							func(updater persistence.SIPUpdater) error {
								s, err := updater(&datatypes.SIP{})
								assert.NilError(t, err)
								assert.DeepEqual(t, s.Name, name)
								assert.DeepEqual(t, s.Status, enums.SIPStatusIngested)
								assert.DeepEqual(t, s.CompletedAt, sql.NullTime{Valid: true, Time: completedAt})
								assert.DeepEqual(t, s.AIPID, uuid.NullUUID{Valid: true, UUID: aipUUID})
								return nil
							},
						),
					).
					Return(nil, nil)
			},
		},
		{
			name: "Fails to update a SIP (invalid status)",
			params: &updateSIPLocalActivityParams{
				UUID:        sipUUID,
				Name:        name,
				Status:      "",
				CompletedAt: completedAt,
				AIPUUID:     aipUUID.String(),
			},
			mockCalls: func(ctx context.Context, svc *ingest_fake.MockService) {
				svc.EXPECT().
					UpdateSIP(
						ctx,
						sipUUID,
						mockutil.Func(
							"should fail to update sip",
							func(updater persistence.SIPUpdater) error {
								_, err := updater(&datatypes.SIP{})
								assert.ErrorContains(t, err, "invalid status: ")
								return nil
							},
						),
					).
					Return(nil, errors.New("invalid status: "))
			},
			wantErr: "invalid status: ",
		},
		{
			name: "Fails to update a SIP (invalid AIP UUID)",
			params: &updateSIPLocalActivityParams{
				UUID:        sipUUID,
				Name:        name,
				Status:      enums.SIPStatusIngested,
				CompletedAt: completedAt,
				AIPUUID:     "invalid-uuid",
			},
			mockCalls: func(ctx context.Context, svc *ingest_fake.MockService) {
				svc.EXPECT().
					UpdateSIP(
						ctx,
						sipUUID,
						mockutil.Func(
							"should fail to update sip",
							func(updater persistence.SIPUpdater) error {
								_, err := updater(&datatypes.SIP{})
								assert.ErrorContains(t, err, "invalid AIP UUID: invalid-uuid")
								return nil
							},
						),
					).
					Return(nil, errors.New("invalid AIP UUID: invalid-uuid"))
			},
			wantErr: "invalid AIP UUID: invalid-uuid",
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctx := t.Context()
			svc := ingest_fake.NewMockService(gomock.NewController(t))
			if tt.mockCalls != nil {
				tt.mockCalls(ctx, svc)
			}

			re, err := updateSIPLocalActivity(ctx, svc, tt.params)
			if tt.wantErr != "" {
				assert.ErrorContains(t, err, tt.wantErr)
				return
			}
			assert.NilError(t, err)
			assert.DeepEqual(t, re, &updateSIPLocalActivityResult{})
		})
	}
}
