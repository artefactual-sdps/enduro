package sftp

import (
	"context"
	"fmt"
	"io"
)

// AuthError represents an SFTP authentication error.
type AuthError struct {
	Message string
}

// Error implements the error interface.
func (e *AuthError) Error() string {
	return fmt.Sprintf("auth: %s", e.Message)
}

// NewAuthError returns a pointer to a new AuthError from the underlying |e|
// error message.
func NewAuthError(e error) error {
	return &AuthError{Message: e.Error()}
}

// A Client manages the transmission of data over SFTP.
//
// Implementations of the Client interface handle the connection details,
// authentication, and other intricacies associated with different SFTP
// servers and protocols.
type Client interface {
	// Delete removes dest from the SFTP server.
	Delete(ctx context.Context, dest string) error
	// Upload asynchronously copies data from the src reader to the specified
	// dest on the SFTP server.
	UploadFile(ctx context.Context, src io.Reader, dest string) (remotePath string, upload AsyncUpload, err error)
	UploadDirectory(ctx context.Context, srcPath string) (remotePath string, upload AsyncUpload, err error)
}

// AsyncUpload provides information about an upload happening asynchronously in
// a separate goroutine.
type AsyncUpload interface {
	// Bytes returns the number of bytes copied to the SFTP destination.
	Bytes() int64
	// Close closes SFTP connection used for the upload. Close must be called
	// when the upload is complete to prevent memory leaks.
	Close() error
	// Done returns a channel that receives a true value when the upload is
	// complete.  A done signal should not be sent on error.
	Done() chan bool
	// Done returns a channel that receives an error if the upload encounters
	// an error.
	Err() chan error
	// Write implements the io.Writer interface and adds len(p) to the count of
	// bytes uploaded.
	Write(p []byte) (int, error)
}
