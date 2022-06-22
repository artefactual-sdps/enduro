package activities

import (
	"context"
	"errors"
	"net/http"
	"os"
	"time"

	goastorage "github.com/artefactual-labs/enduro/internal/api/gen/storage"
)

type UploadActivityParams struct {
	AIPPath string
	AIPID   string
	Name    string
}

type UploadActivity struct {
	storageClient *goastorage.Client
}

func NewUploadActivity(storageClient *goastorage.Client) *UploadActivity {
	return &UploadActivity{
		storageClient: storageClient,
	}
}

func (a *UploadActivity) Execute(ctx context.Context, params *UploadActivityParams) error {
	childCtx, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()
	res, err := a.storageClient.Submit(childCtx, &goastorage.SubmitPayload{
		AipID: params.AIPID,
		Name:  params.Name,
	})
	if err != nil {
		return err
	}

	// Upload to MinIO using the upload pre-signed URL.
	{
		f, err := os.Open(params.AIPPath)
		if err != nil {
			return err
		}
		defer f.Close()

		uploadReq, err := http.NewRequestWithContext(ctx, http.MethodPut, res.URL, f)
		if err != nil {
			return nil
		}
		fi, err := f.Stat()
		if err != nil {
			return err
		}
		uploadReq.ContentLength = fi.Size()
		if err != nil {
			return err
		}

		minioClient := &http.Client{}
		resp, err := minioClient.Do(uploadReq)
		if err != nil {
			return err
		}
		if resp.StatusCode != http.StatusOK {
			return errors.New("unexpected status code returned")
		}
	}

	childCtx, cancel = context.WithTimeout(ctx, time.Second*5)
	defer cancel()
	err = a.storageClient.Update(childCtx, &goastorage.UpdatePayload{AipID: params.AIPID})

	return err
}
