package localact_test

import (
	"database/sql"
	"errors"
	"math/rand"
	"testing"
	"time"

	"github.com/google/uuid"
	"go.artefactual.dev/tools/mockutil"
	temporalsdk_testsuite "go.temporal.io/sdk/testsuite"
	"go.uber.org/mock/gomock"
	"gotest.tools/v3/assert"

	"github.com/artefactual-sdps/enduro/internal/childwf"
	"github.com/artefactual-sdps/enduro/internal/datatypes"
	"github.com/artefactual-sdps/enduro/internal/enums"
	ingest_fake "github.com/artefactual-sdps/enduro/internal/ingest/fake"
	"github.com/artefactual-sdps/enduro/internal/workflow/localact"
)

func TestSavePreprocessingTasksActivity(t *testing.T) {
	t.Parallel()

	wUUID := uuid.New()
	taskUUID := uuid.MustParse("52fdfc07-2182-454f-963f-5f0f9a621d72")
	startedAt := time.Date(2024, 6, 13, 17, 50, 13, 0, time.UTC)
	completedAt := time.Date(2024, 6, 13, 17, 50, 14, 0, time.UTC)

	type test struct {
		name      string
		params    localact.SavePreprocessingTasksActivityParams
		mockCalls func(m *ingest_fake.MockServiceMockRecorder)
		want      *localact.SavePreprocessingTasksActivityResult
		wantErr   string
	}
	for _, tt := range []test{
		{
			name: "Saves a preprocessing task",
			params: localact.SavePreprocessingTasksActivityParams{
				WorkflowUUID: wUUID,
				Tasks: []childwf.Task{
					{
						Name:        "Validate SIP structure",
						Message:     "SIP structure matches validation criteria",
						Outcome:     enums.PreprocessingTaskOutcomeSuccess,
						StartedAt:   startedAt,
						CompletedAt: completedAt,
					},
				},
			},
			mockCalls: func(m *ingest_fake.MockServiceMockRecorder) {
				m.CreateTasks(
					mockutil.Context(),
					[]*datatypes.Task{
						{
							UUID:         taskUUID,
							Name:         "Validate SIP structure",
							Status:       enums.TaskStatusDone,
							StartedAt:    sql.NullTime{Time: startedAt, Valid: true},
							CompletedAt:  sql.NullTime{Time: completedAt, Valid: true},
							Note:         "SIP structure matches validation criteria",
							WorkflowUUID: wUUID,
						},
					},
				).Return(nil)
			},
			want: &localact.SavePreprocessingTasksActivityResult{
				Count: 1,
			},
		},
		{
			name: "Errors when a required value is missing",
			params: localact.SavePreprocessingTasksActivityParams{
				WorkflowUUID: wUUID,
				Tasks: []childwf.Task{
					{
						Message:     "SIP structure matches validation criteria",
						Outcome:     enums.PreprocessingTaskOutcomeSuccess,
						StartedAt:   startedAt,
						CompletedAt: completedAt,
					},
				},
			},
			mockCalls: func(m *ingest_fake.MockServiceMockRecorder) {
				m.CreateTasks(
					mockutil.Context(),
					[]*datatypes.Task{
						{
							UUID:         taskUUID,
							Status:       enums.TaskStatusDone,
							StartedAt:    sql.NullTime{Time: startedAt, Valid: true},
							CompletedAt:  sql.NullTime{Time: completedAt, Valid: true},
							Note:         "SIP structure matches validation criteria",
							WorkflowUUID: wUUID,
						},
					},
				).Return(errors.New(
					"task: create: invalid data error: field Name is required",
				))
			},
			wantErr: "SavePreprocessingTasksActivity: task: create: invalid data error: field Name is required",
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ts := &temporalsdk_testsuite.WorkflowTestSuite{}
			env := ts.NewTestActivityEnvironment()
			svc := ingest_fake.NewMockService(gomock.NewController(t))
			if tt.mockCalls != nil {
				tt.mockCalls(svc.EXPECT())
			}
			tt.params.Ingestsvc = svc
			tt.params.RNG = rand.New(rand.NewSource(1)) // #nosec: G404

			enc, err := env.ExecuteLocalActivity(
				localact.SavePreprocessingTasksActivity,
				tt.params,
			)
			if tt.wantErr != "" {
				assert.Error(
					t,
					err,
					"activity error (type: SavePreprocessingTasksActivity, scheduledEventID: 0, startedEventID: 0, identity: ): "+tt.wantErr,
				)
				return
			}
			assert.NilError(t, err)

			var res localact.SavePreprocessingTasksActivityResult
			_ = enc.Get(&res)
			assert.DeepEqual(t, &res, tt.want)
		})
	}
}
