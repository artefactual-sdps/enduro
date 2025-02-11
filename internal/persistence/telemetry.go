package persistence

import (
	"context"
	"fmt"

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

func (w *wrapper) UpdateSIP(ctx context.Context, id int, updater SIPUpdater) (*datatypes.SIP, error) {
	ctx, span := w.tracer.Start(ctx, "UpdateSIP")
	defer span.End()
	span.SetAttributes(attribute.Int("id", id))

	r, err := w.wrapped.UpdateSIP(ctx, id, updater)
	if err != nil {
		telemetry.RecordError(span, err)
		return nil, updateError(err, "UpdateSIP")
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

func (w *wrapper) CreatePreservationAction(ctx context.Context, pa *datatypes.PreservationAction) error {
	ctx, span := w.tracer.Start(ctx, "CreatePreservationAction")
	defer span.End()

	err := w.wrapped.CreatePreservationAction(ctx, pa)
	if err != nil {
		telemetry.RecordError(span, err)
		return updateError(err, "CreatePreservationAction")
	}

	return nil
}

func (w *wrapper) CreatePreservationTask(ctx context.Context, pt *datatypes.PreservationTask) error {
	ctx, span := w.tracer.Start(ctx, "CreatePreservationTask")
	defer span.End()

	err := w.wrapped.CreatePreservationTask(ctx, pt)
	if err != nil {
		telemetry.RecordError(span, err)
		return updateError(err, "CreatePreservationTask")
	}

	return nil
}

func (w *wrapper) UpdatePreservationTask(
	ctx context.Context,
	id int,
	updater PresTaskUpdater,
) (*datatypes.PreservationTask, error) {
	ctx, span := w.tracer.Start(ctx, "UpdatePreservationTask")
	defer span.End()
	span.SetAttributes(attribute.Int("id", id))

	r, err := w.wrapped.UpdatePreservationTask(ctx, id, updater)
	if err != nil {
		telemetry.RecordError(span, err)
		return nil, updateError(err, "UpdatePreservationTask")
	}

	return r, nil
}
