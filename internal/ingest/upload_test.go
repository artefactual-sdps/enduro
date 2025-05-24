package ingest_test

import (
	"context"
	"fmt"
	"io"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"go.artefactual.dev/tools/mockutil"
	temporalsdk_api_enums "go.temporal.io/api/enums/v1"
	temporalsdk_client "go.temporal.io/sdk/client"
	goa "goa.design/goa/v3/pkg"
	"gocloud.dev/blob/memblob"
	"gotest.tools/v3/assert"

	goaingest "github.com/artefactual-sdps/enduro/internal/api/gen/ingest"
	"github.com/artefactual-sdps/enduro/internal/datatypes"
	"github.com/artefactual-sdps/enduro/internal/enums"
	"github.com/artefactual-sdps/enduro/internal/ingest"
)

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
		svc, psvc, tc := testSvc(t, b, 102400000)
		ctx := context.Background()
		r := io.NopCloser(strings.NewReader(multipartBody))

		// TODO: Fix UUID generation or find how to ignore in Temporal client mock.
		sipUUID := uuid.New()

		psvc.EXPECT().CreateSIP(
			ctx,
			mockutil.Eq(
				&datatypes.SIP{
					UUID:   sipUUID,
					Name:   "first.txt",
					Status: enums.SIPStatusQueued,
				},
				cmpopts.IgnoreFields(datatypes.SIP{}, "UUID"),
			),
		).Return(nil)

		tc.On(
			"ExecuteWorkflow",
			mock.AnythingOfType("*context.timerCtx"),
			temporalsdk_client.StartWorkflowOptions{
				ID:                    fmt.Sprintf("processing-workflow-%s", sipUUID.String()),
				TaskQueue:             "test",
				WorkflowIDReusePolicy: temporalsdk_api_enums.WORKFLOW_ID_REUSE_POLICY_ALLOW_DUPLICATE,
			},
			ingest.ProcessingWorkflowName,
			&ingest.ProcessingWorkflowRequest{
				SIPUUID: sipUUID,
				SIPName: "first.txt",
				Type:    enums.WorkflowTypeCreateAip,
				Key:     fmt.Sprintf("%s%s", ingest.SIPPrefix, sipUUID.String()),
			},
		).Return(nil, nil)

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
		svc, _, _ := testSvc(t, b, 102400000)
		ctx := context.Background()
		r := io.NopCloser(strings.NewReader(multipartBody))

		err := svc.Goa().UploadSip(ctx, &goaingest.UploadSipPayload{ContentType: "invalid type"}, r)
		assert.Equal(t, err.(*goa.ServiceError).Name, "invalid_media_type")
		assert.ErrorContains(t, err, "invalid media type")
	})

	t.Run("Returns invalid_multipart_request if request size is bigger than maximum size", func(t *testing.T) {
		t.Parallel()

		b := memblob.OpenBucket(nil)
		svc, _, _ := testSvc(t, b, 1)
		ctx := context.Background()
		r := io.NopCloser(strings.NewReader(multipartBody))

		err := svc.Goa().
			UploadSip(ctx, &goaingest.UploadSipPayload{ContentType: "multipart/form-data; boundary=foobar"}, r)
		assert.Equal(t, err.(*goa.ServiceError).Name, "invalid_multipart_request")
		assert.ErrorContains(t, err, "invalid multipart request")
	})

	t.Run("Returns invalid_multipart_request if missing file part", func(t *testing.T) {
		t.Parallel()

		b := memblob.OpenBucket(nil)
		svc, _, _ := testSvc(t, b, 102400000)
		ctx := context.Background()
		r := io.NopCloser(strings.NewReader("--foobar--"))

		err := svc.Goa().
			UploadSip(ctx, &goaingest.UploadSipPayload{ContentType: "multipart/form-data; boundary=foobar"}, r)
		assert.Equal(t, err.(*goa.ServiceError).Name, "invalid_multipart_request")
		assert.ErrorContains(t, err, "missing file part in upload")
	})
}
