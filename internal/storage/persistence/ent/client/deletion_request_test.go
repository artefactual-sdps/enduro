package client_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/google/uuid"
	"gotest.tools/v3/assert"

	"github.com/artefactual-sdps/enduro/internal/storage/enums"
	"github.com/artefactual-sdps/enduro/internal/storage/persistence"
	"github.com/artefactual-sdps/enduro/internal/storage/persistence/ent/db"
	"github.com/artefactual-sdps/enduro/internal/storage/types"
)

func initialDataForDeletionRequestTests(t *testing.T, ctx context.Context, entc *db.Client) {
	t.Helper()

	initialDataForTaskTests(t, ctx, entc)
}

func TestCreateDeletionRequest(t *testing.T) {
	t.Parallel()

	drUUID := uuid.New()
	requestedAt := time.Now()

	type test struct {
		name    string
		dr      *types.DeletionRequest
		want    *db.DeletionRequest
		wantErr string
	}

	for _, tt := range []test{
		{
			name: "Creates a Deletion Request",
			dr: &types.DeletionRequest{
				UUID:         drUUID,
				AIPUUID:      aipID,
				Reason:       "Reason",
				Requester:    "requester@example.com",
				RequesterIss: "issuer",
				RequesterSub: "sub",
				RequestedAt:  requestedAt,
				WorkflowDBID: 1,
			},
			want: &db.DeletionRequest{
				ID:           1,
				UUID:         drUUID,
				AipID:        1,
				Reason:       "Reason",
				Requester:    "requester@example.com",
				RequesterIss: "issuer",
				RequesterSub: "sub",
				RequestedAt:  requestedAt,
				Status:       enums.DeletionRequestStatusPending,
				WorkflowID:   1,
			},
		},
		{
			name:    "Fails to create a Deletion Request without AIP UUID",
			dr:      &types.DeletionRequest{UUID: drUUID},
			wantErr: "create deletion request: db: aip not found",
		},
		{
			name: "Fails to create a Deletion Request without WorkflowDBID",
			dr: &types.DeletionRequest{
				UUID:    drUUID,
				AIPUUID: aipID,
			},
			wantErr: "create deletion request: db: validator failed for field \"DeletionRequest.workflow_id\": value out of range",
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			entc, c := setUpClient(t)
			initialDataForDeletionRequestTests(t, ctx, entc)

			err := c.CreateDeletionRequest(ctx, tt.dr)
			if tt.wantErr != "" {
				assert.Error(t, err, tt.wantErr)
				return
			}
			assert.NilError(t, err)

			dbdr := entc.DeletionRequest.GetX(ctx, tt.dr.DBID)
			assert.DeepEqual(
				t,
				dbdr,
				tt.want,
				cmpopts.IgnoreFields(db.DeletionRequest{}, "config", "Edges", "selectValues"),
			)
		})
	}
}

