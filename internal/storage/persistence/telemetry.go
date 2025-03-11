package persistence

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel/trace"

	goastorage "github.com/artefactual-sdps/enduro/internal/api/gen/storage"
	"github.com/artefactual-sdps/enduro/internal/storage/enums"
	"github.com/artefactual-sdps/enduro/internal/storage/types"
	"github.com/artefactual-sdps/enduro/internal/telemetry"
)

type wrapper struct {
	wrapped Storage
	tracer  trace.Tracer
}

var _ Storage = (*wrapper)(nil)

// WithTelemetry enriches Storage by adding instrumentation and context.
func WithTelemetry(wrapped Storage, tracer trace.Tracer) *wrapper {
	return &wrapper{wrapped, tracer}
}

func updateError(err error, name string) error {
	if err == nil {
		return nil
	}

	return fmt.Errorf("%s: %w", name, err)
}

func (w *wrapper) CreateAIP(ctx context.Context, aip *goastorage.AIP) (*goastorage.AIP, error) {
	ctx, span := w.tracer.Start(ctx, "CreateAIP")
	defer span.End()

	r, err := w.wrapped.CreateAIP(ctx, aip)
	if err != nil {
		telemetry.RecordError(span, err)
		return nil, updateError(err, "CreateAIP")
	}

	return r, nil
}

func (w *wrapper) ListAIPs(ctx context.Context, payload *goastorage.ListAipsPayload) (*goastorage.AIPs, error) {
	ctx, span := w.tracer.Start(ctx, "ListAIPs")
	defer span.End()

	r, err := w.wrapped.ListAIPs(ctx, payload)
	if err != nil {
		telemetry.RecordError(span, err)
		return nil, updateError(err, "ListAIPs")
	}

	return r, nil
}

func (w *wrapper) ReadAIP(ctx context.Context, aipID uuid.UUID) (*goastorage.AIP, error) {
	ctx, span := w.tracer.Start(ctx, "ReadAIP")
	defer span.End()

	r, err := w.wrapped.ReadAIP(ctx, aipID)
	if err != nil {
		telemetry.RecordError(span, err)
		return nil, updateError(err, "ReadAIP")
	}

	return r, nil
}

func (w *wrapper) UpdateAIPStatus(ctx context.Context, aipID uuid.UUID, status enums.AIPStatus) error {
	ctx, span := w.tracer.Start(ctx, "UpdateAIPStatus")
	defer span.End()

	err := w.wrapped.UpdateAIPStatus(ctx, aipID, status)
	if err != nil {
		telemetry.RecordError(span, err)
		return updateError(err, "UpdateAIPStatus")
	}

	return nil
}

func (w *wrapper) UpdateAIPLocationID(ctx context.Context, aipID, locationID uuid.UUID) error {
	ctx, span := w.tracer.Start(ctx, "UpdateAIPLocationID")
	defer span.End()

	err := w.wrapped.UpdateAIPLocationID(ctx, aipID, locationID)
	if err != nil {
		telemetry.RecordError(span, err)
		return updateError(err, "UpdateAIPLocationID")
	}

	return nil
}

func (w *wrapper) CreateLocation(
	ctx context.Context,
	location *goastorage.Location,
	config *types.LocationConfig,
) (*goastorage.Location, error) {
	ctx, span := w.tracer.Start(ctx, "CreateLocation")
	defer span.End()

	r, err := w.wrapped.CreateLocation(ctx, location, config)
	if err != nil {
		telemetry.RecordError(span, err)
		return nil, updateError(err, "CreateLocation")
	}

	return r, nil
}

func (w *wrapper) ListLocations(ctx context.Context) (goastorage.LocationCollection, error) {
	ctx, span := w.tracer.Start(ctx, "ListLocations")
	defer span.End()

	r, err := w.wrapped.ListLocations(ctx)
	if err != nil {
		telemetry.RecordError(span, err)
		return nil, updateError(err, "ListLocations")
	}

	return r, nil
}

func (w *wrapper) ReadLocation(ctx context.Context, locationID uuid.UUID) (*goastorage.Location, error) {
	ctx, span := w.tracer.Start(ctx, "ReadLocation")
	defer span.End()

	r, err := w.wrapped.ReadLocation(ctx, locationID)
	if err != nil {
		telemetry.RecordError(span, err)
		return nil, updateError(err, "ReadLocation")
	}

	return r, nil
}

func (w *wrapper) LocationAIPs(ctx context.Context, locationID uuid.UUID) (goastorage.AIPCollection, error) {
	ctx, span := w.tracer.Start(ctx, "LocationAIPs")
	defer span.End()

	r, err := w.wrapped.LocationAIPs(ctx, locationID)
	if err != nil {
		telemetry.RecordError(span, err)
		return nil, updateError(err, "LocationAIPs")
	}

	return r, nil
}
