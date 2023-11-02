package am_test

import (
	"errors"
	"net/http"
	"testing"

	"github.com/go-logr/logr"
	"github.com/google/uuid"
	"go.artefactual.dev/amclient"
	"go.artefactual.dev/amclient/amclienttest"
	"go.artefactual.dev/tools/mockutil"
	temporalsdk_activity "go.temporal.io/sdk/activity"
	temporalsdk_temporal "go.temporal.io/sdk/temporal"
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
			name: "Returns a complete result, a SIP ID, and a Path",
			mockr: func(m *amclienttest.MockTransferServiceMockRecorder, statusCode int) {
				m.Status(
					mockutil.Context(),
					transferID,
				).Return(
					&amclient.TransferStatusResponse{Status: "COMPLETE", SIPID: sipID, Path: path},
					&amclient.Response{},
					nil,
				)
			},
			want:         am.PollTransferActivityResult{SIPID: sipID, Path: path},
			retryableErr: true,
		},
		{
			name: "Returns a complete result but the SIPID is BACKLOG",
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
			wantErr: "Archivematica transfer sent to backlog",
		},
		{
			name: "Returns an unknown transfer status error",
			mockr: func(m *amclienttest.MockTransferServiceMockRecorder, statusCode int) {
				m.Status(
					mockutil.Context(),
					transferID,
				).Return(
					&amclient.TransferStatusResponse{Status: "UNKNOWN"},
					&amclient.Response{},
					nil,
				)
			},
			wantErr: "Unknown Archivematica transfer status: UNKNOWN",
		},
		{
			name: "Returns a non-retryable status failed error",
			mockr: func(m *amclienttest.MockTransferServiceMockRecorder, statusCode int) {
				m.Status(
					mockutil.Context(),
					transferID,
				).Return(
					&amclient.TransferStatusResponse{Status: "FAILED"},
					&amclient.Response{},
					nil,
				)
			},
			wantErr: "Archivematica transfer status: FAILED",
		},
		{
			name:       "Returns a non-retryable invalid credentials error",
			mockr:      httpError,
			statusCode: http.StatusUnauthorized,
			wantErr:    "invalid Archivematica credentials",
		},
		{
			name:       "Returns a non-retryable insufficient permissions error",
			mockr:      httpError,
			statusCode: http.StatusForbidden,
			wantErr:    "insufficient Archivematica permissions",
		},
		{
			name:       "Returns a non-retryable not found error",
			mockr:      httpError,
			statusCode: http.StatusNotFound,
			wantErr:    "Archivematica transfer not found",
		},
		// 	TODO: continue polling
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
				am.NewPollTransferActivity(logr.Discard(), &am.Config{}, amts).Execute,
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

				var e *temporalsdk_temporal.ApplicationError
				if ok := errors.As(err, &e); ok {
					assert.Assert(t, e.NonRetryable() != tt.retryableErr)
				}

				return
			}

			var r am.PollTransferActivityResult
			enc.Get(&r)
			assert.DeepEqual(t, r, tt.want)
		})
	}
}
