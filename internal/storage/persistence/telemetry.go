package persistence

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel/trace"

	goastorage "github.com/artefactual-sdps/enduro/internal/api/gen/storage"
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

func (w *wrapper) CreatePackage(ctx context.Context, pkg *goastorage.Package) (*goastorage.Package, error) {
	ctx, span := w.tracer.Start(ctx, "CreatePackage")
	defer span.End()

	r, err := w.wrapped.CreatePackage(ctx, pkg)
	if err != nil {
		telemetry.RecordError(span, err)
		return nil, updateError(err, "CreatePackage")
	}

	return r, nil
}

func (w *wrapper) ListPackages(ctx context.Context) (goastorage.PackageCollection, error) {
	ctx, span := w.tracer.Start(ctx, "ListPackages")
	defer span.End()

	r, err := w.wrapped.ListPackages(ctx)
	if err != nil {
		telemetry.RecordError(span, err)
		return nil, updateError(err, "ListPackages")
	}

	return r, nil
}

func (w *wrapper) ReadPackage(ctx context.Context, aipID uuid.UUID) (*goastorage.Package, error) {
	ctx, span := w.tracer.Start(ctx, "ReadPackage")
	defer span.End()

	r, err := w.wrapped.ReadPackage(ctx, aipID)
	if err != nil {
		telemetry.RecordError(span, err)
		return nil, updateError(err, "ReadPackage")
	}

	return r, nil
}

func (w *wrapper) UpdatePackageStatus(ctx context.Context, aipID uuid.UUID, status types.PackageStatus) error {
	ctx, span := w.tracer.Start(ctx, "UpdatePackageStatus")
	defer span.End()

	err := w.wrapped.UpdatePackageStatus(ctx, aipID, status)
	if err != nil {
		telemetry.RecordError(span, err)
		return updateError(err, "UpdatePackageStatus")
	}

	return nil
}

func (w *wrapper) UpdatePackageLocationID(ctx context.Context, aipID, locationID uuid.UUID) error {
	ctx, span := w.tracer.Start(ctx, "UpdatePackageLocationID")
	defer span.End()

	err := w.wrapped.UpdatePackageLocationID(ctx, aipID, locationID)
	if err != nil {
		telemetry.RecordError(span, err)
		return updateError(err, "UpdatePackageLocationID")
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

func (w *wrapper) LocationPackages(ctx context.Context, locationID uuid.UUID) (goastorage.PackageCollection, error) {
	ctx, span := w.tracer.Start(ctx, "LocationPackages")
	defer span.End()

	r, err := w.wrapped.LocationPackages(ctx, locationID)
	if err != nil {
		telemetry.RecordError(span, err)
		return nil, updateError(err, "LocationPackages")
	}

	return r, nil
}
