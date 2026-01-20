package ingest_test

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"go.artefactual.dev/tools/mockutil"
	"go.artefactual.dev/tools/ref"
	temporalsdk_api_enums "go.temporal.io/api/enums/v1"
	temporalsdk_client "go.temporal.io/sdk/client"
	temporalsdk_mocks "go.temporal.io/sdk/mocks"
	"gotest.tools/v3/assert"

	"github.com/artefactual-sdps/enduro/internal/api/auth"
	goaingest "github.com/artefactual-sdps/enduro/internal/api/gen/ingest"
	"github.com/artefactual-sdps/enduro/internal/datatypes"
	"github.com/artefactual-sdps/enduro/internal/entfilter"
	"github.com/artefactual-sdps/enduro/internal/enums"
	"github.com/artefactual-sdps/enduro/internal/ingest"
	"github.com/artefactual-sdps/enduro/internal/persistence"
	persistence_fake "github.com/artefactual-sdps/enduro/internal/persistence/fake"
	"github.com/artefactual-sdps/enduro/internal/timerange"
)

func TestAddBatch(t *testing.T) {
	t.Parallel()

	sourceID := uuid.MustParse("123e4567-e89b-12d3-a456-426614174000")
	batchUUID := uuid.MustParse("52fdfc07-2182-454f-963f-5f0f9a621d72")
	userUUID := uuid.MustParse("9566c74d-1003-4c4d-bbbb-0407d1e2c649")
	keys := []string{"sip1.zip", "sip2.zip", "sip3.zip"}
	identifier := ref.New("custom-identifier")
	batch := &datatypes.Batch{
		UUID:       batchUUID,
		Identifier: fmt.Sprintf("Batch-%s", batchUUID.String()),
		Status:     enums.BatchStatusQueued,
		SIPSCount:  len(keys),
	}
	batchWithUploader := &datatypes.Batch{
		UUID:       batchUUID,
		Identifier: fmt.Sprintf("Batch-%s", batchUUID.String()),
		Status:     enums.BatchStatusQueued,
		SIPSCount:  len(keys),
		Uploader: &datatypes.User{
			UUID:    userUUID,
			Email:   "nobody@example.com",
			Name:    "Test User",
			OIDCIss: "http://keycloak:7470/realms/artefactual",
			OIDCSub: "1234567890",
		},
	}

	for _, tt := range []struct {
		name    string
		payload *goaingest.AddBatchPayload
		claims  *auth.Claims
		mock    func(context.Context, *persistence_fake.MockService, *temporalsdk_mocks.Client)
		want    *goaingest.AddBatchResult
		wantErr string
	}{
		{
			name:    "Returns not valid error (missing payload)",
			wantErr: "missing payload",
		},
		{
			name:    "Returns not valid error (invalid SourceID)",
			payload: &goaingest.AddBatchPayload{SourceID: "invalid"},
			wantErr: "invalid SourceID",
		},
		{
			name:    "Returns not valid error (missing keys)",
			payload: &goaingest.AddBatchPayload{SourceID: sourceID.String()},
			wantErr: "empty Keys",
		},
		{
			name:    "Returns not valid error (invalid claims Iss)",
			payload: &goaingest.AddBatchPayload{SourceID: sourceID.String(), Keys: keys},
			claims:  &auth.Claims{},
			wantErr: "invalid user claims: missing Iss",
		},
		{
			name:    "Returns not valid error (invalid claims Sub)",
			payload: &goaingest.AddBatchPayload{SourceID: sourceID.String(), Keys: keys},
			claims:  &auth.Claims{Iss: "http://keycloak:7470/realms/artefactual"},
			wantErr: "invalid user claims: missing Sub",
		},
		{
			name:    "Returns persistence error",
			payload: &goaingest.AddBatchPayload{SourceID: sourceID.String(), Keys: keys, Identifier: identifier},
			mock: func(ctx context.Context, psvc *persistence_fake.MockService, tc *temporalsdk_mocks.Client) {
				psvc.EXPECT().CreateBatch(
					ctx,
					&datatypes.Batch{
						UUID:       batchUUID,
						Identifier: *identifier,
						Status:     enums.BatchStatusQueued,
						SIPSCount:  len(keys),
					},
				).Return(errors.New("persistence error"))
			},
			wantErr: "internal error",
		},
		{
			name:    "Returns Temporal error",
			payload: &goaingest.AddBatchPayload{SourceID: sourceID.String(), Keys: keys},
			mock: func(ctx context.Context, psvc *persistence_fake.MockService, tc *temporalsdk_mocks.Client) {
				psvc.EXPECT().
					CreateBatch(ctx, batch).
					Return(nil)

				tc.On(
					"ExecuteWorkflow",
					mock.AnythingOfType("*context.timerCtx"),
					temporalsdk_client.StartWorkflowOptions{
						ID:                    fmt.Sprintf("%s-%s", ingest.BatchWorkflowName, batchUUID.String()),
						TaskQueue:             "test",
						WorkflowIDReusePolicy: temporalsdk_api_enums.WORKFLOW_ID_REUSE_POLICY_ALLOW_DUPLICATE,
					},
					ingest.BatchWorkflowName,
					&ingest.BatchWorkflowRequest{
						Batch:       *batch,
						SIPSourceID: sourceID,
						Keys:        keys,
					},
				).Return(nil, errors.New("temporal error"))

				psvc.EXPECT().DeleteBatch(ctx, batchUUID)
			},
			wantErr: "internal error",
		},
		{
			name:    "Uploads a SIP",
			payload: &goaingest.AddBatchPayload{SourceID: sourceID.String(), Keys: keys},
			claims:  nil,
			mock: func(ctx context.Context, psvc *persistence_fake.MockService, tc *temporalsdk_mocks.Client) {
				psvc.EXPECT().
					CreateBatch(mockutil.Context(), batch).
					Return(nil)

				tc.On(
					"ExecuteWorkflow",
					mock.AnythingOfType("*context.timerCtx"),
					temporalsdk_client.StartWorkflowOptions{
						ID:                    fmt.Sprintf("%s-%s", ingest.BatchWorkflowName, batchUUID.String()),
						TaskQueue:             "test",
						WorkflowIDReusePolicy: temporalsdk_api_enums.WORKFLOW_ID_REUSE_POLICY_ALLOW_DUPLICATE,
					},
					ingest.BatchWorkflowName,
					&ingest.BatchWorkflowRequest{
						Batch:       *batch,
						SIPSourceID: sourceID,
						Keys:        keys,
					},
				).Return(nil, nil)
			},
			want: &goaingest.AddBatchResult{UUID: batchUUID.String()},
		},
		{
			name:    "Uploads a SIP and creates a user",
			payload: &goaingest.AddBatchPayload{SourceID: sourceID.String(), Keys: keys},
			claims: &auth.Claims{
				Email: "nobody@example.com",
				Name:  "Test User",
				Iss:   "http://keycloak:7470/realms/artefactual",
				Sub:   "1234567890",
			},
			mock: func(ctx context.Context, psvc *persistence_fake.MockService, tc *temporalsdk_mocks.Client) {
				psvc.EXPECT().CreateBatch(
					mockutil.Context(),
					batchWithUploader,
				).Return(nil)

				tc.On(
					"ExecuteWorkflow",
					mock.AnythingOfType("*context.timerCtx"),
					temporalsdk_client.StartWorkflowOptions{
						ID:                    fmt.Sprintf("%s-%s", ingest.BatchWorkflowName, batchUUID.String()),
						TaskQueue:             "test",
						WorkflowIDReusePolicy: temporalsdk_api_enums.WORKFLOW_ID_REUSE_POLICY_ALLOW_DUPLICATE,
					},
					ingest.BatchWorkflowName,
					&ingest.BatchWorkflowRequest{
						Batch:       *batchWithUploader,
						SIPSourceID: sourceID,
						Keys:        keys,
					},
				).Return(nil, nil)
			},
			want: &goaingest.AddBatchResult{UUID: batchUUID.String()},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			svc, psvc, tc := testSvc(t, nil, 0)
			ctx := t.Context()
			if tt.mock != nil {
				tt.mock(ctx, psvc, tc)
			}

			if tt.claims != nil {
				ctx = auth.WithUserClaims(ctx, tt.claims)
			}

			re, err := svc.AddBatch(ctx, tt.payload)
			if tt.wantErr != "" {
				assert.Error(t, err, tt.wantErr)
				return
			}

			assert.NilError(t, err)
			assert.DeepEqual(t, re, tt.want)
		})
	}
}

