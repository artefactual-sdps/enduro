package am_test

import (
	"net/http"
	"testing"
	"time"

	"github.com/go-logr/logr"
	"github.com/google/uuid"
	"github.com/jonboulle/clockwork"
	"go.artefactual.dev/amclient"
	"go.artefactual.dev/amclient/amclienttest"
	"go.artefactual.dev/tools/mockutil"
	temporal_tools "go.artefactual.dev/tools/temporal"
	temporalsdk_activity "go.temporal.io/sdk/activity"
	temporalsdk_testsuite "go.temporal.io/sdk/testsuite"
	"go.uber.org/mock/gomock"
	"gotest.tools/v3/assert"

	"github.com/artefactual-sdps/enduro/internal/am"
	fake_package "github.com/artefactual-sdps/enduro/internal/package_/fake"
)

var (
	http200Resp = http.Response{StatusCode: http.StatusOK, Status: "200 OK"}
	http400Resp = http.Response{StatusCode: http.StatusBadRequest, Status: "400 Bad request"}
)

func TestPollTransferActivity(t *testing.T) {
	t.Parallel()
	transferID := uuid.New().String()
	presActionID := uint(1)
	sipID := uuid.New().String()
	path := "/var/archivematica/fake/sip"

	jobs := []amclient.Job{
		{
			ID:           "e6e01ebb-a8f4-459d-b9a9-c6a8103e4750",
			Name:         "Extract zipped transfer",
			Status:       amclient.JobStatusComplete,
			Microservice: "Approve transfer",
			LinkID:       "541f5994-73b0-45bb-9cb5-367c06a21be7",
			Tasks: []amclient.Task{
				{
					ID:          "11566538-66c5-4a20-aa70-77f7a9fa83d5",
					ExitCode:    0,
					Filename:    "Images-94ade01c-49ce-49e0-9cc3-805575c676d0",
					CreatedAt:   amclient.TaskDateTime{Time: time.Date(2024, time.January, 18, 1, 27, 49, 0, time.UTC)},
					StartedAt:   amclient.TaskDateTime{Time: time.Date(2024, time.January, 18, 1, 27, 49, 0, time.UTC)},
					CompletedAt: amclient.TaskDateTime{Time: time.Date(2024, time.January, 18, 1, 27, 49, 0, time.UTC)},
					Duration:    amclient.TaskDuration(time.Second / 2),
				},
			},
		},
		{
			ID:           "2bcdb038-8861-4ea7-a7bb-01d58efac38c",
			Name:         "Set transfer type: Standard",
			Status:       amclient.JobStatusComplete,
			Microservice: "Verify transfer compliance",
			LinkID:       "045c43ae-d6cf-44f7-97d6-c8a602748565",
			Tasks: []amclient.Task{
				{
					ID:        "53666170-0397-4962-8736-23295444b036",
					ExitCode:  0,
					FileID:    "",
					Filename:  "Images-94ade01c-49ce-49e0-9cc3-805575c676d0",
					CreatedAt: amclient.TaskDateTime{Time: time.Date(2024, time.January, 18, 1, 27, 49, 0, time.UTC)},
					Duration:  amclient.TaskDuration(time.Second / 2),
				},
			},
		},
	}

	type test struct {
		name         string
		params       *am.PollTransferActivityParams
		tfrRec       func(*amclienttest.MockTransferServiceMockRecorder)
		jobRec       func(*amclienttest.MockJobsServiceMockRecorder)
		pkgRec       func(*fake_package.MockServiceMockRecorder)
		want         am.PollTransferActivityResult
		wantErr      string
		retryableErr bool
	}
	for _, tt := range []test{
		{
			name: "Polls twice then returns successfully",
			params: &am.PollTransferActivityParams{
				PresActionID: presActionID,
				TransferID:   transferID,
			},
			tfrRec: func(m *amclienttest.MockTransferServiceMockRecorder) {
				// First poll. AM sometimes returns a "400 Bad Request" error
				// when transfer processing has just started.
				m.Status(
					mockutil.Context(),
					transferID,
				).Return(
					nil,
					&amclient.Response{Response: &http400Resp},
					&amclient.ErrorResponse{Response: &http400Resp},
				)

				// Second poll. AM usually returns a "200 OK" response when a
				// transfer is processing.
				m.Status(
					mockutil.Context(),
					transferID,
				).Return(
					&amclient.TransferStatusResponse{Status: "PROCESSING"},
					&amclient.Response{Response: &http200Resp},
					nil,
				)

				// Third poll. Complete transfer.
				m.Status(
					mockutil.Context(),
					transferID,
				).Return(
					&amclient.TransferStatusResponse{Status: "COMPLETE", SIPID: sipID, Path: path},
					&amclient.Response{Response: &http200Resp},
					nil,
				)
			},
			jobRec: func(m *amclienttest.MockJobsServiceMockRecorder) {
				// Second poll.
				m.List(
					mockutil.Context(),
					transferID,
					&amclient.JobsListRequest{
						Detailed: true,
					},
				).Return(
					jobs,
					&amclient.Response{Response: &http200Resp},
					nil,
				)

				// Third poll. These jobs were saved on the previous poll, so
				// they shouldn't be saved again.
				m.List(
					mockutil.Context(),
					transferID,
					&amclient.JobsListRequest{
						Detailed: true,
					},
				).Return(
					jobs,
					&amclient.Response{Response: &http200Resp},
					nil,
				)
			},
			pkgRec: func(m *fake_package.MockServiceMockRecorder) {
				// Second poll.
				for _, job := range jobs {
					pt := am.ConvertJobToPreservationTask(job)
					pt.PreservationActionID = presActionID
					m.CreatePreservationTask(mockutil.Context(), &pt).Return(nil)

				}
			},
			want: am.PollTransferActivityResult{
				SIPID:         sipID,
				Path:          path,
				PresTaskCount: 2,
			},
		},
		{
			name: "Non-retryable error because SIP is in BACKLOG",
			tfrRec: func(m *amclienttest.MockTransferServiceMockRecorder) {
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
			name: "Non-retryable error from an unknown response status",
			tfrRec: func(m *amclienttest.MockTransferServiceMockRecorder) {
				m.Status(
					mockutil.Context(),
					transferID,
				).Return(
					&amclient.TransferStatusResponse{Status: "UNKNOWN"},
					nil,
					nil,
				)
			},
			wantErr: "Unknown Archivematica response status: UNKNOWN",
		},
		{
			name: "Non-retryable error because transfer failed",
			tfrRec: func(m *amclienttest.MockTransferServiceMockRecorder) {
				m.Status(
					mockutil.Context(),
					transferID,
				).Return(
					&amclient.TransferStatusResponse{Status: "FAILED"},
					nil,
					nil,
				)
			},
			wantErr: "Archivematica response status: FAILED",
		},
		{
			name: "Retryable error on 500 Internal Server Error",
			tfrRec: func(m *amclienttest.MockTransferServiceMockRecorder) {
				httpResp := http.Response{
					StatusCode: http.StatusInternalServerError,
					Status:     "500 Internal Server Error",
				}
				m.Status(
					mockutil.Context(),
					transferID,
				).Return(
					nil,
					&amclient.Response{Response: &httpResp},
					&amclient.ErrorResponse{Response: &httpResp},
				)
			},
			wantErr:      "Archivematica error: 500 Internal Server Error",
			retryableErr: true,
		},
		{
			name: "Non-retryable error from http invalid credentials",
			tfrRec: func(m *amclienttest.MockTransferServiceMockRecorder) {
				httpResp := http.Response{StatusCode: http.StatusUnauthorized}
				m.Status(
					mockutil.Context(),
					transferID,
				).Return(
					nil,
					&amclient.Response{Response: &httpResp},
					&amclient.ErrorResponse{Response: &httpResp},
				)
			},
			wantErr: "invalid Archivematica credentials",
		},
		{
			name: "Non-retryable error from http insufficient permissions",
			tfrRec: func(m *amclienttest.MockTransferServiceMockRecorder) {
				httpResp := http.Response{StatusCode: http.StatusForbidden}
				m.Status(
					mockutil.Context(),
					transferID,
				).Return(
					nil,
					&amclient.Response{Response: &httpResp},
					&amclient.ErrorResponse{Response: &httpResp},
				)
			},
			wantErr: "insufficient Archivematica permissions",
		},
		{
			name: "Non-retryable error from http not found response",
			tfrRec: func(m *amclienttest.MockTransferServiceMockRecorder) {
				httpResp := http.Response{StatusCode: http.StatusNotFound}
				m.Status(
					mockutil.Context(),
					transferID,
				).Return(
					nil,
					&amclient.Response{Response: &httpResp},
					&amclient.ErrorResponse{Response: &httpResp},
				)
			},
			wantErr: "Archivematica resource not found",
		},
	} {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ts := &temporalsdk_testsuite.WorkflowTestSuite{}
			env := ts.NewTestActivityEnvironment()
			ctrl := gomock.NewController(t)

			trfSvc := amclienttest.NewMockTransferService(ctrl)
			if tt.tfrRec != nil {
				tt.tfrRec(trfSvc.EXPECT())
			}

			jobSvc := amclienttest.NewMockJobsService(ctrl)
			if tt.jobRec != nil {
				tt.jobRec(jobSvc.EXPECT())
			}

			pkgSvc := fake_package.NewMockService(ctrl)
			if tt.pkgRec != nil {
				tt.pkgRec(pkgSvc.EXPECT())
			}

			env.RegisterActivityWithOptions(
				am.NewPollTransferActivity(
					logr.Discard(),
					&am.Config{PollInterval: time.Millisecond * 10},
					clockwork.NewFakeClock(),
					trfSvc,
					jobSvc,
					pkgSvc,
				).Execute,
				temporalsdk_activity.RegisterOptions{
					Name: am.PollTransferActivityName,
				},
			)

			enc, err := env.ExecuteActivity(
				am.PollTransferActivityName,
				am.PollTransferActivityParams{
					PresActionID: presActionID,
					TransferID:   transferID,
				},
			)
			if tt.wantErr != "" {
				assert.ErrorContains(t, err, tt.wantErr)
				assert.Assert(t, temporal_tools.NonRetryableError(err) != tt.retryableErr)

				return
			}
			assert.NilError(t, err)

			var r am.PollTransferActivityResult
			enc.Get(&r)
			assert.DeepEqual(t, r, tt.want)
		})
	}
}
