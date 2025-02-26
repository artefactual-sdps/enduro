package ingest_test

import (
	"context"
	"io"
	"strings"
	"testing"

	"go.artefactual.dev/tools/bucket"
	goa "goa.design/goa/v3/pkg"
	"gocloud.dev/blob/memblob"
	"gotest.tools/v3/assert"

	goaingest "github.com/artefactual-sdps/enduro/internal/api/gen/ingest"
	"github.com/artefactual-sdps/enduro/internal/ingest"
)

func TestConfig(t *testing.T) {
	t.Parallel()

	t.Run("Returns error if config is invalid", func(t *testing.T) {
		t.Parallel()

		c := ingest.UploadConfig{
			Bucket: bucket.Config{
				URL:    "s3blob://my-bucket",
				Bucket: "my-bucket",
				Region: "planet-earth",
			},
		}
		err := c.Validate()
		assert.ErrorContains(t, err, "URL and rest of the [upload.bucket] configuration options are mutually exclusive")
	})

	t.Run("Validates if only URL is provided", func(t *testing.T) {
		t.Parallel()

		c := ingest.UploadConfig{
			Bucket: bucket.Config{
				URL: "s3blob://my-bucket",
			},
		}
		err := c.Validate()
		assert.NilError(t, err)
	})

	t.Run("Validates if only bucket options are provided", func(t *testing.T) {
		t.Parallel()

		c := ingest.UploadConfig{
			Bucket: bucket.Config{
				Bucket: "my-bucket",
				Region: "planet-earth",
			},
		}
		err := c.Validate()
		assert.NilError(t, err)
	})
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

func TestUpload(t *testing.T) {
	t.Parallel()

	t.Run("Writes only the first multipart of the request to the bucket", func(t *testing.T) {
		t.Parallel()

		b := memblob.OpenBucket(nil)
		svc, _ := testSvc(t, b, 102400000)
		ctx := context.Background()
		r := io.NopCloser(strings.NewReader(multipartBody))

		err := svc.Goa().
			UploadSip(ctx, &goaingest.UploadSipPayload{ContentType: "multipart/form-data; boundary=foobar"}, r)
		assert.NilError(t, err)

		data, err := b.ReadAll(ctx, "first.txt")
		assert.NilError(t, err)
		assert.Equal(t, string(data), "first")

		_, err = b.ReadAll(ctx, "second.txt")
		assert.ErrorContains(t, err, `blob (key "second.txt") (code=NotFound): blob not found`)
	})

	t.Run("Returns invalid_media_type if media type cannot be parsed", func(t *testing.T) {
		t.Parallel()

		b := memblob.OpenBucket(nil)
		svc, _ := testSvc(t, b, 102400000)
		ctx := context.Background()
		r := io.NopCloser(strings.NewReader(multipartBody))

		err := svc.Goa().UploadSip(ctx, &goaingest.UploadSipPayload{ContentType: "invalid type"}, r)
		assert.Equal(t, err.(*goa.ServiceError).Name, "invalid_media_type")
		assert.ErrorContains(t, err, "invalid media type")
	})

	t.Run("Returns invalid_multipart_request if request size is bigger than maximum size", func(t *testing.T) {
		t.Parallel()

		b := memblob.OpenBucket(nil)
		svc, _ := testSvc(t, b, 1)
		ctx := context.Background()
		r := io.NopCloser(strings.NewReader(multipartBody))

		err := svc.Goa().
			UploadSip(ctx, &goaingest.UploadSipPayload{ContentType: "multipart/form-data; boundary=foobar"}, r)
		assert.Equal(t, err.(*goa.ServiceError).Name, "invalid_multipart_request")
		assert.ErrorContains(t, err, "invalid multipart request")
	})
}
