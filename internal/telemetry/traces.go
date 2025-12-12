package telemetry

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/go-logr/logr"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	otelsdk_resource "go.opentelemetry.io/otel/sdk/resource"
	otelsdk_trace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.37.0"
	"go.opentelemetry.io/otel/trace"
	"go.opentelemetry.io/otel/trace/noop"
)

// TracerProvider provides Tracers that are used by instrumentation code to
// trace computational workflows.
func TracerProvider(
	ctx context.Context,
	logger logr.Logger,
	cfg Config,
	appName, appVersion string,
) (trace.TracerProvider, ShutdownProvider, error) {
	logger = logger.WithValues("enabled", cfg.Traces.Enabled, "addr", cfg.Traces.Address)
	if !cfg.Traces.Enabled || cfg.Traces.Address == "" {
		logger.V(1).Info("Tracing system is disabled.")
		shutdown := func(context.Context) error { return nil }
		return noop.NewTracerProvider(), shutdown, nil
	}

	if _, _, err := net.SplitHostPort(cfg.Traces.Address); err != nil {
		return nil, nil, fmt.Errorf("TracerProvider: invalid traces address %q: %w", cfg.Traces.Address, err)
	}

	exporter, err := otlptracegrpc.New(ctx,
		otlptracegrpc.WithEndpoint(cfg.Traces.Address),
		otlptracegrpc.WithInsecure(),
		otlptracegrpc.WithRetry(otlptracegrpc.RetryConfig{
			Enabled:         true,
			InitialInterval: time.Second,
			MaxInterval:     30 * time.Second,
			MaxElapsedTime:  0,
		}),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("TracerProvider: exporter init: %v", err)
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

	logger.V(1).Info("Using OTel gRPC tracer provider; exporter will retry on failure.")

	return tp, shutdown, nil
}

func RecordError(span trace.Span, err error) {
	span.SetStatus(codes.Error, "Operation failed.")
	span.RecordError(err)
}
