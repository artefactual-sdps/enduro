package storage_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/mock"
	"go.artefactual.dev/tools/ref"
	temporalapi_enums "go.temporal.io/api/enums/v1"
	temporalsdk_client "go.temporal.io/sdk/client"
	temporalsdk_mocks "go.temporal.io/sdk/mocks"
	"gotest.tools/v3/assert"

	"github.com/artefactual-sdps/enduro/internal/api/auth"
	goastorage "github.com/artefactual-sdps/enduro/internal/api/gen/storage"
	"github.com/artefactual-sdps/enduro/internal/storage"
	"github.com/artefactual-sdps/enduro/internal/storage/enums"
	"github.com/artefactual-sdps/enduro/internal/storage/persistence/fake"
	"github.com/artefactual-sdps/enduro/internal/storage/types"
)

func TestRequestAipDeletion(t *testing.T) {
	t.Parallel()

	type test struct {
		name    string
		claims  *auth.Claims
		payload *goastorage.RequestAipDeletionPayload
		mock    func(context.Context, *fake.MockStorage, *temporalsdk_mocks.Client)
		wantErr string
	}

	for _, tt := range []test{
		{
			name:    "Fails to request AIP deletion (not authenticated)",
			wantErr: "authentication is required",
		},
		{
			name:    "Fails to request AIP deletion (missing email claim)",
			claims:  &auth.Claims{},
			wantErr: "email claim is required",
		},
		{
			name: "Fails to request AIP deletion (missing sub claim)",
			claims: &auth.Claims{
				Email: "requester@example.com",
			},
			wantErr: "sub claim is required",
		},
		{
			name: "Fails to request AIP deletion (missing iss claim)",
			claims: &auth.Claims{
				Email: "requester@example.com",
				Sub:   "subject",
			},
			wantErr: "iss claim is required",
		},
		{
			name: "Fails to request AIP deletion (invalid UUID)",
			claims: &auth.Claims{
				Email: "requester@example.com",
				ISS:   "issuer",
				Sub:   "subject",
			},
			payload: &goastorage.RequestAipDeletionPayload{
				UUID: "invalid-uuid",
			},
			wantErr: "invalid UUID",
		},
		{
			name: "Fails to request AIP deletion (invalid reason)",
			claims: &auth.Claims{
				Email: "requester@example.com",
				ISS:   "issuer",
				Sub:   "subject",
			},
			payload: &goastorage.RequestAipDeletionPayload{
				UUID: aipID.String(),
			},
			wantErr: "invalid reason",
		},
		{
			name: "Fails to request AIP deletion (AIP not found)",
			claims: &auth.Claims{
				Email: "requester@example.com",
				ISS:   "issuer",
				Sub:   "subject",
			},
			payload: &goastorage.RequestAipDeletionPayload{
				UUID:   aipID.String(),
				Reason: "Reason",
			},
			mock: func(ctx context.Context, s *fake.MockStorage, tc *temporalsdk_mocks.Client) {
				s.EXPECT().ReadAIP(ctx, aipID).Return(nil, &goastorage.AIPNotFound{})
			},
			wantErr: "AIP not found.",
		},
		{
			name: "Fails to request AIP deletion (AIP not stored)",
			claims: &auth.Claims{
				Email: "requester@example.com",
				ISS:   "issuer",
				Sub:   "subject",
			},
			payload: &goastorage.RequestAipDeletionPayload{
				UUID:   aipID.String(),
				Reason: "Reason",
			},
			mock: func(ctx context.Context, s *fake.MockStorage, tc *temporalsdk_mocks.Client) {
				s.EXPECT().ReadAIP(ctx, aipID).Return(&goastorage.AIP{Status: enums.AIPStatusPending.String()}, nil)
			},
			wantErr: "AIP is not stored",
		},
		{
			name: "Fails to request AIP deletion (init workflow failure)",
			claims: &auth.Claims{
				Email: "requester@example.com",
				ISS:   "issuer",
				Sub:   "subject",
			},
			payload: &goastorage.RequestAipDeletionPayload{
				UUID:   aipID.String(),
				Reason: "Reason",
			},
			mock: func(ctx context.Context, s *fake.MockStorage, tc *temporalsdk_mocks.Client) {
				s.EXPECT().ReadAIP(ctx, aipID).Return(&goastorage.AIP{Status: enums.AIPStatusStored.String()}, nil)
				tc.On(
					"ExecuteWorkflow",
					mock.AnythingOfType("*context.timerCtx"),
					temporalsdk_client.StartWorkflowOptions{
						ID:                    fmt.Sprintf("%s-%s", storage.StorageDeleteWorkflowName, aipID),
						TaskQueue:             "global",
						WorkflowIDReusePolicy: temporalapi_enums.WORKFLOW_ID_REUSE_POLICY_ALLOW_DUPLICATE_FAILED_ONLY,
					},
					storage.StorageDeleteWorkflowName,
					&storage.StorageDeleteWorkflowRequest{
						AIPID:     aipID,
						Reason:    "Reason",
						UserEmail: "requester@example.com",
						UserISS:   "issuer",
						UserSub:   "subject",
						TaskQueue: "global",
					},
				).Return(nil, errors.New("temporal error"))
			},
			wantErr: "cannot perform operation",
		},
		{
			name: "Requests AIP deletion",
			claims: &auth.Claims{
				Email: "requester@example.com",
				ISS:   "issuer",
				Sub:   "subject",
			},
			payload: &goastorage.RequestAipDeletionPayload{
				UUID:   aipID.String(),
				Reason: "Reason",
			},
			mock: func(ctx context.Context, s *fake.MockStorage, tc *temporalsdk_mocks.Client) {
				s.EXPECT().ReadAIP(ctx, aipID).Return(&goastorage.AIP{Status: enums.AIPStatusStored.String()}, nil)
				tc.On(
					"ExecuteWorkflow",
					mock.AnythingOfType("*context.timerCtx"),
					temporalsdk_client.StartWorkflowOptions{
						ID:                    fmt.Sprintf("%s-%s", storage.StorageDeleteWorkflowName, aipID),
						TaskQueue:             "global",
						WorkflowIDReusePolicy: temporalapi_enums.WORKFLOW_ID_REUSE_POLICY_ALLOW_DUPLICATE_FAILED_ONLY,
					},
					storage.StorageDeleteWorkflowName,
					&storage.StorageDeleteWorkflowRequest{
						AIPID:     aipID,
						Reason:    "Reason",
						UserEmail: "requester@example.com",
						UserISS:   "issuer",
						UserSub:   "subject",
						TaskQueue: "global",
					},
				).Return(nil, nil)
			},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctx := auth.WithUserClaims(context.Background(), tt.claims)
			attrs := &setUpAttrs{}
			svc := setUpService(t, attrs)

			if tt.mock != nil {
				tt.mock(ctx, attrs.persistenceMock, attrs.temporalClientMock)
			}

			err := svc.RequestAipDeletion(ctx, tt.payload)
			if tt.wantErr != "" {
				assert.Error(t, err, tt.wantErr)
				return
			}
			assert.NilError(t, err)
		})
	}
}

