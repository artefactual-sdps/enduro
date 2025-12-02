package persistence_test

import (
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"go.artefactual.dev/tools/mockutil"
	"go.opentelemetry.io/otel/trace/noop"
	"go.uber.org/mock/gomock"
	"gotest.tools/v3/assert"

	"github.com/artefactual-sdps/enduro/internal/storage/enums"
	"github.com/artefactual-sdps/enduro/internal/storage/persistence"
	"github.com/artefactual-sdps/enduro/internal/storage/persistence/fake"
	"github.com/artefactual-sdps/enduro/internal/storage/types"
)

func TestUpdateAIP(t *testing.T) {
	t.Parallel()

	aipID := uuid.MustParse("123e4567-e89b-12d3-a456-426614174000")

	type test struct {
		name    string
		mock    func(*fake.MockStorage)
		want    *types.AIP
		wantErr string
	}
	for _, tc := range []test{
		{
			name: "Updates an AIP successfully",
			mock: func(svc *fake.MockStorage) {
				svc.EXPECT().
					UpdateAIP(
						mockutil.Context(),
						aipID,
						mockutil.Func(
							"updater function",
							func(updater persistence.AIPUpdater) error {
								updater(&types.AIP{})
								return nil
							},
						),
					).
					Return(
						&types.AIP{
							UUID: aipID,
							Name: "test",
						},
						nil,
					)
			},
			want: &types.AIP{
				UUID: aipID,
				Name: "test",
			},
		},
		{
			name: "Errors when updating an AIP",
			mock: func(svc *fake.MockStorage) {
				svc.EXPECT().
					UpdateAIP(
						mockutil.Context(),
						aipID,
						gomock.Any(),
					).
					Return(nil, errors.New("update aip: not found"))
			},
			wantErr: "UpdateAIP: update aip: not found",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			svc := fake.NewMockStorage(gomock.NewController(t))
			if tc.mock != nil {
				tc.mock(svc)
			}

			tracer := noop.NewTracerProvider().Tracer("test")
			w := persistence.WithTelemetry(svc, tracer)

			got, err := w.UpdateAIP(
				t.Context(),
				aipID,
				func(aip *types.AIP) (*types.AIP, error) {
					aip.Name = "test"
					return aip, nil
				})
			if tc.wantErr != "" {
				assert.Error(t, err, tc.wantErr)
				return
			}

			assert.NilError(t, err)
			assert.DeepEqual(t, got, tc.want)
		})
	}
}

func TestReadWorkflow(t *testing.T) {
	t.Parallel()

	wfID := uuid.New()
	aipID := uuid.New()
	startedAt := time.Now().Truncate(time.Second)

	type test struct {
		name    string
		dbID    int
		mock    func(*fake.MockStorage)
		want    *types.Workflow
		wantErr string
	}
	for _, tc := range []test{
		{
			name: "Reads a workflow successfully",
			dbID: 1,
			mock: func(svc *fake.MockStorage) {
				svc.EXPECT().
					ReadWorkflow(mockutil.Context(), 1).
					Return(
						&types.Workflow{
							DBID:      1,
							UUID:      wfID,
							Type:      enums.WorkflowTypeDeleteAip,
							Status:    enums.WorkflowStatusInProgress,
							StartedAt: startedAt,
							AIPUUID:   aipID,
						},
						nil,
					)
			},
			want: &types.Workflow{
				DBID:      1,
				UUID:      wfID,
				Type:      enums.WorkflowTypeDeleteAip,
				Status:    enums.WorkflowStatusInProgress,
				StartedAt: startedAt,
				AIPUUID:   aipID,
			},
		},
		{
			name: "Errors when reading a workflow",
			dbID: 1,
			mock: func(svc *fake.MockStorage) {
				svc.EXPECT().
					ReadWorkflow(mockutil.Context(), 1).
					Return(nil, errors.New("read workflow: not found"))
			},
			wantErr: "ReadWorkflow: read workflow: not found",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			svc := fake.NewMockStorage(gomock.NewController(t))
			if tc.mock != nil {
				tc.mock(svc)
			}

			tracer := noop.NewTracerProvider().Tracer("test")
			w := persistence.WithTelemetry(svc, tracer)

			wf, err := w.ReadWorkflow(t.Context(), tc.dbID)
			if tc.wantErr != "" {
				assert.Error(t, err, tc.wantErr)
				return
			}

			assert.NilError(t, err)
			assert.DeepEqual(t, wf, tc.want)
		})
	}
}

