package entclient

import (
	"context"
	"time"

	"github.com/artefactual-sdps/enduro/internal/datatypes"
)

func (c *client) CreatePreservationAction(ctx context.Context, pa *datatypes.PreservationAction) error {
	// Validate required fields.
	if pa.WorkflowID == "" {
		return newRequiredFieldError("WorkflowID")
	}
	if pa.SIPID == 0 {
		return newRequiredFieldError("SIPID")
	}

	// TODO: Validate Type & Status enums.

	// Handle nullable fields.
	var startedAt *time.Time
	if pa.StartedAt.Valid {
		startedAt = &pa.StartedAt.Time
	}

	var completedAt *time.Time
	if pa.CompletedAt.Valid {
		completedAt = &pa.CompletedAt.Time
	}

	q := c.ent.PreservationAction.Create().
		SetWorkflowID(pa.WorkflowID).
		SetType(int8(pa.Type)).     // #nosec G115 -- constrained value.
		SetStatus(int8(pa.Status)). // #nosec G115 -- constrained value.
		SetNillableStartedAt(startedAt).
		SetNillableCompletedAt(completedAt).
		SetSipID(pa.SIPID)

	r, err := q.Save(ctx)
	if err != nil {
		return newDBErrorWithDetails(err, "create preservation action")
	}

	// Update value of pt with data from DB (e.g. ID).
	*pa = *convertPreservationAction(r)

	return nil
}
