package am_test

import (
	"context"
	"net/http"
	"testing"
	"time"

	"entgo.io/ent/dialect/sql"
	"github.com/google/uuid"
	"github.com/jonboulle/clockwork"
	"go.artefactual.dev/amclient"
	"go.artefactual.dev/amclient/amclienttest"
	"go.artefactual.dev/tools/mockutil"
	temporal_tools "go.artefactual.dev/tools/temporal"
	"go.uber.org/mock/gomock"
	"gotest.tools/v3/assert"

	"github.com/artefactual-sdps/enduro/internal/am"
	"github.com/artefactual-sdps/enduro/internal/datatypes"
	"github.com/artefactual-sdps/enduro/internal/enums"
	fake_package "github.com/artefactual-sdps/enduro/internal/package_/fake"
)

func TestJobTracker(t *testing.T) {
	t.Parallel()

	paID := uint(1)
	unitID := uuid.New().String()

	clock := clockwork.NewFakeClock()
	httpError := func(m *amclienttest.MockJobsServiceMockRecorder, statusCode int) {
		m.List(
			mockutil.Context(),
			unitID,
			&amclient.JobsListRequest{
				Detailed: true,
			},
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
					ID:          "c134198c-9485-4f68-8d94-4da1e03b5e1b",
					ExitCode:    0,
					CreatedAt:   amclient.TaskDateTime{Time: time.Date(2024, time.January, 18, 1, 27, 49, 0, time.UTC)},
					StartedAt:   amclient.TaskDateTime{Time: time.Date(2024, time.January, 18, 1, 27, 49, 0, time.UTC)},
					CompletedAt: amclient.TaskDateTime{Time: time.Date(2024, time.January, 18, 1, 27, 49, 0, time.UTC)},
					Duration:    amclient.TaskDuration(time.Second / 2),
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
					ID:          "6f5beca3-71ad-446c-8f19-3bc4dea16c9b",
					ExitCode:    0,
					CreatedAt:   amclient.TaskDateTime{Time: time.Date(2024, time.January, 1, 1, 27, 49, 0, time.UTC)},
					StartedAt:   amclient.TaskDateTime{Time: time.Date(2024, time.January, 18, 1, 27, 49, 0, time.UTC)},
					CompletedAt: amclient.TaskDateTime{Time: time.Date(2024, time.January, 18, 1, 27, 49, 0, time.UTC)},
					Duration:    amclient.TaskDuration(time.Minute),
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
					&amclient.JobsListRequest{
						Detailed: true,
					},
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
					&amclient.JobsListRequest{
						Detailed: true,
					},
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

func TestConvertJobToPreservationTask(t *testing.T) {
	t.Parallel()

	type test struct {
		name string
		job  amclient.Job
		want datatypes.PreservationTask
	}

	for _, tt := range []test{
		{
			name: "Returns preservation task with computed time range",
			job: amclient.Job{
				ID:           "f60018ac-da79-4769-9509-c6c41d5efe7e",
				LinkID:       "70669a5b-01e4-4ea0-ac70-10292f87da05",
				Microservice: "Verify SIP compliance",
				Name:         "Move to processing directory",
				Status:       amclient.JobStatusComplete,
				Tasks: []amclient.Task{
					{
						ID:          "c134198c-9485-4f68-8d94-4da1e03b5e1b",
						ExitCode:    0,
						CreatedAt:   amclient.TaskDateTime{Time: time.Date(2024, time.January, 18, 1, 27, 49, 0, time.UTC)},
						StartedAt:   amclient.TaskDateTime{Time: time.Date(2024, time.January, 18, 1, 27, 49, 0, time.UTC)},
						CompletedAt: amclient.TaskDateTime{Time: time.Date(2025, time.January, 18, 1, 27, 49, 0, time.UTC)},
						Duration:    amclient.TaskDuration(time.Second / 2),
					},
					{
						ID:          "6e5edf16-ff93-47c0-a7d1-e623c110fa09",
						ExitCode:    0,
						CreatedAt:   amclient.TaskDateTime{Time: time.Date(2024, time.January, 18, 1, 27, 49, 0, time.UTC)},
						StartedAt:   amclient.TaskDateTime{Time: time.Date(2025, time.January, 18, 1, 27, 49, 0, time.UTC)},
						CompletedAt: amclient.TaskDateTime{Time: time.Date(2026, time.January, 18, 1, 27, 49, 0, time.UTC)},
						Duration:    amclient.TaskDuration(time.Second / 2),
					},
				},
			},
			want: datatypes.PreservationTask{
				TaskID: "f60018ac-da79-4769-9509-c6c41d5efe7e",
				Name:   "Move to processing directory",
				Status: enums.PreservationTaskStatus(enums.PackageStatusDone),
				StartedAt: sql.NullTime{
					Time:  time.Date(2024, time.January, 18, 1, 27, 49, 0, time.UTC),
					Valid: true,
				},
				CompletedAt: sql.NullTime{
					Time:  time.Date(2026, time.January, 18, 1, 27, 49, 0, time.UTC),
					Valid: true,
				},
			},
		},
		{
			name: "Returns NULL completedAt if job is still processing",
			job: amclient.Job{
				ID:           "c2128d39-2ace-47c5-8cac-39ded8d9c9ef",
				LinkID:       "208d441b-6938-44f9-b54a-bd73f05bc764",
				Microservice: "Verify SIP compliance",
				Name:         "Verify SIP compliance",
				Status:       amclient.JobStatusProcessing,
				Tasks: []amclient.Task{
					{
						ID:          "6f5beca3-71ad-446c-8f19-3bc4dea16c9b",
						ExitCode:    0,
						CreatedAt:   amclient.TaskDateTime{Time: time.Date(2024, time.January, 18, 1, 27, 49, 0, time.UTC)},
						StartedAt:   amclient.TaskDateTime{Time: time.Date(2024, time.January, 18, 1, 27, 49, 0, time.UTC)},
						CompletedAt: amclient.TaskDateTime{Time: time.Time{}},
						Duration:    amclient.TaskDuration(time.Second / 2),
					},
					{
						ID:          "6f5beca3-71ad-446c-8f19-3bc4dea16c9b",
						ExitCode:    0,
						CreatedAt:   amclient.TaskDateTime{Time: time.Date(2025, time.January, 18, 1, 27, 49, 0, time.UTC)},
						StartedAt:   amclient.TaskDateTime{Time: time.Date(2025, time.January, 18, 1, 27, 49, 0, time.UTC)},
						CompletedAt: amclient.TaskDateTime{Time: time.Time{}},
						Duration:    amclient.TaskDuration(time.Second / 2),
					},
				},
			},
			want: datatypes.PreservationTask{
				TaskID: "c2128d39-2ace-47c5-8cac-39ded8d9c9ef",
				Name:   "Verify SIP compliance",
				Status: enums.PreservationTaskStatus(enums.PackageStatusInProgress),
				StartedAt: sql.NullTime{
					Time:  time.Date(2024, time.January, 18, 1, 27, 49, 0, time.UTC),
					Valid: true,
				},
				CompletedAt: sql.NullTime{},
			},
		},
		{
			name: "Returns NULL timestamps in the job has no tasks",
			job:  amclient.Job{},
			want: datatypes.PreservationTask{},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := am.ConvertJobToPreservationTask(tt.job)

			assert.DeepEqual(t, got, tt.want)
		})
	}
}