func TestCreateDeletionRequest(t *testing.T) {
	t.Parallel()

	drID := uuid.New()
	aipID := uuid.New()
	createdAt := time.Now().Truncate(time.Second)

	type test struct {
		name    string
		params  *types.DeletionRequest
		mock    func(*fake.MockStorage)
		wantErr string
	}
	for _, tc := range []test{
		{
			name: "Creates a deletion request successfully",
			params: &types.DeletionRequest{
				UUID:        drID,
				Requester:   "requester@example.com",
				Reason:      "No longer needed",
				Status:      enums.DeletionRequestStatusPending,
				RequestedAt: createdAt,
				AIPUUID:     aipID,
			},
			mock: func(svc *fake.MockStorage) {
				svc.EXPECT().
					CreateDeletionRequest(
						mockutil.Context(),
						&types.DeletionRequest{
							UUID:        drID,
							Requester:   "requester@example.com",
							Reason:      "No longer needed",
							Status:      enums.DeletionRequestStatusPending,
							RequestedAt: createdAt,
							AIPUUID:     aipID,
						},
					).
					Return(nil)
			},
		},
		{
			name: "Errors when creating a deletion request",
			params: &types.DeletionRequest{
				Requester:   "requester@example.com",
				Reason:      "No longer needed",
				Status:      enums.DeletionRequestStatusPending,
				RequestedAt: createdAt,
				AIPUUID:     aipID,
			},
			mock: func(svc *fake.MockStorage) {
				svc.EXPECT().
					CreateDeletionRequest(
						mockutil.Context(),
						&types.DeletionRequest{
							Requester:   "requester@example.com",
							Reason:      "No longer needed",
							Status:      enums.DeletionRequestStatusPending,
							RequestedAt: createdAt,
							AIPUUID:     aipID,
						},
					).
					Return(errors.New("create deletion request: db: aip not found"))
			},
			wantErr: "CreateDeletionRequest: create deletion request: db: aip not found",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			svc := fake.NewMockStorage(gomock.NewController(t))
			if tc.mock != nil {
				tc.mock(svc)
			}

			tracer := noop.NewTracerProvider().Tracer("test")
			w := persistence.WithTelemetry(svc, tracer)

			err := w.CreateDeletionRequest(t.Context(), tc.params)
			if tc.wantErr != "" {
				assert.Error(t, err, tc.wantErr)
				return
			}

			assert.NilError(t, err)
		})
	}
}

func TestListDeletionRequests(t *testing.T) {
	t.Parallel()

	drID := uuid.New()
	aipID := uuid.New()

	type test struct {
		name    string
		filter  *persistence.DeletionRequestFilter
		mock    func(*fake.MockStorage)
		want    []*types.DeletionRequest
		wantErr string
	}
	for _, tc := range []test{
		{
			name:   "Lists deletion requests successfully",
			filter: &persistence.DeletionRequestFilter{},
			mock: func(svc *fake.MockStorage) {
				svc.EXPECT().
					ListDeletionRequests(
						mockutil.Context(),
						&persistence.DeletionRequestFilter{},
					).
					Return(
						[]*types.DeletionRequest{
							{
								DBID:         1,
								UUID:         drID,
								Requester:    "requester@example.com",
								Reason:       "No longer needed",
								Status:       enums.DeletionRequestStatusPending,
								AIPUUID:      aipID,
								WorkflowDBID: 10,
							},
						},
						nil,
					)
			},
			want: []*types.DeletionRequest{
				{
					DBID:         1,
					UUID:         drID,
					Requester:    "requester@example.com",
					Reason:       "No longer needed",
					Status:       enums.DeletionRequestStatusPending,
					AIPUUID:      aipID,
					WorkflowDBID: 10,
				},
			},
		},
		{
			name:   "Errors when listing deletion requests",
			filter: &persistence.DeletionRequestFilter{},
			mock: func(svc *fake.MockStorage) {
				svc.EXPECT().
					ListDeletionRequests(
						mockutil.Context(),
						&persistence.DeletionRequestFilter{},
					).
					Return(nil, errors.New("list deletion requests: internal error"))
			},
			wantErr: "ListDeletionRequests: list deletion requests: internal error",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			svc := fake.NewMockStorage(gomock.NewController(t))
			if tc.mock != nil {
				tc.mock(svc)
			}

			tracer := noop.NewTracerProvider().Tracer("test")
			w := persistence.WithTelemetry(svc, tracer)

			got, err := w.ListDeletionRequests(t.Context(), tc.filter)
			if tc.wantErr != "" {
				assert.Error(t, err, tc.wantErr)
				return
			}

			assert.NilError(t, err)
			assert.DeepEqual(t, got, tc.want)
		})
	}
}

