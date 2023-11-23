package am

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/go-logr/logr"

	"github.com/artefactual-sdps/enduro/internal/sftp"
)

const UploadTransferActivityName = "UploadTransferActivity"

type UploadTransferActivityParams struct {
	SourcePath string
}

type UploadTransferActivityResult struct {
	BytesCopied int64
	RemotePath  string
}

type UploadTransferActivity struct {
	client sftp.Client
	logger logr.Logger
}

func NewUploadTransferActivity(logger logr.Logger, client sftp.Client) *UploadTransferActivity {
	return &UploadTransferActivity{client: client, logger: logger}
}

func (a *UploadTransferActivity) Execute(ctx context.Context, params *UploadTransferActivityParams) (*UploadTransferActivityResult, error) {
	a.logger.V(1).Info("Execute UploadTransferActivity", "SourcePath", params.SourcePath)

	src, err := os.Open(params.SourcePath)
	if err != nil {
		return nil, fmt.Errorf("%s: %v", UploadTransferActivityName, err)
	}
	defer src.Close()

	bytes, path, err := a.client.Upload(ctx, src, filepath.Base(params.SourcePath))
	if err != nil {
		return nil, fmt.Errorf("%s: %v", UploadTransferActivityName, err)
	}

	return &UploadTransferActivityResult{
		BytesCopied: bytes,
		RemotePath:  path,
	}, nil
}