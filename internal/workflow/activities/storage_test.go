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
	storage_fake "github.com/artefactual-sdps/enduro/internal/storage/fake"
	"github.com/artefactual-sdps/enduro/internal/workflow/activities"
)

var (
	aipID      = uuid.MustParse("488c64cc-d89b-4916-9131-c94152dfb12e")
	locationID = uuid.MustParse("a06a155c-9cf0-4416-a2b6-e90e58ef3186")
	objectKey  = uuid.MustParse("e2630293-a714-4787-ab6d-e68254a6fb6a")
)

func TestCreatePackageActivity(t *testing.T) {
	t.Parallel()

	type test struct {
		name      string
		params    *activities.CreateStoragePackageActivityParams
		mockCalls func(m *storage_fake.MockClientMockRecorder)
		want      *activities.CreateStoragePackageActivityResult
		wantErr   string
	}
	for _, tt := range []test{
		{
			name: "Creates a new package",
			params: &activities.CreateStoragePackageActivityParams{
				Name:       "Package 1",
				AIPID:      aipID.String(),
				ObjectKey:  objectKey.String(),
				Status:     "stored",
				LocationID: &locationID,
			},
			mockCalls: func(m *storage_fake.MockClientMockRecorder) {
				m.Create(
					mockutil.Context(),
					&goastorage.CreatePayload{
						Name:       "Package 1",
						AipID:      aipID.String(),
						ObjectKey:  objectKey.String(),
						Status:     "stored",
						LocationID: ref.New(locationID),
					},
				).Return(
					&goastorage.Package{
						Name:       "Package 1",
						AipID:      aipID,
						ObjectKey:  objectKey,
						Status:     "stored",
						LocationID: ref.New(locationID),
						CreatedAt:  "2024-05-03 16:02:25",
					},
					nil,
				)
			},
			want: &activities.CreateStoragePackageActivityResult{
				CreatedAt: "2024-05-03 16:02:25",
			},
		},
		{
			name: "Errors on invalid AIP ID",
			params: &activities.CreateStoragePackageActivityParams{
				Name:       "Package 1",
				AIPID:      "12345",
				ObjectKey:  objectKey.String(),
				Status:     "stored",
				LocationID: &locationID,
			},
			mockCalls: func(m *storage_fake.MockClientMockRecorder) {
				m.Create(
					mockutil.Context(),
					&goastorage.CreatePayload{
						Name:       "Package 1",
						AipID:      "12345",
						ObjectKey:  objectKey.String(),
						Status:     "stored",
						LocationID: ref.New(locationID),
					},
				).Return(
					nil, goastorage.MakeNotValid(errors.New("invalid aip_id")),
				)
			},
			wantErr: "activity error (type: create-storage-package-activity, scheduledEventID: 0, startedEventID: 0, identity: ): create-storage-package-activity: invalid aip_id",
		},
		{
			name: "Errors on invalid authorization",
			params: &activities.CreateStoragePackageActivityParams{
				Name:       "Package 1",
				AIPID:      aipID.String(),
				ObjectKey:  objectKey.String(),
				Status:     "stored",
				LocationID: &locationID,
			},
			mockCalls: func(m *storage_fake.MockClientMockRecorder) {
				m.Create(
					mockutil.Context(),
					&goastorage.CreatePayload{
						Name:       "Package 1",
						AipID:      aipID.String(),
						ObjectKey:  objectKey.String(),
						Status:     "stored",
						LocationID: ref.New(locationID),
					},
				).Return(
					nil, goastorage.Unauthorized("unauthorized"),
				)
			},
			wantErr: "activity error (type: create-storage-package-activity, scheduledEventID: 0, startedEventID: 0, identity: ): create-storage-package-activity: Invalid token",
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ts := &temporalsdk_testsuite.WorkflowTestSuite{}
			env := ts.NewTestActivityEnvironment()
			mockClient := storage_fake.NewMockClient(gomock.NewController(t))
			if tt.mockCalls != nil {
				tt.mockCalls(mockClient.EXPECT())
			}

			env.RegisterActivityWithOptions(
				activities.NewCreateStoragePackageActivity(mockClient).Execute,
				temporalsdk_activity.RegisterOptions{
					Name: activities.CreateStoragePackageActivityName,
				},
			)

			enc, err := env.ExecuteActivity(activities.CreateStoragePackageActivityName, tt.params)
			if tt.wantErr != "" {
				assert.Error(t, err, tt.wantErr)
				assert.Assert(t, temporal.NonRetryableError(err))
				return
			}
			assert.NilError(t, err)

			var res activities.CreateStoragePackageActivityResult
			_ = enc.Get(&res)
			assert.DeepEqual(t, &res, tt.want)
		})
	}
}
