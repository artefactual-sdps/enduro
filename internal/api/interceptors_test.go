package api

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/go-logr/logr"
	"github.com/go-logr/logr/funcr"
	"github.com/google/go-cmp/cmp/cmpopts"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	otelsdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	goa "goa.design/goa/v3/pkg"
	"gotest.tools/v3/assert"

	goaabout "github.com/artefactual-sdps/enduro/internal/api/gen/about"
	goaingest "github.com/artefactual-sdps/enduro/internal/api/gen/ingest"
	goastorage "github.com/artefactual-sdps/enduro/internal/api/gen/storage"
)

func TestServerErrorHandlerLogsAndSanitizesInternalError(t *testing.T) {
	t.Parallel()

	cause := errors.New("database unavailable")
	returned := goastorage.MakeInternalError(fmt.Errorf("load AIP data: %v", cause))
	var logged string
	logger := funcr.New(
		func(_, args string) { logged = args },
		funcr.Options{},
	)
	endpoint := goastorage.WrapShowAipEndpoint(
		func(context.Context, any) (any, error) { return nil, returned },
		newStorageServerInterceptors(logger),
	)

	_, err := endpoint(context.Background(), "payload")

	var serr *goa.ServiceError
	assert.Assert(t, errors.As(err, &serr))
	assert.DeepEqual(t, serr, &goa.ServiceError{
		Name:    "internal_error",
		ID:      returned.ID,
		Message: apiInternalErrorMsg,
		Fault:   true,
	}, cmpopts.IgnoreUnexported(goa.ServiceError{}))
	assert.Assert(t, strings.Contains(logged, "load AIP data: database unavailable"))
	assert.Assert(t, strings.Contains(logged, `"service"="storage"`))
	assert.Assert(t, strings.Contains(logged, `"method"="ShowAip"`))
	assert.Assert(t, strings.Contains(logged, fmt.Sprintf(`"error_id"=%q`, serr.ID)))
}

func TestServerErrorHandlerClassifiesRawError(t *testing.T) {
	t.Parallel()

	cause := errors.New("unexpected failure")
	var logged string
	logger := funcr.New(
		func(_, args string) { logged = args },
		funcr.Options{},
	)
	endpoint := goaabout.WrapAboutEndpoint(
		func(context.Context, any) (any, error) { return nil, cause },
		newAboutServerInterceptors(logger),
	)

	_, err := endpoint(context.Background(), "payload")

	var serr *goa.ServiceError
	assert.Assert(t, errors.As(err, &serr))
	assert.DeepEqual(t, serr, &goa.ServiceError{
		Name:    "internal_error",
		Message: apiInternalErrorMsg,
		Fault:   true,
	},
		cmpopts.IgnoreFields(goa.ServiceError{}, "ID"),
		cmpopts.IgnoreUnexported(goa.ServiceError{}),
	)
	assert.Assert(t, strings.Contains(logged, cause.Error()))
}

func TestServerErrorHandlerPassesThroughOtherFault(t *testing.T) {
	t.Parallel()

	cause := errors.New("unexpected dependency fault")
	returned := goa.NewServiceError(cause, "dependency_fault", false, false, true)
	var logged string
	logger := funcr.New(
		func(_, args string) { logged = args },
		funcr.Options{},
	)
	endpoint := goaabout.WrapAboutEndpoint(
		func(context.Context, any) (any, error) {
			return nil, returned
		},
		newAboutServerInterceptors(logger),
	)

	_, err := endpoint(context.Background(), "payload")

	assert.DeepEqual(t, err, returned, cmpopts.IgnoreUnexported(goa.ServiceError{}))
	assert.Equal(t, logged, "")
}

func TestServerErrorHandlerPassesThroughDomainError(t *testing.T) {
	t.Parallel()

	domainErr := &goastorage.AIPNotFound{Message: "AIP not found"}
	var logged string
	logger := funcr.New(
		func(_, args string) { logged = args },
		funcr.Options{},
	)
	endpoint := goastorage.WrapShowAipEndpoint(
		func(context.Context, any) (any, error) { return nil, domainErr },
		newStorageServerInterceptors(logger),
	)

	_, err := endpoint(context.Background(), "payload")

	assert.DeepEqual(t, err, domainErr)
	assert.Equal(t, logged, "")
}

