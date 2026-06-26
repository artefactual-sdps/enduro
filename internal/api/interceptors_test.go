package api

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-logr/logr"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	otelsdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	goahttp "goa.design/goa/v3/http"
	goa "goa.design/goa/v3/pkg"
	"gotest.tools/v3/assert"

	goaabout "github.com/artefactual-sdps/enduro/internal/api/gen/about"
	storagesvr "github.com/artefactual-sdps/enduro/internal/api/gen/http/storage/server"
	goaingest "github.com/artefactual-sdps/enduro/internal/api/gen/ingest"
	goastorage "github.com/artefactual-sdps/enduro/internal/api/gen/storage"
)

func newTestSpan(t *testing.T) (*tracetest.SpanRecorder, context.Context, func()) {
	t.Helper()

	recorder := tracetest.NewSpanRecorder()
	tp := otelsdktrace.NewTracerProvider(otelsdktrace.WithSpanProcessor(recorder))
	t.Cleanup(func() { _ = tp.Shutdown(t.Context()) })

	ctx, span := tp.Tracer("test").Start(context.Background(), "request")

	return recorder, ctx, func() { span.End() }
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
	assert.Equal(t, serr.Name, "internal_error")
	assert.Equal(t, serr.Message, apiOperationTimeoutMsg)
	assert.Equal(t, serr.Timeout, true)
	assert.Equal(t, serr.Temporary, false)
	assert.Equal(t, serr.Fault, true)
}

func TestOperationTimeoutEncodesAsServerTimeout(t *testing.T) {
	t.Parallel()

	for _, tt := range []struct {
		name   string
		encode func(context.Context, http.ResponseWriter, error) error
	}{
		{
			name: "undeclared internal error",
			encode: storagesvr.EncodeListAipsError(
				goahttp.ResponseEncoder,
				nil,
			),
		},
		{
			name: "declared internal error",
			encode: storagesvr.EncodeDownloadAipRequestError(
				goahttp.ResponseEncoder,
				nil,
			),
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := goa.NewServiceError(
				errors.New(apiOperationTimeoutMsg),
				"internal_error",
				true,
				false,
				true,
			)
			rec := httptest.NewRecorder()

			assert.NilError(t, tt.encode(context.Background(), rec, err))
			assert.Equal(t, rec.Code, http.StatusInternalServerError)

			var body map[string]any
			assert.NilError(t, json.NewDecoder(rec.Body).Decode(&body))
			assert.Equal(t, body["name"], "internal_error")
			assert.Equal(t, body["message"], apiOperationTimeoutMsg)
			assert.Equal(t, body["timeout"], true)
			assert.Equal(t, body["temporary"], false)
			assert.Equal(t, body["fault"], true)
		})
	}
}

func TestOperationTimeoutUsesDeadline(t *testing.T) {
	t.Parallel()

	endpoint := goastorage.WrapDownloadAipRequestEndpoint(
		func(ctx context.Context, _ any) (any, error) {
			deadline, ok := ctx.Deadline()
			assert.Assert(t, ok)
			assert.Assert(t, time.Until(deadline) <= defaultAPIOperationTimeout)
			return "ok", nil
		},
		newStorageServerInterceptors(logr.Discard()),
	)

	res, err := endpoint(context.Background(), "payload")
	assert.NilError(t, err)
	assert.Equal(t, res, "ok")
}

func TestOperationTimeoutUsesDeadlineForNormalMethods(t *testing.T) {
	t.Parallel()

	endpoint := goastorage.WrapShowAipEndpoint(
		func(ctx context.Context, _ any) (any, error) {
			deadline, ok := ctx.Deadline()
			assert.Assert(t, ok)
			assert.Assert(t, time.Until(deadline) <= defaultAPIOperationTimeout)
			return "ok", nil
		},
		newStorageServerInterceptors(logr.Discard()),
	)

	res, err := endpoint(context.Background(), "payload")
	assert.NilError(t, err)
	assert.Equal(t, res, "ok")
}

func TestOperationTimeoutUsesDeadlineForAboutMethods(t *testing.T) {
	t.Parallel()

	endpoint := goaabout.WrapAboutEndpoint(
		func(ctx context.Context, _ any) (any, error) {
			deadline, ok := ctx.Deadline()
			assert.Assert(t, ok)
			assert.Assert(t, time.Until(deadline) <= defaultAPIOperationTimeout)
			return "ok", nil
		},
		newAboutServerInterceptors(logr.Discard()),
	)

	res, err := endpoint(context.Background(), "payload")
	assert.NilError(t, err)
	assert.Equal(t, res, "ok")
}

func TestOperationTimeoutSkipsBodyReturningMethods(t *testing.T) {
	t.Parallel()

	endpoint := goastorage.WrapDownloadAipEndpoint(
		func(ctx context.Context, _ any) (any, error) {
			_, ok := ctx.Deadline()
			assert.Assert(t, !ok)
			return "ok", nil
		},
		newStorageServerInterceptors(logr.Discard()),
	)

	res, err := endpoint(context.Background(), "payload")
	assert.NilError(t, err)
	assert.Equal(t, res, "ok")
}

func TestOperationTimeoutSkipsIngestUploadAndDownload(t *testing.T) {
	t.Parallel()

	interceptors := newIngestServerInterceptors(logr.Discard())

	for _, endpoint := range []goa.Endpoint{
		goaingest.WrapUploadSipEndpoint(noDeadlineEndpoint(t), interceptors),
		goaingest.WrapDownloadSipEndpoint(noDeadlineEndpoint(t), interceptors),
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

func noDeadlineEndpoint(t *testing.T) goa.Endpoint {
	t.Helper()

	return func(ctx context.Context, _ any) (any, error) {
		_, ok := ctx.Deadline()
		assert.Assert(t, !ok)
		return "ok", nil
	}
}

func newTestStorageServerInterceptors(timeout time.Duration) *storageServerInterceptors {
	i := newOperationInterceptors(logr.Discard())
	i.operationTimeout = timeout

	return &storageServerInterceptors{operationInterceptors: i}
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
	assert.Equal(t, serr.Message, apiOperationTimeoutMsg)

	spans := recorder.Ended()
	assert.Equal(t, len(spans), 1)
	assert.Equal(t, spans[0].Status().Code, codes.Error)
	assert.Assert(t, spanHasEvent(spans[0].Events(), "api.operation.timeout"))
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
