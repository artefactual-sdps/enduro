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
	"github.com/artefactual-sdps/enduro/internal/package_"
	fake_package "github.com/artefactual-sdps/enduro/internal/package_/fake"
)

func TestPollIngestActivity(t *testing.T) {
	clock := clockwork.NewFakeClock()
	nullTime := sql.NullTime{Time: clock.Now(), Valid: true}
	path := "/var/archivematica/fake/sip"
	presActionID := uint(2)
	sipID := uuid.New().String()

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

	jobs := []amclient.Job{
		{
			ID:           "7b7f7abd-e9c9-4c2e-9837-a65fa68cfcc8",
			Name:         "Move to processing directory",
			Status:       amclient.JobStatusComplete,
			Microservice: "Verify SIP compliance",
			LinkID:       "70669a5b-01e4-4ea0-ac70-10292f87da05",
			Tasks: []amclient.Task{
				{
					ID:       "9dc0b71a-cbb1-40f4-9fa4-647cc16c8ed5",
					ExitCode: 0,
				},
			},
		},
		{
			ID:           "43222c5f-89c3-469a-9167-5330f9e33e46",
			Name:         "Verify SIP compliance",
			Status:       amclient.JobStatusComplete,
			Microservice: "Verify SIP compliance",
			LinkID:       "208d441b-6938-44f9-b54a-bd73f05bc764",
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
		statusCode   int
		ingRec       func(*amclienttest.MockIngestServiceMockRecorder, int)
		jobRec       func(*amclienttest.MockJobsServiceMockRecorder)
		pkgRec       func(*fake_package.MockServiceMockRecorder)
		want         am.PollIngestActivityResult
		wantErr      string
		retryableErr bool
	}
	for _, tt := range []test{
		{
			name: "Polls twice then returns successfully",
			ingRec: func(m *amclienttest.MockIngestServiceMockRecorder, statusCode int) {
				// Poll 1: AM sometimes returns a "400 Bad Request" error when
				// ingest has just started.
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

				// Poll 2: AM usually returns a "200 OK" response when a
				// ingest is still processing.
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

				// Poll 3: ingest complete.
				m.Status(
					mockutil.Context(),
					sipID,
				).Return(
					&amclient.IngestStatusResponse{Status: "COMPLETE", SIPID: sipID, Path: path},
					nil,
					nil,
				)
			},
			jobRec: func(m *amclienttest.MockJobsServiceMockRecorder) {
				// Poll 2: ingest in progress (one job complete).
				m.List(
					mockutil.Context(),
					sipID,
					&amclient.JobsListRequest{},
				).Return(
					jobs[:1],
					&amclient.Response{Response: &http200Resp},
					nil,
				)

				// Poll 3: ingest is complete (two jobs complete).
				m.List(
					mockutil.Context(),
					sipID,
					&amclient.JobsListRequest{},
				).Return(
					jobs,
					&amclient.Response{Response: &http200Resp},
					nil,
				)
			},
			pkgRec: func(m *fake_package.MockServiceMockRecorder) {
				tasks := make([]*package_.PreservationTask, len(jobs))
				for i, job := range jobs {
					pt := am.ConvertJobToPreservationTask(job)
					pt.PreservationActionID = presActionID
					pt.CompletedAt = nullTime
					pt.StartedAt = nullTime
					tasks[i] = &pt
				}

				// Poll 2: save first job.
				m.CreatePreservationTask(mockutil.Context(), tasks[0]).Return(nil)

				// Poll 3: save second job.
				m.CreatePreservationTask(mockutil.Context(), tasks[1]).Return(nil)
			},
			want: am.PollIngestActivityResult{
				Status:        "COMPLETE",
				PresTaskCount: 2,
			},
		},
		{
			name: "Non-retryable error from an unknown response status",
			ingRec: func(m *amclienttest.MockIngestServiceMockRecorder, statusCode int) {
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
			ingRec: func(m *amclienttest.MockIngestServiceMockRecorder, statusCode int) {
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
			name: "Retryable error on 500 Internal Server Error",
			ingRec: func(m *amclienttest.MockIngestServiceMockRecorder, sc int) {
				httpResp := http.Response{
					StatusCode: http.StatusInternalServerError,
					Status:     "500 Internal Server Error",
				}
				m.Status(
					mockutil.Context(),
					sipID,
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
			name:       "Non-retryable error from http invalid credentials",
			ingRec:     httpError,
			statusCode: http.StatusUnauthorized,
			wantErr:    "invalid Archivematica credentials",
		},
		{
			name:       "Non-retryable error from http insufficient permissions",
			ingRec:     httpError,
			statusCode: http.StatusForbidden,
			wantErr:    "insufficient Archivematica permissions",
		},
		{
			name:       "Non-retryable error from http not found response",
			ingRec:     httpError,
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

			ingSvc := amclienttest.NewMockIngestService(ctrl)
			if tt.ingRec != nil {
				tt.ingRec(ingSvc.EXPECT(), tt.statusCode)
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
				am.NewPollIngestActivity(
					logr.Discard(),
					&am.Config{PollInterval: time.Millisecond * 10},
					clock,
					ingSvc,
					jobSvc,
					pkgSvc,
				).Execute,
				temporalsdk_activity.RegisterOptions{
					Name: am.PollIngestActivityName,
				},
			)

			enc, err := env.ExecuteActivity(
				am.PollIngestActivityName,
				am.PollIngestActivityParams{
					PresActionID: presActionID,
					SIPID:        sipID,
				},
			)
			if tt.wantErr != "" {
				assert.ErrorContains(t, err, tt.wantErr)
				assert.Assert(t, temporal_tools.NonRetryableError(err) != tt.retryableErr)

				return
			}
			assert.NilError(t, err)

			var r am.PollIngestActivityResult
			enc.Get(&r)
			assert.DeepEqual(t, r, tt.want)
		})
	}
}
