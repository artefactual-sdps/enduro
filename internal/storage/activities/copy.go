package activities

import (
	"context"
	"io"

	"github.com/artefactual-labs/enduro/internal/storage"
)

type CopyToPermanentLocationActivity struct {
	storagesvc storage.Service
}

func NewCopyToPermanentLocationActivity(storagesvc storage.Service) *CopyToPermanentLocationActivity {
	return &CopyToPermanentLocationActivity{storagesvc: storagesvc}
}

func (a *CopyToPermanentLocationActivity) Execute(ctx context.Context, params *storage.CopyToPermanentLocationActivityParams) error {
	p, err := a.storagesvc.ReadPackage(ctx, params.AIPID)
	if err != nil {
		return err
	}

	reader, err := a.storagesvc.Bucket().NewReader(ctx, p.ObjectKey, nil)
	if err != nil {
		return err
	}
	defer reader.Close()

	l, err := a.storagesvc.Location(params.Location)
	if err != nil {
		return err
	}

	bucket, err := l.OpenBucket()
	if err != nil {
		return err
	}
	defer bucket.Close()

	writer, err := bucket.NewWriter(ctx, p.ObjectKey, nil)
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
