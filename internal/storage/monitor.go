package storage

import (
	"context"
	"time"

	"go.artefactual.dev/tools/ref"

	"github.com/artefactual-sdps/enduro/internal/api/auth"
	goastorage "github.com/artefactual-sdps/enduro/internal/api/gen/storage"
)

func (s *serviceImpl) MonitorRequest(
	ctx context.Context,
	payload *goastorage.MonitorRequestPayload,
) (*goastorage.MonitorRequestResult, error) {
	res := &goastorage.MonitorRequestResult{}

	ticket, err := s.ticketProvider.Request(ctx, auth.UserClaimsFromContext(ctx))
	if err != nil {
		s.logger.Error(err, "failed to request ticket")
		return nil, ErrInternalError
	}

	// A ticket is not provided when authentication is disabled.
	// Do not set the ticket cookie in that case.
	if ticket != "" {
		res.Ticket = &ticket
	}

	return res, nil
}

func (s *serviceImpl) Monitor(
	ctx context.Context,
	payload *goastorage.MonitorPayload,
	stream goastorage.MonitorServerStream,
) error {
	defer stream.Close()

	// Verify the ticket and update the claims.
	var claims auth.Claims
	if err := s.ticketProvider.Check(ctx, payload.Ticket, &claims); err != nil {
		s.logger.Error(err, "failed to check ticket", "ticket", payload.Ticket)
		return ErrInternalError
	}

	// Subscribe to the event service.
	sub, err := s.evsvc.Subscribe(ctx)
	if err != nil {
		s.logger.Error(err, "failed to subscribe to event service")
		return ErrInternalError
	}
	defer sub.Close()

	// Say hello to be nice.
	event := &goastorage.StoragePingEvent{Message: ref.New("Hello")}
	if err := stream.Send(&goastorage.StorageEvent{Value: event}); err != nil {
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
			event := &goastorage.StoragePingEvent{Message: ref.New("Ping")}
			if err := stream.Send(&goastorage.StorageEvent{Value: event}); err != nil {
				// Consider send errors as client disconnections.
				s.logger.V(1).Info("Failed to send ping event.", "err", err)
				return nil
			}

		case event, ok := <-sub.C():
			if !ok || event == nil {
				return nil
			}

			// Check the event type and the user attributes before sending.
			switch event.Value.(type) {
			case *goastorage.StoragePingEvent:
				// Is this event even sent through this channel?
			case *goastorage.LocationCreatedEvent:
				if !claims.CheckAttributes([]string{auth.StorageLocationsListAttr}) {
					continue
				}
			case *goastorage.AIPCreatedEvent:
				if !claims.CheckAttributes([]string{auth.StorageAIPSListAttr}) {
					continue
				}
			case *goastorage.AIPUpdatedEvent,
				*goastorage.AIPStatusUpdatedEvent,
				*goastorage.AIPLocationUpdatedEvent:
				if !claims.CheckAttributes([]string{auth.StorageAIPSListAttr}) &&
					!claims.CheckAttributes([]string{auth.StorageAIPSReadAttr}) {
					continue
				}
			case *goastorage.AIPWorkflowCreatedEvent,
				*goastorage.AIPWorkflowUpdatedEvent,
				*goastorage.AIPTaskCreatedEvent,
				*goastorage.AIPTaskUpdatedEvent:
				if !claims.CheckAttributes([]string{auth.StorageAIPSWorkflowsListAttr}) {
					continue
				}
			default:
				// Do not send the event if the type is not considered.
				continue
			}

			if err := stream.Send(event); err != nil {
				// Consider send errors as client disconnections.
				s.logger.V(1).Info("Failed to send event.", "err", err)
				return nil
			}
		}
	}
}
