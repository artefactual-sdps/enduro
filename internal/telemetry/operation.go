package telemetry

import (
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

func SetGoaOperationAttributes(
	span trace.Span,
	service string,
	method string,
	timeout time.Duration,
	hasTimeout bool,
) {
	attrs := []attribute.KeyValue{
		attribute.String("goa.service", service),
		attribute.String("goa.method", method),
	}
	if hasTimeout {
		attrs = append(attrs, operationTimeoutAttr(timeout))
	}

	span.SetAttributes(attrs...)
}

func SetOperationDuration(span trace.Span, duration time.Duration) {
	span.SetAttributes(operationDurationAttr(duration))
}

func RecordOperationTimeout(
	span trace.Span,
	err error,
	msg string,
	elapsed time.Duration,
	timeout time.Duration,
) {
	span.RecordError(err)
	span.SetStatus(codes.Error, msg)
	span.AddEvent("api.operation.timeout", trace.WithAttributes(
		operationDurationAttr(elapsed),
		operationTimeoutAttr(timeout),
	))
}

func RecordSlowOperation(span trace.Span, elapsed, threshold time.Duration) {
	span.AddEvent("api.operation.slow", trace.WithAttributes(
		operationDurationAttr(elapsed),
		operationSlowThresholdAttr(threshold),
	))
}

func operationDurationAttr(duration time.Duration) attribute.KeyValue {
	return attribute.Int64("api.operation.duration_ms", duration.Milliseconds())
}

func operationTimeoutAttr(timeout time.Duration) attribute.KeyValue {
	return attribute.Int64("api.operation.timeout_ms", timeout.Milliseconds())
}

func operationSlowThresholdAttr(threshold time.Duration) attribute.KeyValue {
	return attribute.Int64("api.operation.slow_threshold_ms", threshold.Milliseconds())
}
