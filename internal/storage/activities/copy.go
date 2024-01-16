package activities

import (
	"context"
	"io"

	"github.com/artefactual-sdps/enduro/internal/storage"
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

	reader, _, err := a.storagesvc.PackageReader(ctx, p)
	if err != nil {
		return err
	}
	defer reader.Close()

	l, err := a.storagesvc.Location(ctx, params.LocationID)
	if err != nil {
		return err
	}

	bucket, err := l.OpenBucket(ctx)
	if err != nil {
		return err
	}
	defer bucket.Close()

	writer, err := bucket.NewWriter(ctx, params.AIPID.String(), nil)
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