func TestShowBatch(t *testing.T) {
	t.Parallel()

	batchUUID := uuid.New()
	userUUID := uuid.New()
	createdAt := time.Date(2024, 9, 25, 9, 31, 10, 0, time.UTC)
	startedAt := time.Date(2024, 9, 25, 9, 31, 11, 0, time.UTC)
	completedAt := time.Date(2024, 9, 25, 9, 31, 12, 0, time.UTC)

	for _, tt := range []struct {
		name    string
		payload *goaingest.ShowBatchPayload
		mock    func(context.Context, *persistence_fake.MockService)
		want    *goaingest.Batch
		wantErr string
	}{
		{
			name:    "Returns not valid error (invalid UUID)",
			payload: &goaingest.ShowBatchPayload{UUID: "invalid"},
			wantErr: "invalid UUID",
		},
		{
			name:    "Returns not found error",
			payload: &goaingest.ShowBatchPayload{UUID: batchUUID.String()},
			mock: func(ctx context.Context, psvc *persistence_fake.MockService) {
				psvc.EXPECT().ReadBatch(ctx, batchUUID).Return(nil, persistence.ErrNotFound)
			},
			wantErr: "Batch not found.",
		},
		{
			name:    "Returns internal error",
			payload: &goaingest.ShowBatchPayload{UUID: batchUUID.String()},
			mock: func(ctx context.Context, psvc *persistence_fake.MockService) {
				psvc.EXPECT().ReadBatch(ctx, batchUUID).Return(nil, persistence.ErrInternal)
			},
			wantErr: "internal error",
		},
		{
			name:    "Returns batch",
			payload: &goaingest.ShowBatchPayload{UUID: batchUUID.String()},
			mock: func(ctx context.Context, psvc *persistence_fake.MockService) {
				psvc.EXPECT().ReadBatch(ctx, batchUUID).Return(&datatypes.Batch{
					UUID:        batchUUID,
					Identifier:  "batch-identifier",
					Status:      enums.BatchStatusIngested,
					SIPSCount:   3,
					CreatedAt:   createdAt,
					StartedAt:   startedAt,
					CompletedAt: completedAt,
					Uploader: &datatypes.User{
						UUID:  userUUID,
						Email: "nobody@example.com",
						Name:  "Test User",
					},
				}, nil)
			},
			want: &goaingest.Batch{
				UUID:          batchUUID,
				Identifier:    "batch-identifier",
				Status:        enums.BatchStatusIngested.String(),
				SipsCount:     3,
				CreatedAt:     createdAt.Format(time.RFC3339),
				StartedAt:     ref.New(startedAt.Format(time.RFC3339)),
				CompletedAt:   ref.New(completedAt.Format(time.RFC3339)),
				UploaderUUID:  ref.New(userUUID),
				UploaderEmail: ref.New("nobody@example.com"),
				UploaderName:  ref.New("Test User"),
			},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			svc, psvc, _ := testSvc(t, nil, 0)
			ctx := t.Context()
			if tt.mock != nil {
				tt.mock(ctx, psvc)
			}

			got, err := svc.ShowBatch(ctx, tt.payload)
			if tt.wantErr != "" {
				assert.Error(t, err, tt.wantErr)
				return
			}

			assert.NilError(t, err)
			assert.DeepEqual(t, got, tt.want)
		})
	}
}

