package activities

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"time"

	goahttp "goa.design/goa/v3/http"

	"github.com/artefactual-labs/enduro/internal/aipstore"
	"github.com/artefactual-labs/enduro/internal/api/gen/http/storage/client"
	"github.com/artefactual-labs/enduro/internal/api/gen/storage"
)

type UploadActivity struct {
	config aipstore.Config
}

func NewUploadActivity(config aipstore.Config) *UploadActivity {
	return &UploadActivity{config: config}
}

func (a *UploadActivity) Execute(ctx context.Context, AIPPath string) error {
	doer := &http.Client{Timeout: time.Second}
	c := client.NewClient("http", "enduro:9000", doer, goahttp.RequestEncoder, goahttp.ResponseDecoder, false)

	submitEndpoint := c.Submit()
	submitData, err := client.BuildSubmitPayload("{\"key\": \"foobar\"}")
	if err != nil {
		return err
	}
	submitResponseData, err := submitEndpoint(ctx, submitData)
	if err != nil {
		return err
	}

	sr, ok := submitResponseData.(*storage.SubmitResult)
	if !ok {
		return errors.New("unexpected value from Submit endpoint")
	}

	f, err := os.Open(AIPPath)
	if err != nil {
		return err
	}
	defer f.Close() // XXX is this needed if client.Do closes it?

	httpClient := &http.Client{} // XXX set timeout here?
	uploadReq, err := http.NewRequest(http.MethodPut, sr.URL, f)
	if err != nil {
		return nil
	}
	fi, err := f.Stat()
	if err != nil {
		return err
	}
	uploadReq.ContentLength = fi.Size() // XXX why is this not done by NewRequest?
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
	_, ok = updateResponseData.(*storage.UpdateResult)
	if !ok {
		return errors.New("unexpected value from Update endpoint")
	}
	return nil
}
