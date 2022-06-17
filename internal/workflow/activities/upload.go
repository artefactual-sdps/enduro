package activities

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"time"

	goahttp "goa.design/goa/v3/http"

	"github.com/artefactual-labs/enduro/internal/api/gen/http/storage/client"
	goastorage "github.com/artefactual-labs/enduro/internal/api/gen/storage"
)

type UploadActivity struct {
}

func NewUploadActivity() *UploadActivity {
	return &UploadActivity{}
}

func (a *UploadActivity) Execute(ctx context.Context, AIPPath string) error {
	doer := &http.Client{Timeout: time.Second}
	c := client.NewClient("http", "enduro:9000", doer, goahttp.RequestEncoder, goahttp.ResponseDecoder, false)

	submitEndpoint := c.Submit()
	submitResponseData, err := submitEndpoint(ctx, nil)
	if err != nil {
		return err
	}

	sr, ok := submitResponseData.(*goastorage.SubmitResult)
	if !ok {
		return errors.New("unexpected value from Submit endpoint")
	}

	f, err := os.Open(AIPPath)
	if err != nil {
		return err
	}
	defer f.Close()

	httpClient := &http.Client{}
	uploadReq, err := http.NewRequest(http.MethodPut, sr.URL, f)
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
	_, err = httpClient.Do(uploadReq)
	if err != nil {
		return err
	}

	updateEndpoint := c.Update()
	updateData, err := client.BuildUpdatePayload(fmt.Sprintf("{\"workflow_id\": \"%s\"}", sr.WorkflowID))
	if err != nil {
		return err
	}
	updateResponseData, err := updateEndpoint(ctx, updateData)
	if err != nil {
		return err
	}
	_, ok = updateResponseData.(*goastorage.UpdateResult)
	if !ok {
		return errors.New("unexpected value from Update endpoint")
	}
	return nil
}
