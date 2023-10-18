package am

import (
	"context"
	"fmt"
	"os"

	"github.com/artefactual-sdps/enduro/internal/sftp"
)

const UploadTransferActivityName = "UploadTransferActivity"

type UploadTransferActivityParams struct {
	FullPath string
	Filename string
}

type UploadTransferActivityResult struct {
	BytesCopied int64
	RemotePath  string
}

type UploadTransferActivity struct {
	sftpSvc sftp.Service
}

func NewUploadTransferActivity(svc sftp.Service) *UploadTransferActivity {
	return &UploadTransferActivity{
		sftpSvc: svc,
	}
}

func (a *UploadTransferActivity) Execute(ctx context.Context, params *UploadTransferActivityParams) (*UploadTransferActivityResult, error) {
	src, err := os.Open(params.FullPath)
	if err != nil {
		return nil, fmt.Errorf("upload transfer: %v", err)
	}
	defer src.Close()

	bytes, path, err := a.sftpSvc.Upload(src, params.Filename)
	if err != nil {
		return nil, fmt.Errorf("upload transfer: %v", err)
	}

	return &UploadTransferActivityResult{
		BytesCopied: bytes,
		RemotePath:  path,
	}, nil
}
