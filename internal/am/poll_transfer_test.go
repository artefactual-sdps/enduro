package am_test

import (
	"database/sql"
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
	ttime, err := time.Parse(time.RFC3339, "2023-12-05T12:02:00Z")
	if err != nil {
		t.Fatal("Invalid test time")
	}
	nullTime := sql.NullTime{Time: ttime, Valid: true}

	jobs := []amclient.Job{
		{
			ID:           "f60018ac-da79-4769-9509-c6c41d5efe7e",
			LinkID:       "70669a5b-01e4-4ea0-ac70-10292f87da05",
			Microservice: "Verify SIP compliance",
			Name:         "Move to processing directory",
			Status:       amclient.JobStatusComplete,
			Tasks: []amclient.Task{
				{
					ID:       "c134198c-9485-4f68-8d94-4da1e03b5e1b",
					ExitCode: 0,
				},
			},
		},
		{
			ID:           "c2128d39-2ace-47c5-8cac-39ded8d9c9ef",
			LinkID:       "208d441b-6938-44f9-b54a-bd73f05bc764",
			Microservice: "Verify SIP compliance",
			Name:         "Verify SIP compliance",
			Status:       amclient.JobStatusComplete,
			Tasks: []amclient.Task{
				{
					ID:       "6f5beca3-71ad-446c-8f19-3bc4dea16c9b",
					ExitCode: 0,
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
					&amclient.JobsListRequest{},
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
					&amclient.JobsListRequest{},
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
					pt.CompletedAt = nullTime
					pt.StartedAt = nullTime
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
					clockwork.NewFakeClockAt(ttime),
					jobSvc,
					pkgSvc,
					trfSvc,
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
