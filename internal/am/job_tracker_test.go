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
	"go.opentelemetry.io/otel/trace/noop"
	"go.uber.org/mock/gomock"
	"gotest.tools/v3/assert"

	"github.com/artefactual-sdps/enduro/internal/am"
	"github.com/artefactual-sdps/enduro/internal/datatypes"
	"github.com/artefactual-sdps/enduro/internal/enums"
	ingest_fake "github.com/artefactual-sdps/enduro/internal/ingest/fake"
	"github.com/artefactual-sdps/enduro/internal/persistence"
)

func TestJobTracker(t *testing.T) {
	t.Parallel()

	wUUID := uuid.New()
	taskUUID := uuid.New()
	unitID := uuid.New().String()
	startedAt := time.Date(2024, time.January, 18, 1, 27, 49, 0, time.UTC)
	completedAt := time.Date(2024, time.January, 18, 1, 27, 51, 0, time.UTC)

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
			ID:     taskUUID.String(),
			LinkID: "3229e01f-adf3-4294-85f7-4acb01b3fbcf",
			Name:   "Extract zipped bag transfer",
			Status: amclient.JobStatusComplete,
			Tasks: []amclient.Task{
				{
					ID:          "c134198c-9485-4f68-8d94-4da1e03b5e1b",
					ExitCode:    0,
					CreatedAt:   amclient.TaskDateTime{Time: time.Date(2024, time.January, 18, 1, 27, 49, 0, time.UTC)},
					StartedAt:   amclient.TaskDateTime{Time: startedAt},
					CompletedAt: amclient.TaskDateTime{Time: completedAt},
					Duration:    amclient.TaskDuration(time.Second / 2),
				},
			},
		},
		{
			ID:     "c2128d39-2ace-47c5-8cac-39ded8d9c9ef",
			LinkID: "208d441b-6938-44f9-b54a-bd73f05bc764",
			Name:   "Verify bag, and restructure for compliance",
			Status: amclient.JobStatusComplete,
		},
	}

	type test struct {
		name       string
		jobRec     func(*amclienttest.MockJobsServiceMockRecorder, int)
		ingestRec  func(*ingest_fake.MockServiceMockRecorder)
		statusCode int

		want         int
		wantErr      string
		retryableErr bool
	}
	for _, tt := range []test{
		{
			name: "Updates workflow tasks",
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
			ingestRec: func(m *ingest_fake.MockServiceMockRecorder) {
				m.CreateTasks(mockutil.Context(), gomock.Any()).
					DoAndReturn(func(_ context.Context, seq persistence.TaskSequence) error {
						var got []*datatypes.Task
						seq(func(t *datatypes.Task) bool {
							got = append(got, t)
							return true
						})
						assert.Equal(t, len(got), 1)
						assert.DeepEqual(t, got[0], &datatypes.Task{
							ID:           0,
							UUID:         taskUUID,
							Name:         "Extract zipped bag transfer",
							Status:       enums.TaskStatusDone,
							StartedAt:    sql.NullTime{Time: startedAt, Valid: true},
							CompletedAt:  sql.NullTime{Time: completedAt, Valid: true},
							WorkflowUUID: wUUID,
						})
						return nil
					})
			},
			want: 1,
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
			ingestsvc := ingest_fake.NewMockService(ctrl)
			if tt.ingestRec != nil {
				tt.ingestRec(ingestsvc.EXPECT())
			}

			jt := am.NewJobTracker(clock, jobsSvc, ingestsvc, wUUID, noop.Tracer{})
			got, err := jt.SaveTasks(context.Background(), unitID)
			if tt.wantErr != "" {
				assert.Error(t, err, tt.wantErr)
				assert.Assert(t, temporal_tools.NonRetryableError(err) != tt.retryableErr)

				return
			}

			assert.Equal(t, got, tt.want)
		})
	}
}

func TestConvertJobToTask(t *testing.T) {
	t.Parallel()

	taskUUID := uuid.New()

	type test struct {
		name    string
		job     amclient.Job
		want    *datatypes.Task
		wantErr string
	}

	for _, tt := range []test{
		{
			name: "Returns task with computed time range",
			job: amclient.Job{
				ID:     taskUUID.String(),
				LinkID: "70669a5b-01e4-4ea0-ac70-10292f87da05",
				Name:   "Move to processing directory",
				Status: amclient.JobStatusComplete,
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
			want: &datatypes.Task{
				UUID:   taskUUID,
				Name:   "Move to processing directory",
				Status: enums.TaskStatusDone,
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
				ID:     taskUUID.String(),
				LinkID: "208d441b-6938-44f9-b54a-bd73f05bc764",
				Name:   "Verify SIP compliance",
				Status: amclient.JobStatusProcessing,
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
			want: &datatypes.Task{
				UUID:   taskUUID,
				Name:   "Verify SIP compliance",
				Status: enums.TaskStatusInProgress,
				StartedAt: sql.NullTime{
					Time:  time.Date(2024, time.January, 18, 1, 27, 49, 0, time.UTC),
					Valid: true,
				},
				CompletedAt: sql.NullTime{},
			},
		},
		{
			name: "Returns NULL timestamps if the job has no tasks",
			job: amclient.Job{
				ID:     taskUUID.String(),
				LinkID: "208d441b-6938-44f9-b54a-bd73f05bc764",
				Name:   "Verify SIP compliance",
				Status: amclient.JobStatusProcessing,
			},
			want: &datatypes.Task{
				UUID:   taskUUID,
				Name:   "Verify SIP compliance",
				Status: enums.TaskStatusInProgress,
			},
		},
		{
			name: "Errors on invalid jod ID",
			job: amclient.Job{
				ID:     "invalid-uuid",
				LinkID: "70669a5b-01e4-4ea0-ac70-10292f87da05",
				Name:   "Move to processing directory",
				Status: amclient.JobStatusComplete,
			},
			wantErr: "unable to parse task UUID from job ID: \"invalid-uuid\"",
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := am.ConvertJobToTask(tt.job)
			if tt.wantErr != "" {
				assert.ErrorContains(t, err, tt.wantErr)
				return
			}
			assert.NilError(t, err)
			assert.DeepEqual(t, got, tt.want)
		})
	}
}
