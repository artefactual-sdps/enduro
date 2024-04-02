package package_

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	goapackage "github.com/artefactual-sdps/enduro/internal/api/gen/package_"
	"github.com/artefactual-sdps/enduro/internal/datatypes"
	"github.com/artefactual-sdps/enduro/internal/enums"
	"github.com/artefactual-sdps/enduro/internal/event"
)

func (svc *packageImpl) CreatePreservationTask(ctx context.Context, pt *datatypes.PreservationTask) error {
	err := svc.perSvc.CreatePreservationTask(ctx, pt)
	if err != nil {
		return fmt.Errorf("preservation task: create: %v", err)
	}

	ev := &goapackage.PreservationTaskCreatedEvent{
		ID:   pt.ID,
		Item: preservationTaskToGoa(pt),
	}
	event.PublishEvent(ctx, svc.evsvc, ev)

	return nil
}

func (svc *packageImpl) CompletePreservationTask(
	ctx context.Context,
	id uint,
	status enums.PreservationTaskStatus,
	completedAt time.Time,
	note *string,
) error {
	pt, err := svc.perSvc.UpdatePreservationTask(
		ctx,
		id,
		func(pt *datatypes.PreservationTask) (*datatypes.PreservationTask, error) {
			pt.Status = status
			pt.CompletedAt = sql.NullTime{
				Time:  completedAt,
				Valid: true,
			}
			if note != nil {
				pt.Note = *note
			}

			return pt, nil
		},
	)
	if err != nil {
		return fmt.Errorf("error updating preservation task: %v", err)
	}

	ev := &goapackage.PreservationTaskUpdatedEvent{
		ID:   id,
		Item: preservationTaskToGoa(pt),
	}
	event.PublishEvent(ctx, svc.evsvc, ev)

	return nil
}
