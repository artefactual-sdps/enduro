package sftp

import (
	"context"
	"io"
)

// A Client manages the transmission of data over SFTP.
//
// Implementations of the Client interface handle the connection details,
// authentication, and other intricacies associated with different SFTP
// servers and protocols.
type Client interface {
	// Upload transfers data from the provided source reader to a specified
	// destination on the SFTP server.
	Upload(ctx context.Context, src io.Reader, dest string) (bytes int64, remotePath string, err error)
}
