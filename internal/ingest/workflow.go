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

func (svc *ingestImpl) CreateWorkflow(
	ctx context.Context,
	w *datatypes.Workflow,
) error {
	err := svc.perSvc.CreateWorkflow(ctx, w)
	if err != nil {
		return fmt.Errorf("workflow: create: %v", err)
	}

	ev := &goaingest.SIPWorkflowCreatedEvent{
		ID:   uint(w.ID), // #nosec G115 -- constrained value.
		Item: workflowToGoa(w),
	}
	event.PublishEvent(ctx, svc.evsvc, ev)

	return nil
}

func (svc *ingestImpl) SetWorkflowStatus(
	ctx context.Context,
	ID int,
	status enums.WorkflowStatus,
) error {
	query := `UPDATE workflow SET status = ? WHERE id = ?`
	args := []any{
		status,
		ID,
	}

	_, err := svc.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("error updating workflow: %w", err)
	}

	if item, err := svc.readWorkflow(ctx, ID); err == nil {
		event.PublishEvent(
			ctx,
			svc.evsvc,
			&goaingest.SIPWorkflowUpdatedEvent{ID: item.ID, Item: item},
		)
	}

	return nil
}

func (svc *ingestImpl) CompleteWorkflow(
	ctx context.Context,
	ID int,
	status enums.WorkflowStatus,
	completedAt time.Time,
) error {
	query := `UPDATE workflow SET status = ?, completed_at = ? WHERE id = ?`
	args := []any{
		status,
		completedAt,
		ID,
	}

	_, err := svc.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("error updating workflow: %w", err)
	}

	if item, err := svc.readWorkflow(ctx, ID); err == nil {
		event.PublishEvent(
			ctx,
			svc.evsvc,
			&goaingest.SIPWorkflowUpdatedEvent{ID: item.ID, Item: item},
		)
	}

	return nil
}

func (svc *ingestImpl) readWorkflow(
	ctx context.Context,
	ID int,
) (*goaingest.SIPWorkflow, error) {
	query := `
		SELECT
			workflow.id,
			workflow.temporal_id,
			workflow.type,
			workflow.status,
			CONVERT_TZ(workflow.started_at, @@session.time_zone, '+00:00') AS started_at,
			CONVERT_TZ(workflow.completed_at, @@session.time_zone, '+00:00') AS completed_at,
			sip.uuid as sip_uuid
		FROM workflow
		LEFT JOIN sip ON (workflow.sip_id = sip.id)
		WHERE workflow.id = ?
	`

	args := []any{ID}
	w := datatypes.Workflow{}
	if err := svc.db.GetContext(ctx, &w, query, args...); err != nil {
		return nil, err
	}

	return workflowToGoa(&w), nil
}