func TestUpdateDeletionRequest(t *testing.T) {
	t.Parallel()

	drUUID := uuid.New()
	requestedAt := time.Now()
	reviewedAt := time.Now()

	type test struct {
		name    string
		updater persistence.DeletionRequestUpdater
		dbID    int
		want    *types.DeletionRequest
		wantErr string
	}

	for _, tt := range []test{
		{
			name: "Updates a Deletion Request",
			updater: func(dr *types.DeletionRequest) (*types.DeletionRequest, error) {
				dr.Reviewer = "reviewer@example.com"
				dr.ReviewerIss = "issuer"
				dr.ReviewerSub = "sub2"
				dr.ReviewedAt = reviewedAt
				dr.Status = enums.DeletionRequestStatusApproved
				return dr, nil
			},
			want: &types.DeletionRequest{
				DBID:         1,
				UUID:         drUUID,
				Reason:       "Reason",
				Requester:    "requester@example.com",
				RequesterIss: "issuer",
				RequesterSub: "sub",
				RequestedAt:  requestedAt,
				Reviewer:     "reviewer@example.com",
				ReviewerIss:  "issuer",
				ReviewerSub:  "sub2",
				ReviewedAt:   reviewedAt,
				Status:       enums.DeletionRequestStatusApproved,
				AIPUUID:      aipID,
				WorkflowDBID: 1,
			},
		},
		{
			name: "Updates a Deletion Request (ignores immutable fields)",
			updater: func(dr *types.DeletionRequest) (*types.DeletionRequest, error) {
				dr.Reason = "Updated reason"
				dr.Requester = "updated-requester@example.com"
				dr.RequesterIss = "updated-issuer"
				dr.RequesterSub = "updated-sub"
				dr.RequestedAt = requestedAt.Add(time.Minute)
				return dr, nil
			},
			want: &types.DeletionRequest{
				DBID:         1,
				UUID:         drUUID,
				Reason:       "Reason",
				Requester:    "requester@example.com",
				RequesterIss: "issuer",
				RequesterSub: "sub",
				RequestedAt:  requestedAt,
				Status:       enums.DeletionRequestStatusPending,
				AIPUUID:      aipID,
				WorkflowDBID: 1,
			},
		},
		{
			name:    "Fails to update a Deletion Request (not found)",
			updater: nil,
			dbID:    1234,
			wantErr: "update deletion request: db: deletion_request not found",
		},
		{
			name: "Fails to update a Deletion Request (updater error)",
			updater: func(dr *types.DeletionRequest) (*types.DeletionRequest, error) {
				return nil, errors.New("updater error")
			},
			wantErr: "update deletion request: updater error",
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			entc, c := setUpClient(t)
			initialDataForDeletionRequestTests(t, ctx, entc)

			if tt.dbID == 0 {
				dbdr := entc.DeletionRequest.Create().
					SetUUID(drUUID).
					SetRequester("requester@example.com").
					SetRequesterIss("issuer").
					SetRequesterSub("sub").
					SetRequestedAt(requestedAt).
					SetReason("Reason").
					SetAipID(1).
					SetWorkflowID(1).
					SaveX(ctx)

				tt.dbID = dbdr.ID
			}

			dr, err := c.UpdateDeletionRequest(ctx, tt.dbID, tt.updater)
			if tt.wantErr != "" {
				assert.Error(t, err, tt.wantErr)
				return
			}

			assert.NilError(t, err)
			assert.DeepEqual(t, dr, tt.want)
		})
	}
}

func TestReadAipPendingDeletionRequest(t *testing.T) {
	t.Parallel()

	drUUID := uuid.New()
	requestedAt := time.Now()

	type test struct {
		name    string
		aipID   uuid.UUID
		setup   func(ctx context.Context, entc *db.Client)
		want    *types.DeletionRequest
		wantErr string
	}

	for _, tt := range []test{
		{
			name:  "Finds pending deletion request",
			aipID: aipID,
			setup: func(ctx context.Context, entc *db.Client) {
				entc.DeletionRequest.Create().
					SetUUID(drUUID).
					SetRequester("requester@example.com").
					SetRequesterIss("issuer").
					SetRequesterSub("sub").
					SetRequestedAt(requestedAt).
					SetReason("Reason").
					SetAipID(1).
					SetWorkflowID(1).
					SaveX(ctx)
			},
			want: &types.DeletionRequest{
				DBID:         1,
				UUID:         drUUID,
				Reason:       "Reason",
				Requester:    "requester@example.com",
				RequesterIss: "issuer",
				RequesterSub: "sub",
				RequestedAt:  requestedAt,
				Status:       enums.DeletionRequestStatusPending,
				WorkflowDBID: 1,
			},
		},
		{
			name:    "No pending deletion request found",
			aipID:   aipID,
			wantErr: "read AIP pending deletion request: db: deletion_request not found",
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			entc, c := setUpClient(t)
			initialDataForDeletionRequestTests(t, ctx, entc)

			if tt.setup != nil {
				tt.setup(ctx, entc)
			}

			dr, err := c.ReadAipPendingDeletionRequest(ctx, tt.aipID)
			if tt.wantErr != "" {
				assert.Error(t, err, tt.wantErr)
				return
			}

			assert.NilError(t, err)
			assert.DeepEqual(t, dr, tt.want)
		})
	}
}
