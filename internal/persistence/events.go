package persistence

import (
	"context"

	goapackage "github.com/artefactual-sdps/enduro/internal/api/gen/package_"
	"github.com/artefactual-sdps/enduro/internal/event"
	"github.com/artefactual-sdps/enduro/internal/package_"
)

type eventManager struct {
	evsvc event.EventService
	inner Service
}

var _ Service = (*eventManager)(nil)

// WithEvents decorates a persistence service implementation with event
// publication to evsvc.
func WithEvents(evsvc event.EventService, inner Service) *eventManager {
	return &eventManager{evsvc: evsvc, inner: inner}
}

// CreatePackage creates and persists a new package then publishes a "package
// created" event on success.
func (m *eventManager) CreatePackage(ctx context.Context, pkg *package_.Package) (*package_.Package, error) {
	pkg, err := m.inner.CreatePackage(ctx, pkg)
	if err != nil {
		return nil, err
	}

	// Publish a "package created" event.
	ev := &goapackage.EnduroPackageCreatedEvent{
		ID:   uint(pkg.ID),
		Item: pkg.Goa(),
	}
	event.PublishEvent(ctx, m.evsvc, ev)

	return pkg, nil
}