func TestUpdateDeletionRequest(t *testing.T) {
	t.Parallel()

	dbID := 1
	drUUID := uuid.New()
	aipID := uuid.New()
	requestedAt := time.Now().Truncate(time.Second)
	reviewedAt := time.Now().Truncate(time.Second)

	type test struct {
		name    string
		id      int
		updater persistence.DeletionRequestUpdater
		mock    func(*fake.MockStorage)
		want    *types.DeletionRequest
		wantErr string
	}
	for _, tc := range []test{
		{
			name: "Updates a deletion request successfully",
			id:   dbID,
			updater: func(dr *types.DeletionRequest) (*types.DeletionRequest, error) {
				dr.Reviewer = "reviewer@example.com"
				dr.ReviewerIss = "issuer"
				dr.ReviewerSub = "subject"
				dr.ReviewedAt = reviewedAt
				dr.Status = enums.DeletionRequestStatusApproved
				return dr, nil
			},
			mock: func(svc *fake.MockStorage) {
				svc.EXPECT().
					UpdateDeletionRequest(
						mockutil.Context(),
						dbID,
						mockutil.Func(
							"updater function",
							func(updater persistence.DeletionRequestUpdater) error {
								dr := &types.DeletionRequest{
									DBID:         dbID,
									UUID:         drUUID,
									Requester:    "requester@example.com",
									RequesterIss: "issuer",
									RequesterSub: "subject",
									Reason:       "Test reason",
									Status:       enums.DeletionRequestStatusPending,
									RequestedAt:  requestedAt,
									AIPUUID:      aipID,
									WorkflowDBID: 10,
								}

								if _, err := updater(dr); err != nil {
									return err
								}

								return nil
							},
						),
					).
					Return(
						&types.DeletionRequest{
							DBID:         dbID,
							UUID:         drUUID,
							Requester:    "requester@example.com",
							RequesterIss: "issuer",
							RequesterSub: "subject",
							Reviewer:     "reviewer@example.com",
							ReviewerIss:  "issuer",
							ReviewerSub:  "subject",
							Reason:       "Test reason",
							Status:       enums.DeletionRequestStatusApproved,
							RequestedAt:  requestedAt,
							ReviewedAt:   reviewedAt,
							AIPUUID:      aipID,
							WorkflowDBID: 10,
						},
						nil,
					)
			},
			want: &types.DeletionRequest{
				DBID:         dbID,
				UUID:         drUUID,
				Requester:    "requester@example.com",
				RequesterIss: "issuer",
				RequesterSub: "subject",
				Reviewer:     "reviewer@example.com",
				ReviewerIss:  "issuer",
				ReviewerSub:  "subject",
				Reason:       "Test reason",
				Status:       enums.DeletionRequestStatusApproved,
				RequestedAt:  requestedAt,
				ReviewedAt:   reviewedAt,
				AIPUUID:      aipID,
				WorkflowDBID: 10,
			},
		},
		{
			name: "Errors when updating a deletion request (not found)",
			id:   999,
			updater: func(dr *types.DeletionRequest) (*types.DeletionRequest, error) {
				return dr, nil
			},
			mock: func(svc *fake.MockStorage) {
				svc.EXPECT().
					UpdateDeletionRequest(
						mockutil.Context(),
						999,
						mockutil.Func(
							"updater function",
							func(updater persistence.DeletionRequestUpdater) error {
								return nil
							},
						),
					).
					Return(nil, errors.New("update deletion request: not found"))
			},
			wantErr: "UpdateDeletionRequest: update deletion request: not found",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			svc := fake.NewMockStorage(gomock.NewController(t))
			if tc.mock != nil {
				tc.mock(svc)
			}

			tracer := noop.NewTracerProvider().Tracer("test")
			w := persistence.WithTelemetry(svc, tracer)

			got, err := w.UpdateDeletionRequest(t.Context(), tc.id, tc.updater)
			if tc.wantErr != "" {
				assert.Error(t, err, tc.wantErr)
				return
			}

			assert.NilError(t, err)
			assert.DeepEqual(t, got, tc.want)
		})
	}
}

func TestReadDeletionRequest(t *testing.T) {
	t.Parallel()

	drID := uuid.New()
	aipID := uuid.New()

	type test struct {
		name    string
		id      uuid.UUID
		mock    func(*fake.MockStorage)
		want    *types.DeletionRequest
		wantErr string
	}
	for _, tc := range []test{
		{
			name: "Reads a deletion request successfully",
			id:   drID,
			mock: func(svc *fake.MockStorage) {
				svc.EXPECT().
					ReadDeletionRequest(mockutil.Context(), drID).
					Return(
						&types.DeletionRequest{
							DBID:         1,
							UUID:         drID,
							Requester:    "requester@example.com",
							Reason:       "No longer needed",
							Status:       enums.DeletionRequestStatusPending,
							AIPUUID:      aipID,
							WorkflowDBID: 10,
						},
						nil,
					)
			},
			want: &types.DeletionRequest{
				DBID:         1,
				UUID:         drID,
				Requester:    "requester@example.com",
				Reason:       "No longer needed",
				Status:       enums.DeletionRequestStatusPending,
				AIPUUID:      aipID,
				WorkflowDBID: 10,
			},
		},
		{
			name: "Errors when reading a deletion request",
			id:   drID,
			mock: func(svc *fake.MockStorage) {
				svc.EXPECT().
					ReadDeletionRequest(mockutil.Context(), drID).
					Return(nil, errors.New("read deletion request: not found"))
			},
			wantErr: "ReadDeletionRequest: read deletion request: not found",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			svc := fake.NewMockStorage(gomock.NewController(t))
			if tc.mock != nil {
				tc.mock(svc)
			}

			tracer := noop.NewTracerProvider().Tracer("test")
			w := persistence.WithTelemetry(svc, tracer)

			got, err := w.ReadDeletionRequest(t.Context(), tc.id)
			if tc.wantErr != "" {
				assert.Error(t, err, tc.wantErr)
				return
			}

			assert.NilError(t, err)
			assert.DeepEqual(t, got, tc.want)
		})
	}
}
