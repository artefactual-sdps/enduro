package am_test

import (
	"errors"
	"fmt"
	"os"
	"testing"

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
	filename := "transfer.zip"
	td := tfs.NewDir(t, "enduro-upload-transfer-test",
		tfs.WithFile(filename, "Testing 1-2-3!"),
	)

	type test struct {
		name            string
		params          am.UploadTransferActivityParams
		recorder        func(*sftp_fake.MockClientMockRecorder)
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
			recorder: func(m *sftp_fake.MockClientMockRecorder) {
				var t *os.File
				m.Upload(
					mockutil.Context(),
					gomock.AssignableToTypeOf(t),
					filename,
				).Return(int64(14), "/transfer_dir/"+filename, nil)
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
			recorder: func(m *sftp_fake.MockClientMockRecorder) {
				var t *os.File
				m.Upload(
					mockutil.Context(),
					gomock.AssignableToTypeOf(t),
					filename,
				).Return(
					0,
					"",
					errors.New("ssh: failed to connect: dial tcp 127.0.0.1:2200: connect: connection refused"),
				)
			},
			wantErr: "activity error (type: UploadTransferActivity, scheduledEventID: 0, startedEventID: 0, identity: ): UploadTransferActivity: ssh: failed to connect: dial tcp 127.0.0.1:2200: connect: connection refused",
		},
		{
			name: "Non-retryable error when authentication fails",
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
					&sftp.AuthError{
						Message: "ssh: handshake failed: ssh: unable to authenticate, attempted methods [none publickey], no supported methods remain",
					},
				)
			},
			wantErr:         "activity error (type: UploadTransferActivity, scheduledEventID: 0, startedEventID: 0, identity: ): UploadTransferActivity: auth: ssh: handshake failed: ssh: unable to authenticate, attempted methods [none publickey], no supported methods remain",
			wantNonRetryErr: true,
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
