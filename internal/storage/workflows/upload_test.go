package workflows

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	temporalsdk_testsuite "go.temporal.io/sdk/testsuite"

	"github.com/artefactual-sdps/enduro/internal/storage"
)

func TestStorageUploadWorkflow(t *testing.T) {
	s := temporalsdk_testsuite.WorkflowTestSuite{}
	env := s.NewTestWorkflowEnvironment()

	// Signal handler
	env.RegisterDelayedCallback(
		func() {
			env.SignalWorkflow(
				storage.UploadDoneSignalName,
				storage.UploadDoneSignal{},
			)
		},
		0,
	)

	env.ExecuteWorkflow(
		NewStorageUploadWorkflow().Execute,
		storage.StorageUploadWorkflowRequest{
			AIPID: uuid.NewString(),
		},
	)

	require.True(t, env.IsWorkflowCompleted())
	err := env.GetWorkflowResult(nil)
	require.NoError(t, err)
}
