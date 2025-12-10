package ingest

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	goaingest "github.com/artefactual-sdps/enduro/internal/api/gen/ingest"
	"github.com/artefactual-sdps/enduro/internal/datatypes"
	"github.com/artefactual-sdps/enduro/internal/enums"
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
		UUID: w.UUID,
		Item: workflowToGoa(w),
	}
	PublishEvent(ctx, svc.evsvc, ev)

	return nil
}

func (svc *ingestImpl) SetWorkflowStatus(
	ctx context.Context,
	ID int,
	status enums.WorkflowStatus,
) error {
	w, err := svc.perSvc.UpdateWorkflow(ctx, ID, func(w *datatypes.Workflow) (*datatypes.Workflow, error) {
		w.Status = status
		return w, nil
	})
	if err != nil {
		return fmt.Errorf("error updating workflow: %w", err)
	}

	PublishEvent(
		ctx,
		svc.evsvc,
		&goaingest.SIPWorkflowUpdatedEvent{UUID: w.UUID, Item: workflowToGoa(w)},
	)

	return nil
}

func (svc *ingestImpl) CompleteWorkflow(
	ctx context.Context,
	ID int,
	status enums.WorkflowStatus,
	completedAt time.Time,
) error {
	w, err := svc.perSvc.UpdateWorkflow(ctx, ID, func(w *datatypes.Workflow) (*datatypes.Workflow, error) {
		w.Status = status
		w.CompletedAt = sql.NullTime{
			Time:  completedAt,
			Valid: !completedAt.IsZero(),
		}
		return w, nil
	})
	if err != nil {
		return fmt.Errorf("error updating workflow: %w", err)
	}

	PublishEvent(
		ctx,
		svc.evsvc,
		&goaingest.SIPWorkflowUpdatedEvent{UUID: w.UUID, Item: workflowToGoa(w)},
	)

	return nil
}
