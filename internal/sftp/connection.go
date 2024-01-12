package sftp

import (
	"errors"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

// connection represents an SFTP connection and the underlying SSH connection.
type connection struct {
	*sftp.Client
	sshClient *ssh.Client
}

// Close closes the SFTP connection then the underlying SSH connection.
func (c *connection) Close() error {
	var errs error

	if c.Client != nil {
		if err := c.Client.Close(); err != nil {
			errs = errors.Join(err, errs)
		}
	}

	if c.sshClient != nil {
		if err := c.sshClient.Close(); err != nil {
			errs = errors.Join(err, errs)
		}
	}

	return errs
}
