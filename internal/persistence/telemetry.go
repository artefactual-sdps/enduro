package persistence

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	"github.com/artefactual-sdps/enduro/internal/datatypes"
	"github.com/artefactual-sdps/enduro/internal/telemetry"
)

type wrapper struct {
	wrapped Service
	tracer  trace.Tracer
}

var _ Service = (*wrapper)(nil)

// WithTelemetry enriches Service by adding instrumentation and context.
func WithTelemetry(wrapped Service, tracer trace.Tracer) *wrapper {
	return &wrapper{wrapped, tracer}
}

func updateError(err error, name string) error {
	if err == nil {
		return nil
	}

	return fmt.Errorf("%s: %w", name, err)
}

func (w *wrapper) CreateSIP(ctx context.Context, p *datatypes.SIP) error {
	ctx, span := w.tracer.Start(ctx, "CreateSIP")
	defer span.End()

	err := w.wrapped.CreateSIP(ctx, p)
	if err != nil {
		telemetry.RecordError(span, err)
		return updateError(err, "CreateSIP")
	}

	return nil
}

func (w *wrapper) UpdateSIP(ctx context.Context, id uuid.UUID, updater SIPUpdater) (*datatypes.SIP, error) {
	ctx, span := w.tracer.Start(ctx, "UpdateSIP")
	defer span.End()
	span.SetAttributes(attribute.String("id", id.String()))

	r, err := w.wrapped.UpdateSIP(ctx, id, updater)
	if err != nil {
		telemetry.RecordError(span, err)
		return nil, updateError(err, "UpdateSIP")
	}

	return r, nil
}

func (w *wrapper) DeleteSIP(ctx context.Context, id int) error {
	ctx, span := w.tracer.Start(ctx, "DeleteSIP")
	defer span.End()
	span.SetAttributes(attribute.Int("id", id))

	err := w.wrapped.DeleteSIP(ctx, id)
	if err != nil {
		telemetry.RecordError(span, err)
		return updateError(err, "DeleteSIP")
	}

	return nil
}

func (w *wrapper) ReadSIP(ctx context.Context, id uuid.UUID) (*datatypes.SIP, error) {
	ctx, span := w.tracer.Start(ctx, "ReadSIP")
	defer span.End()
	span.SetAttributes(attribute.String("id", id.String()))

	r, err := w.wrapped.ReadSIP(ctx, id)
	if err != nil {
		telemetry.RecordError(span, err)
		return nil, updateError(err, "ReadSIP")
	}

	return r, nil
}

func (w *wrapper) ListSIPs(ctx context.Context, f *SIPFilter) ([]*datatypes.SIP, *Page, error) {
	ctx, span := w.tracer.Start(ctx, "ListSIPs")
	defer span.End()

	r, pg, err := w.wrapped.ListSIPs(ctx, f)
	if err != nil {
		telemetry.RecordError(span, err)
		return nil, nil, updateError(err, "ListSIPs")
	}

	return r, pg, nil
}

func (w *wrapper) CreateWorkflow(ctx context.Context, workflow *datatypes.Workflow) error {
	ctx, span := w.tracer.Start(ctx, "CreateWorkflow")
	defer span.End()

	err := w.wrapped.CreateWorkflow(ctx, workflow)
	if err != nil {
		telemetry.RecordError(span, err)
		return updateError(err, "CreateWorkflow")
	}

	return nil
}

func (w *wrapper) CreateTask(ctx context.Context, task *datatypes.Task) error {
	ctx, span := w.tracer.Start(ctx, "CreateTask")
	defer span.End()

	err := w.wrapped.CreateTask(ctx, task)
	if err != nil {
		telemetry.RecordError(span, err)
		return updateError(err, "CreateTask")
	}

	return nil
}

func (w *wrapper) UpdateTask(
	ctx context.Context,
	id int,
	updater TaskUpdater,
) (*datatypes.Task, error) {
	ctx, span := w.tracer.Start(ctx, "UpdateTask")
	defer span.End()
	span.SetAttributes(attribute.Int("id", id))

	r, err := w.wrapped.UpdateTask(ctx, id, updater)
	if err != nil {
		telemetry.RecordError(span, err)
		return nil, updateError(err, "UpdateTask")
	}

	return r, nil
}

func (w *wrapper) CreateUser(ctx context.Context, user *datatypes.User) error {
	ctx, span := w.tracer.Start(ctx, "CreateUser")
	defer span.End()

	err := w.wrapped.CreateUser(ctx, user)
	if err != nil {
		telemetry.RecordError(span, err)
		return updateError(err, "CreateUser")
	}

	return nil
}

func (w *wrapper) ReadUser(ctx context.Context, id uuid.UUID) (*datatypes.User, error) {
	ctx, span := w.tracer.Start(ctx, "ReadUser")
	defer span.End()
	span.SetAttributes(attribute.String("id", id.String()))

	u, err := w.wrapped.ReadUser(ctx, id)
	if err != nil {
		telemetry.RecordError(span, err)
		return nil, updateError(err, "ReadUser")
	}

	return u, nil
}

func (w *wrapper) ReadOIDCUser(ctx context.Context, iss, sub string) (*datatypes.User, error) {
	ctx, span := w.tracer.Start(ctx, "ReadOIDCUser")
	defer span.End()

	u, err := w.wrapped.ReadOIDCUser(ctx, iss, sub)
	if err != nil {
		telemetry.RecordError(span, err)
		return nil, updateError(err, "ReadOIDCUser")
	}

	return u, nil
}

func (w *wrapper) ListUsers(ctx context.Context, f *UserFilter) ([]*datatypes.User, *Page, error) {
	ctx, span := w.tracer.Start(ctx, "ListUsers")
	defer span.End()

	r, p, err := w.wrapped.ListUsers(ctx, f)
	if err != nil {
		telemetry.RecordError(span, err)
		return nil, nil, updateError(err, "ListUsers")
	}

	return r, p, nil
}
