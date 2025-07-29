package client

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/artefactual-sdps/enduro/internal/storage/persistence"
	"github.com/artefactual-sdps/enduro/internal/storage/persistence/ent/db"
	"github.com/artefactual-sdps/enduro/internal/storage/persistence/ent/db/task"
	"github.com/artefactual-sdps/enduro/internal/storage/persistence/ent/db/workflow"
	"github.com/artefactual-sdps/enduro/internal/storage/types"
)

func (c *Client) CreateTask(ctx context.Context, t *types.Task) error {
	tx, err := c.c.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("create task: %v", err)
	}

	// TODO: Use UUIDs in the entire service interface, using t.WorkflowUUID here
	// to find the workflow DBID. This will make more sense then, for now we need
	// it the other way around to populate the workflow UUID in the task.
	workflow, err := tx.Workflow.Query().Where(workflow.ID(t.WorkflowDBID)).Only(ctx)
	if err != nil {
		return rollback(tx, fmt.Errorf("create task: %v", err))
	}

	q := tx.Task.Create().
		SetUUID(t.UUID).
		SetName(t.Name).
		SetStatus(t.Status).
		SetNote(t.Note).
		SetWorkflowID(t.WorkflowDBID)

	if !t.StartedAt.IsZero() {
		q.SetStartedAt(t.StartedAt)
	}
	if !t.CompletedAt.IsZero() {
		q.SetCompletedAt(t.CompletedAt)
	}

	dbt, err := q.Save(ctx)
	if err != nil {
		return rollback(tx, fmt.Errorf("create task: %v", err))
	}
	if err = tx.Commit(); err != nil {
		return rollback(tx, fmt.Errorf("create task: %v", err))
	}

	t.DBID = dbt.ID
	t.WorkflowUUID = workflow.UUID

	return nil
}

func (c *Client) UpdateTask(ctx context.Context, id int, upd persistence.TaskUpdater) (*types.Task, error) {
	tx, err := c.c.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("update task: %v", err)
	}

	dbt, err := tx.Task.Query().WithWorkflow(func(q *db.WorkflowQuery) {
		q.Select(workflow.FieldUUID)
	}).Where(task.ID(id)).Only(ctx)
	if err != nil {
		return nil, rollback(tx, fmt.Errorf("update task: %v", err))
	}

	// Keep track of the workflow UUID to include it in the task after conversion.
	wUUID := dbt.Edges.Workflow.UUID

	up, err := upd(convertTask(dbt))
	if err != nil {
		return nil, rollback(tx, fmt.Errorf("update task: %v", err))
	}

	q := tx.Task.UpdateOneID(id).
		SetUUID(up.UUID).
		SetName(up.Name).
		SetStatus(up.Status).
		SetNote(up.Note)

	if !up.StartedAt.IsZero() {
		q.SetStartedAt(up.StartedAt)
	}
	if !up.CompletedAt.IsZero() {
		q.SetCompletedAt(up.CompletedAt)
	}

	dbt, err = q.Save(ctx)
	if err != nil {
		return nil, rollback(tx, fmt.Errorf("update task: %v", err))
	}
	if err = tx.Commit(); err != nil {
		return nil, rollback(tx, fmt.Errorf("update task: %v", err))
	}

	t := convertTask(dbt)
	t.WorkflowUUID = wUUID

	return t, nil
}

func convertTask(dbt *db.Task) *types.Task {
	var wUUID uuid.UUID
	if dbt.Edges.Workflow != nil {
		wUUID = dbt.Edges.Workflow.UUID
	}
	return &types.Task{
		DBID:         dbt.ID,
		UUID:         dbt.UUID,
		Name:         dbt.Name,
		Status:       dbt.Status,
		StartedAt:    dbt.StartedAt,
		CompletedAt:  dbt.CompletedAt,
		Note:         dbt.Note,
		WorkflowDBID: dbt.WorkflowID,
		WorkflowUUID: wUUID,
	}
}
