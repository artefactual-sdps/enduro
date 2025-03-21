package client

import (
	"context"

	"github.com/artefactual-sdps/enduro/internal/storage/persistence"
	"github.com/artefactual-sdps/enduro/internal/storage/persistence/ent/db"
	"github.com/artefactual-sdps/enduro/internal/storage/types"
)

func (c *Client) CreateWorkflow(ctx context.Context, w *types.Workflow) error {
	q := c.c.Workflow.Create().
		SetUUID(w.UUID).
		SetTemporalID(w.TemporalID).
		SetType(w.Type).
		SetStatus(w.Status).
		SetAipID(w.AIPID)

	if !w.StartedAt.IsZero() {
		q.SetStartedAt(w.StartedAt)
	}
	if !w.CompletedAt.IsZero() {
		q.SetCompletedAt(w.CompletedAt)
	}

	dbw, err := q.Save(ctx)
	if err != nil {
		return err
	}

	w.ID = dbw.ID

	return nil
}

func (c *Client) UpdateWorkflow(ctx context.Context, id int, upd persistence.WorkflowUpdater) (*types.Workflow, error) {
	tx, err := c.c.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}

	w, err := tx.Workflow.Get(ctx, id)
	if err != nil {
		return nil, rollback(tx, err)
	}

	up, err := upd(convertWorkflow(w))
	if err != nil {
		return nil, rollback(tx, err)
	}

	q := tx.Workflow.UpdateOneID(id).
		SetUUID(up.UUID).
		SetTemporalID(up.TemporalID).
		SetType(up.Type).
		SetStatus(up.Status).
		SetAipID(up.AIPID)

	if !up.StartedAt.IsZero() {
		q.SetStartedAt(up.StartedAt)
	}
	if !up.CompletedAt.IsZero() {
		q.SetCompletedAt(up.CompletedAt)
	}

	w, err = q.Save(ctx)
	if err != nil {
		return nil, rollback(tx, err)
	}
	if err = tx.Commit(); err != nil {
		return nil, rollback(tx, err)
	}

	return convertWorkflow(w), nil
}

func convertWorkflow(dbw *db.Workflow) *types.Workflow {
	return &types.Workflow{
		ID:          dbw.ID,
		UUID:        dbw.UUID,
		TemporalID:  dbw.TemporalID,
		Type:        dbw.Type,
		Status:      dbw.Status,
		StartedAt:   dbw.StartedAt,
		CompletedAt: dbw.CompletedAt,
		AIPID:       dbw.AipID,
	}
}
