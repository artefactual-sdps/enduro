package am

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"

	"github.com/artefactual-sdps/enduro/internal/sftp"
)

const DeleteTransferActivityName = "DeleteTransferActivity"

type DeleteTransferActivityParams struct {
	Destination string
}

type DeleteTransferActivity struct {
	client sftp.Client
	logger logr.Logger
}

type DeleteTransferActivityResult struct{}

func NewDeleteTransferActivity(logger logr.Logger, client sftp.Client) *DeleteTransferActivity {
	return &DeleteTransferActivity{client: client, logger: logger}
}

func (a *DeleteTransferActivity) Execute(
	ctx context.Context,
	params *DeleteTransferActivityParams,
) (*DeleteTransferActivityResult, error) {
	a.logger.V(1).Info("Execute DeleteTransferActivity",
		"destination", params.Destination,
	)

	err := a.client.Delete(ctx, params.Destination)
	if err != nil {
		return &DeleteTransferActivityResult{}, fmt.Errorf("delete transfer: path: %q: %v", params.Destination, err)
	}

	return &DeleteTransferActivityResult{}, nil
}
