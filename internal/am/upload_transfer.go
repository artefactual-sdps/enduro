package am

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/go-logr/logr"
	"go.artefactual.dev/tools/temporal"
	temporalsdk_activity "go.temporal.io/sdk/activity"

	"github.com/artefactual-sdps/enduro/internal/sftp"
)

const UploadTransferActivityName = "UploadTransferActivity"

type UploadTransferActivityParams struct {
	// Local path of the source file.
	SourcePath string
}

type UploadTransferActivityResult struct {
	// Bytes copied to the remote file over the SFTP connection.
	BytesCopied int64
	// Full path of the destination file including `remoteDir` config path.
	RemoteFullPath string
	// Path of the destination file relative to the `remoteDir` config path.
	RemoteRelativePath string
}

// UploadTransferActivity uploads a transfer via the SFTP client, and sends
// a periodic Temporal Heartbeat at the given heartRate.
type UploadTransferActivity struct {
	client    sftp.Client
	logger    logr.Logger
	heartRate time.Duration
}

// NewUploadTransferActivity initializes and returns a new
// UploadTransferActivity.
func NewUploadTransferActivity(
	logger logr.Logger,
	client sftp.Client,
	heartRate time.Duration,
) *UploadTransferActivity {
	return &UploadTransferActivity{
		client:    client,
		logger:    logger,
		heartRate: heartRate,
	}
}

// Execute copies the source transfer to the destination via SFTP.
func (a *UploadTransferActivity) Execute(ctx context.Context, params *UploadTransferActivityParams) (*UploadTransferActivityResult, error) {
	a.logger.V(1).Info("Execute UploadTransferActivity", "SourcePath", params.SourcePath)

	src, err := os.Open(params.SourcePath)
	if err != nil {
		return nil, fmt.Errorf("%s: %v", UploadTransferActivityName, err)
	}
	defer src.Close()

	filename := filepath.Base(params.SourcePath)
	path, upload, err := a.client.Upload(ctx, src, filename)
	if err != nil {
		e := fmt.Errorf("%s: %v", UploadTransferActivityName, err)

		switch err.(type) {
		case *sftp.AuthError:
			return nil, temporal.NewNonRetryableError(e)
		default:
			return nil, e
		}
	}

	fi, err := src.Stat()
	if err != nil {
		return nil, fmt.Errorf("%s: %v", UploadTransferActivityName, err)
	}

	// Block (with a heartbeat) until ctx is cancelled, the upload is done, or
	// it stops with an error.
	err = a.Heartbeat(ctx, upload, fi.Size())
	if err != nil {
		return nil, err
	}

	return &UploadTransferActivityResult{
		BytesCopied:        upload.Bytes(),
		RemoteFullPath:     path,
		RemoteRelativePath: filename,
	}, nil
}

// Heartbeat sends a periodic Temporal heartbeat, which includes the number of
// bytes uploaded, until the upload is complete, cancelled or returns an error.
func (a *UploadTransferActivity) Heartbeat(ctx context.Context, upload sftp.AsyncUpload, fileSize int64) error {
	ticker := time.NewTicker(a.heartRate)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case err := <-upload.Err():
			return err
		case <-upload.Done():
			return nil
		case <-ticker.C:
			temporalsdk_activity.RecordHeartbeat(ctx,
				fmt.Sprintf("Uploaded %d bytes of %d.", upload.Bytes(), fileSize),
			)
		}
	}
}
