package sftp_test

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/gliderlabs/ssh"
	gosftp "github.com/pkg/sftp"
	gossh "golang.org/x/crypto/ssh"
	"gotest.tools/v3/assert"
	tfs "gotest.tools/v3/fs"

	"github.com/artefactual-sdps/enduro/internal/sftp"
)

// ServerAddress is the test SFTP server address.
const serverAddress = "127.0.0.1:2222"

// pubkeyHandler returns a handler that checks the client's public key against
// the keys in the authorized_keys file.
func pubKeyHandler(t *testing.T, ctx ssh.Context, key ssh.PublicKey) bool {
	file, err := os.Open("./testdata/authorized_keys")
	if err != nil {
		t.Fatalf("SFTP server: couldn't open authorized_keys file: %s", err)
	}

	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		allowed, _, _, _, err := ssh.ParseAuthorizedKey([]byte(scanner.Text()))
		if err != nil {
			t.Fatalf("SFTP server: couldn't parse authorized keys: %s", err)
		}
		if ssh.KeysEqual(key, allowed) {
			return true
		}
	}

	t.Log("SFTP server: unknown key provided.")
	return false
}

// HostKeySigner signs messages from the server to the client and allows the
// client to confirm the host key signature.
func hostKeySigner() (gossh.Signer, error) {
	keyfile := "./testdata/serverkeys/test_rsa"

	key, err := os.ReadFile(keyfile)
	if err != nil {
		return nil, fmt.Errorf("read keyfile %q, %v\n", keyfile, err)
	}

	signer, err := gossh.ParsePrivateKey(key)
	if err != nil {
		return nil, fmt.Errorf("parse private key: %v\n", err)
	}

	return signer, nil
}

// SftpHandler starts the SFTP subsystem.
func sftpHandler(sess ssh.Session) {
	debugStream := io.Discard
	serverOptions := []gosftp.ServerOption{
		gosftp.WithDebug(debugStream),
	}
	server, err := gosftp.NewServer(
		sess,
		serverOptions...,
	)
	if err != nil {
		log.Fatalf("SFTP server init error: %s", err)
	}
	if err := server.Serve(); err == io.EOF {
		server.Close()
		fmt.Println("SFTP client exited session.")
	} else if err != nil {
		fmt.Println("SFTP server completed with error:", err)
	}
}

// StartSFTPServer starts a test SFTP server, and returns a pointer to the
// server.
func startSFTPServer(t *testing.T, addr string) *ssh.Server {
	t.Helper()

	var err error

	srv := ssh.Server{
		Addr: addr,
		Handler: func(s ssh.Session) {
			authorizedKey := gossh.MarshalAuthorizedKey(s.PublicKey())
			io.WriteString(s, fmt.Sprintf("public key used by %s:\n", s.User()))
			s.Write(authorizedKey)
		},
		PublicKeyHandler: func(ctx ssh.Context, key ssh.PublicKey) bool {
			return pubKeyHandler(t, ctx, key)
		},
		SubsystemHandlers: map[string]ssh.SubsystemHandler{
			"sftp": sftpHandler,
		},
	}

	signer, err := hostKeySigner()
	if err != nil {
		t.Fatalf("SFTP server: couldn't create host key signer: %v", err)
	}
	srv.AddHostKey(signer)

	errCh := make(chan error, 1)
	go func() {
		errCh <- srv.ListenAndServe()
	}()

	// Wait for the server to be ready
	func() {
		for {
			select {
			case err := <-errCh:
				t.Fatalf("SFTP server: failed to start: %v", err)
			default:
				conn, err := net.DialTimeout("tcp", addr, 1*time.Second)
				if err == nil {
					conn.Close()
					return
				}
				time.Sleep(10 * time.Millisecond)
			}
		}
	}()

	t.Cleanup(func() { srv.Close() })
	return &srv
}

