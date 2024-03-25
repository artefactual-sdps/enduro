package telemetry

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	otelsdk_resource "go.opentelemetry.io/otel/sdk/resource"
	otelsdk_trace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
	"go.opentelemetry.io/otel/trace"
	"go.opentelemetry.io/otel/trace/noop"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// TracerProvider provides Tracers that are used by instrumentation code to
// trace computational workflows.
func TracerProvider(
	ctx context.Context,
	logger logr.Logger,
	cfg Config,
	appName, appVersion string,
) (trace.TracerProvider, ShutdownProvider, error) {
	if !cfg.Traces.Enabled || cfg.Traces.Address == "" {
		logger.V(1).
			Info("Tracing system is disabled.", "enabled", cfg.Traces.Enabled, "addr", cfg.Traces.Address, "sampling-ration", cfg.Traces.SamplingRatio)
		shutdown := func(context.Context) error { return nil }
		return noop.NewTracerProvider(), shutdown, nil
	}

	conn, err := grpc.DialContext(
		ctx,
		cfg.Traces.Address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("TracerProvider: connect to collector: %v", err)
	}

	exporter, err := otlptracegrpc.New(ctx, otlptracegrpc.WithGRPCConn(conn))
	if err != nil {
		return nil, nil, fmt.Errorf("TracerProvider: new exporter: %v", err)
	}

	resource, err := otelsdk_resource.Merge(
		otelsdk_resource.Default(),
		otelsdk_resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName(appName),
			semconv.ServiceVersion(appVersion),
		),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("TracerProvider: new resource: %v", err)
	}

	var ratio float64 = 1
	if cfg.Traces.SamplingRatio != nil {
		ratio = *cfg.Traces.SamplingRatio
	}
	sampler := otelsdk_trace.ParentBased(otelsdk_trace.TraceIDRatioBased(ratio))

	tp := otelsdk_trace.NewTracerProvider(
		otelsdk_trace.WithSampler(sampler),
		otelsdk_trace.WithResource(resource),
		otelsdk_trace.WithBatcher(exporter),
	)
	shutdown := func(context.Context) error { return tp.Shutdown(ctx) }

	otel.SetTextMapPropagator(propagation.TraceContext{})

	logger.V(1).Info("Using OTel gRPC tracer provider.", "addr", cfg.Traces.Address, "ratio", ratio)

	return tp, shutdown, nil
}

func RecordError(span trace.Span, err error) {
	span.SetStatus(codes.Error, "Operation failed.")
	span.RecordError(err)
}
