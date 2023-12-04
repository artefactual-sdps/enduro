package am_test

import (
	"errors"
	"fmt"
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

func TestDeleteTransferActivity(t *testing.T) {
	t.Parallel()

	filename := "fake_bag"
	activityErr := "activity error (type: DeleteTransferActivity, scheduledEventID: 0, startedEventID: 0, identity: ): "
	td := tfs.NewDir(t, "enduro-delete-transfer-test",
		tfs.WithFile(filename, "Testing 1-2-3!"),
	)

	type test struct {
		name     string
		params   am.DeleteTransferActivityParams
		recorder func(*sftp_fake.MockClientMockRecorder, am.DeleteTransferActivityParams)
		errMsg   string
	}
	for _, tt := range []test{
		{
			name: "Deletes transfer",
			params: am.DeleteTransferActivityParams{
				Destination: td.Path(),
			},
			recorder: func(m *sftp_fake.MockClientMockRecorder, params am.DeleteTransferActivityParams) {
				m.Delete(
					mockutil.Context(),
					params.Destination,
				).Return(nil)
			},
		},
		{
			name: "Errors when file does not exist",
			params: am.DeleteTransferActivityParams{
				Destination: td.Join("missing"),
			},
			recorder: func(m *sftp_fake.MockClientMockRecorder, params am.DeleteTransferActivityParams) {
				m.Delete(
					mockutil.Context(),
					params.Destination,
				).Return(
					errors.New("SFTP: unable to remove file \"test.txt\": file does not exist"),
				)
			},
			errMsg: fmt.Sprintf("delete transfer: path: %q: %v", td.Join("missing"), errors.New("SFTP: unable to remove file \"test.txt\": file does not exist")),
		},
		{
			name: "Errors when Delete fails",
			params: am.DeleteTransferActivityParams{
				Destination: td.Join(filename),
			},
			recorder: func(m *sftp_fake.MockClientMockRecorder, params am.DeleteTransferActivityParams) {
				m.Delete(
					mockutil.Context(),
					params.Destination,
				).Return(
					errors.New("SSH: failed to connect: dial tcp 127.0.0.1:2200: connect: connection refused"),
				)
			},
			errMsg: fmt.Sprintf("delete transfer: path: %q: %v", td.Join(filename), errors.New("SSH: failed to connect: dial tcp 127.0.0.1:2200: connect: connection refused")),
		},
	} {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ts := &temporalsdk_testsuite.WorkflowTestSuite{}
			env := ts.NewTestActivityEnvironment()
			msvc := sftp_fake.NewMockClient(gomock.NewController(t))

			if tt.recorder != nil {
				tt.recorder(msvc.EXPECT(), tt.params)
			}

			env.RegisterActivityWithOptions(
				am.NewDeleteTransferActivity(logr.Discard(), msvc).Execute,
				temporalsdk_activity.RegisterOptions{
					Name: am.DeleteTransferActivityName,
				},
			)

			_, err := env.ExecuteActivity(am.DeleteTransferActivityName, tt.params)
			if tt.errMsg != "" {
				assert.Error(t, err, activityErr+tt.errMsg)
				return
			}

			assert.NilError(t, err)
		})
	}
}