func TestReviewBatch(t *testing.T) {
	t.Parallel()

	batchUUID := uuid.New()

	for _, tt := range []struct {
		name    string
		payload *goaingest.ReviewBatchPayload
		mock    func(context.Context, *persistence_fake.MockService, *temporalsdk_mocks.Client)
		wantErr string
	}{
		{
			name:    "Returns not valid error (invalid UUID)",
			payload: &goaingest.ReviewBatchPayload{UUID: "invalid"},
			wantErr: "invalid UUID",
		},
		{
			name:    "Returns not found error",
			payload: &goaingest.ReviewBatchPayload{UUID: batchUUID.String()},
			mock: func(ctx context.Context, psvc *persistence_fake.MockService, _ *temporalsdk_mocks.Client) {
				psvc.EXPECT().ReadBatch(ctx, batchUUID).Return(nil, persistence.ErrNotFound)
			},
			wantErr: "Batch not found.",
		},
		{
			name:    "Returns internal error (persistence failure)",
			payload: &goaingest.ReviewBatchPayload{UUID: batchUUID.String()},
			mock: func(ctx context.Context, psvc *persistence_fake.MockService, _ *temporalsdk_mocks.Client) {
				psvc.EXPECT().ReadBatch(ctx, batchUUID).Return(nil, persistence.ErrInternal)
			},
			wantErr: "internal error",
		},
		{
			name:    "Returns not valid error (not pending)",
			payload: &goaingest.ReviewBatchPayload{UUID: batchUUID.String()},
			mock: func(ctx context.Context, psvc *persistence_fake.MockService, _ *temporalsdk_mocks.Client) {
				psvc.EXPECT().ReadBatch(ctx, batchUUID).Return(&datatypes.Batch{
					UUID:   batchUUID,
					Status: enums.BatchStatusIngested,
				}, nil)
			},
			wantErr: "batch is not awaiting user review",
		},
		{
			name:    "Returns internal error (signal failure)",
			payload: &goaingest.ReviewBatchPayload{UUID: batchUUID.String(), Continue: true},
			mock: func(ctx context.Context, psvc *persistence_fake.MockService, tc *temporalsdk_mocks.Client) {
				psvc.EXPECT().ReadBatch(ctx, batchUUID).Return(&datatypes.Batch{
					UUID:   batchUUID,
					Status: enums.BatchStatusPending,
				}, nil)
				tc.On(
					"SignalWorkflow",
					ctx,
					ingest.BatchWorkflowID(batchUUID),
					"",
					ingest.BatchDecisionSignalName,
					ingest.BatchDecisionSignal{Continue: true},
				).Return(errors.New("temporal error"))
			},
			wantErr: "internal error",
		},
		{
			name:    "Signals batch decision",
			payload: &goaingest.ReviewBatchPayload{UUID: batchUUID.String(), Continue: false},
			mock: func(ctx context.Context, psvc *persistence_fake.MockService, tc *temporalsdk_mocks.Client) {
				psvc.EXPECT().ReadBatch(ctx, batchUUID).Return(&datatypes.Batch{
					UUID:   batchUUID,
					Status: enums.BatchStatusPending,
				}, nil)
				tc.On(
					"SignalWorkflow",
					ctx,
					ingest.BatchWorkflowID(batchUUID),
					"",
					ingest.BatchDecisionSignalName,
					ingest.BatchDecisionSignal{Continue: false},
				).Return(nil)
			},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			svc, psvc, tc := testSvc(t, nil, 0)
			ctx := t.Context()
			if tt.mock != nil {
				tt.mock(ctx, psvc, tc)
			}

			err := svc.ReviewBatch(ctx, tt.payload)
			if tt.wantErr != "" {
				assert.Error(t, err, tt.wantErr)
				return
			}

			assert.NilError(t, err)
		})
	}
}

