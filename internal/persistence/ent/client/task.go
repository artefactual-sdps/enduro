package client

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/artefactual-sdps/enduro/internal/datatypes"
	"github.com/artefactual-sdps/enduro/internal/persistence"
	"github.com/artefactual-sdps/enduro/internal/persistence/ent/db"
	"github.com/artefactual-sdps/enduro/internal/persistence/ent/db/task"
	"github.com/artefactual-sdps/enduro/internal/persistence/ent/db/workflow"
)

func (c *client) CreateTask(ctx context.Context, task *datatypes.Task) error {
	// Validate required fields.
	if task.UUID == uuid.Nil {
		return newRequiredFieldError("UUID")
	}
	if task.Name == "" {
		return newRequiredFieldError("Name")
	}
	if task.WorkflowUUID == uuid.Nil {
		return newRequiredFieldError("WorkflowUUID")
	}
	// TODO: Validate Status.

	// Handle nullable fields.
	var startedAt *time.Time
	if task.StartedAt.Valid {
		startedAt = &task.StartedAt.Time
	}

	var completedAt *time.Time
	if task.CompletedAt.Valid {
		completedAt = &task.CompletedAt.Time
	}

	tx, err := c.ent.BeginTx(ctx, nil)
	if err != nil {
		return newDBErrorWithDetails(err, "create task")
	}

	wDBID, err := tx.Workflow.Query().Where(workflow.UUID(task.WorkflowUUID)).OnlyID(ctx)
	if err != nil {
		return rollback(tx, newDBErrorWithDetails(err, "create task"))
	}

	q := tx.Task.Create().
		SetUUID(task.UUID).
		SetName(task.Name).
		SetStatus(int8(task.Status)). // #nosec G115 -- constrained value.
		SetNillableStartedAt(startedAt).
		SetNillableCompletedAt(completedAt).
		SetNote(task.Note).
		SetWorkflowID(wDBID)

	dbt, err := q.Save(ctx)
	if err != nil {
		return rollback(tx, newDBErrorWithDetails(err, "create task"))
	}
	if err = tx.Commit(); err != nil {
		return rollback(tx, newDBErrorWithDetails(err, "create task"))
	}

	task.ID = dbt.ID

	return nil
}

func (c *client) CreateTasks(ctx context.Context, seq persistence.TaskSequence) error {
	tx, err := c.ent.BeginTx(ctx, nil)
	if err != nil {
		return newDBErrorWithDetails(err, "create tasks")
	}

	workflowIDs := make(map[uuid.UUID]int)

	for ck := range chunk(seq) {
		if len(ck) == 0 {
			continue
		}

		builders := make([]*db.TaskCreate, 0, len(ck))
		originals := make([]*datatypes.Task, 0, len(ck))

		for _, task := range ck {
			if task.UUID == uuid.Nil {
				return rollback(tx, newRequiredFieldError("UUID"))
			}
			if task.Name == "" {
				return rollback(tx, newRequiredFieldError("Name"))
			}
			if task.WorkflowUUID == uuid.Nil {
				return rollback(tx, newRequiredFieldError("WorkflowUUID"))
			}

			var startedAt *time.Time
			if task.StartedAt.Valid {
				startedAt = &task.StartedAt.Time
			}

			var completedAt *time.Time
			if task.CompletedAt.Valid {
				completedAt = &task.CompletedAt.Time
			}

			wID, ok := workflowIDs[task.WorkflowUUID]
			if !ok {
				wID, err = tx.Workflow.Query().
					Where(workflow.UUID(task.WorkflowUUID)).
					OnlyID(ctx)
				if err != nil {
					return rollback(tx, newDBErrorWithDetails(err, "create tasks"))
				}
				workflowIDs[task.WorkflowUUID] = wID
			}

			builder := tx.Task.Create().
				SetUUID(task.UUID).
				SetName(task.Name).
				SetStatus(int8(task.Status)). // #nosec G115 -- constrained value.
				SetNillableStartedAt(startedAt).
				SetNillableCompletedAt(completedAt).
				SetNote(task.Note).
				SetWorkflowID(wID)

			builders = append(builders, builder)
			originals = append(originals, task)
		}

		created, err := tx.Task.CreateBulk(builders...).Save(ctx)
		if err != nil {
			return rollback(tx, newDBErrorWithDetails(err, "create tasks"))
		}

		if len(created) != len(originals) {
			return rollback(
				tx,
				newDBErrorWithDetails(
					fmt.Errorf("create tasks: created %d rows, expected %d", len(created), len(originals)),
					"create tasks",
				),
			)
		}

		for i, dbt := range created {
			originals[i].ID = dbt.ID
		}
	}

	if err = tx.Commit(); err != nil {
		return rollback(tx, newDBErrorWithDetails(err, "create tasks"))
	}

	return nil
}

func (c *client) UpdateTask(
	ctx context.Context,
	id int,
	updater persistence.TaskUpdater,
) (*datatypes.Task, error) {
	tx, err := c.ent.BeginTx(ctx, nil)
	if err != nil {
		return nil, newDBErrorWithDetails(err, "update task")
	}

	task, err := tx.Task.Query().WithWorkflow(func(q *db.WorkflowQuery) {
		q.Select(workflow.FieldUUID)
	}).Where(task.ID(id)).Only(ctx)
	if err != nil {
		return nil, rollback(tx, newDBError(err))
	}

	// Keep track of the workflow UUID to include it in the task after conversion.
	wUUID := task.Edges.Workflow.UUID

	up, err := updater(convertTask(task))
	if err != nil {
		return nil, rollback(tx, newUpdaterError(err))
	}

	q := tx.Task.UpdateOneID(id).
		SetName(up.Name).
		SetStatus(int8(up.Status)). // #nosec G115 -- constrained value.
		SetNote(up.Note)

	// Set nullable column values.
	if up.StartedAt.Valid {
		q.SetStartedAt(up.StartedAt.Time)
	}
	if up.CompletedAt.Valid {
		q.SetCompletedAt(up.CompletedAt.Time)
	}

	task, err = q.Save(ctx)
	if err != nil {
		return nil, rollback(tx, newDBError(err))
	}
	if err = tx.Commit(); err != nil {
		return nil, rollback(tx, newDBError(err))
	}

	t := convertTask(task)
	t.WorkflowUUID = wUUID

	return t, nil
}
