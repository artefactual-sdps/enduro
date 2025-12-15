package client

import (
	"context"
	"time"

	entsql "entgo.io/ent/dialect/sql"
	"github.com/google/uuid"

	"github.com/artefactual-sdps/enduro/internal/datatypes"
	"github.com/artefactual-sdps/enduro/internal/persistence"
	"github.com/artefactual-sdps/enduro/internal/persistence/ent/db/sip"
	"github.com/artefactual-sdps/enduro/internal/persistence/ent/db/workflow"
)

func (c *client) CreateWorkflow(ctx context.Context, w *datatypes.Workflow) error {
	// Validate required fields.
	if w.UUID == uuid.Nil {
		return newRequiredFieldError("UUID")
	}
	if w.TemporalID == "" {
		return newRequiredFieldError("TemporalID")
	}
	if w.SIPUUID == uuid.Nil {
		return newRequiredFieldError("SIPUUID")
	}
	if !w.Type.IsValid() {
		return newInvalidFieldError("Type", w.Type.String())
	}

	// Handle nullable fields.
	var startedAt *time.Time
	if w.StartedAt.Valid {
		startedAt = &w.StartedAt.Time
	}
	var completedAt *time.Time
	if w.CompletedAt.Valid {
		completedAt = &w.CompletedAt.Time
	}

	tx, err := c.ent.BeginTx(ctx, nil)
	if err != nil {
		return newDBErrorWithDetails(err, "create workflow")
	}

	sipDBID, err := tx.SIP.Query().Where(sip.UUID(w.SIPUUID)).OnlyID(ctx)
	if err != nil {
		return rollback(tx, newDBErrorWithDetails(err, "create workflow"))
	}

	q := tx.Workflow.Create().
		SetUUID(w.UUID).
		SetTemporalID(w.TemporalID).
		SetType(w.Type).
		SetStatus(int8(w.Status)). // #nosec G115 -- constrained value.
		SetNillableStartedAt(startedAt).
		SetNillableCompletedAt(completedAt).
		SetSipID(sipDBID)

	dbw, err := q.Save(ctx)
	if err != nil {
		return rollback(tx, newDBErrorWithDetails(err, "create workflow"))
	}
	if err = tx.Commit(); err != nil {
		return rollback(tx, newDBErrorWithDetails(err, "create workflow"))
	}

	w.ID = dbw.ID

	return nil
}

func (c *client) UpdateWorkflow(
	ctx context.Context,
	id int,
	updater persistence.WorkflowUpdater,
) (*datatypes.Workflow, error) {
	tx, err := c.ent.BeginTx(ctx, nil)
	if err != nil {
		return nil, newDBError(err)
	}

	dbw, err := tx.Workflow.Query().
		Where(workflow.ID(id)).
		WithSip().
		WithTasks().
		Only(ctx)
	if err != nil {
		return nil, rollback(tx, newDBError(err))
	}

	up, err := updater(convertWorkflow(dbw))
	if err != nil {
		return nil, rollback(tx, newUpdaterError(err))
	}

	q := tx.Workflow.UpdateOneID(dbw.ID)

	if up.Status.IsValid() {
		q.SetStatus(int8(up.Status)) // #nosec G115 -- constrained value.
	}
	if up.StartedAt.Valid {
		q.SetStartedAt(up.StartedAt.Time.UTC())
	} else {
		q.ClearStartedAt()
	}
	if up.CompletedAt.Valid {
		q.SetCompletedAt(up.CompletedAt.Time.UTC())
	} else {
		q.ClearCompletedAt()
	}

	dbw, err = q.Save(ctx)
	if err != nil {
		return nil, rollback(tx, newDBError(err))
	}

	dbw, err = tx.Workflow.Query().
		Where(workflow.ID(dbw.ID)).
		WithSip().
		WithTasks().
		Only(ctx)
	if err != nil {
		return nil, rollback(tx, newDBError(err))
	}

	if err = tx.Commit(); err != nil {
		return nil, rollback(tx, newDBError(err))
	}

	return convertWorkflow(dbw), nil
}

func (c *client) ReadWorkflow(ctx context.Context, id int) (*datatypes.Workflow, error) {
	dbw, err := c.ent.Workflow.Query().
		Where(workflow.ID(id)).
		WithSip().
		WithTasks().
		Only(ctx)
	if err != nil {
		return nil, newDBError(err)
	}

	return convertWorkflow(dbw), nil
}

func (c *client) ListWorkflowsBySIP(ctx context.Context, sipUUID uuid.UUID) ([]*datatypes.Workflow, error) {
	dbw, err := c.ent.Workflow.Query().
		WithSip().
		WithTasks().
		Where(workflow.HasSipWith(sip.UUID(sipUUID))).
		Order(workflow.ByStartedAt(entsql.OrderDesc())).
		All(ctx)
	if err != nil {
		return nil, newDBError(err)
	}

	res := make([]*datatypes.Workflow, len(dbw))
	for i, w := range dbw {
		res[i] = convertWorkflow(w)
	}

	return res, nil
}
