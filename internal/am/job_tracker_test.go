package am_test

import (
	"context"
	"database/sql"
	"net/http"
	"testing"

	"github.com/google/uuid"
	"github.com/jonboulle/clockwork"
	"go.artefactual.dev/amclient"
	"go.artefactual.dev/amclient/amclienttest"
	"go.artefactual.dev/tools/mockutil"
	temporal_tools "go.artefactual.dev/tools/temporal"
	"go.uber.org/mock/gomock"
	"gotest.tools/v3/assert"

	"github.com/artefactual-sdps/enduro/internal/am"
	fake_package "github.com/artefactual-sdps/enduro/internal/package_/fake"
)

func TestJobTracker(t *testing.T) {
	t.Parallel()

	paID := uint(1)
	unitID := uuid.New().String()

	clock := clockwork.NewFakeClock()
	nullTime := sql.NullTime{Time: clock.Now(), Valid: true}
	httpError := func(m *amclienttest.MockJobsServiceMockRecorder, statusCode int) {
		m.List(
			mockutil.Context(),
			unitID,
			&amclient.JobsListRequest{},
		).Return(
			nil,
			&amclient.Response{Response: &http.Response{StatusCode: statusCode}},
			&amclient.ErrorResponse{Response: &http.Response{StatusCode: statusCode}},
		)
	}

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
		name       string
		jobRec     func(*amclienttest.MockJobsServiceMockRecorder, int)
		pkgRec     func(*fake_package.MockServiceMockRecorder)
		statusCode int

		want         int
		wantErr      string
		retryableErr bool
	}
	for _, tt := range []test{
		{
			name: "Updates preservation action tasks",
			jobRec: func(m *amclienttest.MockJobsServiceMockRecorder, statusCode int) {
				m.List(
					mockutil.Context(),
					unitID,
					&amclient.JobsListRequest{},
				).Return(
					jobs,
					&amclient.Response{
						Response: &http.Response{StatusCode: http.StatusOK, Status: "200 OK"},
					},
					nil,
				)
			},
			pkgRec: func(m *fake_package.MockServiceMockRecorder) {
				for _, job := range jobs {
					pt := am.ConvertJobToPreservationTask(job)
					pt.PreservationActionID = paID
					pt.StartedAt = nullTime
					pt.CompletedAt = nullTime
					m.CreatePreservationTask(mockutil.Context(), &pt).Return(nil)
				}
			},
			want: 2,
		},
		{
			name: "Retryable error when AM returns 400 Bad Request",
			jobRec: func(m *amclienttest.MockJobsServiceMockRecorder, statusCode int) {
				// AM sometimes returns a "400 Bad Request" error when a
				// transfer is processing.
				m.List(
					mockutil.Context(),
					unitID,
					&amclient.JobsListRequest{},
				).Return(
					nil,
					&amclient.Response{
						Response: &http.Response{StatusCode: http.StatusBadRequest, Status: "400 Bad request"},
					},
					&amclient.ErrorResponse{
						Response: &http.Response{StatusCode: http.StatusBadRequest, Status: "400 Bad request"},
					},
				)
			},
			wantErr:      am.ErrBadRequest.Error(),
			retryableErr: true,
		},
		{
			name:       "Non-retryable error from http invalid credentials",
			jobRec:     httpError,
			statusCode: http.StatusUnauthorized,
			wantErr:    "invalid Archivematica credentials",
		},
	} {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			jobsSvc := amclienttest.NewMockJobsService(ctrl)
			if tt.jobRec != nil {
				tt.jobRec(jobsSvc.EXPECT(), tt.statusCode)
			}
			pkgSvc := fake_package.NewMockService(ctrl)
			if tt.pkgRec != nil {
				tt.pkgRec(pkgSvc.EXPECT())
			}

			pa := am.NewJobTracker(clock, jobsSvc, pkgSvc, paID)
			got, err := pa.SavePreservationTasks(context.Background(), unitID)
			if tt.wantErr != "" {
				assert.Error(t, err, tt.wantErr)
				assert.Assert(t, temporal_tools.NonRetryableError(err) != tt.retryableErr)

				return
			}

			assert.Equal(t, got, tt.want)
		})
	}
}
