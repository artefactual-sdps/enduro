package ingest_test

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	temporalsdk_api_enums "go.temporal.io/api/enums/v1"
	temporalsdk_client "go.temporal.io/sdk/client"
	temporalsdk_mocks "go.temporal.io/sdk/mocks"
	"gocloud.dev/blob/memblob"
	"gotest.tools/v3/assert"

	goaingest "github.com/artefactual-sdps/enduro/internal/api/gen/ingest"
	"github.com/artefactual-sdps/enduro/internal/datatypes"
	"github.com/artefactual-sdps/enduro/internal/enums"
	"github.com/artefactual-sdps/enduro/internal/ingest"
	persistence_fake "github.com/artefactual-sdps/enduro/internal/persistence/fake"
)

const txtMultipartBody = `Content-Type: multipart/form-data; boundary="foobar"

--foobar
Content-Disposition: form-data; name="field1"; filename="first.txt"
Content-Type: text/plain

first
--foobar--
`

const zipMmultipartBody = `Content-Type: multipart/form-data; boundary="foobar"

--foobar
Content-Disposition: form-data; name="field1"; filename="first.zip"
Content-Type: application/zip

<binary zip data>
--foobar--
`

func TestUpload(t *testing.T) {
	t.Parallel()

	uuid0 := uuid.MustParse("52fdfc07-2182-454f-963f-5f0f9a621d72")
	key := fmt.Sprintf("%sfirst-%s.zip", ingest.SIPPrefix, uuid0.String())

	for _, tt := range []struct {
		name          string
		mock          func(context.Context, *persistence_fake.MockService, *temporalsdk_mocks.Client)
		multipartBody string
		contentType   string
		maxUploadSize int64
		want          *goaingest.UploadSipResult
		wantErr       string
	}{
		{
			name:          "Returns invalid_media_type if media type cannot be parsed",
			multipartBody: zipMmultipartBody,
			contentType:   "invalid type",
			maxUploadSize: 102400000,
			wantErr:       "invalid media type",
		},
		{
			name:          "Returns invalid_multipart_request if request size is bigger than maximum size",
			multipartBody: zipMmultipartBody,
			contentType:   "multipart/form-data; boundary=foobar",
			maxUploadSize: 0,
			wantErr:       "invalid multipart request",
		},
		{
			name:          "Returns invalid_multipart_request if missing file part",
			multipartBody: "--foobar--",
			contentType:   "multipart/form-data; boundary=foobar",
			maxUploadSize: 102400000,
			wantErr:       "missing file part in upload",
		},
		{
			name:          "Returns invalid_multipart_request if unable to identify format",
			multipartBody: txtMultipartBody,
			contentType:   "multipart/form-data; boundary=foobar",
			maxUploadSize: 102400000,
			wantErr:       "unable to identify format",
		},
		{
			name: "Returns persistence error",
			mock: func(ctx context.Context, psvc *persistence_fake.MockService, tc *temporalsdk_mocks.Client) {
				psvc.EXPECT().CreateSIP(
					ctx,
					&datatypes.SIP{
						UUID:   uuid0,
						Name:   "first.zip",
						Status: enums.SIPStatusQueued,
					},
				).Return(errors.New("persistence error"))
			},
			multipartBody: zipMmultipartBody,
			contentType:   "multipart/form-data; boundary=foobar",
			maxUploadSize: 102400000,
			wantErr:       "persistence error",
		},
		{
			name: "Returns Temporal error",
			mock: func(ctx context.Context, psvc *persistence_fake.MockService, tc *temporalsdk_mocks.Client) {
				psvc.EXPECT().CreateSIP(
					ctx,
					&datatypes.SIP{
						UUID:   uuid0,
						Name:   "first.zip",
						Status: enums.SIPStatusQueued,
					},
				).DoAndReturn(func(ctx context.Context, s *datatypes.SIP) error {
					s.ID = 1
					return nil
				})

				tc.On(
					"ExecuteWorkflow",
					mock.AnythingOfType("*context.timerCtx"),
					temporalsdk_client.StartWorkflowOptions{
						ID:                    fmt.Sprintf("processing-workflow-%s", uuid0.String()),
						TaskQueue:             "test",
						WorkflowIDReusePolicy: temporalsdk_api_enums.WORKFLOW_ID_REUSE_POLICY_ALLOW_DUPLICATE,
					},
					ingest.ProcessingWorkflowName,
					&ingest.ProcessingWorkflowRequest{
						SIPUUID:   uuid0,
						SIPName:   "first.zip",
						Type:      enums.WorkflowTypeCreateAip,
						Key:       key,
						Extension: ".zip",
					},
				).Return(nil, errors.New("temporal error"))

				psvc.EXPECT().DeleteSIP(ctx, 1)
			},
			multipartBody: zipMmultipartBody,
			contentType:   "multipart/form-data; boundary=foobar",
			maxUploadSize: 102400000,
			wantErr:       "temporal error",
		},
		{
			name: "Uploads a SIP",
			mock: func(ctx context.Context, psvc *persistence_fake.MockService, tc *temporalsdk_mocks.Client) {
				psvc.EXPECT().CreateSIP(
					ctx,
					&datatypes.SIP{
						UUID:   uuid0,
						Name:   "first.zip",
						Status: enums.SIPStatusQueued,
					},
				).Return(nil)

				tc.On(
					"ExecuteWorkflow",
					mock.AnythingOfType("*context.timerCtx"),
					temporalsdk_client.StartWorkflowOptions{
						ID:                    fmt.Sprintf("processing-workflow-%s", uuid0.String()),
						TaskQueue:             "test",
						WorkflowIDReusePolicy: temporalsdk_api_enums.WORKFLOW_ID_REUSE_POLICY_ALLOW_DUPLICATE,
					},
					ingest.ProcessingWorkflowName,
					&ingest.ProcessingWorkflowRequest{
						SIPUUID:   uuid0,
						SIPName:   "first.zip",
						Type:      enums.WorkflowTypeCreateAip,
						Key:       key,
						Extension: ".zip",
					},
				).Return(nil, nil)
			},
			multipartBody: zipMmultipartBody,
			contentType:   "multipart/form-data; boundary=foobar",
			maxUploadSize: 102400000,
			want:          &goaingest.UploadSipResult{UUID: uuid0.String()},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			b := memblob.OpenBucket(nil)
			r := io.NopCloser(strings.NewReader(tt.multipartBody))
			svc, psvc, tc := testSvc(t, b, tt.maxUploadSize)
			ctx := t.Context()
			if tt.mock != nil {
				tt.mock(ctx, psvc, tc)
			}

			re, err := svc.Goa().UploadSip(ctx, &goaingest.UploadSipPayload{ContentType: tt.contentType}, r)
			if tt.wantErr != "" {
				assert.Error(t, err, tt.wantErr)

				// On any error, the blob should not exist.
				_, err = b.ReadAll(ctx, key)
				assert.ErrorContains(t, err, fmt.Sprintf("blob (key %q) (code=NotFound): blob not found", key))

				return
			}
			assert.NilError(t, err)
			assert.DeepEqual(t, re, tt.want)

			// Make sure the blob has been uploaded.
			data, err := b.ReadAll(ctx, key)
			assert.NilError(t, err)
			assert.Equal(t, string(data), "<binary zip data>")
		})
	}
}
