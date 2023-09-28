package am

import (
	context "context"
	"net/http"
	"time"

	"github.com/artefactual-sdps/enduro/internal/temporal"
	"go.artefactual.dev/amclient"
)

const defaultMaxElapsedTime = time.Minute * 10

type PollTransferParams struct {
	TransferID string
}
type PollTransferActivity struct {
	cfg *Config
}
type PollTransferResponse struct {
	SIPID string
}

func (a *PollTransferActivity) Execute(ctx context.Context, params *PollTransferParams) (*PollTransferResponse, error) {
	client := http.Client{}

	c := amclient.NewClient(&client, a.cfg.Address, a.cfg.User, a.cfg.Key)
	resp, httpResp, err := c.Transfer.Status(ctx, params.TransferID)
	if err != nil {
		return nil, temporal.NonRetryableError(err)
	}
	if amclient.CheckResponse(httpResp.Response) != nil {
		return nil, temporal.NonRetryableError(err)
	}

	if resp.SIPID == "" {
		return nil, temporal.ContinuePollingError()
	}

	return &PollTransferResponse{SIPID: resp.SIPID}, err
}
