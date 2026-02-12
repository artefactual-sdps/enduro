package activities_test

import (
	"errors"
	"testing"

	"github.com/google/uuid"
	"go.artefactual.dev/tools/mockutil"
	"go.artefactual.dev/tools/ref"
	"go.artefactual.dev/tools/temporal"
	temporalsdk_activity "go.temporal.io/sdk/activity"
	temporalsdk_testsuite "go.temporal.io/sdk/testsuite"
	"go.uber.org/mock/gomock"
	"gotest.tools/v3/assert"

	goastorage "github.com/artefactual-sdps/enduro/internal/api/gen/storage"
	ingest_fake "github.com/artefactual-sdps/enduro/internal/ingest/fake"
	"github.com/artefactual-sdps/enduro/internal/workflow/activities"
)

var (
	aipID      = uuid.MustParse("488c64cc-d89b-4916-9131-c94152dfb12e")
	locationID = uuid.MustParse("a06a155c-9cf0-4416-a2b6-e90e58ef3186")
	objectKey  = uuid.MustParse("e2630293-a714-4787-ab6d-e68254a6fb6a")
)

func TestCreateAIPActivity(t *testing.T) {
	t.Parallel()

	type test struct {
		name      string
		params    *activities.CreateStorageAIPActivityParams
		mockCalls func(m *ingest_fake.MockStorageClientMockRecorder)
		want      *activities.CreateStorageAIPActivityResult
		wantErr   string
	}
	for _, tt := range []test{
		{
			name: "Creates a new AIP",
			params: &activities.CreateStorageAIPActivityParams{
				Name:       "AIP 1",
				AIPID:      aipID.String(),
				ObjectKey:  objectKey.String(),
				Status:     "stored",
				LocationID: &locationID,
			},
			mockCalls: func(m *ingest_fake.MockStorageClientMockRecorder) {
				m.CreateAip(
					mockutil.Context(),
					&goastorage.CreateAipPayload{
						Name:         "AIP 1",
						UUID:         aipID.String(),
						ObjectKey:    objectKey.String(),
						Status:       "stored",
						LocationUUID: ref.New(locationID),
					},
				).Return(
					&goastorage.AIP{
						Name:         "AIP 1",
						UUID:         aipID,
						ObjectKey:    objectKey,
						Status:       "stored",
						LocationUUID: ref.New(locationID),
						CreatedAt:    "2024-05-03 16:02:25",
					},
					nil,
				)
			},
			want: &activities.CreateStorageAIPActivityResult{
				CreatedAt: "2024-05-03 16:02:25",
			},
		},
		{
			name: "Errors on invalid AIP ID",
			params: &activities.CreateStorageAIPActivityParams{
				Name:       "AIP 1",
				AIPID:      "12345",
				ObjectKey:  objectKey.String(),
				Status:     "stored",
				LocationID: &locationID,
			},
			mockCalls: func(m *ingest_fake.MockStorageClientMockRecorder) {
				m.CreateAip(
					mockutil.Context(),
					&goastorage.CreateAipPayload{
						Name:         "AIP 1",
						UUID:         "12345",
						ObjectKey:    objectKey.String(),
						Status:       "stored",
						LocationUUID: ref.New(locationID),
					},
				).Return(
					nil, goastorage.MakeNotValid(errors.New("invalid aip_id")),
				)
			},
			wantErr: "activity error (type: create-storage-aip-activity, scheduledEventID: 0, startedEventID: 0, identity: ): create-storage-aip-activity: invalid aip_id",
		},
		{
			name: "Errors on invalid authorization",
			params: &activities.CreateStorageAIPActivityParams{
				Name:       "AIP 1",
				AIPID:      aipID.String(),
				ObjectKey:  objectKey.String(),
				Status:     "stored",
				LocationID: &locationID,
			},
			mockCalls: func(m *ingest_fake.MockStorageClientMockRecorder) {
				m.CreateAip(
					mockutil.Context(),
					&goastorage.CreateAipPayload{
						Name:         "AIP 1",
						UUID:         aipID.String(),
						ObjectKey:    objectKey.String(),
						Status:       "stored",
						LocationUUID: ref.New(locationID),
					},
				).Return(
					nil, goastorage.Unauthorized("Unauthorized"),
				)
			},
			wantErr: "activity error (type: create-storage-aip-activity, scheduledEventID: 0, startedEventID: 0, identity: ): create-storage-aip-activity: Unauthorized",
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ts := &temporalsdk_testsuite.WorkflowTestSuite{}
			env := ts.NewTestActivityEnvironment()
			mockClient := ingest_fake.NewMockStorageClient(gomock.NewController(t))
			if tt.mockCalls != nil {
				tt.mockCalls(mockClient.EXPECT())
			}

			env.RegisterActivityWithOptions(
				activities.NewCreateStorageAIPActivity(mockClient).Execute,
				temporalsdk_activity.RegisterOptions{
					Name: activities.CreateStorageAIPActivityName,
				},
			)

			enc, err := env.ExecuteActivity(activities.CreateStorageAIPActivityName, tt.params)
			if tt.wantErr != "" {
				assert.Error(t, err, tt.wantErr)
				assert.Assert(t, temporal.NonRetryableError(err))
				return
			}
			assert.NilError(t, err)

			var res activities.CreateStorageAIPActivityResult
			_ = enc.Get(&res)
			assert.DeepEqual(t, &res, tt.want)
		})
	}
}
