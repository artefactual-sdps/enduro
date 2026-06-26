package api

import (
	"context"
	"errors"
	"time"

	"github.com/go-logr/logr"
	"go.opentelemetry.io/otel/trace"
	goa "goa.design/goa/v3/pkg"

	goaabout "github.com/artefactual-sdps/enduro/internal/api/gen/about"
	goaingest "github.com/artefactual-sdps/enduro/internal/api/gen/ingest"
	goastorage "github.com/artefactual-sdps/enduro/internal/api/gen/storage"
	"github.com/artefactual-sdps/enduro/internal/telemetry"
)

const (
	defaultAPIOperationTimeout = 5 * time.Second
	apiOperationSlowThreshold  = 2 * time.Second
	apiOperationTimeoutMsg     = "API operation timed out"
)

type operationInfo interface {
	Service() string
	Method() string
	CallType() goa.InterceptorCallType
	RawPayload() any
}

type operationKey struct {
	service string
	method  string
}

type operationRule struct {
	timeout       time.Duration
	slowThreshold time.Duration
	skipTimeout   bool
	skipLogging   bool
}

// operationInterceptors is shared by all Goa services so service-level
// operation budgets, logging, and span annotations stay consistent. The rule
// map keeps service/method exceptions explicit because Goa reports some
// body-returning endpoints as unary.
type operationInterceptors struct {
	logger           logr.Logger
	operationTimeout time.Duration
	slowThreshold    time.Duration
	rules            map[operationKey]operationRule
}

func newOperationInterceptors(logger logr.Logger) *operationInterceptors {
	return &operationInterceptors{
		logger:           logger,
		operationTimeout: defaultAPIOperationTimeout,
		slowThreshold:    apiOperationSlowThreshold,
		rules: map[operationKey]operationRule{
			{service: "ingest", method: "Monitor"}: {
				skipTimeout: true,
				skipLogging: true,
			},
			{service: "ingest", method: "UploadSip"}: {
				skipTimeout: true,
			},
			{service: "ingest", method: "DownloadSip"}: {
				skipTimeout: true,
			},
			{service: "storage", method: "Monitor"}: {
				skipTimeout: true,
				skipLogging: true,
			},
			{service: "storage", method: "DownloadAip"}: {
				skipTimeout: true,
			},
			{service: "storage", method: "AipDeletionReport"}: {
				skipTimeout: true,
			},
		},
	}
}

type aboutServerInterceptors struct {
	*operationInterceptors
}

func newAboutServerInterceptors(logger logr.Logger) *aboutServerInterceptors {
	return &aboutServerInterceptors{operationInterceptors: newOperationInterceptors(logger)}
}

func (i *aboutServerInterceptors) OperationTimeout(
	ctx context.Context,
	info *goaabout.OperationTimeoutInfo,
	next goa.Endpoint,
) (any, error) {
	return i.handle(ctx, info, next)
}

type ingestServerInterceptors struct {
	*operationInterceptors
}

func newIngestServerInterceptors(logger logr.Logger) *ingestServerInterceptors {
	return &ingestServerInterceptors{operationInterceptors: newOperationInterceptors(logger)}
}

func (i *ingestServerInterceptors) OperationTimeout(
	ctx context.Context,
	info *goaingest.OperationTimeoutInfo,
	next goa.Endpoint,
) (any, error) {
	return i.handle(ctx, info, next)
}

type storageServerInterceptors struct {
	*operationInterceptors
}

func newStorageServerInterceptors(logger logr.Logger) *storageServerInterceptors {
	return &storageServerInterceptors{operationInterceptors: newOperationInterceptors(logger)}
}

func (i *storageServerInterceptors) OperationTimeout(
	ctx context.Context,
	info *goastorage.OperationTimeoutInfo,
	next goa.Endpoint,
) (any, error) {
	return i.handle(ctx, info, next)
}

func (i *operationInterceptors) handle(
	ctx context.Context,
	info operationInfo,
	next goa.Endpoint,
) (any, error) {
	rule := i.rule(info)
	if info.CallType() != goa.InterceptorUnary || rule.skipLogging {
		return next(ctx, info.RawPayload())
	}

	parentCtx := ctx
	timeout, hasTimeout := rule.timeout, !rule.skipTimeout && rule.timeout > 0
	span := trace.SpanFromContext(ctx)
	telemetry.SetGoaOperationAttributes(span, info.Service(), info.Method(), timeout, hasTimeout)
	if hasTimeout {
		// Apply a default budget to normal API methods. Goa reports
		// body-returning endpoints as unary, so methods that continue work
		// after the endpoint returns must be excluded by rule.
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, timeout)
		defer cancel()
	}

	start := time.Now()
	res, err := next(ctx, info.RawPayload())
	elapsed := time.Since(start)
	telemetry.SetOperationDuration(span, elapsed)

	if hasTimeout &&
		errors.Is(ctx.Err(), context.DeadlineExceeded) &&
		!errors.Is(parentCtx.Err(), context.DeadlineExceeded) {
		// Only translate deadlines introduced here. Parent request cancellation
		// should keep its original error semantics.
		logErr := err
		if logErr == nil {
			logErr = ctx.Err()
		}
		telemetry.RecordOperationTimeout(
			span,
			logErr,
			apiOperationTimeoutMsg,
			elapsed,
			timeout,
		)
		i.logger.Error(
			logErr,
			"API operation timed out.",
			"service", info.Service(),
			"method", info.Method(),
			"duration", elapsed,
			"timeout", timeout,
		)

		return nil, goa.NewServiceError(
			errors.New(apiOperationTimeoutMsg),
			"internal_error",
			true,
			false,
			true,
		)
	}

	if elapsed >= rule.slowThreshold {
		telemetry.RecordSlowOperation(span, elapsed, rule.slowThreshold)
		i.logger.V(1).Info(
			"API operation completed slowly.",
			"service", info.Service(),
			"method", info.Method(),
			"duration", elapsed,
			"threshold", rule.slowThreshold,
		)
	}

	return res, err
}

func (i *operationInterceptors) rule(info operationInfo) operationRule {
	rule := operationRule{
		timeout:       i.operationTimeout,
		slowThreshold: i.slowThreshold,
	}

	if override, ok := i.rules[operationKey{
		service: info.Service(),
		method:  info.Method(),
	}]; ok {
		if override.timeout != 0 {
			rule.timeout = override.timeout
		}
		if override.slowThreshold != 0 {
			rule.slowThreshold = override.slowThreshold
		}
		rule.skipTimeout = override.skipTimeout
		rule.skipLogging = override.skipLogging
	}

	return rule
}