func TestListBatches(t *testing.T) {
	t.Parallel()

	batchUUID1 := uuid.MustParse("e2ace0da-8697-453d-9ea1-4c9b62309e54")
	batchUUID2 := uuid.MustParse("ffdb12f4-1735-4022-b746-a9bf4a32109b")
	batchUUID3 := uuid.MustParse("52fdfc07-2182-454f-963f-5f0f9a621d72")
	uploaderUUID := uuid.MustParse("0b075937-458c-43d9-b46c-222a072d62a9")

	testBatches := []*datatypes.Batch{
		{
			ID:          1,
			UUID:        batchUUID1,
			Identifier:  "Batch 1",
			SIPSCount:   2,
			Status:      enums.BatchStatusIngested,
			CreatedAt:   time.Date(2024, 9, 25, 9, 31, 10, 0, time.UTC),
			StartedAt:   time.Date(2024, 9, 25, 9, 31, 11, 0, time.UTC),
			CompletedAt: time.Date(2024, 9, 25, 9, 31, 12, 0, time.UTC),
			Uploader: &datatypes.User{
				UUID:  uploaderUUID,
				Email: "uploader@example.com",
				Name:  "Test Uploader",
			},
		},
		{
			ID:         2,
			UUID:       batchUUID2,
			Identifier: "Batch 2",
			SIPSCount:  3,
			Status:     enums.BatchStatusProcessing,
			CreatedAt:  time.Date(2024, 10, 1, 17, 13, 26, 0, time.UTC),
			StartedAt:  time.Date(2024, 10, 1, 17, 13, 27, 0, time.UTC),
		},
		{
			ID:         3,
			UUID:       batchUUID3,
			Identifier: "Batch 3",
			SIPSCount:  1,
			Status:     enums.BatchStatusFailed,
			CreatedAt:  time.Date(2024, 10, 1, 17, 13, 26, 0, time.UTC),
		},
	}

	for _, tt := range []struct {
		name         string
		payload      *goaingest.ListBatchesPayload
		mockRecorder func(*persistence_fake.MockServiceMockRecorder)
		want         *goaingest.Batches
		wantErr      string
	}{
		{
			name: "Returns all batches",
			mockRecorder: func(mr *persistence_fake.MockServiceMockRecorder) {
				mr.ListBatches(
					mockutil.Context(),
					&persistence.BatchFilter{
						Sort: entfilter.NewSort().AddCol("id", true),
					},
				).Return(
					testBatches,
					&persistence.Page{Limit: 20, Total: 3},
					nil,
				)
			},
			want: &goaingest.Batches{
				Items: goaingest.BatchCollection{
					{
						UUID:          batchUUID1,
						Identifier:    "Batch 1",
						SipsCount:     2,
						Status:        enums.BatchStatusIngested.String(),
						CreatedAt:     "2024-09-25T09:31:10Z",
						StartedAt:     ref.New("2024-09-25T09:31:11Z"),
						CompletedAt:   ref.New("2024-09-25T09:31:12Z"),
						UploaderUUID:  ref.New(uploaderUUID),
						UploaderEmail: ref.New("uploader@example.com"),
						UploaderName:  ref.New("Test Uploader"),
					},
					{
						UUID:       batchUUID2,
						Identifier: "Batch 2",
						SipsCount:  3,
						Status:     enums.BatchStatusProcessing.String(),
						CreatedAt:  "2024-10-01T17:13:26Z",
						StartedAt:  ref.New("2024-10-01T17:13:27Z"),
					},
					{
						UUID:       batchUUID3,
						Identifier: "Batch 3",
						SipsCount:  1,
						Status:     enums.BatchStatusFailed.String(),
						CreatedAt:  "2024-10-01T17:13:26Z",
					},
				},
				Page: &goaingest.EnduroPage{
					Limit: 20,
					Total: 3,
				},
			},
		},
		{
			name: "Returns filtered batches",
			payload: &goaingest.ListBatchesPayload{
				Identifier:          ref.New("Batch 1"),
				EarliestCreatedTime: ref.New("2024-09-25T09:30:00Z"),
				LatestCreatedTime:   ref.New("2024-09-25T09:40:00Z"),
				Status:              ref.New(enums.BatchStatusIngested.String()),
				UploaderUUID:        ref.New("0b075937-458c-43d9-b46c-222a072d62a9"),
				Limit:               ref.New(10),
				Offset:              ref.New(1),
			},
			mockRecorder: func(mr *persistence_fake.MockServiceMockRecorder) {
				mr.ListBatches(
					mockutil.Context(),
					&persistence.BatchFilter{
						Identifier: ref.New("Batch 1"),
						Status:     ref.New(enums.BatchStatusIngested),
						CreatedAt: &timerange.Range{
							Start: time.Date(2024, 9, 25, 9, 30, 0, 0, time.UTC),
							End:   time.Date(2024, 9, 25, 9, 40, 0, 0, time.UTC),
						},
						UploaderID: ref.New(uuid.MustParse("0b075937-458c-43d9-b46c-222a072d62a9")),
						Sort:       entfilter.NewSort().AddCol("id", true),
						Page: persistence.Page{
							Limit:  10,
							Offset: 1,
						},
					},
				).Return(
					testBatches[0:1],
					&persistence.Page{Limit: 10, Total: 1},
					nil,
				)
			},
			want: &goaingest.Batches{
				Items: goaingest.BatchCollection{
					{
						UUID:          batchUUID1,
						Identifier:    "Batch 1",
						SipsCount:     2,
						Status:        enums.BatchStatusIngested.String(),
						CreatedAt:     "2024-09-25T09:31:10Z",
						StartedAt:     ref.New("2024-09-25T09:31:11Z"),
						CompletedAt:   ref.New("2024-09-25T09:31:12Z"),
						UploaderUUID:  ref.New(uploaderUUID),
						UploaderEmail: ref.New("uploader@example.com"),
						UploaderName:  ref.New("Test Uploader"),
					},
				},
				Page: &goaingest.EnduroPage{
					Limit: 10,
					Total: 1,
				},
			},
		},
		{
			name: "Errors on an internal service error",
			payload: &goaingest.ListBatchesPayload{
				Identifier: ref.New("Batch 42"),
			},
			mockRecorder: func(mr *persistence_fake.MockServiceMockRecorder) {
				mr.ListBatches(
					mockutil.Context(),
					&persistence.BatchFilter{
						Identifier: ref.New("Batch 42"),
						Sort:       entfilter.NewSort().AddCol("id", true),
					},
				).Return(nil, nil, persistence.ErrInternal)
			},
			wantErr: "internal error",
		},
		{
			name: "Errors on a bad uploader_uuid",
			payload: &goaingest.ListBatchesPayload{
				UploaderUUID: ref.New("invalid"),
			},
			wantErr: "uploader_uuid: invalid UUID",
		},
		{
			name: "Errors on a bad status",
			payload: &goaingest.ListBatchesPayload{
				Status: ref.New("meditating"),
			},
			wantErr: "status: invalid value",
		},
		{
			name: "Errors on a bad earliest_created_time",
			payload: &goaingest.ListBatchesPayload{
				EarliestCreatedTime: ref.New("2024-15-15T25:83:52Z"),
			},
			wantErr: "created at: time range: cannot parse start time",
		},
		{
			name: "Errors on a bad latest_created_time",
			payload: &goaingest.ListBatchesPayload{
				LatestCreatedTime: ref.New("2024-15-15T25:83:52Z"),
			},
			wantErr: "created at: time range: cannot parse end time",
		},
		{
			name: "Errors on a bad created at range",
			payload: &goaingest.ListBatchesPayload{
				EarliestCreatedTime: ref.New("2024-10-01T17:43:52Z"),
				LatestCreatedTime:   ref.New("2023-10-01T17:43:52Z"),
			},
			wantErr: "created at: time range: end cannot be before start",
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			svc, psvc, _ := testSvc(t, nil, 0)

			if tt.mockRecorder != nil {
				tt.mockRecorder(psvc.EXPECT())
			}

			got, err := svc.ListBatches(t.Context(), tt.payload)
			if tt.wantErr != "" {
				assert.Error(t, err, tt.wantErr)
				return
			}

			assert.NilError(t, err)
			assert.DeepEqual(t, got, tt.want)
		})
	}
}