func TestServerErrorHandlerLogsTimeoutOnce(t *testing.T) {
	t.Parallel()

	var logs []string
	logger := funcr.New(
		func(_, args string) { logs = append(logs, args) },
		funcr.Options{},
	)
	interceptors := newAboutServerInterceptors(logger)
	interceptors.operationTimeout = time.Nanosecond
	endpoint := goaabout.WrapAboutEndpoint(
		func(ctx context.Context, _ any) (any, error) {
			<-ctx.Done()
			return nil, ctx.Err()
		},
		interceptors,
	)

	_, err := endpoint(context.Background(), "payload")

	var serr *goa.ServiceError
	assert.Assert(t, errors.As(err, &serr))
	assert.DeepEqual(t, serr, &goa.ServiceError{
		Name:    "internal_error",
		Message: apiInternalErrorMsg,
		Timeout: true,
		Fault:   true,
	},
		cmpopts.IgnoreFields(goa.ServiceError{}, "ID"),
		cmpopts.IgnoreUnexported(goa.ServiceError{}),
	)
	assert.Equal(t, len(logs), 1)
	assert.Assert(t, strings.Contains(logs[0], apiOperationTimeoutMsg))
	assert.Assert(t, strings.Contains(logs[0], fmt.Sprintf(`"error_id"=%q`, serr.ID)))
}

func TestOperationTimeoutMapsDeadlineExceeded(t *testing.T) {
	t.Parallel()

	endpoint := goastorage.WrapDownloadAipRequestEndpoint(
		func(ctx context.Context, _ any) (any, error) {
			<-ctx.Done()
			return nil, ctx.Err()
		},
		newTestStorageServerInterceptors(time.Nanosecond),
	)

	_, err := endpoint(context.Background(), "payload")
	var serr *goa.ServiceError
	assert.Assert(t, errors.As(err, &serr))
	assert.DeepEqual(t, serr, &goa.ServiceError{
		Name:    "internal_error",
		Message: apiInternalErrorMsg,
		Timeout: true,
		Fault:   true,
	},
		cmpopts.IgnoreFields(goa.ServiceError{}, "ID"),
		cmpopts.IgnoreUnexported(goa.ServiceError{}),
	)
}

func TestOperationTimeoutStorageTransferRules(t *testing.T) {
	t.Parallel()

	for _, tt := range []struct {
		name         string
		wrap         func(goa.Endpoint, *storageServerInterceptors) goa.Endpoint
		wantDeadline bool
	}{
		{
			name: "download ticket request gets deadline",
			wrap: func(next goa.Endpoint, interceptors *storageServerInterceptors) goa.Endpoint {
				return goastorage.WrapDownloadAipRequestEndpoint(next, interceptors)
			},
			wantDeadline: true,
		},
		{
			name: "deletion report ticket request gets deadline",
			wrap: func(next goa.Endpoint, interceptors *storageServerInterceptors) goa.Endpoint {
				return goastorage.WrapAipDeletionReportRequestEndpoint(next, interceptors)
			},
			wantDeadline: true,
		},
		{
			name: "download body skips deadline",
			wrap: func(next goa.Endpoint, interceptors *storageServerInterceptors) goa.Endpoint {
				return goastorage.WrapDownloadAipEndpoint(next, interceptors)
			},
			wantDeadline: false,
		},
		{
			name: "deletion report body skips deadline",
			wrap: func(next goa.Endpoint, interceptors *storageServerInterceptors) goa.Endpoint {
				return goastorage.WrapAipDeletionReportEndpoint(next, interceptors)
			},
			wantDeadline: false,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			endpoint := tt.wrap(
				deadlineEndpoint(t, tt.wantDeadline),
				newStorageServerInterceptors(logr.Discard()),
			)

			res, err := endpoint(context.Background(), "payload")
			assert.NilError(t, err)
			assert.Equal(t, res, "ok")
		})
	}
}

func TestOperationTimeoutUsesDeadlineForAboutMethods(t *testing.T) {
	t.Parallel()

	endpoint := goaabout.WrapAboutEndpoint(
		deadlineEndpoint(t, true),
		newAboutServerInterceptors(logr.Discard()),
	)

	res, err := endpoint(context.Background(), "payload")
	assert.NilError(t, err)
	assert.Equal(t, res, "ok")
}

