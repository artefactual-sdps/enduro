package sftp_test

import (
	"bufio"
	"fmt"
	"io"
	"log"
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

// PubkeyHandler handles checking the client's public key against the keys in
// the authorized_keys file.
func pubkeyHandler(ctx ssh.Context, key ssh.PublicKey) bool {
	file, err := os.Open("./testdata/authorized_keys")
	if err != nil {
		log.Fatalln("SSH: couldn't open authorized_keys file.")
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		allowed, _, _, _, err := ssh.ParseAuthorizedKey([]byte(scanner.Text()))
		if err != nil {
			log.Fatalln("SSH: couldn't parse authorized key.")
		}
		if ssh.KeysEqual(key, allowed) {
			return true
		}
	}

	log.Println("SSH: unknown key provided.")
	return false
}

// HostKeySigner signs messages from the server to the client and allows the
// client to confirm the host key signature.
func hostKeySigner() (gossh.Signer, error) {
	keyfile := "./testdata/serverkeys/test_rsa"

	key, err := os.ReadFile(keyfile)
	if err != nil {
		return nil, fmt.Errorf("couldn't read keyfile %q, %v\n", keyfile, err)
	}

	signer, err := gossh.ParsePrivateKey(key)
	if err != nil {
		return nil, fmt.Errorf("couldn't parse private key: %v\n", err)
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
		log.Printf("sftp server init error: %s\n", err)
		return
	}
	if err := server.Serve(); err == io.EOF {
		server.Close()
		fmt.Println("sftp client exited session.")
	} else if err != nil {
		fmt.Println("sftp server completed with error:", err)
	}
}

// StartSFTPServer starts a test SFTP server, and returns a pointer to the
// server. The caller must call Close() to shut down the server when done with
// it.
func startSFTPServer() (*ssh.Server, error) {
	srv := ssh.Server{
		Addr: "127.0.0.1:2222",
		Handler: func(s ssh.Session) {
			authorizedKey := gossh.MarshalAuthorizedKey(s.PublicKey())
			io.WriteString(s, fmt.Sprintf("public key used by %s:\n", s.User()))
			s.Write(authorizedKey)
		},
		PublicKeyHandler: pubkeyHandler,
		SubsystemHandlers: map[string]ssh.SubsystemHandler{
			"sftp": sftpHandler,
		},
	}

	signer, err := hostKeySigner()
	if err != nil {
		return nil, err
	}
	srv.AddHostKey(signer)

	go func() {
		err = srv.ListenAndServe()
	}()

	return &srv, err
}

func TestGoClient(t *testing.T) {
	srv, err := startSFTPServer()
	if err != nil {
		t.Fatalf("Failed to start SFTP server: %v", err)
	}
	t.Cleanup(func() { srv.Close() })

	// Give the server 100ms to start.
	time.Sleep(100 * time.Millisecond)

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
				Host:           "127.0.0.1",
				Port:           "2222",
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
				Host:           "127.0.0.1",
				Port:           "2222",
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
				Host:           "127.0.0.1",
				Port:           "2222",
				KnownHostsFile: "./testdata/known_hosts",
				PrivateKey: sftp.PrivateKey{
					Path:       "./testdata/clientkeys/test_pass_rsa",
					Passphrase: "wrong",
				},
			},
			wantErr: "SSH: failed to parse private key with passphrase: x509: decryption password incorrect",
		},
		{
			name: "Errors when the SFTP server isn't there",
			cfg: sftp.Config{
				Host:           "127.0.0.1",
				Port:           "2200",
				KnownHostsFile: "./testdata/known_hosts",
				PrivateKey: sftp.PrivateKey{
					Path: "./testdata/clientkeys/test_ed25519",
				},
			},
			wantErr: "SSH: failed to connect: dial tcp 127.0.0.1:2200: connect: connection refused",
		},
		{
			name: "Errors when the private key is not recognized",
			cfg: sftp.Config{
				Host:           "127.0.0.1",
				Port:           "2222",
				KnownHostsFile: "./testdata/known_hosts",
				PrivateKey: sftp.PrivateKey{
					Path: "./testdata/clientkeys/test_unk_ed25519",
				},
			},
			wantErr: "SSH: failed to connect: ssh: handshake failed: ssh: unable to authenticate, attempted methods [none publickey], no supported methods remain",
		},
		{
			name: "Errors when the host key is not in known_hosts",
			cfg: sftp.Config{
				Host:           "127.0.0.1",
				Port:           "2222",
				KnownHostsFile: "./testdata/empty_file",
				PrivateKey: sftp.PrivateKey{
					Path: "./testdata/clientkeys/test_ed25519",
				},
			},
			wantErr: "SSH: failed to connect: ssh: handshake failed: knownhosts: key is unknown",
		},
		{
			name: "Errors when the known_hosts file doesn't exist",
			cfg: sftp.Config{
				Host:           "127.0.0.1",
				Port:           "2222",
				KnownHostsFile: "./testdata/missing",
				PrivateKey: sftp.PrivateKey{
					Path: "./testdata/clientkeys/test_ed25519",
				},
			},
			wantErr: "SSH: couldn't parse known_hosts_file: open ./testdata/missing: no such file or directory",
		},
	} {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			sftpc := sftp.NewGoClient(tc.cfg)
			src := strings.NewReader("Testing 1-2-3")
			dest := tfs.NewDir(t, "sftp_test")

			bytes, err := sftpc.Upload(src, dest.Join("test.txt"))

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
