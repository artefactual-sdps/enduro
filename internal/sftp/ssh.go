package sftp

import (
	"context"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-logr/logr"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/knownhosts"
)

// sshConnect connects to an SSH server using the given configuration and
// returns a client connection.
//
// Only private key authentication is currently supported, with or without a
// passphrase.
func sshConnect(ctx context.Context, logger logr.Logger, cfg Config) (*ssh.Client, error) {
	// Load private key for authentication.
	keyBytes, err := os.ReadFile(filepath.Clean(cfg.PrivateKey.Path)) // #nosec G304 -- File data is validated below
	if err != nil {
		return nil, NewAuthError(fmt.Errorf("ssh: read private key: %v", err))
	}

	// Create a signer from the private key, with or without a passphrase.
	var signer ssh.Signer
	if cfg.PrivateKey.Passphrase != "" {
		signer, err = ssh.ParsePrivateKeyWithPassphrase(keyBytes, []byte(cfg.PrivateKey.Passphrase))
		if err != nil {
			return nil, NewAuthError(fmt.Errorf("ssh: parse private key with passphrase: %v", err))
		}
	} else {
		signer, err = ssh.ParsePrivateKey(keyBytes)
		if err != nil {
			return nil, NewAuthError(fmt.Errorf("ssh: parse private key: %v", err))
		}
	}

	// Check that the host key is in the client's known_hosts file.
	hostcallback, err := knownhosts.New(filepath.Clean(cfg.KnownHostsFile))
	if err != nil {
		return nil, NewAuthError(fmt.Errorf("ssh: parse known_hosts: %v", err))
	}

	// Configure the SSH client.
	sshConfig := &ssh.ClientConfig{
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
		HostKeyCallback: hostcallback,
		Timeout:         5 * time.Second,
		User:            cfg.User,
	}

	// Connect to the server.
	address := net.JoinHostPort(cfg.Host, cfg.Port)

	dialer := &net.Dialer{}
	conn, err := dialer.DialContext(ctx, "tcp", address)
	if err != nil {
		logger.V(2).Info("ssh: dial failed", "address", address, "user", cfg.User)
		return nil, fmt.Errorf("ssh: connect: %v", err)
	}

	sshConn, chans, reqs, err := ssh.NewClientConn(conn, address, sshConfig)
	if err != nil {
		if strings.Contains(err.Error(), "ssh: unable to authenticate") ||
			strings.Contains(err.Error(), "knownhosts: key is unknown") {
			logger.V(2).Info("ssh: authentication failed", "address", address, "user", cfg.User)
			return nil, NewAuthError(err)
		}
		logger.V(2).Info("ssh: failed to connect", "address", address, "user", cfg.User)
		return nil, err
	}

	return ssh.NewClient(sshConn, chans, reqs), nil
}
