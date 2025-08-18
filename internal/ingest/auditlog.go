package ingest

import (
	"context"
	"log/slog"

	goaingest "github.com/artefactual-sdps/enduro/internal/api/gen/ingest"
)

// WithAuditLogger configures and starts an audit logger that subscribes to the
// event service and logs published events to the specified logger.
func (svc *ingestImpl) WithAuditLogger(ctx context.Context, logger *slog.Logger) {
	if logger == nil {
		return
	}

	svc.auditLogger = logger
	svc.logAuditEvents(ctx)
}

func (svc *ingestImpl) logAuditEvents(ctx context.Context) {
	sub, err := svc.evsvc.Subscribe(ctx)
	if err != nil {
		svc.logger.Error(err, "audit log: failed to subscribe to event service")
		return
	}

	go func() {
	loop:
		for {
			select {
			case <-ctx.Done():
				if err := ctx.Err(); err != nil {
					svc.logger.Error(err, "audit log: context canceled")
				}
				break loop
			case event, ok := <-sub.C():
				if !ok || event == nil {
					break loop
				}

				if ev, ok := event.IngestValue.(*goaingest.SIPCreatedEvent); ok {
					svc.logSIPCreatedEvent(ev)
				}
			}
		}

		sub.Close()
	}()
}

func (svc *ingestImpl) logSIPCreatedEvent(e *goaingest.SIPCreatedEvent) {
	var userID string
	if e.Item.UploaderUUID != nil {
		userID = e.Item.UploaderUUID.String()
	}

	svc.auditLogger.Info(
		"SIP deposited",
		"type", "SIP.deposit",
		"objectID", e.UUID.String(),
		"userID", userID,
	)
}
