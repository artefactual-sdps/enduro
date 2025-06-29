package sftp_test

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/gliderlabs/ssh"
	"github.com/go-logr/logr"
	gosftp "github.com/pkg/sftp"
	gossh "golang.org/x/crypto/ssh"
	"gotest.tools/v3/assert"
	tfs "gotest.tools/v3/fs"

	"github.com/artefactual-sdps/enduro/internal/sftp"
)

// pubkeyHandler returns a handler that checks the client's public key against
// the keys in the authorized_keys file.
func pubKeyHandler(t *testing.T, key ssh.PublicKey) bool {
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

// StartSFTPServer starts a test SFTP server, and returns its host and port.
func startSFTPServer(t *testing.T) (string, string) {
	t.Helper()

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	assert.NilError(t, err)

	addr := ln.Addr().String()
	host, port, err := net.SplitHostPort(addr)
	assert.NilError(t, err)

	srv := ssh.Server{
		Addr: addr,
		Handler: func(s ssh.Session) {
			authorizedKey := gossh.MarshalAuthorizedKey(s.PublicKey())
			io.WriteString(s, fmt.Sprintf("public key used by %s:\n", s.User()))
			s.Write(authorizedKey)
		},
		PublicKeyHandler: func(ctx ssh.Context, key ssh.PublicKey) bool {
			return pubKeyHandler(t, key)
		},
		SubsystemHandlers: map[string]ssh.SubsystemHandler{
			"sftp": sftpHandler,
		},
	}
	t.Cleanup(func() { _ = srv.Close() })

	signer, err := hostKeySigner()
	if err != nil {
		t.Fatalf("SFTP server: couldn't create host key signer: %v", err)
	}
	srv.AddHostKey(signer)

	errCh := make(chan error, 1)
	go func() {
		errCh <- srv.Serve(ln)
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

	return host, port
}

func TestUploadFile(t *testing.T) {
	t.Parallel()

	host, port := startSFTPServer(t)

	// Start a listener on an open port and use the address to test a bad SFTP
	// server address.
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("Couldn't start listener: %v", err)
	}
	defer ln.Close()
	badHost, badPort, _ := net.SplitHostPort(ln.Addr().String())

	type params struct {
		src  io.Reader
		dest string
	}
	type results struct {
		Bytes int64
		Paths []tfs.PathOp
	}

	type test struct {
		name    string
		cfg     sftp.Config
		params  params
		want    results
		wantErr error
	}
	for _, tc := range []test{
		{
			name: "Uploads a file using private key auth",
			cfg: sftp.Config{
				Host:           host,
				Port:           port,
				KnownHostsFile: knownHostsFile(t, host, port),
				PrivateKey: sftp.PrivateKey{
					Path: "./testdata/clientkeys/test_ed25519",
				},
			},
			params: params{
				src:  strings.NewReader("Testing 1-2-3"),
				dest: "test.txt",
			},
			want: results{
				Bytes: 13,
				Paths: []tfs.PathOp{tfs.WithFile("test.txt", "Testing 1-2-3")},
			},
		},
		{
			name: "Uploads a file using private key + password auth",
			cfg: sftp.Config{
				Host:           host,
				Port:           port,
				KnownHostsFile: knownHostsFile(t, host, port),
				PrivateKey: sftp.PrivateKey{
					Path:       "./testdata/clientkeys/test_pass_rsa",
					Passphrase: "Backpack-Spirits6-Bronzing",
				},
			},
			params: params{
				src:  strings.NewReader("Testing 1-2-3"),
				dest: "test.txt",
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
				KnownHostsFile: knownHostsFile(t, host, port),
				PrivateKey: sftp.PrivateKey{
					Path:       "./testdata/clientkeys/test_pass_rsa",
					Passphrase: "wrong",
				},
			},
			params: params{
				src:  strings.NewReader("Testing 1-2-3"),
				dest: "test.txt",
			},
			wantErr: &sftp.AuthError{
				Message: "ssh: parse private key with passphrase: x509: decryption password incorrect",
			},
		},
		{
			name: "Errors when the SFTP server isn't there",
			cfg: sftp.Config{
				Host:           badHost,
				Port:           badPort,
				KnownHostsFile: knownHostsFile(t, host, port),
				PrivateKey: sftp.PrivateKey{
					Path: "./testdata/clientkeys/test_ed25519",
				},
			},
			params: params{
				src:  strings.NewReader("Testing 1-2-3"),
				dest: "test.txt",
			},
			wantErr: fmt.Errorf(
				"ssh: connect: dial tcp %s:%s: connect: connection refused",
				badHost, badPort,
			),
		},
		{
			name: "Errors when the private key is not recognized",
			cfg: sftp.Config{
				Host:           host,
				Port:           port,
				KnownHostsFile: knownHostsFile(t, host, port),
				PrivateKey: sftp.PrivateKey{
					Path: "./testdata/clientkeys/test_unk_ed25519",
				},
			},
			wantErr: &sftp.AuthError{
				Message: "ssh: handshake failed: ssh: unable to authenticate, attempted methods [none publickey], no supported methods remain",
			},
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
			wantErr: &sftp.AuthError{
				Message: "ssh: handshake failed: knownhosts: key is unknown",
			},
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
			wantErr: &sftp.AuthError{
				Message: "ssh: parse known_hosts: open testdata/missing: no such file or directory",
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Use a unique RemoteDir for each test.
			remoteDir := tfs.NewDir(t, "sftp_test_remote")
			tc.cfg.RemoteDir = remoteDir.Path()

			client := sftp.NewGoClient(logr.Discard(), tc.cfg)
			remotePath, upload, err := client.UploadFile(context.Background(), tc.params.src, tc.params.dest)
			if tc.wantErr != nil {
				assert.Error(t, err, tc.wantErr.Error())
				assert.Assert(t, reflect.TypeOf(err) == reflect.TypeOf(tc.wantErr))
				return
			}
			assert.NilError(t, err)

			assert.Equal(t, remotePath, tc.cfg.RemoteDir+"/"+tc.params.dest)
			assert.Equal(t, upload.Bytes(), int64(0)) // Upload hasn't started yet.

			select {
			case <-upload.Done():
				assert.Equal(t, upload.Bytes(), tc.want.Bytes)
				assert.Assert(t, tfs.Equal(remoteDir.Path(), tfs.Expected(t, tc.want.Paths...)))
			case err = <-upload.Err():
				t.Fatal(err)
			}
		})
	}
}

func TestUploadDirectory(t *testing.T) {
	t.Parallel()

	testTransferDir := tfs.NewDir(t, "test_transfer",
		tfs.WithFile("test.txt", "Testing 1-2-3"),
		tfs.WithFile("test2.txt", "Testing 4-5-6-7"),
	)

	host, port := startSFTPServer(t)

	type params struct {
		src string
	}

	type test struct {
		name      string
		cfg       sftp.Config
		params    params
		wantBytes int64
		wantErr   error
	}
	for _, tc := range []test{
		{
			name: "Uploads a directory using private key auth",
			cfg: sftp.Config{
				Host:           host,
				Port:           port,
				KnownHostsFile: knownHostsFile(t, host, port),
				PrivateKey: sftp.PrivateKey{
					Path: "./testdata/clientkeys/test_ed25519",
				},
			},
			params: params{
				src: testTransferDir.Path(),
			},
			wantBytes: 28,
		},
		{
			name: "Uploads a file using private key + password auth",
			cfg: sftp.Config{
				Host:           host,
				Port:           port,
				KnownHostsFile: knownHostsFile(t, host, port),
				PrivateKey: sftp.PrivateKey{
					Path:       "./testdata/clientkeys/test_pass_rsa",
					Passphrase: "Backpack-Spirits6-Bronzing",
				},
			},
			params: params{
				src: testTransferDir.Path(),
			},
			wantBytes: 28,
		},
		{
			name: "Errors when the key passphrase is wrong",
			cfg: sftp.Config{
				Host:           host,
				Port:           port,
				KnownHostsFile: knownHostsFile(t, host, port),
				PrivateKey: sftp.PrivateKey{
					Path:       "./testdata/clientkeys/test_pass_rsa",
					Passphrase: "wrong",
				},
			},
			params: params{
				src: testTransferDir.Path(),
			},
			wantErr: &sftp.AuthError{
				Message: "ssh: parse private key with passphrase: x509: decryption password incorrect",
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Use a unique RemoteDir for each test.
			remoteDir := tfs.NewDir(t, "sftp_test_remote")
			tc.cfg.RemoteDir = remoteDir.Path()

			client := sftp.NewGoClient(logr.Discard(), tc.cfg)
			remotePath, upload, err := client.UploadDirectory(context.Background(), tc.params.src)
			if tc.wantErr != nil {
				assert.Error(t, err, tc.wantErr.Error())
				assert.Assert(t, reflect.TypeOf(err) == reflect.TypeOf(tc.wantErr))
				return
			}

			assert.NilError(t, err)
			assert.Equal(t, remotePath, tc.cfg.RemoteDir+"/"+filepath.Base(tc.params.src))

			select {
			case <-upload.Done():
				assert.Equal(t, upload.Bytes(), tc.wantBytes)
			case err = <-upload.Err():
				t.Fatal(err)
			}
		})
	}
}

func TestDelete(t *testing.T) {
	t.Parallel()

	type params struct {
		fsOps       []tfs.PathOp // The state of the filesystem served by the SFTP server.
		restrictDir string       // Set 0o555 on dir to reproduce permission issues.
		file        string       // The file that we will delete.
	}

	type test struct {
		name    string
		params  params
		wantFs  []tfs.PathOp
		wantErr string
	}
	for _, tc := range []test{
		{
			name: "Deletes a file",
			params: params{
				fsOps: []tfs.PathOp{
					tfs.WithFile("test.txt", ""),
				},
				file: "test.txt",
			},
			wantFs: []tfs.PathOp{},
		},
		{
			name: "Errors when file doesn't exist",
			params: params{
				fsOps: []tfs.PathOp{
					// File test.txt must be non-existent.
				},
				file: "test.txt",
			},
			wantErr: "SFTP: unable to remove \"test.txt\": file does not exist",
		},
		{
			name: "Errors when there are insufficient permissions",
			params: params{
				fsOps: []tfs.PathOp{
					tfs.WithDir("restricted",
						tfs.WithFile("test.txt", ""),
					),
				},
				restrictDir: "restricted",
				file:        "restricted/test.txt",
			},
			wantErr: "SFTP: unable to remove \"restricted/test.txt\": permission denied",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			host, port := startSFTPServer(t)

			cfg := sftp.Config{
				Host:           host,
				Port:           port,
				KnownHostsFile: knownHostsFile(t, host, port),
				PrivateKey: sftp.PrivateKey{
					Path: "./testdata/clientkeys/test_ed25519",
				},
			}

			// Use a unique RemoteDir for each test.
			remoteDir := tfs.NewDir(t, "sftp_test_remote", tc.params.fsOps...)
			cfg.RemoteDir = remoteDir.Path()
			if tc.params.restrictDir != "" {
				err := os.Chmod(remoteDir.Join(tc.params.restrictDir), 0o555)
				assert.NilError(t, err)
			}

			client := sftp.NewGoClient(logr.Discard(), cfg)
			err := client.Delete(context.Background(), tc.params.file)
			if tc.wantErr != "" {
				assert.Error(t, err, tc.wantErr)
				return
			}

			assert.NilError(t, err)
			assert.Assert(t, tfs.Equal(remoteDir.Path(), tfs.Expected(t, tc.wantFs...)))
		})
	}
}

// knownHostsFile returns the path to a known_hosts file with the given host:port.
func knownHostsFile(t *testing.T, host, port string) string {
	t.Helper()

	blob, err := os.ReadFile("./testdata/known_hosts")
	assert.NilError(t, err)

	addr := fmt.Sprintf("[%s]:%s", host, port)
	blob = bytes.Replace(blob, []byte("[127.0.0.1]:2222"), []byte(addr), 1)

	return tfs.NewFile(t, "", tfs.WithBytes(blob)).Path()
}
