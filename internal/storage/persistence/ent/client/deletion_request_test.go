package client_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/google/uuid"
	"go.artefactual.dev/tools/ref"
	"gotest.tools/v3/assert"

	"github.com/artefactual-sdps/enduro/internal/storage"
	"github.com/artefactual-sdps/enduro/internal/storage/enums"
	"github.com/artefactual-sdps/enduro/internal/storage/persistence"
	"github.com/artefactual-sdps/enduro/internal/storage/persistence/ent/db"
	"github.com/artefactual-sdps/enduro/internal/storage/types"
)

func initialDataForDeletionRequestTests(t *testing.T, ctx context.Context, entc *db.Client) {
	t.Helper()

	aip := entc.AIP.Create().
		SetName("AIP 1").
		SetAipID(aipID).
		SetObjectKey(objectKey).
		SetStatus(enums.AIPStatusStored).
		SaveX(ctx)

	entc.Workflow.Create().
		SetUUID(wUUID).
		SetTemporalID("temporal-id").
		SetType(enums.WorkflowTypeDeleteAip).
		SetStatus(enums.WorkflowStatusCanceled).
		SetAipID(aip.ID).
		SaveX(ctx)

	entc.Workflow.Create().
		SetUUID(uuid.New()).
		SetTemporalID("temporal-id").
		SetType(enums.WorkflowTypeDeleteAip).
		SetStatus(enums.WorkflowStatusDone).
		SetAipID(aip.ID).
		SaveX(ctx)
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

			ctx := t.Context()
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

func TestListDeletionRequests(t *testing.T) {
	t.Parallel()

	drUUID1 := uuid.New()
	drUUID2 := uuid.New()
	requestedAt1 := time.Date(2025, 10, 29, 10, 10, 10, 0, time.UTC)
	requestedAt2 := time.Date(2025, 10, 29, 11, 11, 11, 0, time.UTC)

	type test struct {
		name   string
		filter *persistence.DeletionRequestFilter
		want   []*types.DeletionRequest
	}

	for _, tc := range []test{
		{
			name: "Lists all Deletion Requests when filter is nil",
			want: []*types.DeletionRequest{
				{
					DBID:         1,
					UUID:         drUUID1,
					Reason:       "Reason 1",
					Requester:    "requester@example.com",
					RequesterIss: "issuer",
					RequesterSub: "sub",
					RequestedAt:  requestedAt1,
					Status:       enums.DeletionRequestStatusPending,
					AIPUUID:      aipID,
					WorkflowDBID: 1,
				},
				{
					DBID:         2,
					UUID:         drUUID2,
					Reason:       "Reason 2",
					Requester:    "requester@example.com",
					RequesterIss: "issuer",
					RequesterSub: "sub",
					RequestedAt:  requestedAt2,
					Status:       enums.DeletionRequestStatusApproved,
					AIPUUID:      aipID,
					WorkflowDBID: 2,
					ReportKey: storage.ReportPrefix +
						"aip_deletion_report_123e4567-e89b-12d3-a456-426614174000.pdf",
				},
			},
		},
		{
			name: "Lists Deletion Requests by Status",
			filter: &persistence.DeletionRequestFilter{
				Status: ref.New(enums.DeletionRequestStatusPending),
			},
			want: []*types.DeletionRequest{
				{
					DBID:         1,
					UUID:         drUUID1,
					Reason:       "Reason 1",
					Requester:    "requester@example.com",
					RequesterIss: "issuer",
					RequesterSub: "sub",
					RequestedAt:  requestedAt1,
					Status:       enums.DeletionRequestStatusPending,
					AIPUUID:      aipID,
					WorkflowDBID: 1,
				},
			},
		},
		{
			name: "Lists Deletion Requests by AIP UUID",
			filter: &persistence.DeletionRequestFilter{
				AIPUUID: ref.New(aipID),
			},
			want: []*types.DeletionRequest{
				{
					DBID:         1,
					UUID:         drUUID1,
					Reason:       "Reason 1",
					Requester:    "requester@example.com",
					RequesterIss: "issuer",
					RequesterSub: "sub",
					RequestedAt:  requestedAt1,
					Status:       enums.DeletionRequestStatusPending,
					AIPUUID:      aipID,
					WorkflowDBID: 1,
				},
				{
					DBID:         2,
					UUID:         drUUID2,
					Reason:       "Reason 2",
					Requester:    "requester@example.com",
					RequesterIss: "issuer",
					RequesterSub: "sub",
					RequestedAt:  requestedAt2,
					Status:       enums.DeletionRequestStatusApproved,
					AIPUUID:      aipID,
					WorkflowDBID: 2,
					ReportKey: storage.ReportPrefix +
						"aip_deletion_report_123e4567-e89b-12d3-a456-426614174000.pdf",
				},
			},
		},
		{
			name: "Returns no results for non-matching filter",
			filter: &persistence.DeletionRequestFilter{
				AIPUUID: ref.New(uuid.New()),
			},
			want: []*types.DeletionRequest{},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctx := t.Context()
			entc, c := setUpClient(t)
			initialDataForDeletionRequestTests(t, ctx, entc)

			entc.DeletionRequest.Create().
				SetUUID(drUUID1).
				SetRequester("requester@example.com").
				SetRequesterIss("issuer").
				SetRequesterSub("sub").
				SetRequestedAt(requestedAt1).
				SetReason("Reason 1").
				SetStatus(enums.DeletionRequestStatusPending).
				SetAipID(1).
				SetWorkflowID(1).
				SaveX(ctx)

			entc.DeletionRequest.Create().
				SetUUID(drUUID2).
				SetRequester("requester@example.com").
				SetRequesterIss("issuer").
				SetRequesterSub("sub").
				SetRequestedAt(requestedAt2).
				SetReason("Reason 2").
				SetStatus(enums.DeletionRequestStatusApproved).
				SetAipID(1).
				SetWorkflowID(2).
				SetReportKey(storage.ReportPrefix +
					"aip_deletion_report_123e4567-e89b-12d3-a456-426614174000.pdf").
				SaveX(ctx)

			r, err := c.ListDeletionRequests(ctx, tc.filter)

			assert.NilError(t, err)
			assert.DeepEqual(t, r, tc.want)
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
				dr.ReportKey = storage.ReportPrefix +
					"aip_deletion_report_123e4567-e89b-12d3-a456-426614174000.pdf"
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
				ReportKey: storage.ReportPrefix +
					"aip_deletion_report_123e4567-e89b-12d3-a456-426614174000.pdf",
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

			ctx := t.Context()
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

func TestReadDeletionRequests(t *testing.T) {
	t.Parallel()

	drUUID := uuid.MustParse("123e4567-e89b-12d3-a456-426614174000")
	requestedAt := time.Date(2025, 10, 29, 10, 10, 10, 0, time.UTC)
	reviewedAt := time.Date(2025, 10, 30, 11, 11, 11, 0, time.UTC)

	type test struct {
		name    string
		id      uuid.UUID
		want    *types.DeletionRequest
		wantErr string
	}

	for _, tc := range []test{
		{
			name: "Reads a deletion request",
			id:   drUUID,
			want: &types.DeletionRequest{
				DBID:         1,
				UUID:         drUUID,
				Requester:    "requester@example.com",
				RequesterIss: "issuer",
				RequesterSub: "sub",
				Reviewer:     "reviewer@example.com",
				ReviewerIss:  "issuer",
				ReviewerSub:  "sub",
				Reason:       "Test reason",
				Status:       enums.DeletionRequestStatusApproved,
				RequestedAt:  requestedAt,
				ReviewedAt:   reviewedAt,
				AIPUUID:      aipID,
				WorkflowDBID: 1,
				ReportKey: storage.ReportPrefix +
					"aip_deletion_report_123e4567-e89b-12d3-a456-426614174000.pdf",
			},
		},
		{
			name:    "Returns a deletion request not found error",
			id:      uuid.MustParse("223e4567-e89b-12d3-a456-426614174000"),
			wantErr: "read deletion request: db: deletion_request not found",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctx := t.Context()
			entc, c := setUpClient(t)
			initialDataForDeletionRequestTests(t, ctx, entc)

			entc.DeletionRequest.Create().
				SetUUID(drUUID).
				SetRequester("requester@example.com").
				SetRequesterIss("issuer").
				SetRequesterSub("sub").
				SetReviewer("reviewer@example.com").
				SetReviewerIss("issuer").
				SetReviewerSub("sub").
				SetReason("Test reason").
				SetStatus(enums.DeletionRequestStatusApproved).
				SetRequestedAt(requestedAt).
				SetReviewedAt(reviewedAt).
				SetAipID(1).
				SetWorkflowID(1).
				SetReportKey(
					storage.ReportPrefix + "aip_deletion_report_123e4567-e89b-12d3-a456-426614174000.pdf",
				).
				SaveX(ctx)

			r, err := c.ReadDeletionRequest(ctx, tc.id)
			if tc.wantErr != "" {
				assert.Error(t, err, tc.wantErr)
				return
			}

			assert.NilError(t, err)
			assert.DeepEqual(t, r, tc.want)
		})
	}
}
