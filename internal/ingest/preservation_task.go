package ingest

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	goaingest "github.com/artefactual-sdps/enduro/internal/api/gen/ingest"
	"github.com/artefactual-sdps/enduro/internal/datatypes"
	"github.com/artefactual-sdps/enduro/internal/enums"
	"github.com/artefactual-sdps/enduro/internal/event"
)

func (svc *ingestImpl) CreatePreservationTask(ctx context.Context, pt *datatypes.PreservationTask) error {
	err := svc.perSvc.CreatePreservationTask(ctx, pt)
	if err != nil {
		return fmt.Errorf("preservation task: create: %v", err)
	}

	event.PublishEvent(ctx, svc.evsvc, &goaingest.SIPPreservationTaskCreatedEvent{
		ID:   uint(pt.ID), // #nosec G115 -- constrained value.
		Item: preservationTaskToGoa(pt),
	})

	return nil
}

func (svc *ingestImpl) CompletePreservationTask(
	ctx context.Context,
	id int,
	status enums.PreservationTaskStatus,
	completedAt time.Time,
	note *string,
) error {
	if id < 0 {
		return fmt.Errorf("%w: ID", ErrInvalid)
	}

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

	event.PublishEvent(ctx, svc.evsvc, &goaingest.SIPPreservationTaskUpdatedEvent{
		ID:   uint(id), // #nosec G115 -- range validated.
		Item: preservationTaskToGoa(pt),
	})

	return nil
}
