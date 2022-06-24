package workflows

import (
	"context"
	"io"

	temporalsdk_workflow "go.temporal.io/sdk/workflow"

	"github.com/artefactual-labs/enduro/internal/storage"
)

type StorageMoveWorkflow struct {
	storagesvc storage.Service
}

func NewStorageMoveWorkflow(storagesvc storage.Service) *StorageMoveWorkflow {
	return &StorageMoveWorkflow{
		storagesvc: storagesvc,
	}
}

func copyToPermanentLocation(ctx context.Context, storagesvc storage.Service, AIPID, location string) error {
	p, err := storagesvc.ReadPackage(ctx, AIPID)
	if err != nil {
		return err
	}

	reader, err := storagesvc.Bucket().NewReader(ctx, p.ObjectKey, nil)
	if err != nil {
		return err
	}
	defer reader.Close()

	l, err := storagesvc.Location(location)
	if err != nil {
		return err
	}

	bucket, err := l.OpenBucket()
	if err != nil {
		return err
	}
	defer bucket.Close()

	// XXX: what key should we use for the permanent location?
	writer, err := bucket.NewWriter(ctx, p.AIPID, nil)
	if err != nil {
		return err
	}

	_, copyErr := io.Copy(writer, reader)
	closeErr := writer.Close()

	if copyErr != nil {
		return copyErr
	}
	if closeErr != nil {
		return closeErr
	}

	return nil
}

func (w *StorageMoveWorkflow) Execute(ctx temporalsdk_workflow.Context, req storage.StorageMoveWorkflowRequest) error {
	// XXX: how do we get a regular context from the temporal one?
	childCtx := context.Background()
	err := copyToPermanentLocation(childCtx, w.storagesvc, req.AIPID, req.Location)
	if err != nil {
		return err
	}

	// XXX: should we delete the package from the internal aips bucket here?

	err = w.storagesvc.UpdatePackageLocation(childCtx, req.Location, req.AIPID)
	if err != nil {
		return err
	}

	err = w.storagesvc.UpdatePackageStatus(childCtx, storage.StatusStored, req.AIPID)
	if err != nil {
		return err
	}

	return nil
}
