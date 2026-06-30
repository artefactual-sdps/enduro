package ingest

import (
	"context"
	"time"

	goaingest "github.com/artefactual-sdps/enduro/internal/api/gen/ingest"
	"github.com/artefactual-sdps/enduro/internal/auth"
)

func (svc *ingestImpl) Monitor(
	ctx context.Context,
	payload *goaingest.MonitorPayload,
	stream goaingest.MonitorServerStream,
) error {
	defer stream.Close()

	claims := auth.UserClaimsFromContext(ctx)

	// Subscribe to the event service.
	sub, err := svc.evsvc.Subscribe(ctx)
	if err != nil {
		svc.logger.Error(err, "failed to subscribe to event service")
		return ErrInternalError
	}
	defer sub.Close()

	// Say hello to be nice.
	event := &goaingest.IngestPingEvent{Message: new("Hello")}
	if err := stream.SendWithContext(ctx, &goaingest.IngestEvent{Value: NewEventValue(event)}); err != nil {
		// Consider send errors as client disconnections.
		svc.logger.V(1).Info("Failed to send hello event.", "err", err)
		return nil
	}

	// We'll use this ticker to ping the client once in a while to detect stale
	// connections. I'm not entirely sure this is needed, it may depend on the
	// client or the various middlewares.
	ticker := time.NewTicker(time.Second * 10)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil

		case <-ticker.C:
			event := &goaingest.IngestPingEvent{Message: new("Ping")}
			if err := stream.SendWithContext(ctx, &goaingest.IngestEvent{Value: NewEventValue(event)}); err != nil {
				// Consider send errors as client disconnections.
				svc.logger.V(1).Info("Failed to send ping event.", "err", err)
				return nil
			}

		case event, ok := <-sub.C():
			if !ok || event == nil {
				return nil
			}

			// Check the event type and the user attributes before sending.
			switch event.Value.Kind() {
			case goaingest.ValueKindIngestPingEvent:
				// Is this event even sent through this channel?
			case goaingest.ValueKindSipCreatedEvent:
				if !claims.CheckAttributes([]string{auth.IngestSIPSListAttr}) {
					continue
				}
			case goaingest.ValueKindSipUpdatedEvent, goaingest.ValueKindSipStatusUpdatedEvent:
				if !claims.CheckAttributes([]string{auth.IngestSIPSListAttr}) &&
					!claims.CheckAttributes([]string{auth.IngestSIPSReadAttr}) {
					continue
				}
			case goaingest.ValueKindSipWorkflowCreatedEvent,
				goaingest.ValueKindSipWorkflowUpdatedEvent,
				goaingest.ValueKindSipTaskCreatedEvent,
				goaingest.ValueKindSipTaskUpdatedEvent:
				if !claims.CheckAttributes([]string{auth.IngestSIPSWorkflowsListAttr}) {
					continue
				}
			case goaingest.ValueKindBatchCreatedEvent:
				if !claims.CheckAttributes([]string{auth.IngestBatchesListAttr}) {
					continue
				}
			case goaingest.ValueKindBatchUpdatedEvent:
				if !claims.CheckAttributes([]string{auth.IngestBatchesListAttr}) &&
					!claims.CheckAttributes([]string{auth.IngestBatchesReadAttr}) {
					continue
				}
			default:
				// Do not send the event if the type is not considered.
				continue
			}

			if err := stream.SendWithContext(ctx, event); err != nil {
				// Consider send errors as client disconnections.
				svc.logger.V(1).Info("Failed to send event.", "err", err)
				return nil
			}
		}
	}
}
