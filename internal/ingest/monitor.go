package ingest

import (
	"context"
	"errors"
	"time"

	"go.artefactual.dev/tools/ref"

	"github.com/artefactual-sdps/enduro/internal/api/auth"
	goaingest "github.com/artefactual-sdps/enduro/internal/api/gen/ingest"
)

func (w *goaWrapper) MonitorRequest(
	ctx context.Context,
	payload *goaingest.MonitorRequestPayload,
) (*goaingest.MonitorRequestResult, error) {
	res := &goaingest.MonitorRequestResult{}

	ticket, err := w.ticketProvider.Request(ctx, auth.UserClaimsFromContext(ctx))
	if err != nil {
		w.logger.Error(err, "failed to request ticket")
		return nil, goaingest.MakeNotAvailable(errors.New("cannot perform operation"))
	}

	// A ticket is not provided when authentication is disabled.
	// Do not set the ticket cookie in that case.
	if ticket != "" {
		res.Ticket = &ticket
	}

	return res, nil
}

// Monitor ingest activity. It implements goaingest.Service.
func (w *goaWrapper) Monitor(
	ctx context.Context,
	payload *goaingest.MonitorPayload,
	stream goaingest.MonitorServerStream,
) error {
	defer stream.Close()

	// Verify the ticket and update the claims.
	var claims auth.Claims
	if err := w.ticketProvider.Check(ctx, payload.Ticket, &claims); err != nil {
		w.logger.V(1).Info("failed to check ticket", "err", err)
		return goaingest.MakeNotAvailable(errors.New("cannot perform operation"))
	}

	// Subscribe to the event service.
	sub, err := w.evsvc.Subscribe(ctx)
	if err != nil {
		return err
	}
	defer sub.Close()

	// Say hello to be nice.
	event := &goaingest.IngestPingEvent{Message: ref.New("Hello")}
	if err := stream.Send(&goaingest.IngestEvent{IngestValue: event}); err != nil {
		return err
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
			event := &goaingest.IngestPingEvent{Message: ref.New("Ping")}
			if err := stream.Send(&goaingest.IngestEvent{IngestValue: event}); err != nil {
				return nil
			}

		case event, ok := <-sub.C():
			if !ok || event == nil {
				return nil
			}

			// Check the event type and the user attributes before sending.
			switch event.IngestValue.(type) {
			case *goaingest.IngestPingEvent:
				// Is this event even sent through this channel?
			case *goaingest.SIPCreatedEvent:
				if !claims.CheckAttributes([]string{auth.IngestSIPSListAttr}) {
					continue
				}
			case *goaingest.SIPUpdatedEvent, *goaingest.SIPStatusUpdatedEvent:
				if !claims.CheckAttributes([]string{auth.IngestSIPSListAttr}) &&
					!claims.CheckAttributes([]string{auth.IngestSIPSReadAttr}) {
					continue
				}
			case *goaingest.SIPWorkflowCreatedEvent,
				*goaingest.SIPWorkflowUpdatedEvent,
				*goaingest.SIPTaskCreatedEvent,
				*goaingest.SIPTaskUpdatedEvent:
				if !claims.CheckAttributes([]string{auth.IngestSIPSWorkflowsListAttr}) {
					continue
				}
			default:
				// Do not send the event if the type is not considered.
				continue
			}

			if err := stream.Send(event); err != nil {
				return err
			}
		}
	}
}
