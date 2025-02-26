package activities

import (
	"context"
	"errors"
	"net/http"
	"os"
	"time"

	goastorage "github.com/artefactual-sdps/enduro/internal/api/gen/storage"
)

type UploadActivityParams struct {
	AIPPath string
	AIPID   string
	Name    string
}

type UploadActivity struct {
	storageClient *goastorage.Client
}

type UploadActivityResult struct{}

func NewUploadActivity(storageClient *goastorage.Client) *UploadActivity {
	return &UploadActivity{
		storageClient: storageClient,
	}
}

func (a *UploadActivity) Execute(ctx context.Context, params *UploadActivityParams) (*UploadActivityResult, error) {
	childCtx, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()
	res, err := a.storageClient.SubmitAip(childCtx, &goastorage.SubmitAipPayload{
		UUID: params.AIPID,
		Name: params.Name,
	})
	if err != nil {
		return &UploadActivityResult{}, err
	}

	// Upload to MinIO using the upload pre-signed URL.
	{
		f, err := os.Open(params.AIPPath)
		if err != nil {
			return &UploadActivityResult{}, err
		}
		defer f.Close() //#nosec G307 -- Errors returned by Close() here do not require specific handling.

		uploadReq, err := http.NewRequestWithContext(ctx, http.MethodPut, res.URL, f)
		if err != nil {
			return &UploadActivityResult{}, err
		}

		fi, err := f.Stat()
		if err != nil {
			return &UploadActivityResult{}, err
		}

		uploadReq.ContentLength = fi.Size()

		minioClient := &http.Client{}
		resp, err := minioClient.Do(uploadReq)
		if err != nil {
			return &UploadActivityResult{}, err
		}
		if resp.StatusCode != http.StatusOK {
			return &UploadActivityResult{}, errors.New("unexpected status code returned")
		}
	}

	childCtx, cancel = context.WithTimeout(ctx, time.Second*5)
	defer cancel()
	err = a.storageClient.UpdateAip(childCtx, &goastorage.UpdateAipPayload{UUID: params.AIPID})

	return &UploadActivityResult{}, err
}
