package ingest_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"go.artefactual.dev/tools/mockutil"
	"gotest.tools/v3/assert"

	"github.com/artefactual-sdps/enduro/internal/datatypes"
	"github.com/artefactual-sdps/enduro/internal/enums"
	"github.com/artefactual-sdps/enduro/internal/persistence"
	persistence_fake "github.com/artefactual-sdps/enduro/internal/persistence/fake"
)

func TestCreateWorkflow(t *testing.T) {
	t.Parallel()

	sipUUID := uuid.New()
	temporalID := "processing-workflow-720db1d4-825c-4911-9a20-61c212cf23ff"
	startedAt := sql.NullTime{
		Time:  time.Date(2024, 6, 3, 8, 51, 35, 0, time.UTC),
		Valid: true,
	}
	completedAt := sql.NullTime{
		Time:  time.Date(2024, 6, 3, 8, 52, 18, 0, time.UTC),
		Valid: true,
	}

	type test struct {
		name    string
		w       datatypes.Workflow
		mock    func(*persistence_fake.MockService, datatypes.Workflow) *persistence_fake.MockService
		want    datatypes.Workflow
		wantErr string
	}
	for _, tt := range []test{
		{
			name: "Creates a workflow",
			w: datatypes.Workflow{
				TemporalID: temporalID,
				SIPUUID:    sipUUID,
				Type:       enums.WorkflowTypeCreateAip,
			},
			want: datatypes.Workflow{
				ID:         11,
				TemporalID: temporalID,
				Type:       enums.WorkflowTypeCreateAip,
				Status:     enums.WorkflowStatusUnspecified,
				StartedAt: sql.NullTime{
					Time:  time.Date(2024, 6, 3, 9, 4, 23, 0, time.UTC),
					Valid: true,
				},
				SIPUUID: sipUUID,
			},
			mock: func(svc *persistence_fake.MockService, w datatypes.Workflow) *persistence_fake.MockService {
				svc.EXPECT().
					CreateWorkflow(mockutil.Context(), &w).
					DoAndReturn(
						func(ctx context.Context, w *datatypes.Workflow) error {
							w.ID = 11
							w.StartedAt = sql.NullTime{
								Time:  time.Date(2024, 6, 3, 9, 4, 23, 0, time.UTC),
								Valid: true,
							}
							return nil
						},
					)
				return svc
			},
		},
		{
			name: "Creates a workflow with optional values",
			w: datatypes.Workflow{
				TemporalID:  temporalID,
				Type:        enums.WorkflowTypeCreateAip,
				Status:      enums.WorkflowStatusDone,
				StartedAt:   startedAt,
				CompletedAt: completedAt,
				SIPUUID:     sipUUID,
			},
			want: datatypes.Workflow{
				ID:          11,
				TemporalID:  temporalID,
				Type:        enums.WorkflowTypeCreateAip,
				Status:      enums.WorkflowStatusDone,
				StartedAt:   startedAt,
				CompletedAt: completedAt,
				SIPUUID:     sipUUID,
			},
			mock: func(svc *persistence_fake.MockService, w datatypes.Workflow) *persistence_fake.MockService {
				svc.EXPECT().
					CreateWorkflow(mockutil.Context(), &w).
					DoAndReturn(
						func(ctx context.Context, w *datatypes.Workflow) error {
							w.ID = 11
							return nil
						},
					)
				return svc
			},
		},
		{
			name: "Errors when TemporalID is missing",
			w: datatypes.Workflow{
				SIPUUID: sipUUID,
			},
			wantErr: "workflow: create: invalid data error: field \"TemporalID\" is required",
			mock: func(svc *persistence_fake.MockService, w datatypes.Workflow) *persistence_fake.MockService {
				svc.EXPECT().
					CreateWorkflow(mockutil.Context(), &w).
					DoAndReturn(
						func(ctx context.Context, w *datatypes.Workflow) error {
							return errors.New("invalid data error: field \"TemporalID\" is required")
						},
					)
				return svc
			},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ingestsvc, perSvc, _ := testSvc(t, nil, 0)
			if tt.mock != nil {
				tt.mock(perSvc, tt.w)
			}

			w := tt.w
			err := ingestsvc.CreateWorkflow(t.Context(), &w)

			if tt.wantErr != "" {
				assert.Error(t, err, tt.wantErr)
				return
			}

			assert.NilError(t, err)
			assert.DeepEqual(t, w, tt.want)
		})
	}
}

