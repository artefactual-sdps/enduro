package am_test

import (
	"net/http"
	"testing"
	"time"

	"github.com/go-logr/logr"
	"github.com/google/uuid"
	"go.artefactual.dev/amclient"
	"go.artefactual.dev/amclient/amclienttest"
	"go.artefactual.dev/tools/mockutil"
	temporal_tools "go.artefactual.dev/tools/temporal"
	temporalsdk_activity "go.temporal.io/sdk/activity"
	temporalsdk_testsuite "go.temporal.io/sdk/testsuite"
	"go.uber.org/mock/gomock"
	"gotest.tools/v3/assert"

	"github.com/artefactual-sdps/enduro/internal/am"
)

func TestPollTransferActivity(t *testing.T) {
	transferID := uuid.New().String()
	sipID := uuid.New().String()
	path := "/var/archivematica/fake/sip"
	httpError := func(m *amclienttest.MockTransferServiceMockRecorder, statusCode int) {
		m.Status(
			mockutil.Context(),
			transferID,
		).Return(
			nil,
			&amclient.Response{Response: &http.Response{StatusCode: statusCode}},
			&amclient.ErrorResponse{Response: &http.Response{StatusCode: statusCode}},
		)
	}

	type test struct {
		name         string
		statusCode   int
		mockr        func(*amclienttest.MockTransferServiceMockRecorder, int)
		want         am.PollTransferActivityResult
		wantErr      string
		retryableErr bool
	}
	for _, tt := range []test{
		{
			name: "Polls twice then returns successfully",
			mockr: func(m *amclienttest.MockTransferServiceMockRecorder, statusCode int) {
				// AM sometimes returns a "400 Bad Request" error when a
				// transfer is processing.
				m.Status(
					mockutil.Context(),
					transferID,
				).Return(
					nil,
					&amclient.Response{
						Response: &http.Response{StatusCode: http.StatusBadRequest, Status: "400 Bad request"},
					},
					&amclient.ErrorResponse{
						Response: &http.Response{StatusCode: http.StatusBadRequest, Status: "400 Bad request"},
					},
				)

				// AM sometimes returns a "200 OK" response when a transfer is
				// processing.
				m.Status(
					mockutil.Context(),
					transferID,
				).Return(
					&amclient.TransferStatusResponse{Status: "PROCESSING"},
					&amclient.Response{
						Response: &http.Response{StatusCode: http.StatusOK, Status: "200 OK"},
					},
					nil,
				)
				m.Status(
					mockutil.Context(),
					transferID,
				).Return(
					&amclient.TransferStatusResponse{Status: "COMPLETE", SIPID: sipID, Path: path},
					nil,
					nil,
				)
			},
			want:         am.PollTransferActivityResult{SIPID: sipID, Path: path},
			retryableErr: true,
		},
		{
			name: "Non-retryable error because SIP is in BACKLOG",
			mockr: func(m *amclienttest.MockTransferServiceMockRecorder, statusCode int) {
				m.Status(
					mockutil.Context(),
					transferID,
				).Return(
					&amclient.TransferStatusResponse{Status: "COMPLETE", SIPID: "BACKLOG", Path: path},
					&amclient.Response{},
					nil,
				)
			},
			wantErr: "Archivematica SIP sent to backlog",
		},
		{
			name: "Non-retryable error from unknown transfer status",
			mockr: func(m *amclienttest.MockTransferServiceMockRecorder, statusCode int) {
				m.Status(
					mockutil.Context(),
					transferID,
				).Return(
					&amclient.TransferStatusResponse{Status: "UNKNOWN"},
					nil,
					nil,
				)
			},
			wantErr: "Unknown Archivematica transfer status: UNKNOWN",
		},
		{
			name: "Non-retryable error because transfer failed",
			mockr: func(m *amclienttest.MockTransferServiceMockRecorder, statusCode int) {
				m.Status(
					mockutil.Context(),
					transferID,
				).Return(
					&amclient.TransferStatusResponse{Status: "FAILED"},
					nil,
					nil,
				)
			},
			wantErr: "Archivematica transfer status: FAILED",
		},
		{
			name:       "Non-retryable error from http invalid credentials",
			mockr:      httpError,
			statusCode: http.StatusUnauthorized,
			wantErr:    "invalid Archivematica credentials",
		},
		{
			name:       "Non-retryable error from http insufficient permissions",
			mockr:      httpError,
			statusCode: http.StatusForbidden,
			wantErr:    "insufficient Archivematica permissions",
		},
		{
			name:       "Non-retryable error from http not found response",
			mockr:      httpError,
			statusCode: http.StatusNotFound,
			wantErr:    "Archivematica resource not found",
		},
	} {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ts := &temporalsdk_testsuite.WorkflowTestSuite{}
			env := ts.NewTestActivityEnvironment()
			ctrl := gomock.NewController(t)
			amts := amclienttest.NewMockTransferService(ctrl)

			if tt.mockr != nil {
				tt.mockr(amts.EXPECT(), tt.statusCode)
			}

			env.RegisterActivityWithOptions(
				am.NewPollTransferActivity(
					logr.Discard(),
					&am.Config{PollInterval: time.Millisecond * 10},
					amts,
				).Execute,
				temporalsdk_activity.RegisterOptions{
					Name: am.PollTransferActivityName,
				},
			)

			enc, err := env.ExecuteActivity(
				am.PollTransferActivityName,
				am.PollTransferActivityParams{TransferID: transferID},
			)
			if tt.wantErr != "" {
				assert.ErrorContains(t, err, tt.wantErr)
				assert.Assert(t, temporal_tools.NonRetryableError(err))

				return
			}
			assert.NilError(t, err)

			var r am.PollTransferActivityResult
			enc.Get(&r)
			assert.DeepEqual(t, r, tt.want)
		})
	}
}
