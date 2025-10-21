package client

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/artefactual-sdps/enduro/internal/storage/persistence"
	"github.com/artefactual-sdps/enduro/internal/storage/persistence/ent/db"
	"github.com/artefactual-sdps/enduro/internal/storage/persistence/ent/db/aip"
	"github.com/artefactual-sdps/enduro/internal/storage/persistence/ent/db/deletionrequest"
	"github.com/artefactual-sdps/enduro/internal/storage/types"
)

func (c *Client) CreateDeletionRequest(ctx context.Context, dr *types.DeletionRequest) error {
	tx, err := c.c.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("create deletion request: %v", err)
	}

	aipDBID, err := tx.AIP.Query().Where(aip.AipID(dr.AIPUUID)).OnlyID(ctx)
	if err != nil {
		return rollback(tx, fmt.Errorf("create deletion request: %v", err))
	}

	q := tx.DeletionRequest.Create().
		SetUUID(dr.UUID).
		SetRequester(dr.Requester).
		SetRequesterIss(dr.RequesterIss).
		SetRequesterSub(dr.RequesterSub).
		SetReason(dr.Reason).
		SetAipID(aipDBID).
		SetWorkflowID(dr.WorkflowDBID)

	if !dr.RequestedAt.IsZero() {
		q.SetRequestedAt(dr.RequestedAt)
	}

	dbdr, err := q.Save(ctx)
	if err != nil {
		return rollback(tx, fmt.Errorf("create deletion request: %v", err))
	}
	if err = tx.Commit(); err != nil {
		return rollback(tx, fmt.Errorf("create deletion request: %v", err))
	}

	dr.DBID = dbdr.ID

	return nil
}

func (c *Client) ListDeletionRequests(
	ctx context.Context,
	f *persistence.DeletionRequestFilter,
) ([]*types.DeletionRequest, error) {
	query := c.c.DeletionRequest.Query().WithAip()

	if f != nil {
		if f.AIPUUID != nil {
			query = query.Where(deletionrequest.HasAipWith(aip.AipID(*f.AIPUUID)))
		}
		if f.Status != nil {
			query = query.Where(deletionrequest.StatusEQ(*f.Status))
		}
	}

	r, err := query.All(ctx)
	if err != nil {
		return nil, fmt.Errorf("list deletion requests: %v", err)
	}

	drs := make([]*types.DeletionRequest, len(r))
	for i, dr := range r {
		drs[i] = convertDeletionRequest(dr)
	}

	return drs, nil
}

func (c *Client) UpdateDeletionRequest(
	ctx context.Context,
	id int,
	upd persistence.DeletionRequestUpdater,
) (*types.DeletionRequest, error) {
	tx, err := c.c.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("update deletion request: %v", err)
	}

	dr, err := tx.DeletionRequest.Query().
		Where(deletionrequest.ID(id)).
		WithAip().
		Only(ctx)
	if err != nil {
		return nil, rollback(tx, fmt.Errorf("update deletion request: %v", err))
	}

	// Get the UUID of the related AIP.
	aipUUID := dr.Edges.Aip.AipID

	up, err := upd(convertDeletionRequest(dr))
	if err != nil {
		return nil, rollback(tx, fmt.Errorf("update deletion request: %v", err))
	}

	q := tx.DeletionRequest.UpdateOneID(id).
		SetReviewer(up.Reviewer).
		SetReviewerIss(up.ReviewerIss).
		SetReviewerSub(up.ReviewerSub).
		SetStatus(up.Status)

	if !up.ReviewedAt.IsZero() {
		q.SetReviewedAt(up.ReviewedAt)
	}

	dr, err = q.Save(ctx)
	if err != nil {
		return nil, rollback(tx, fmt.Errorf("update deletion request: %v", err))
	}
	if err = tx.Commit(); err != nil {
		return nil, rollback(tx, fmt.Errorf("update deletion request: %v", err))
	}

	r := convertDeletionRequest(dr)

	// Add the AIP UUID to the returned deletion request.
	r.AIPUUID = aipUUID

	return r, nil
}

func (c *Client) ReadDeletionRequest(
	ctx context.Context,
	id uuid.UUID,
) (*types.DeletionRequest, error) {
	dr, err := c.c.DeletionRequest.Query().
		Where(deletionrequest.UUID(id)).
		Only(ctx)
	if err != nil {
		return nil, fmt.Errorf("read deletion request: %v", err)
	}

	return convertDeletionRequest(dr), nil
}

func convertDeletionRequest(dbdr *db.DeletionRequest) *types.DeletionRequest {
	dr := &types.DeletionRequest{
		DBID:         dbdr.ID,
		UUID:         dbdr.UUID,
		Requester:    dbdr.Requester,
		RequesterIss: dbdr.RequesterIss,
		RequesterSub: dbdr.RequesterSub,
		Reviewer:     dbdr.Reviewer,
		ReviewerIss:  dbdr.ReviewerIss,
		ReviewerSub:  dbdr.ReviewerSub,
		Reason:       dbdr.Reason,
		Status:       dbdr.Status,
		RequestedAt:  dbdr.RequestedAt,
		ReviewedAt:   dbdr.ReviewedAt,
		WorkflowDBID: dbdr.WorkflowID,
	}

	if dbdr.Edges.Aip != nil {
		dr.AIPUUID = dbdr.Edges.Aip.AipID
	}

	return dr
}