func TestReviewAipDeletion(t *testing.T) {
	t.Parallel()

	type test struct {
		name    string
		claims  *auth.Claims
		payload *goastorage.ReviewAipDeletionPayload
		mock    func(context.Context, *fake.MockStorage, *temporalsdk_mocks.Client)
		wantErr string
	}

	for _, tt := range []test{
		{
			name:    "Fails to review AIP deletion (not authenticated)",
			wantErr: "authentication is required",
		},
		{
			name:    "Fails to review AIP deletion (missing email claim)",
			claims:  &auth.Claims{},
			wantErr: "email claim is required",
		},
		{
			name: "Fails to review AIP deletion (missing sub claim)",
			claims: &auth.Claims{
				Email: "reviewer@example.com",
			},
			wantErr: "sub claim is required",
		},
		{
			name: "Fails to review AIP deletion (missing iss claim)",
			claims: &auth.Claims{
				Email: "reviewer@example.com",
				Sub:   "subject-2",
			},
			wantErr: "iss claim is required",
		},
		{
			name: "Fails to review AIP deletion (invalid UUID)",
			claims: &auth.Claims{
				Email: "reviewer@example.com",
				ISS:   "issuer",
				Sub:   "subject-2",
			},
			payload: &goastorage.ReviewAipDeletionPayload{
				UUID: "invalid-uuid",
			},
			wantErr: "invalid UUID",
		},
		{
			name: "Fails to review AIP deletion (AIP not found)",
			claims: &auth.Claims{
				Email: "reviewer@example.com",
				ISS:   "issuer",
				Sub:   "subject-2",
			},
			payload: &goastorage.ReviewAipDeletionPayload{
				UUID:     aipID.String(),
				Approved: false,
			},
			mock: func(ctx context.Context, s *fake.MockStorage, tc *temporalsdk_mocks.Client) {
				s.EXPECT().ReadAIP(ctx, aipID).Return(nil, &goastorage.AIPNotFound{})
			},
			wantErr: "AIP not found.",
		},
		{
			name: "Fails to review AIP deletion (AIP not pending)",
			claims: &auth.Claims{
				Email: "reviewer@example.com",
				ISS:   "issuer",
				Sub:   "subject-2",
			},
			payload: &goastorage.ReviewAipDeletionPayload{
				UUID:     aipID.String(),
				Approved: false,
			},
			mock: func(ctx context.Context, s *fake.MockStorage, tc *temporalsdk_mocks.Client) {
				s.EXPECT().ReadAIP(ctx, aipID).Return(&goastorage.AIP{Status: enums.AIPStatusStored.String()}, nil)
			},
			wantErr: "AIP is not awaiting user review",
		},
		{
			name: "Fails to review AIP deletion (deletion request read error)",
			claims: &auth.Claims{
				Email: "reviewer@example.com",
				ISS:   "issuer",
				Sub:   "subject-2",
			},
			payload: &goastorage.ReviewAipDeletionPayload{
				UUID:     aipID.String(),
				Approved: false,
			},
			mock: func(ctx context.Context, s *fake.MockStorage, tc *temporalsdk_mocks.Client) {
				s.EXPECT().ReadAIP(ctx, aipID).Return(&goastorage.AIP{Status: enums.AIPStatusPending.String()}, nil)
				s.EXPECT().ReadAipPendingDeletionRequest(ctx, aipID).Return(nil, errors.New("persistence error"))
			},
			wantErr: "cannot perform operation",
		},
		{
			name: "Fails to review AIP deletion (reviewer matches requester)",
			claims: &auth.Claims{
				Email: "requester@example.com",
				ISS:   "issuer",
				Sub:   "subject",
			},
			payload: &goastorage.ReviewAipDeletionPayload{
				UUID:     aipID.String(),
				Approved: false,
			},
			mock: func(ctx context.Context, s *fake.MockStorage, tc *temporalsdk_mocks.Client) {
				s.EXPECT().ReadAIP(ctx, aipID).Return(&goastorage.AIP{Status: enums.AIPStatusPending.String()}, nil)
				s.EXPECT().ReadAipPendingDeletionRequest(ctx, aipID).Return(&types.DeletionRequest{
					RequesterISS: "issuer",
					RequesterSub: "subject",
				}, nil)
			},
			wantErr: "requester cannot review their own request",
		},
		{
			name: "Fails to review AIP deletion (signal workflow failure)",
			claims: &auth.Claims{
				Email: "reviewer@example.com",
				ISS:   "issuer",
				Sub:   "subject-2",
			},
			payload: &goastorage.ReviewAipDeletionPayload{
				UUID:     aipID.String(),
				Approved: false,
			},
			mock: func(ctx context.Context, s *fake.MockStorage, tc *temporalsdk_mocks.Client) {
				s.EXPECT().ReadAIP(ctx, aipID).Return(&goastorage.AIP{Status: enums.AIPStatusPending.String()}, nil)
				s.EXPECT().ReadAipPendingDeletionRequest(ctx, aipID).Return(&types.DeletionRequest{
					RequesterISS: "issuer",
					RequesterSub: "subject",
				}, nil)
				tc.On(
					"SignalWorkflow",
					mock.AnythingOfType("*context.valueCtx"),
					fmt.Sprintf("%s-%s", storage.StorageDeleteWorkflowName, aipID),
					"",
					storage.DeletionDecisionSignalName,
					storage.DeletionDecisionSignal{
						Status:    enums.DeletionRequestStatusRejected,
						UserEmail: "reviewer@example.com",
						UserISS:   "issuer",
						UserSub:   "subject-2",
					},
				).Return(errors.New("temporal error"))
			},
			wantErr: "cannot perform operation",
		},
		{
			name: "Reviews AIP deletion",
			claims: &auth.Claims{
				Email: "reviewer@example.com",
				ISS:   "issuer",
				Sub:   "subject-2",
			},
			payload: &goastorage.ReviewAipDeletionPayload{
				UUID:     aipID.String(),
				Approved: true,
			},
			mock: func(ctx context.Context, s *fake.MockStorage, tc *temporalsdk_mocks.Client) {
				s.EXPECT().ReadAIP(ctx, aipID).Return(&goastorage.AIP{Status: enums.AIPStatusPending.String()}, nil)
				s.EXPECT().ReadAipPendingDeletionRequest(ctx, aipID).Return(&types.DeletionRequest{
					RequesterISS: "issuer",
					RequesterSub: "subject",
				}, nil)
				tc.On(
					"SignalWorkflow",
					mock.AnythingOfType("*context.valueCtx"),
					fmt.Sprintf("%s-%s", storage.StorageDeleteWorkflowName, aipID),
					"",
					storage.DeletionDecisionSignalName,
					storage.DeletionDecisionSignal{
						Status:    enums.DeletionRequestStatusApproved,
						UserEmail: "reviewer@example.com",
						UserISS:   "issuer",
						UserSub:   "subject-2",
					},
				).Return(nil)
			},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctx := auth.WithUserClaims(context.Background(), tt.claims)
			attrs := &setUpAttrs{}
			svc := setUpService(t, attrs)

			if tt.mock != nil {
				tt.mock(ctx, attrs.persistenceMock, attrs.temporalClientMock)
			}

			err := svc.ReviewAipDeletion(ctx, tt.payload)
			if tt.wantErr != "" {
				assert.Error(t, err, tt.wantErr)
				return
			}
			assert.NilError(t, err)
		})
	}
}

