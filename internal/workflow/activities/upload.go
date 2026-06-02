package activities

import (
	"context"
	"errors"
	"io"
	"os"
	"time"

	"gocloud.dev/blob"

	goastorage "github.com/artefactual-sdps/enduro/internal/api/gen/storage"
	"github.com/artefactual-sdps/enduro/internal/ingest"
	"github.com/artefactual-sdps/enduro/internal/storage/enums"
)

type UploadActivityParams struct {
	AIPPath string
	AIPID   string
	Name    string
}

type UploadActivity struct {
	storageClient    ingest.StorageClient
	aipStagingBucket *blob.Bucket
}

type UploadActivityResult struct{}

func NewUploadActivity(storageClient ingest.StorageClient, aipStagingBucket *blob.Bucket) *UploadActivity {
	return &UploadActivity{
		storageClient:    storageClient,
		aipStagingBucket: aipStagingBucket,
	}
}

func (a *UploadActivity) Execute(ctx context.Context, params *UploadActivityParams) (*UploadActivityResult, error) {
	if err := a.uploadToAIPStagingBucket(ctx, params); err != nil {
		return &UploadActivityResult{}, err
	}

	childCtx, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()
	_, err := a.storageClient.CreateAip(childCtx, &goastorage.CreateAipPayload{
		UUID:      params.AIPID,
		Name:      params.Name,
		ObjectKey: params.AIPID,
		Status:    enums.AIPStatusPending.String(),
	})

	return &UploadActivityResult{}, err
}

func (a *UploadActivity) uploadToAIPStagingBucket(ctx context.Context, params *UploadActivityParams) error {
	f, err := os.Open(params.AIPPath)
	if err != nil {
		return err
	}
	defer f.Close()

	w, err := a.aipStagingBucket.NewWriter(ctx, params.AIPID, nil)
	if err != nil {
		return err
	}

	_, copyErr := io.Copy(w, f)
	closeErr := w.Close()

	return errors.Join(copyErr, closeErr)
}
