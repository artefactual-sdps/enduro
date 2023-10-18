package sftp

import (
	"fmt"
	"net"
	"os"
	"time"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/knownhosts"
)

// SSHConnect connects to an SSH server using the given configuration and
// returns a client connection.
//
// Only private key authentication is currently supported, with or without a
// passphrase.
func SSHConnect(cfg Config) (*ssh.Client, error) {
	// Load private key for authentication.
	keyPath := os.ExpandEnv(cfg.PrivateKey.Path)
	keyBytes, err := os.ReadFile(keyPath) // #nosec G304 -- File data is validated below
	if err != nil {
		return nil, fmt.Errorf("failed to read private key: %w", err)
	}

	// Create a signer from the private key, with or without a passphrase.
	var signer ssh.Signer
	if cfg.PrivateKey.Passphrase != "" {
		signer, err = ssh.ParsePrivateKeyWithPassphrase(keyBytes, []byte(cfg.PrivateKey.Passphrase))
		if err != nil {
			return nil, fmt.Errorf("failed to parse private key with passphrase: %w", err)
		}
	} else {
		signer, err = ssh.ParsePrivateKey(keyBytes)
		if err != nil {
			return nil, fmt.Errorf("failed to parse private key: %w", err)
		}
	}

	// Check that the host key is in the client's known_hosts file.
	hostcallback, err := knownhosts.New(os.ExpandEnv(cfg.KnownHostsFile))
	if err != nil {
		return nil, fmt.Errorf("couldn't parse known_hosts_file: %w", err)
	}

	// Configure the SSH client.
	sshConfig := &ssh.ClientConfig{
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
		HostKeyCallback: hostcallback,
		Timeout:         5 * time.Second,
	}

	// Connect to the server.
	address := net.JoinHostPort(cfg.Host, cfg.Port)
	conn, err := ssh.Dial("tcp", address, sshConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to connect: %w", err)
	}

	return conn, nil
}