func TestCancelAipDeletion(t *testing.T) {
	t.Parallel()

	type test struct {
		name    string
		claims  *auth.Claims
		payload *goastorage.CancelAipDeletionPayload
		mock    func(context.Context, *fake.MockStorage, *temporalsdk_mocks.Client)
		wantErr string
	}

	for _, tt := range []test{
		{
			name: "Fails when user is not authenticated",
			payload: &goastorage.CancelAipDeletionPayload{
				UUID: aipID.String(),
			},
			wantErr: "authentication is required",
		},
		{
			name: "Fails on invalid AIP UUID",
			claims: &auth.Claims{
				Email: "reviewer@example.com",
				ISS:   "issuer",
				Sub:   "subject-2",
			},
			payload: &goastorage.CancelAipDeletionPayload{
				UUID: "invalid-uuid",
			},
			wantErr: "invalid UUID",
		},
		{
			name: "Fails when deletion request is not found",
			claims: &auth.Claims{
				Email: "reviewer@example.com",
				ISS:   "issuer",
				Sub:   "subject-2",
			},
			payload: &goastorage.CancelAipDeletionPayload{
				UUID: aipID.String(),
			},
			mock: func(ctx context.Context, s *fake.MockStorage, tc *temporalsdk_mocks.Client) {
				s.EXPECT().ReadAipPendingDeletionRequest(ctx, aipID).Return(nil, errors.New("db: deletion_request not found"))
			},
			wantErr: "db: deletion_request not found",
		},
		{
			name: "Fails on deletion request read error",
			claims: &auth.Claims{
				Email: "reviewer@example.com",
				ISS:   "issuer",
				Sub:   "subject-2",
			},
			payload: &goastorage.CancelAipDeletionPayload{
				UUID: aipID.String(),
			},
			mock: func(ctx context.Context, s *fake.MockStorage, tc *temporalsdk_mocks.Client) {
				s.EXPECT().ReadAipPendingDeletionRequest(ctx, aipID).Return(nil, errors.New("db: persistence error"))
			},
			wantErr: "db: persistence error",
		},
		{
			name: "Fails if auth user is not the requester",
			claims: &auth.Claims{
				Email: "requester@example.com",
				ISS:   "issuer",
				Sub:   "subject",
			},
			payload: &goastorage.CancelAipDeletionPayload{
				UUID: aipID.String(),
			},
			mock: func(ctx context.Context, s *fake.MockStorage, tc *temporalsdk_mocks.Client) {
				s.EXPECT().ReadAipPendingDeletionRequest(ctx, aipID).Return(&types.DeletionRequest{
					RequesterISS: "issuer",
					RequesterSub: "subject-2",
				}, nil)
			},
			wantErr: "Forbidden",
		},
		{
			name: "Fails on signal workflow failure",
			claims: &auth.Claims{
				Email: "reviewer@example.com",
				ISS:   "issuer",
				Sub:   "subject",
			},
			payload: &goastorage.CancelAipDeletionPayload{
				UUID: aipID.String(),
			},
			mock: func(ctx context.Context, s *fake.MockStorage, tc *temporalsdk_mocks.Client) {
				s.EXPECT().ReadAipPendingDeletionRequest(ctx, aipID).Return(
					&types.DeletionRequest{
						RequesterISS: "issuer",
						RequesterSub: "subject",
					}, nil,
				)
				tc.On(
					"SignalWorkflow",
					mock.AnythingOfType("*context.valueCtx"),
					fmt.Sprintf("%s-%s", storage.StorageDeleteWorkflowName, aipID),
					"",
					storage.DeletionDecisionSignalName,
					storage.DeletionDecisionSignal{
						Status:    enums.DeletionRequestStatusCanceled,
						UserEmail: "reviewer@example.com",
						UserISS:   "issuer",
						UserSub:   "subject",
					},
				).Return(errors.New("temporal error"))
			},
			wantErr: "cannot perform operation",
		},
		{
			name: "Cancels AIP deletion request",
			claims: &auth.Claims{
				Email: "reviewer@example.com",
				ISS:   "issuer",
				Sub:   "subject",
			},
			payload: &goastorage.CancelAipDeletionPayload{
				UUID: aipID.String(),
			},
			mock: func(ctx context.Context, s *fake.MockStorage, tc *temporalsdk_mocks.Client) {
				s.EXPECT().ReadAipPendingDeletionRequest(ctx, aipID).Return(&types.DeletionRequest{
					RequesterISS: "issuer",
					RequesterSub: "subject",
				}, nil)
				tc.On(
					"SignalWorkflow",
					mock.AnythingOfType("*context.valueCtx"),
					fmt.Sprintf("%s-%s", storage.StorageDeleteWorkflowName, aipID),
					"",
					storage.DeletionDecisionSignalName,
					storage.DeletionDecisionSignal{
						Status:    enums.DeletionRequestStatusCanceled,
						UserEmail: "reviewer@example.com",
						UserISS:   "issuer",
						UserSub:   "subject",
					},
				).Return(nil)
			},
		},
		{
			name: "Doesn't cancel deletion request when test flag is set",
			claims: &auth.Claims{
				Email: "reviewer@example.com",
				ISS:   "issuer",
				Sub:   "subject",
			},
			payload: &goastorage.CancelAipDeletionPayload{
				UUID:  aipID.String(),
				Check: ref.New(true),
			},
			mock: func(ctx context.Context, s *fake.MockStorage, tc *temporalsdk_mocks.Client) {
				s.EXPECT().ReadAipPendingDeletionRequest(ctx, aipID).Return(&types.DeletionRequest{
					RequesterISS: "issuer",
					RequesterSub: "subject",
				}, nil)
			},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctx := auth.WithUserClaims(context.Background(), tt.claims)
			attrs := &setUpAttrs{}
			svc := setUpService(t, attrs)

			if tt.mock != nil {
				tt.mock(ctx, attrs.persistenceMock, attrs.temporalClientMock)
			}

			err := svc.CancelAipDeletion(ctx, tt.payload)
			if tt.wantErr != "" {
				assert.Error(t, err, tt.wantErr)
				return
			}
			assert.NilError(t, err)
		})
	}
}
