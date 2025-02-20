package am

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"time"

	"go.artefactual.dev/tools/temporal"
	temporal_tools "go.artefactual.dev/tools/temporal"
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
	heartRate time.Duration
}

// NewUploadTransferActivity initializes and returns a new
// UploadTransferActivity.
func NewUploadTransferActivity(client sftp.Client, heartRate time.Duration) *UploadTransferActivity {
	return &UploadTransferActivity{
		client:    client,
		heartRate: heartRate,
	}
}

// Execute copies the source transfer to the destination via SFTP.
func (a *UploadTransferActivity) Execute(
	ctx context.Context,
	params *UploadTransferActivityParams,
) (*UploadTransferActivityResult, error) {
	logger := temporal_tools.GetLogger(ctx)
	logger.V(1).Info("Execute UploadTransferActivity", "SourcePath", params.SourcePath)

	info, err := os.Stat(params.SourcePath)
	if err != nil {
		return nil, fmt.Errorf("%s: %v", UploadTransferActivityName, err)
	}

	var path string
	var upload sftp.AsyncUpload
	var size int64

	filename := filepath.Base(params.SourcePath)
	if info.IsDir() {
		path, upload, err = a.client.UploadDirectory(ctx, params.SourcePath)
		if err != nil {
			return nil, uploadError(err)
		}

		err := filepath.WalkDir(params.SourcePath, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}

			if !d.IsDir() {
				info, err = os.Stat(path)
				if err != nil {
					return err
				}

				size += info.Size()
			}

			return nil
		})
		if err != nil {
			return nil, err
		}
	} else {
		src, err := os.Open(params.SourcePath)
		if err != nil {
			return nil, fmt.Errorf("%s: %v", UploadTransferActivityName, err)
		}
		defer src.Close()

		path, upload, err = a.client.UploadFile(ctx, src, filename)
		if err != nil {
			return nil, uploadError(err)
		}

		fi, err := src.Stat()
		if err != nil {
			return nil, fmt.Errorf("%s: %v", UploadTransferActivityName, err)
		}

		size = fi.Size()
	}

	// Block (with a heartbeat) until ctx is cancelled, the upload is done, or
	// it stops with an error.
	err = a.Heartbeat(ctx, upload, size)
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

func uploadError(err error) error {
	e := fmt.Errorf("%s: %v", UploadTransferActivityName, err)

	switch err.(type) {
	case *sftp.AuthError:
		return temporal.NewNonRetryableError(e)
	default:
		return e
	}
}
