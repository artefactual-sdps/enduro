package entclient

import (
	"context"
	"time"

	"github.com/artefactual-sdps/enduro/internal/datatypes"
)

func (c *client) CreateWorkflow(ctx context.Context, w *datatypes.Workflow) error {
	// Validate required fields.
	if w.TemporalID == "" {
		return newRequiredFieldError("TemporalID")
	}
	if w.SIPID == 0 {
		return newRequiredFieldError("SIPID")
	}

	// TODO: Validate Type & Status enums.

	// Handle nullable fields.
	var startedAt *time.Time
	if w.StartedAt.Valid {
		startedAt = &w.StartedAt.Time
	}

	var completedAt *time.Time
	if w.CompletedAt.Valid {
		completedAt = &w.CompletedAt.Time
	}

	q := c.ent.Workflow.Create().
		SetTemporalID(w.TemporalID).
		SetType(int8(w.Type)).     // #nosec G115 -- constrained value.
		SetStatus(int8(w.Status)). // #nosec G115 -- constrained value.
		SetNillableStartedAt(startedAt).
		SetNillableCompletedAt(completedAt).
		SetSipID(w.SIPID)

	r, err := q.Save(ctx)
	if err != nil {
		return newDBErrorWithDetails(err, "create workflow")
	}

	// Update value of task with data from DB (e.g. ID).
	*w = *convertWorkflow(r)

	return nil
}