func TestOperationTimeoutSkipsIngestUploadAndDownload(t *testing.T) {
	t.Parallel()

	interceptors := newIngestServerInterceptors(logr.Discard())

	for _, endpoint := range []goa.Endpoint{
		goaingest.WrapUploadSipEndpoint(deadlineEndpoint(t, false), interceptors),
		goaingest.WrapDownloadSipEndpoint(deadlineEndpoint(t, false), interceptors),
	} {
		res, err := endpoint(context.Background(), "payload")
		assert.NilError(t, err)
		assert.Equal(t, res, "ok")
	}
}

func TestOperationTimeoutAnnotatesSpan(t *testing.T) {
	t.Parallel()

	recorder, ctx, endSpan := newTestSpan(t)
	endpoint := goastorage.WrapShowAipEndpoint(
		func(ctx context.Context, _ any) (any, error) {
			return "ok", nil
		},
		newStorageServerInterceptors(logr.Discard()),
	)

	res, err := endpoint(ctx, "payload")
	endSpan()

	assert.NilError(t, err)
	assert.Equal(t, res, "ok")

	spans := recorder.Ended()
	assert.Equal(t, len(spans), 1)
	assert.Equal(t, spanAttribute(t, spans[0].Attributes(), "goa.service").AsString(), "storage")
	assert.Equal(t, spanAttribute(t, spans[0].Attributes(), "goa.method").AsString(), "ShowAip")
	assert.Equal(t,
		spanAttribute(t, spans[0].Attributes(), "api.operation.timeout_ms").AsInt64(),
		defaultAPIOperationTimeout.Milliseconds(),
	)
	assert.Equal(t, spans[0].Status().Code, codes.Unset)
}

func TestOperationTimeoutRecordsSpanError(t *testing.T) {
	t.Parallel()

	recorder, ctx, endSpan := newTestSpan(t)
	endpoint := goastorage.WrapShowAipEndpoint(
		func(ctx context.Context, _ any) (any, error) {
			<-ctx.Done()
			return nil, ctx.Err()
		},
		newTestStorageServerInterceptors(time.Nanosecond),
	)

	_, err := endpoint(ctx, "payload")
	endSpan()

	var serr *goa.ServiceError
	assert.Assert(t, errors.As(err, &serr))
	assert.DeepEqual(t, serr, &goa.ServiceError{
		Name:    "internal_error",
		Message: apiInternalErrorMsg,
		Timeout: true,
		Fault:   true,
	},
		cmpopts.IgnoreFields(goa.ServiceError{}, "ID"),
		cmpopts.IgnoreUnexported(goa.ServiceError{}),
	)

	spans := recorder.Ended()
	assert.Equal(t, len(spans), 1)
	assert.Equal(t, spans[0].Status().Code, codes.Error)
	assert.Assert(t, spanHasEvent(spans[0].Events(), "api.operation.timeout"))
}

func newTestSpan(t *testing.T) (*tracetest.SpanRecorder, context.Context, func()) {
	t.Helper()

	recorder := tracetest.NewSpanRecorder()
	tp := otelsdktrace.NewTracerProvider(otelsdktrace.WithSpanProcessor(recorder))
	t.Cleanup(func() { _ = tp.Shutdown(t.Context()) })

	ctx, span := tp.Tracer("test").Start(context.Background(), "request")

	return recorder, ctx, func() { span.End() }
}

func deadlineEndpoint(t *testing.T, wantDeadline bool) goa.Endpoint {
	t.Helper()

	return func(ctx context.Context, _ any) (any, error) {
		deadline, ok := ctx.Deadline()
		assert.Equal(t, ok, wantDeadline)
		if wantDeadline {
			assert.Assert(t, time.Until(deadline) > 0)
			assert.Assert(t, time.Until(deadline) <= defaultAPIOperationTimeout)
		}
		return "ok", nil
	}
}

func newTestStorageServerInterceptors(timeout time.Duration) *storageServerInterceptors {
	i := newServerInterceptors(logr.Discard())
	i.operationTimeout = timeout

	return &storageServerInterceptors{serverInterceptors: i}
}

func spanAttribute(
	t *testing.T,
	attrs []attribute.KeyValue,
	key string,
) attribute.Value {
	t.Helper()

	for _, attr := range attrs {
		if string(attr.Key) == key {
			return attr.Value
		}
	}

	t.Fatalf("span attribute %q not found", key)
	return attribute.Value{}
}

func spanHasEvent(events []otelsdktrace.Event, name string) bool {
	for _, event := range events {
		if event.Name == name {
			return true
		}
	}

	return false
}
