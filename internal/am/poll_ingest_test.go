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

func TestPollIngestActivity(t *testing.T) {
	sipID := uuid.New().String()
	path := "/var/archivematica/fake/sip"
	httpError := func(m *amclienttest.MockIngestServiceMockRecorder, statusCode int) {
		m.Status(
			mockutil.Context(),
			sipID,
		).Return(
			nil,
			&amclient.Response{Response: &http.Response{StatusCode: statusCode}},
			&amclient.ErrorResponse{Response: &http.Response{StatusCode: statusCode}},
		)
	}

	type test struct {
		name         string
		statusCode   int
		mockr        func(*amclienttest.MockIngestServiceMockRecorder, int)
		want         am.PollIngestActivityResult
		wantErr      string
		retryableErr bool
	}
	for _, tt := range []test{
		{
			name: "Polls twice then returns successfully",
			mockr: func(m *amclienttest.MockIngestServiceMockRecorder, statusCode int) {
				// AM sometimes returns a "400 Bad Request" error when a
				// transfer is processing.
				m.Status(
					mockutil.Context(),
					sipID,
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
					sipID,
				).Return(
					&amclient.IngestStatusResponse{Status: "PROCESSING"},
					&amclient.Response{
						Response: &http.Response{StatusCode: http.StatusOK, Status: "200 OK"},
					},
					nil,
				)
				m.Status(
					mockutil.Context(),
					sipID,
				).Return(
					&amclient.IngestStatusResponse{Status: "COMPLETE", SIPID: sipID, Path: path},
					nil,
					nil,
				)
			},
			want:         am.PollIngestActivityResult{Status: "COMPLETE"},
			retryableErr: true,
		},
		{
			name: "Non-retryable error from an unknown response status",
			mockr: func(m *amclienttest.MockIngestServiceMockRecorder, statusCode int) {
				m.Status(
					mockutil.Context(),
					sipID,
				).Return(
					&amclient.IngestStatusResponse{Status: "UNKNOWN"},
					nil,
					nil,
				)
			},
			wantErr: "Unknown Archivematica response status: UNKNOWN",
		},
		{
			name: "Non-retryable error because ingest failed",
			mockr: func(m *amclienttest.MockIngestServiceMockRecorder, statusCode int) {
				m.Status(
					mockutil.Context(),
					sipID,
				).Return(
					&amclient.IngestStatusResponse{Status: "FAILED"},
					nil,
					nil,
				)
			},
			wantErr: "Archivematica response status: FAILED",
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
			msvc := amclienttest.NewMockIngestService(ctrl)

			if tt.mockr != nil {
				tt.mockr(msvc.EXPECT(), tt.statusCode)
			}

			env.RegisterActivityWithOptions(
				am.NewPollIngestActivity(
					logr.Discard(),
					&am.Config{PollInterval: time.Millisecond * 10},
					msvc,
				).Execute,
				temporalsdk_activity.RegisterOptions{
					Name: am.PollIngestActivityName,
				},
			)

			enc, err := env.ExecuteActivity(
				am.PollIngestActivityName,
				am.PollIngestActivityParams{SIPID: sipID},
			)
			if tt.wantErr != "" {
				assert.ErrorContains(t, err, tt.wantErr)
				assert.Assert(t, temporal_tools.NonRetryableError(err))

				return
			}
			assert.NilError(t, err)

			var r am.PollIngestActivityResult
			enc.Get(&r)
			assert.DeepEqual(t, r, tt.want)
		})
	}
}
