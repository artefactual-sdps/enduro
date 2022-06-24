package workflows

import (
	temporalsdk_workflow "go.temporal.io/sdk/workflow"

	"github.com/artefactual-labs/enduro/internal/storage"
)

type StorageUploadWorkflow struct{}

func NewStorageUploadWorkflow() *StorageUploadWorkflow {
	return &StorageUploadWorkflow{}
}

func (w *StorageUploadWorkflow) Execute(ctx temporalsdk_workflow.Context, req storage.StorageUploadWorkflowRequest) error {
	var signal storage.UploadDoneSignal
	timerFuture := temporalsdk_workflow.NewTimer(ctx, storage.SubmitURLExpirationTime)
	signalChan := temporalsdk_workflow.GetSignalChannel(ctx, storage.UploadDoneSignalName)
	selector := temporalsdk_workflow.NewSelector(ctx)
	selector.AddReceive(signalChan, func(channel temporalsdk_workflow.ReceiveChannel, more bool) {
		_ = channel.Receive(ctx, &signal)
	})
	selector.AddFuture(timerFuture, func(f temporalsdk_workflow.Future) {
		_ = f.Get(ctx, nil)
	})
	selector.Select(ctx)
	return nil
}
