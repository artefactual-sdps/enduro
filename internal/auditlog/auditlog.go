package auditlog

import (
	"context"
	"fmt"
	"io"
	"log/slog"

	"github.com/artefactual-sdps/enduro/internal/event"
)

type EventHandler[T any] func(T) *Event

type Auditlog[T any] struct {
	logger  *slog.Logger
	w       io.WriteCloser
	sub     event.Subscription[T]
	handler EventHandler[T]
	stopCh  chan struct{}
}

// New creates a new Auditlog instance with the provided logger and event
// handler.  It is the caller's responsibility to close any writers used by the
// logger when done logging.
func New[T any](l *slog.Logger, h EventHandler[T]) *Auditlog[T] {
	return &Auditlog[T]{
		logger:  l,
		handler: h,
		stopCh:  make(chan struct{}),
	}
}

// Listen subscribes to the provided event service and starts listening for
// events to log.  If the subscription fails, an error is returned immediately.
// When done logging `Close()` should be called to stop the listener and close
// any open resources.
func (a *Auditlog[T]) Listen(ctx context.Context, svc event.Service[T]) error {
	sub, err := svc.Subscribe(ctx)
	if err != nil {
		return fmt.Errorf("audit log: subscribe: %w", err)
	}
	a.sub = sub

	go func() {
		for {
			select {
			case <-a.stopCh:
				return
			case r, ok := <-a.sub.C():
				if !ok {
					return
				}

				ev := a.handler(r)
				if ev != nil {
					a.logger.Log(ctx, ev.Level, ev.Msg, ev.Args()...)
				}
			}
		}
	}()

	return nil
}

// Close stops the audit log listener and closes any open resources.
func (a *Auditlog[T]) Close() error {
	close(a.stopCh)
	a.sub.Close()

	// Close the log writer if we control it.
	if a.w != nil {
		a.w.Close()
	}

	return nil
}
