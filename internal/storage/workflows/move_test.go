package workflows

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	temporalsdk_activity "go.temporal.io/sdk/activity"
	temporalsdk_testsuite "go.temporal.io/sdk/testsuite"
	"go.uber.org/mock/gomock"

	"github.com/artefactual-sdps/enduro/internal/storage"
	"github.com/artefactual-sdps/enduro/internal/storage/activities"
	"github.com/artefactual-sdps/enduro/internal/storage/fake"
	"github.com/artefactual-sdps/enduro/internal/storage/types"
)

func TestStorageMoveWorkflow(t *testing.T) {
	s := temporalsdk_testsuite.WorkflowTestSuite{}
	env := s.NewTestWorkflowEnvironment()

	aipID := uuid.New()
	locationID := uuid.MustParse("e7452225-53d6-46f3-9f90-d0f2ee18b7cd")

	// Mock services and their expected calls
	ctrl := gomock.NewController(t)
	storagesvc := fake.NewMockService(ctrl)
	storagesvc.EXPECT().Delete(gomock.Any(), aipID)
	storagesvc.EXPECT().UpdatePackageLocationID(gomock.Any(), aipID, locationID)
	storagesvc.EXPECT().UpdatePackageStatus(gomock.Any(), aipID, types.StatusMoving)
	storagesvc.EXPECT().UpdatePackageStatus(gomock.Any(), aipID, types.StatusStored)

	// Worker activities
	env.RegisterActivityWithOptions(activities.NewCopyToPermanentLocationActivity(storagesvc).Execute, temporalsdk_activity.RegisterOptions{Name: storage.CopyToPermanentLocationActivityName})
	env.OnActivity(storage.CopyToPermanentLocationActivityName, mock.Anything, mock.Anything).Return(nil, nil)

	env.ExecuteWorkflow(
		NewStorageMoveWorkflow(storagesvc).Execute,
		storage.StorageMoveWorkflowRequest{
			AIPID:      aipID,
			LocationID: locationID,
			TaskQueue:  "global",
		},
	)

	require.True(t, env.IsWorkflowCompleted())
	err := env.GetWorkflowResult(nil)
	require.NoError(t, err)
}
