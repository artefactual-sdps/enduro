package client

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/artefactual-sdps/enduro/internal/storage/persistence"
	"github.com/artefactual-sdps/enduro/internal/storage/persistence/ent/db"
	"github.com/artefactual-sdps/enduro/internal/storage/persistence/ent/db/aip"
	"github.com/artefactual-sdps/enduro/internal/storage/persistence/ent/db/workflow"
	"github.com/artefactual-sdps/enduro/internal/storage/types"
)

func (c *Client) CreateWorkflow(ctx context.Context, w *types.Workflow) error {
	tx, err := c.c.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("create workflow: %v", err)
	}

	aipDBID, err := tx.AIP.Query().Where(aip.AipID(w.AIPUUID)).OnlyID(ctx)
	if err != nil {
		return rollback(tx, fmt.Errorf("create workflow: %v", err))
	}

	q := tx.Workflow.Create().
		SetUUID(w.UUID).
		SetTemporalID(w.TemporalID).
		SetType(w.Type).
		SetStatus(w.Status).
		SetAipID(aipDBID)

	if !w.StartedAt.IsZero() {
		q.SetStartedAt(w.StartedAt)
	}
	if !w.CompletedAt.IsZero() {
		q.SetCompletedAt(w.CompletedAt)
	}

	dbw, err := q.Save(ctx)
	if err != nil {
		return rollback(tx, fmt.Errorf("create workflow: %v", err))
	}
	if err = tx.Commit(); err != nil {
		return rollback(tx, fmt.Errorf("create workflow: %v", err))
	}

	w.DBID = dbw.ID

	return nil
}

func (c *Client) ReadWorkflow(ctx context.Context, dbID int) (*types.Workflow, error) {
	dbw, err := c.c.Workflow.Query().
		Where(workflow.ID(dbID)).
		Only(ctx)
	if err != nil {
		return nil, fmt.Errorf("read workflow: %v", err)
	}

	w := convertWorkflow(dbw)

	return w, nil
}

func (c *Client) UpdateWorkflow(ctx context.Context, id int, upd persistence.WorkflowUpdater) (*types.Workflow, error) {
	tx, err := c.c.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("update workflow: %v", err)
	}

	dbw, err := tx.Workflow.Query().WithAip(func(q *db.AIPQuery) {
		q.Select(aip.FieldAipID)
	}).Where(workflow.ID(id)).Only(ctx)
	if err != nil {
		return nil, rollback(tx, fmt.Errorf("update workflow: %v", err))
	}

	// Keep track of the AIP UUID to include it in the workflow after conversion.
	aipUUID := dbw.Edges.Aip.AipID

	up, err := upd(convertWorkflow(dbw))
	if err != nil {
		return nil, rollback(tx, fmt.Errorf("update workflow: %v", err))
	}

	q := tx.Workflow.UpdateOneID(id).
		SetUUID(up.UUID).
		SetTemporalID(up.TemporalID).
		SetType(up.Type).
		SetStatus(up.Status)

	if !up.StartedAt.IsZero() {
		q.SetStartedAt(up.StartedAt)
	}
	if !up.CompletedAt.IsZero() {
		q.SetCompletedAt(up.CompletedAt)
	}

	dbw, err = q.Save(ctx)
	if err != nil {
		return nil, rollback(tx, fmt.Errorf("update workflow: %v", err))
	}
	if err = tx.Commit(); err != nil {
		return nil, rollback(tx, fmt.Errorf("update workflow: %v", err))
	}

	w := convertWorkflow(dbw)
	w.AIPUUID = aipUUID

	return w, nil
}

func convertWorkflow(dbw *db.Workflow) *types.Workflow {
	var aipUUID uuid.UUID
	if dbw.Edges.Aip != nil {
		aipUUID = dbw.Edges.Aip.AipID
	}
	return &types.Workflow{
		DBID:        dbw.ID,
		UUID:        dbw.UUID,
		TemporalID:  dbw.TemporalID,
		Type:        dbw.Type,
		Status:      dbw.Status,
		StartedAt:   dbw.StartedAt,
		CompletedAt: dbw.CompletedAt,
		AIPUUID:     aipUUID,
	}
}
