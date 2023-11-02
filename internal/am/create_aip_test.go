package am_test

import (
	"errors"
	"net/http"
	"testing"

	"github.com/go-logr/logr"
	"github.com/google/uuid"
	"go.artefactual.dev/amclient"
	"go.artefactual.dev/amclient/amclienttest"
	temporalsdk_activity "go.temporal.io/sdk/activity"
	temporalsdk_testsuite "go.temporal.io/sdk/testsuite"
	"go.uber.org/mock/gomock"
	"gotest.tools/v3/assert"

	"github.com/artefactual-sdps/enduro/internal/am"
)

func TestPollIngestActivity_Execute(t *testing.T) {
	// Initializations
	logger := logr.Discard()
	cfg := &am.Config{}
	opts := &am.PollIngestActivityParams{UUID: uuid.New().String()}
	// Define the test cases
	tests := []struct {
		name       string
		mockFunc   func(amis *amclienttest.MockIngestServiceMockRecorder)
		errMessage string
	}{
		{
			name: "successful status check",
			mockFunc: func(amisMock *amclienttest.MockIngestServiceMockRecorder) {
				amisMock.Status(gomock.Any(), gomock.Any()).Return(
					&amclient.IngestStatusResponse{Status: "COMPLETE"},
					&amclient.Response{
						Response: &http.Response{StatusCode: http.StatusOK},
					},
					nil,
				)
			},
		},
		{
			name: "Returns an invalid credentials error",
			mockFunc: func(amisMock *amclienttest.MockIngestServiceMockRecorder) {
				amisMock.Status(gomock.Any(), gomock.Any()).Return(
					nil,
					&amclient.Response{
						Response: &http.Response{StatusCode: http.StatusUnauthorized},
					},
					errors.New("status code error"),
				)
			},
			errMessage: "invalid Archivematica credentials",
		},
		{
			name: "Returns an insufficient permissions error",
			mockFunc: func(amisMock *amclienttest.MockIngestServiceMockRecorder) {
				amisMock.Status(gomock.Any(), gomock.Any()).Return(
					nil,
					&amclient.Response{
						Response: &http.Response{StatusCode: http.StatusForbidden},
					},
					errors.New("status code error"),
				)
			},
			errMessage: "insufficient Archivematica permissions",
		},
		{
			name: "Returns a not found error",
			mockFunc: func(amisMock *amclienttest.MockIngestServiceMockRecorder) {
				amisMock.Status(gomock.Any(), gomock.Any()).Return(
					nil,
					&amclient.Response{
						Response: &http.Response{StatusCode: http.StatusNotFound},
					},
					errors.New("status code error"),
				)
			},
			errMessage: "Archivematica transfer not found",
		},
		{
			name: "Returns a continue polling error",
			mockFunc: func(amisMock *amclienttest.MockIngestServiceMockRecorder) {
				amisMock.Status(gomock.Any(), gomock.Any()).Return(
					&amclient.IngestStatusResponse{Status: "gpPROCESSING"},
					&amclient.Response{
						Response: &http.Response{StatusCode: http.StatusOK},
					},
					nil,
				)
			},
			errMessage: "Continue polling",
		},
		{
			name: "Returns a failed error",
			mockFunc: func(amisMock *amclienttest.MockIngestServiceMockRecorder) {
				amisMock.Status(gomock.Any(), gomock.Any()).Return(
					&amclient.IngestStatusResponse{Status: "FAILED"},
					&amclient.Response{
						Response: &http.Response{StatusCode: http.StatusOK},
					},
					nil,
				)
			},
			errMessage: "ingest is in a state that we can't handle",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ts := &temporalsdk_testsuite.WorkflowTestSuite{}
			env := ts.NewTestActivityEnvironment()
			amisMock := amclienttest.NewMockIngestService(gomock.NewController(t))
			pollIngestActivity := am.NewPollIngestActivity(logger, cfg, amisMock)

			env.RegisterActivityWithOptions(
				pollIngestActivity.Execute,
				temporalsdk_activity.RegisterOptions{
					Name: am.PollIngestActivityName,
				},
			)
			tt.mockFunc(amisMock.EXPECT())

			_, err := env.ExecuteActivity(am.PollIngestActivityName, opts)
			if tt.errMessage != "" {
				assert.ErrorContains(t, err, tt.errMessage)
				return
			}

			assert.NilError(t, err)
		})
	}
}
