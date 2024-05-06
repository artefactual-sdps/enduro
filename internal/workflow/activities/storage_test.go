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
				AIPID:      "88e5a5fc-5c74-40e8-aa50-d6542b45f251",
				ObjectKey:  "3776a8f2-5ad5-4b4e-80c7-a888d9229cc1",
				Status:     "stored",
				LocationID: "92c3934d-8911-4ce4-8a44-ffa72a0b5720",
			},
			mockCalls: func(m *storage_fake.MockClientMockRecorder) {
				m.Create(
					mockutil.Context(),
					&goastorage.CreatePayload{
						Name:       "Package 1",
						AipID:      "88e5a5fc-5c74-40e8-aa50-d6542b45f251",
						ObjectKey:  "3776a8f2-5ad5-4b4e-80c7-a888d9229cc1",
						Status:     "stored",
						LocationID: ref.New(uuid.MustParse("92c3934d-8911-4ce4-8a44-ffa72a0b5720")),
					},
				).Return(
					&goastorage.Package{
						Name:       "Package 1",
						AipID:      uuid.MustParse("88e5a5fc-5c74-40e8-aa50-d6542b45f251"),
						ObjectKey:  uuid.MustParse("3776a8f2-5ad5-4b4e-80c7-a888d9229cc1"),
						Status:     "stored",
						LocationID: ref.New(uuid.MustParse("92c3934d-8911-4ce4-8a44-ffa72a0b5720")),
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
			name: "Errors on invalid locationID",
			params: &activities.CreateStoragePackageActivityParams{
				Name:       "Package 1",
				AIPID:      "88e5a5fc-5c74-40e8-aa50-d6542b45f251",
				ObjectKey:  "3776a8f2-5ad5-4b4e-80c7-a888d9229cc1",
				Status:     "stored",
				LocationID: "12345",
			},
			wantErr: "activity error (type: create-storage-package-activity, scheduledEventID: 0, startedEventID: 0, identity: ): create-storage-package-activity: invalid location ID: invalid UUID length: 5",
		},
		{
			name: "Errors on invalid AIP ID",
			params: &activities.CreateStoragePackageActivityParams{
				Name:       "Package 1",
				AIPID:      "12345",
				ObjectKey:  "3776a8f2-5ad5-4b4e-80c7-a888d9229cc1",
				Status:     "stored",
				LocationID: "92c3934d-8911-4ce4-8a44-ffa72a0b5720",
			},
			mockCalls: func(m *storage_fake.MockClientMockRecorder) {
				m.Create(
					mockutil.Context(),
					&goastorage.CreatePayload{
						Name:       "Package 1",
						AipID:      "12345",
						ObjectKey:  "3776a8f2-5ad5-4b4e-80c7-a888d9229cc1",
						Status:     "stored",
						LocationID: ref.New(uuid.MustParse("92c3934d-8911-4ce4-8a44-ffa72a0b5720")),
					},
				).Return(
					nil, goastorage.MakeNotValid(errors.New("invalid request")),
				)
			},
			wantErr: "activity error (type: create-storage-package-activity, scheduledEventID: 0, startedEventID: 0, identity: ): create-storage-package-activity: invalid request",
		},
		{
			name: "Errors on invalid authorization",
			params: &activities.CreateStoragePackageActivityParams{
				Name:       "Package 1",
				AIPID:      "88e5a5fc-5c74-40e8-aa50-d6542b45f251",
				ObjectKey:  "3776a8f2-5ad5-4b4e-80c7-a888d9229cc1",
				Status:     "stored",
				LocationID: "92c3934d-8911-4ce4-8a44-ffa72a0b5720",
			},
			mockCalls: func(m *storage_fake.MockClientMockRecorder) {
				m.Create(
					mockutil.Context(),
					&goastorage.CreatePayload{
						Name:       "Package 1",
						AipID:      "88e5a5fc-5c74-40e8-aa50-d6542b45f251",
						ObjectKey:  "3776a8f2-5ad5-4b4e-80c7-a888d9229cc1",
						Status:     "stored",
						LocationID: ref.New(uuid.MustParse("92c3934d-8911-4ce4-8a44-ffa72a0b5720")),
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
