package am_test

import (
	"errors"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/go-logr/logr"
	"go.artefactual.dev/tools/mockutil"
	"go.artefactual.dev/tools/temporal"
	temporalsdk_activity "go.temporal.io/sdk/activity"
	temporalsdk_testsuite "go.temporal.io/sdk/testsuite"
	"go.uber.org/mock/gomock"
	"gotest.tools/v3/assert"
	tfs "gotest.tools/v3/fs"

	"github.com/artefactual-sdps/enduro/internal/am"
	"github.com/artefactual-sdps/enduro/internal/sftp"
	sftp_fake "github.com/artefactual-sdps/enduro/internal/sftp/fake"
)

func TestUploadTransferActivity(t *testing.T) {
	t.Parallel()

	filename := "transfer.zip"
	td := tfs.NewDir(t, "enduro-upload-transfer-test",
		tfs.WithFile(filename, "Testing 1-2-3!"),
	)

	type test struct {
		name            string
		params          am.UploadTransferActivityParams
		mock            func(*gomock.Controller) (sftp.Client, sftp.AsyncUpload)
		want            am.UploadTransferActivityResult
		wantErr         string
		wantNonRetryErr bool
	}
	for _, tt := range []test{
		{
			name: "Uploads transfer",
			params: am.UploadTransferActivityParams{
				SourcePath: td.Join(filename),
			},
			mock: func(ctrl *gomock.Controller) (sftp.Client, sftp.AsyncUpload) {
				var fp *os.File

				client := sftp_fake.NewMockClient(ctrl)
				upload := sftp_fake.NewMockAsyncUpload(ctrl)

				client.EXPECT().
					Upload(
						mockutil.Context(),
						gomock.AssignableToTypeOf(fp),
						filename,
					).
					Return("/transfer_dir/"+filename, upload, nil)

				doneCh := make(chan bool, 1)
				upload.EXPECT().Done().Return(doneCh).Times(2)

				errCh := make(chan error, 1)
				upload.EXPECT().Err().Return(errCh).Times(2)

				upload.EXPECT().Bytes().DoAndReturn(func() int64 {
					doneCh <- true
					return int64(7)
				})
				upload.EXPECT().Bytes().Return(14)

				return client, upload
			},
			want: am.UploadTransferActivityResult{
				BytesCopied:        int64(14),
				RemoteFullPath:     "/transfer_dir/" + filename,
				RemoteRelativePath: filename,
			},
		},
		{
			name: "Errors when local file can't be read",
			params: am.UploadTransferActivityParams{
				SourcePath: td.Join("missing"),
			},
			wantErr: fmt.Sprintf("activity error (type: UploadTransferActivity, scheduledEventID: 0, startedEventID: 0, identity: ): UploadTransferActivity: open %s: no such file or directory", td.Join("missing")),
		},
		{
			name: "Retryable error when SSH connection fails",
			params: am.UploadTransferActivityParams{
				SourcePath: td.Join(filename),
			},
			mock: func(ctrl *gomock.Controller) (sftp.Client, sftp.AsyncUpload) {
				var fp *os.File

				client := sftp_fake.NewMockClient(ctrl)
				client.EXPECT().
					Upload(
						mockutil.Context(),
						gomock.AssignableToTypeOf(fp),
						filename,
					).
					Return(
						"",
						nil,
						errors.New("ssh: failed to connect: dial tcp 127.0.0.1:2200: connect: connection refused"),
					)

				return client, nil
			},
			wantErr: "activity error (type: UploadTransferActivity, scheduledEventID: 0, startedEventID: 0, identity: ): UploadTransferActivity: ssh: failed to connect: dial tcp 127.0.0.1:2200: connect: connection refused",
		},
		{
			name: "Non-retryable error when authentication fails",
			params: am.UploadTransferActivityParams{
				SourcePath: td.Join(filename),
			},
			mock: func(ctrl *gomock.Controller) (sftp.Client, sftp.AsyncUpload) {
				var fp *os.File

				client := sftp_fake.NewMockClient(ctrl)
				client.EXPECT().
					Upload(
						mockutil.Context(),
						gomock.AssignableToTypeOf(fp),
						filename,
					).
					Return(
						"",
						nil,
						&sftp.AuthError{
							Message: "ssh: handshake failed: ssh: unable to authenticate, attempted methods [none publickey], no supported methods remain",
						},
					)

				return client, nil
			},
			wantErr:         "activity error (type: UploadTransferActivity, scheduledEventID: 0, startedEventID: 0, identity: ): UploadTransferActivity: auth: ssh: handshake failed: ssh: unable to authenticate, attempted methods [none publickey], no supported methods remain",
			wantNonRetryErr: true,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ts := &temporalsdk_testsuite.WorkflowTestSuite{}
			env := ts.NewTestActivityEnvironment()
			ctrl := gomock.NewController(t)

			var client sftp.Client
			if tt.mock != nil {
				client, _ = tt.mock(ctrl)
			}

			env.RegisterActivityWithOptions(
				am.NewUploadTransferActivity(logr.Discard(), client, 2*time.Millisecond).Execute,
				temporalsdk_activity.RegisterOptions{
					Name: am.UploadTransferActivityName,
				},
			)

			fut, err := env.ExecuteActivity(am.UploadTransferActivityName, tt.params)
			if tt.wantErr != "" {
				assert.Error(t, err, tt.wantErr)
				assert.Assert(t, temporal.NonRetryableError(err) == tt.wantNonRetryErr)
				return
			}

			var res am.UploadTransferActivityResult
			err = fut.Get(&res)
			assert.NilError(t, err)
			assert.DeepEqual(t, res, tt.want)
		})
	}
}
