package sftp

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strconv"

	"github.com/dolmen-go/contextio"
	"github.com/go-logr/logr"
	"github.com/pkg/sftp"
)

// GoClient implements the SFTP service using native Go SSH and SFTP packages.
type GoClient struct {
	cfg    Config
	logger logr.Logger
}

var _ Client = (*GoClient)(nil)

// NewGoClient returns a new GoSFTP client with the given configuration.
func NewGoClient(logger logr.Logger, cfg Config) *GoClient {
	cfg.SetDefaults()

	return &GoClient{cfg: cfg, logger: logger}
}

// Delete removes the file or directory from dest. A new SFTP connection is
// opened before removing it, and closed when the delete is complete.
func (c *GoClient) Delete(ctx context.Context, dest string) error {
	remotePath := sftp.Join(c.cfg.RemoteDir, dest)

	conn, err := c.dial(ctx)
	if err != nil {
		return fmt.Errorf("sftp: dial: %v", err)
	}
	defer conn.Close()

	if err := conn.RemoveAll(remotePath); err != nil {
		head := fmt.Sprintf("SFTP: unable to remove %q", dest)
		if errors.Is(err, fs.ErrNotExist) || errors.Is(err, fs.ErrPermission) {
			return fmt.Errorf("%s: %w", head, err)
		}
		if statusErr, ok := err.(*sftp.StatusError); ok {
			return fmt.Errorf("%s: %s", head, formatStatusError(statusErr))
		}
		return fmt.Errorf("%s: %v", head, err)
	}

	return nil
}

// UploadFile asynchronously copies the src data to dest over an SFTP connection.
//
// When UploadFile is called it starts the upload in an asynchronous goroutine, then
// immediately returns the full remote path, and an AsyncUpload struct that
// provides access to the upload status and progress.
//
// When the upload completes, the `AsyncUpload.Done()` channel is sent a `true`
// value. If an error occurs during the upload the error is sent to the
// `AsyncUpload.Error()` channel and the upload is terminated. If a ctx
// cancellation signal is received, the `ctx.Err()` error will be sent to the
// `AsyncUpload.Error()` channel, and the upload is terminated.
func (c *GoClient) UploadFile(ctx context.Context, src io.Reader, dest string) (string, AsyncUpload, error) {
	remotePath := sftp.Join(c.cfg.RemoteDir, dest)

	conn, err := c.dial(ctx)
	if err != nil {
		return "", nil, err
	}

	// Asynchronously upload file.
	upload := NewAsyncUpload(conn)
	go uploadFile(ctx, src, remotePath, &upload)

	return remotePath, &upload, nil
}

// UploadDirectory asynchronously copies a directory to dest over an SFTP connection.
//
// When UploadDirectory is called it starts the upload in an asynchronous goroutine,
// then immediately returns the full remote path, and an AsyncUpload struct that
// provides access to the upload status and progress.
//
// When the upload completes, the `AsyncUpload.Done()` channel is sent a `true`
// value. If an error occurs during the upload the error is sent to the
// `AsyncUpload.Error()` channel and the upload is terminated. If a ctx
// cancellation signal is received, the `ctx.Err()` error will be sent to the
// `AsyncUpload.Error()` channel, and the upload is terminated.
func (c *GoClient) UploadDirectory(ctx context.Context, srcPath string) (string, AsyncUpload, error) {
	transferDir := filepath.Base(srcPath)

	conn, err := c.dial(ctx)
	if err != nil {
		return "", nil, err
	}

	// Asynchronously upload directory of files.
	upload := NewAsyncUpload(conn)
	go uploadDirectory(ctx, srcPath, c.cfg.RemoteDir, &upload)

	return sftp.Join(c.cfg.RemoteDir, transferDir), &upload, nil
}

func uploadFile(ctx context.Context, src io.Reader, remotePath string, upload *AsyncUploadImpl) {
	defer upload.Close()

	remoteCopy(ctx, upload, src, remotePath)

	upload.Done() <- true
}

func uploadDirectory(ctx context.Context, srcPath, remoteDir string, upload *AsyncUploadImpl) {
	defer upload.Close()

	transferDir := filepath.Base(srcPath)

	err := filepath.WalkDir(srcPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(srcPath, path)
		if err != nil {
			return err
		}

		remotePath := sftp.Join(remoteDir, transferDir, relPath)

		if !d.IsDir() {
			f, err := os.Open(path) // #nosec G304 -- trusted file path.
			if err != nil {
				return err
			}
			defer f.Close()

			remoteCopy(ctx, upload, f, remotePath)
		} else {
			err = upload.conn.MkdirAll(remotePath)
			if err != nil {
				return err
			}
		}

		return nil
	})
	if err != nil {
		upload.Err() <- fmt.Errorf("goclient: %v", err)
	}

	upload.Done() <- true
}

// dial connects to an SSH host, creates an SFTP client on the connection, then
// returns conn. When conn is no longer needed, conn.close() must be called to
// prevent leaks.
func (c *GoClient) dial(ctx context.Context) (*connection, error) {
	var (
		conn connection
		err  error
	)

	conn.sshClient, err = sshConnect(ctx, c.logger, c.cfg)
	if err != nil {
		return nil, err
	}

	conn.Client, err = sftp.NewClient(conn.sshClient)
	if err != nil {
		_ = conn.sshClient.Close()
		return nil, fmt.Errorf("start SFTP subsystem: %v", err)
	}

	return &conn, nil
}

// remoteCopy copies data from the src reader to a remote file at dest, and
// updates upload progress asynchronously. Upload status and progress will be
// sent to the upload struct via the `upload.Done()` and `upload.Error()` channels.
func remoteCopy(ctx context.Context, upload *AsyncUploadImpl, src io.Reader, dest string) {
	// Note: Some SFTP servers don't support O_RDWR mode.
	w, err := upload.conn.OpenFile(dest, (os.O_WRONLY | os.O_CREATE | os.O_TRUNC))
	if err != nil {
		upload.Err() <- fmt.Errorf("sftp: open remote file %q: %v", dest, err)
		return
	}
	defer w.Close()

	// Write the number of bytes copied to upload.
	src = contextio.NewReader(ctx, src)
	src = io.TeeReader(src, upload)

	// Use contextio to stop the upload if a context cancellation signal is
	// received.
	_, err = io.Copy(contextio.NewWriter(ctx, w), src)
	if err != nil {
		upload.Err() <- fmt.Errorf("remote copy: %v", err)
	}
}

var statusCodeRegex = regexp.MustCompile(`\(SSH_[A-Z_]+\)$`)

// formatStatusError extracts/formats the SFTP status error code and message.
func formatStatusError(err *sftp.StatusError) string {
	var (
		code    string
		codeMsg = err.FxCode()
	)

	// Find the first match in the error, removing surrounding parentheses.
	matches := statusCodeRegex.FindStringSubmatch(err.Error())
	if len(matches) > 0 {
		code = matches[0][1 : len(matches[0])-1]
	} else {
		code = strconv.FormatUint(uint64(err.Code), 10)
	}

	return fmt.Sprintf("%s (%s)", codeMsg, code)
}
