package client

import (
	"context"

	"github.com/artefactual-sdps/enduro/internal/storage/persistence"
	"github.com/artefactual-sdps/enduro/internal/storage/persistence/ent/db"
	"github.com/artefactual-sdps/enduro/internal/storage/types"
)

func (c *Client) CreateTask(ctx context.Context, t *types.Task) error {
	q := c.c.Task.Create().
		SetUUID(t.UUID).
		SetName(t.Name).
		SetStatus(t.Status).
		SetNote(t.Note).
		SetWorkflowID(t.WorkflowID)

	if !t.StartedAt.IsZero() {
		q.SetStartedAt(t.StartedAt)
	}
	if !t.CompletedAt.IsZero() {
		q.SetCompletedAt(t.CompletedAt)
	}

	dbt, err := q.Save(ctx)
	if err != nil {
		return err
	}

	t.ID = dbt.ID

	return nil
}

func (c *Client) UpdateTask(ctx context.Context, id int, upd persistence.TaskUpdater) (*types.Task, error) {
	tx, err := c.c.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}

	t, err := tx.Task.Get(ctx, id)
	if err != nil {
		return nil, rollback(tx, err)
	}

	up, err := upd(convertTask(t))
	if err != nil {
		return nil, rollback(tx, err)
	}

	q := tx.Task.UpdateOneID(id).
		SetUUID(up.UUID).
		SetName(up.Name).
		SetStatus(up.Status).
		SetNote(up.Note).
		SetWorkflowID(up.WorkflowID)

	if !up.StartedAt.IsZero() {
		q.SetStartedAt(up.StartedAt)
	}
	if !up.CompletedAt.IsZero() {
		q.SetCompletedAt(up.CompletedAt)
	}

	t, err = q.Save(ctx)
	if err != nil {
		return nil, rollback(tx, err)
	}
	if err = tx.Commit(); err != nil {
		return nil, rollback(tx, err)
	}

	return convertTask(t), nil
}

func convertTask(dbt *db.Task) *types.Task {
	return &types.Task{
		ID:          dbt.ID,
		UUID:        dbt.UUID,
		Name:        dbt.Name,
		Status:      dbt.Status,
		StartedAt:   dbt.StartedAt,
		CompletedAt: dbt.CompletedAt,
		Note:        dbt.Note,
		WorkflowID:  dbt.WorkflowID,
	}
}
