package ingest

import (
	"context"
	"fmt"
	"time"

	goaingest "github.com/artefactual-sdps/enduro/internal/api/gen/ingest"
	"github.com/artefactual-sdps/enduro/internal/datatypes"
	"github.com/artefactual-sdps/enduro/internal/enums"
	"github.com/artefactual-sdps/enduro/internal/event"
)

func (svc *ingestImpl) CreatePreservationAction(
	ctx context.Context,
	pa *datatypes.PreservationAction,
) error {
	err := svc.perSvc.CreatePreservationAction(ctx, pa)
	if err != nil {
		return fmt.Errorf("preservation action: create: %v", err)
	}

	ev := &goaingest.SIPPreservationActionCreatedEvent{
		ID:   uint(pa.ID), // #nosec G115 -- constrained value.
		Item: preservationActionToGoa(pa),
	}
	event.PublishEvent(ctx, svc.evsvc, ev)

	return nil
}

func (svc *ingestImpl) SetPreservationActionStatus(
	ctx context.Context,
	ID int,
	status enums.PreservationActionStatus,
) error {
	query := `UPDATE preservation_action SET status = ? WHERE id = ?`
	args := []interface{}{
		status,
		ID,
	}

	_, err := svc.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("error updating preservation action: %w", err)
	}

	if item, err := svc.readPreservationAction(ctx, ID); err == nil {
		event.PublishEvent(
			ctx,
			svc.evsvc,
			&goaingest.SIPPreservationActionUpdatedEvent{ID: item.ID, Item: item},
		)
	}

	return nil
}

func (svc *ingestImpl) CompletePreservationAction(
	ctx context.Context,
	ID int,
	status enums.PreservationActionStatus,
	completedAt time.Time,
) error {
	query := `UPDATE preservation_action SET status = ?, completed_at = ? WHERE id = ?`
	args := []interface{}{
		status,
		completedAt,
		ID,
	}

	_, err := svc.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("error updating preservation action: %w", err)
	}

	if item, err := svc.readPreservationAction(ctx, ID); err == nil {
		event.PublishEvent(
			ctx,
			svc.evsvc,
			&goaingest.SIPPreservationActionUpdatedEvent{ID: item.ID, Item: item},
		)
	}

	return nil
}

func (svc *ingestImpl) readPreservationAction(
	ctx context.Context,
	ID int,
) (*goaingest.SIPPreservationAction, error) {
	query := `
		SELECT
			preservation_action.id,
			preservation_action.workflow_id,
			preservation_action.type,
			preservation_action.status,
			CONVERT_TZ(preservation_action.started_at, @@session.time_zone, '+00:00') AS started_at,
			CONVERT_TZ(preservation_action.completed_at, @@session.time_zone, '+00:00') AS completed_at,
			preservation_action.sip_id
		FROM preservation_action
		LEFT JOIN sip ON (preservation_action.sip_id = sip.id)
		WHERE preservation_action.id = ?
	`

	args := []interface{}{ID}
	pa := datatypes.PreservationAction{}
	if err := svc.db.GetContext(ctx, &pa, query, args...); err != nil {
		return nil, err
	}

	return preservationActionToGoa(&pa), nil
}