func TestSetWorkflowStatus(t *testing.T) {
	t.Parallel()

	workflowUUID := uuid.New()

	type params struct {
		id     int
		status enums.WorkflowStatus
	}
	type test struct {
		name    string
		params  params
		mock    func(*persistence_fake.MockService) *persistence_fake.MockService
		wantErr string
	}
	for _, tt := range []test{
		{
			name: "Updates the workflow status",
			params: params{
				id:     42,
				status: enums.WorkflowStatusDone,
			},
			mock: func(svc *persistence_fake.MockService) *persistence_fake.MockService {
				svc.EXPECT().
					UpdateWorkflow(
						mockutil.Context(),
						42,
						mockutil.Func(
							"should update workflow status",
							func(upd persistence.WorkflowUpdater) error {
								updated, err := upd(&datatypes.Workflow{UUID: workflowUUID})
								if err != nil {
									return err
								}
								assert.Equal(t, updated.Status, enums.WorkflowStatusDone)
								return nil
							},
						),
					).
					DoAndReturn(
						func(
							ctx context.Context,
							id int,
							upd persistence.WorkflowUpdater,
						) (*datatypes.Workflow, error) {
							return upd(&datatypes.Workflow{ID: id, UUID: workflowUUID})
						},
					)
				return svc
			},
		},
		{
			name: "Errors when UpdateWorkflow fails",
			params: params{
				id:     99,
				status: enums.WorkflowStatusDone,
			},
			mock: func(svc *persistence_fake.MockService) *persistence_fake.MockService {
				svc.EXPECT().
					UpdateWorkflow(
						mockutil.Context(),
						99,
						mockutil.Func(
							"should update workflow status",
							func(upd persistence.WorkflowUpdater) error {
								return nil
							},
						),
					).
					Return(nil, errors.New("db error"))
				return svc
			},
			wantErr: "error updating workflow: db error",
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ingestsvc, perSvc, _ := testSvc(t, nil, 0)
			if tt.mock != nil {
				tt.mock(perSvc)
			}

			err := ingestsvc.SetWorkflowStatus(t.Context(), tt.params.id, tt.params.status)

			if tt.wantErr != "" {
				assert.Error(t, err, tt.wantErr)
				return
			}

			assert.NilError(t, err)
		})
	}
}

func TestCompleteWorkflow(t *testing.T) {
	t.Parallel()

	workflowUUID := uuid.New()
	completedAt := time.Date(2024, 6, 3, 8, 52, 18, 0, time.UTC)

	type params struct {
		id          int
		status      enums.WorkflowStatus
		completedAt time.Time
	}
	type test struct {
		name    string
		params  params
		mock    func(*persistence_fake.MockService) *persistence_fake.MockService
		wantErr string
	}
	for _, tt := range []test{
		{
			name: "Completes the workflow",
			params: params{
				id:          101,
				status:      enums.WorkflowStatusDone,
				completedAt: completedAt,
			},
			mock: func(svc *persistence_fake.MockService) *persistence_fake.MockService {
				svc.EXPECT().
					UpdateWorkflow(
						mockutil.Context(),
						101,
						mockutil.Func(
							"should update workflow completion",
							func(upd persistence.WorkflowUpdater) error {
								updated, err := upd(&datatypes.Workflow{UUID: workflowUUID})
								if err != nil {
									return err
								}
								assert.Equal(t, updated.Status, enums.WorkflowStatusDone)
								assert.DeepEqual(t, updated.CompletedAt, sql.NullTime{Time: completedAt, Valid: true})
								return nil
							},
						),
					).
					DoAndReturn(
						func(
							ctx context.Context,
							id int,
							upd persistence.WorkflowUpdater,
						) (*datatypes.Workflow, error) {
							return upd(&datatypes.Workflow{ID: id, UUID: workflowUUID})
						},
					)
				return svc
			},
		},
		{
			name: "Completes the workflow with error status",
			params: params{
				id:          102,
				status:      enums.WorkflowStatusError,
				completedAt: completedAt,
			},
			mock: func(svc *persistence_fake.MockService) *persistence_fake.MockService {
				svc.EXPECT().
					UpdateWorkflow(
						mockutil.Context(),
						102,
						mockutil.Func(
							"should update workflow with error status",
							func(upd persistence.WorkflowUpdater) error {
								updated, err := upd(&datatypes.Workflow{UUID: workflowUUID})
								if err != nil {
									return err
								}
								assert.Equal(t, updated.Status, enums.WorkflowStatusError)
								assert.DeepEqual(t, updated.CompletedAt, sql.NullTime{Time: completedAt, Valid: true})
								return nil
							},
						),
					).
					DoAndReturn(
						func(
							ctx context.Context,
							id int,
							upd persistence.WorkflowUpdater,
						) (*datatypes.Workflow, error) {
							return upd(&datatypes.Workflow{ID: id, UUID: workflowUUID})
						},
					)
				return svc
			},
		},
		{
			name: "Errors when UpdateWorkflow fails",
			params: params{
				id:          103,
				status:      enums.WorkflowStatusDone,
				completedAt: completedAt,
			},
			mock: func(svc *persistence_fake.MockService) *persistence_fake.MockService {
				svc.EXPECT().
					UpdateWorkflow(
						mockutil.Context(),
						103,
						mockutil.Func(
							"should update workflow completion",
							func(upd persistence.WorkflowUpdater) error {
								return nil
							},
						),
					).
					Return(nil, errors.New("db error"))
				return svc
			},
			wantErr: "error updating workflow: db error",
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ingestsvc, perSvc, _ := testSvc(t, nil, 0)
			if tt.mock != nil {
				tt.mock(perSvc)
			}

			err := ingestsvc.CompleteWorkflow(context.Background(), tt.params.id, tt.params.status, tt.params.completedAt)

			if tt.wantErr != "" {
				assert.Error(t, err, tt.wantErr)
				return
			}

			assert.NilError(t, err)
		})
	}
}
