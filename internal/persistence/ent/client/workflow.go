package entclient

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/artefactual-sdps/enduro/internal/datatypes"
	"github.com/artefactual-sdps/enduro/internal/persistence/ent/db/sip"
)

func (c *client) CreateWorkflow(ctx context.Context, w *datatypes.Workflow) error {
	// Validate required fields.
	if w.TemporalID == "" {
		return newRequiredFieldError("TemporalID")
	}
	if w.SIPUUID == uuid.Nil {
		return newRequiredFieldError("SIPUUID")
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
		SetTemporalID(w.TemporalID).
		SetType(int8(w.Type)).     // #nosec G115 -- constrained value.
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
