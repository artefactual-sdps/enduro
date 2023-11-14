package sftp

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/dolmen-go/contextio"
	"github.com/go-logr/logr"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

// GoClient implements the SFTP service using native Go SSH and SFTP packages.
type GoClient struct {
	cfg    Config
	logger logr.Logger

	ssh  *ssh.Client
	sftp *sftp.Client
}

var _ Client = (*GoClient)(nil)

// NewGoClient returns a new GoSFTP client with the given configuration.
func NewGoClient(logger logr.Logger, cfg Config) *GoClient {
	cfg.SetDefaults()

	return &GoClient{cfg: cfg, logger: logger}
}

// Upload writes the data from src to the remote file at dest and returns the
// number of bytes written.  A new SFTP connection is opened before writing, and
// closed when the upload is complete or cancelled.
//
// Upload is not thread safe.
func (c *GoClient) Upload(ctx context.Context, src io.Reader, dest string) (int64, string, error) {
	if err := c.dial(); err != nil {
		return 0, "", err
	}
	defer c.close()

	// SFTP assumes that "/" is used as the directory separator. See:
	// https://datatracker.ietf.org/doc/html/draft-ietf-secsh-filexfer-02#section-6.2
	remotePath := strings.TrimSuffix(c.cfg.RemoteDir, "/") + "/" + dest

	// Note: Some SFTP servers don't support O_RDWR mode.
	w, err := c.sftp.OpenFile(remotePath, (os.O_WRONLY | os.O_CREATE | os.O_TRUNC))
	if err != nil {
		return 0, "", fmt.Errorf("SFTP: open remote file %q: %v", dest, err)
	}
	defer w.Close()

	// Use contextio to stop the upload if a context cancellation signal is
	// received.
	bytes, err := io.Copy(contextio.NewWriter(ctx, w), contextio.NewReader(ctx, src))
	if err != nil {
		return 0, "", fmt.Errorf("SFTP: upload to %q: %v", dest, err)
	}

	return bytes, remotePath, nil
}

// Delete removes the data from dest, a new SFTP connection is opened before
// removing the file, and closed when the delete is complete.
func (c *GoClient) Delete(ctx context.Context, dest string) error {
	if err := c.dial(); err != nil {
		return fmt.Errorf("SFTP: unable to dial: %w", err)
	}
	defer c.close()

	// Delete the file
	if err := c.sftp.Remove(dest); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("SFTP: file does not exist: %w", err)
		} else if os.IsPermission(err) {
			return fmt.Errorf("SFTP: insufficient permissions to delete file: %w", err)
		}
		return fmt.Errorf("SFTP: unable to remove file %q: %v", dest, err)
	}
	return nil
}

// Dial connects to an SSH host then creates an SFTP client on the connection.
// When the clients are no longer needed, close() must be called to prevent
// leaks.
func (c *GoClient) dial() error {
	var err error

	c.ssh, err = sshConnect(c.logger, c.cfg)
	if err != nil {
		return fmt.Errorf("SSH: %v", err)
	}

	c.sftp, err = sftp.NewClient(c.ssh)
	if err != nil {
		return fmt.Errorf("start SFTP subsystem: %v", err)
	}

	return nil
}

// Close closes the SFTP client first, then the SSH client.
func (c *GoClient) close() error {
	var errs error

	if c.sftp != nil {
		if err := c.sftp.Close(); err != nil {
			errs = errors.Join(err, errs)
		}
	}

	if c.ssh != nil {
		if err := c.ssh.Close(); err != nil {
			errs = errors.Join(err, errs)
		}
	}

	return errs
}
