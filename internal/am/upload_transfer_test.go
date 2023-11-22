package am_test

import (
	"errors"
	"fmt"
	"os"
	"testing"

	"github.com/go-logr/logr"
	"go.artefactual.dev/tools/mockutil"
	temporalsdk_activity "go.temporal.io/sdk/activity"
	temporalsdk_testsuite "go.temporal.io/sdk/testsuite"
	"go.uber.org/mock/gomock"
	"gotest.tools/v3/assert"
	tfs "gotest.tools/v3/fs"

	"github.com/artefactual-sdps/enduro/internal/am"
	sftp_fake "github.com/artefactual-sdps/enduro/internal/sftp/fake"
)

func TestUploadTransferActivity(t *testing.T) {
	filename := "transfer.zip"
	td := tfs.NewDir(t, "enduro-upload-transfer-test",
		tfs.WithFile(filename, "Testing 1-2-3!"),
	)

	type test struct {
		name     string
		params   am.UploadTransferActivityParams
		want     am.UploadTransferActivityResult
		recorder func(*sftp_fake.MockClientMockRecorder)
		errMsg   string
	}
	for _, tt := range []test{
		{
			name: "Uploads transfer",
			params: am.UploadTransferActivityParams{
				SourcePath: td.Join(filename),
			},
			recorder: func(m *sftp_fake.MockClientMockRecorder) {
				var t *os.File
				m.Upload(
					mockutil.Context(),
					gomock.AssignableToTypeOf(t),
					filename,
				).Return(int64(14), "/transfer_dir/"+filename, nil)
			},
			want: am.UploadTransferActivityResult{
				BytesCopied: int64(14),
				RemotePath:  "/transfer_dir/" + filename,
			},
		},
		{
			name: "Errors when local file can't be read",
			params: am.UploadTransferActivityParams{
				SourcePath: td.Join("missing"),
			},
			errMsg: fmt.Sprintf("upload transfer: open %s: no such file or directory", td.Join("missing")),
		},
		{
			name: "Errors when upload fails",
			params: am.UploadTransferActivityParams{
				SourcePath: td.Join(filename),
			},
			recorder: func(m *sftp_fake.MockClientMockRecorder) {
				var t *os.File
				m.Upload(
					mockutil.Context(),
					gomock.AssignableToTypeOf(t),
					filename,
				).Return(
					0,
					"",
					errors.New("SSH: failed to connect: dial tcp 127.0.0.1:2200: connect: connection refused"),
				)
			},
			errMsg: "upload transfer: SSH: failed to connect: dial tcp 127.0.0.1:2200: connect: connection refused",
		},
	} {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ts := &temporalsdk_testsuite.WorkflowTestSuite{}
			env := ts.NewTestActivityEnvironment()
			msvc := sftp_fake.NewMockClient(gomock.NewController(t))

			if tt.recorder != nil {
				tt.recorder(msvc.EXPECT())
			}

			env.RegisterActivityWithOptions(
				am.NewUploadTransferActivity(logr.Discard(), msvc).Execute,
				temporalsdk_activity.RegisterOptions{
					Name: am.UploadTransferActivityName,
				},
			)

			fut, err := env.ExecuteActivity(am.UploadTransferActivityName, tt.params)
			if tt.errMsg != "" {
				assert.ErrorContains(t, err, tt.errMsg)
				return
			}

			var res am.UploadTransferActivityResult
			err = fut.Get(&res)
			assert.NilError(t, err)
			assert.DeepEqual(t, res, tt.want)
		})
	}
}