func TestGoClient(t *testing.T) {
	host, port, err := net.SplitHostPort(serverAddress)
	if err != nil {
		t.Fatalf("Bad server address: %s", serverAddress)
	}

	_ = startSFTPServer(t, serverAddress)

	// Start a listener on an open port and use the address to test a bad SFTP
	// server address.
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("Couldn't start listener: %v", err)
	}
	defer listener.Close()
	badHost, badPort, _ := net.SplitHostPort(listener.Addr().String())

	type results struct {
		Bytes int64
		Paths []tfs.PathOp
	}

	type test struct {
		name    string
		cfg     sftp.Config
		want    results
		wantErr string
	}
	for _, tc := range []test{
		{
			name: "Uploads a file using a key with no passphrase",
			cfg: sftp.Config{
				Host:           host,
				Port:           port,
				KnownHostsFile: "./testdata/known_hosts",
				PrivateKey: sftp.PrivateKey{
					Path: "./testdata/clientkeys/test_ed25519",
				},
			},
			want: results{
				Bytes: 13,
				Paths: []tfs.PathOp{tfs.WithFile("test.txt", "Testing 1-2-3")},
			},
		},
		{
			name: "Uploads a file using a key with a passphrase",
			cfg: sftp.Config{
				Host:           host,
				Port:           port,
				KnownHostsFile: "./testdata/known_hosts",
				PrivateKey: sftp.PrivateKey{
					Path:       "./testdata/clientkeys/test_pass_rsa",
					Passphrase: "Backpack-Spirits6-Bronzing",
				},
			},
			want: results{
				Bytes: 13,
				Paths: []tfs.PathOp{tfs.WithFile("test.txt", "Testing 1-2-3")},
			},
		},
		{
			name: "Errors when the key passphrase is wrong",
			cfg: sftp.Config{
				Host:           host,
				Port:           port,
				KnownHostsFile: "./testdata/known_hosts",
				PrivateKey: sftp.PrivateKey{
					Path:       "./testdata/clientkeys/test_pass_rsa",
					Passphrase: "wrong",
				},
			},
			wantErr: "SSH: parse private key with passphrase: x509: decryption password incorrect",
		},
		{
			name: "Errors when the SFTP server isn't there",
			cfg: sftp.Config{
				Host:           badHost,
				Port:           badPort,
				KnownHostsFile: "./testdata/known_hosts",
				PrivateKey: sftp.PrivateKey{
					Path: "./testdata/clientkeys/test_ed25519",
				},
			},
			wantErr: fmt.Sprintf(
				"SSH: connect: dial tcp %s:%s: connect: connection refused",
				badHost, badPort,
			),
		},
		{
			name: "Errors when the private key is not recognized",
			cfg: sftp.Config{
				Host:           host,
				Port:           port,
				KnownHostsFile: "./testdata/known_hosts",
				PrivateKey: sftp.PrivateKey{
					Path: "./testdata/clientkeys/test_unk_ed25519",
				},
			},
			wantErr: "SSH: connect: ssh: handshake failed: ssh: unable to authenticate, attempted methods [none publickey], no supported methods remain",
		},
		{
			name: "Errors when the host key is not in known_hosts",
			cfg: sftp.Config{
				Host:           host,
				Port:           port,
				KnownHostsFile: "./testdata/empty_file",
				PrivateKey: sftp.PrivateKey{
					Path: "./testdata/clientkeys/test_ed25519",
				},
			},
			wantErr: "SSH: connect: ssh: handshake failed: knownhosts: key is unknown",
		},
		{
			name: "Errors when the known_hosts file doesn't exist",
			cfg: sftp.Config{
				Host:           host,
				Port:           port,
				KnownHostsFile: "./testdata/missing",
				PrivateKey: sftp.PrivateKey{
					Path: "./testdata/clientkeys/test_ed25519",
				},
			},
			wantErr: "SSH: parse known_hosts: open testdata/missing: no such file or directory",
		},
	} {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			sftpc := sftp.NewGoClient(tc.cfg)
			src := strings.NewReader("Testing 1-2-3")
			dest := tfs.NewDir(t, "sftp_test")

			bytes, err := sftpc.Upload(context.Background(), src, dest.Join("test.txt"))

			if tc.wantErr != "" {
				assert.Error(t, err, tc.wantErr)
				return
			}

			assert.NilError(t, err)
			assert.Equal(t, bytes, tc.want.Bytes)
			assert.Assert(t, tfs.Equal(dest.Path(), tfs.Expected(t, tc.want.Paths...)))
		})
	}
}
