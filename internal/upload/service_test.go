package upload_test

import (
	"context"
	"io"
	"strings"
	"testing"

	"github.com/go-logr/logr"
	goa "goa.design/goa/v3/pkg"
	_ "gocloud.dev/blob/memblob"
	"gotest.tools/v3/assert"

	"github.com/artefactual-sdps/enduro/internal/api/auth"
	goaupload "github.com/artefactual-sdps/enduro/internal/api/gen/upload"
	"github.com/artefactual-sdps/enduro/internal/ref"
	"github.com/artefactual-sdps/enduro/internal/upload"
)

type setUpAttrs struct {
	logger        *logr.Logger
	config        *upload.Config
	uploadMaxSize *int
	tokenVerifier auth.TokenVerifier
}

func setUpService(t *testing.T, attrs *setUpAttrs) upload.Service {
	t.Helper()

	params := setUpAttrs{
		logger:        ref.New(logr.Discard()),
		config:        ref.New(upload.Config{URL: "mem://my-bucket"}),
		uploadMaxSize: ref.New(upload.UPLOAD_MAX_SIZE),
		tokenVerifier: &auth.OIDCTokenVerifier{},
	}
	if attrs.logger != nil {
		params.logger = attrs.logger
	}
	if attrs.config != nil {
		params.config = attrs.config
	}
	if attrs.uploadMaxSize != nil {
		params.uploadMaxSize = attrs.uploadMaxSize
	}
	if attrs.tokenVerifier != nil {
		params.tokenVerifier = attrs.tokenVerifier
	}

	s, err := upload.NewService(
		*params.logger,
		*params.config,
		*params.uploadMaxSize,
		params.tokenVerifier,
	)
	assert.NilError(t, err)

	return s
}

const multipartBody = `Content-Type: multipart/form-data; boundary="foobar"

--foobar
Content-Disposition: form-data; name="field1"; filename="first.txt"
Content-Type: text/plain

first
--foobar
Content-Disposition: form-data; name="field2"; filename="second.txt"
Content-Type: text/plain

second
--foobar--
`

func TestNewService(t *testing.T) {
	t.Parallel()

	_, err := upload.NewService(
		logr.Discard(),
		upload.Config{URL: "mem://my-bucket"},
		upload.UPLOAD_MAX_SIZE,
		&auth.OIDCTokenVerifier{},
	)
	assert.NilError(t, err)
}

func TestServiceUpload(t *testing.T) {
	t.Parallel()

	t.Run("Writes only the first multipart of the request to the bucket", func(t *testing.T) {
		t.Parallel()

		attrs := &setUpAttrs{}
		svc := setUpService(t, attrs)
		ctx := context.Background()
		r := io.NopCloser(strings.NewReader(multipartBody))

		err := svc.Upload(ctx, &goaupload.UploadPayload{ContentType: "multipart/form-data; boundary=foobar"}, r)
		assert.NilError(t, err)

		b := svc.Bucket()
		data, err := b.ReadAll(ctx, "first.txt")
		assert.NilError(t, err)
		assert.Equal(t, string(data), "first")

		_, err = b.ReadAll(ctx, "second.txt")
		assert.ErrorContains(t, err, `blob (key "second.txt") (code=NotFound): blob not found`)
	})

	t.Run("Returns invalid_media_type if media type cannot be parsed", func(t *testing.T) {
		t.Parallel()

		attrs := &setUpAttrs{
			uploadMaxSize: ref.New(1),
		}
		svc := setUpService(t, attrs)
		ctx := context.Background()
		r := io.NopCloser(strings.NewReader(multipartBody))

		err := svc.Upload(ctx, &goaupload.UploadPayload{ContentType: "invalid type"}, r)
		assert.Equal(t, err.(*goa.ServiceError).Name, "invalid_media_type")
		assert.ErrorContains(t, err, "invalid media type")
	})

	t.Run("Returns invalid_multipart_request if request size is bigger than maximum size", func(t *testing.T) {
		t.Parallel()

		attrs := &setUpAttrs{
			uploadMaxSize: ref.New(1),
		}
		svc := setUpService(t, attrs)
		ctx := context.Background()
		r := io.NopCloser(strings.NewReader(multipartBody))

		err := svc.Upload(ctx, &goaupload.UploadPayload{ContentType: "multipart/form-data; boundary=foobar"}, r)
		assert.Equal(t, err.(*goa.ServiceError).Name, "invalid_multipart_request")
		assert.ErrorContains(t, err, "invalid multipart request")
	})
}
