package sftp

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

// GoClient implements the SFTP service using native Go SSH and SFTP packages.
type GoClient struct {
	cfg Config

	ssh  *ssh.Client
	sftp *sftp.Client
}

var _ Service = (*GoClient)(nil)

// NewGoClient returns a new GoSFTP client with the given configuration.
func NewGoClient(cfg Config) *GoClient {
	cfg.SetDefaults()

	return &GoClient{cfg: cfg}
}

// Upload writes the data from src to the remote file at dest and returns the
// number of bytes written and the full path of the uploaded file.  A new
// SFTP connection is opened before writing, and closed when the upload is
// complete.
func (c *GoClient) Upload(src io.Reader, dest string) (int64, string, error) {
	if err := c.dial(); err != nil {
		return 0, "", err
	}
	defer c.close()

	// Confirm that remote directory exists.
	_, err := c.sftp.Lstat(c.cfg.RemoteDir)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return 0, "", fmt.Errorf("SFTP: remote dir %q doesn't exist", c.cfg.RemoteDir)
		}
		return 0, "", fmt.Errorf("SFTP: couldn't lstat %q: %w", c.cfg.RemoteDir, err)
	}

	// Build remote filepath.
	if c.cfg.RemoteDir != "" {
		dest = filepath.Clean(strings.TrimRight(c.cfg.RemoteDir, "/") + "/" + dest)
	}

	// Use OpenFile with flags because some SFTP servers don't support the
	// O_RDWR mode used by sftp.Create().
	w, err := c.sftp.OpenFile(dest, os.O_WRONLY|os.O_CREATE|os.O_TRUNC)
	if err != nil {
		return 0, "", fmt.Errorf("SFTP: couldn't create remote file %q: %w", dest, err)
	}

	bytes, err := io.Copy(w, src)
	if err != nil {
		return 0, "", fmt.Errorf("SFTP: failed to write to %q: %w", dest, err)
	}

	return bytes, dest, nil
}

// Dial connects to an SSH host then creates an SFTP client on the connection.
// When the clients are no longer needed, close() must be called to prevent
// leaks.
func (c *GoClient) dial() error {
	sshc, err := SSHConnect(c.cfg)
	if err != nil {
		return fmt.Errorf("SSH: %w", err)
	}
	c.ssh = sshc

	sftpc, err := sftp.NewClient(sshc)
	if err != nil {
		return fmt.Errorf("Unable to start SFTP subsystem: %w", err)
	}
	c.sftp = sftpc

	return nil
}

// Close closes the SFTP client first, then the SSH client.
func (c *GoClient) close() error {
	var errs error

	if err := c.sftp.Close(); err != nil {
		errs = errors.Join(err, errs)
	}
	if err := c.ssh.Close(); err != nil {
		errs = errors.Join(err, errs)
	}

	return errs
}
