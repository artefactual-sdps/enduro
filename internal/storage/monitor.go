package storage

import (
	"context"
	"time"

	goastorage "github.com/artefactual-sdps/enduro/internal/api/gen/storage"
	"github.com/artefactual-sdps/enduro/internal/auth"
)

func (s *serviceImpl) Monitor(
	ctx context.Context,
	payload *goastorage.MonitorPayload,
	stream goastorage.MonitorServerStream,
) error {
	defer stream.Close()

	claims := auth.UserClaimsFromContext(ctx)

	// Subscribe to the event service.
	sub, err := s.evsvc.Subscribe(ctx)
	if err != nil {
		s.logger.Error(err, "failed to subscribe to event service")
		return ErrInternalError
	}
	defer sub.Close()

	// Say hello to be nice.
	event := &goastorage.StoragePingEvent{Message: new("Hello")}
	if err := stream.SendWithContext(ctx, &goastorage.StorageEvent{Value: NewEventValue(event)}); err != nil {
		// Consider send errors as client disconnections.
		s.logger.V(1).Info("Failed to send hello event.", "err", err)
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
			event := &goastorage.StoragePingEvent{Message: new("Ping")}
			if err := stream.SendWithContext(ctx, &goastorage.StorageEvent{Value: NewEventValue(event)}); err != nil {
				// Consider send errors as client disconnections.
				s.logger.V(1).Info("Failed to send ping event.", "err", err)
				return nil
			}

		case event, ok := <-sub.C():
			if !ok || event == nil {
				return nil
			}

			// Check the event type and the user attributes before sending.
			switch event.Value.Kind() {
			case goastorage.ValueKindStoragePingEvent:
				// Is this event even sent through this channel?
			case goastorage.ValueKindLocationCreatedEvent:
				if !claims.CheckAttributes([]string{auth.StorageLocationsListAttr}) {
					continue
				}
			case goastorage.ValueKindAipCreatedEvent:
				if !claims.CheckAttributes([]string{auth.StorageAIPSListAttr}) {
					continue
				}
			case goastorage.ValueKindAipUpdatedEvent,
				goastorage.ValueKindAipStatusUpdatedEvent,
				goastorage.ValueKindAipLocationUpdatedEvent:
				if !claims.CheckAttributes([]string{auth.StorageAIPSListAttr}) &&
					!claims.CheckAttributes([]string{auth.StorageAIPSReadAttr}) {
					continue
				}
			case goastorage.ValueKindAipWorkflowCreatedEvent,
				goastorage.ValueKindAipWorkflowUpdatedEvent,
				goastorage.ValueKindAipTaskCreatedEvent,
				goastorage.ValueKindAipTaskUpdatedEvent:
				if !claims.CheckAttributes([]string{auth.StorageAIPSWorkflowsListAttr}) {
					continue
				}
			default:
				// Do not send the event if the type is not considered.
				continue
			}

			if err := stream.SendWithContext(ctx, event); err != nil {
				// Consider send errors as client disconnections.
				s.logger.V(1).Info("Failed to send event.", "err", err)
				return nil
			}
		}
	}
}
