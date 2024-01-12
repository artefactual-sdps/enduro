package sftp

import "sync/atomic"

// AsyncUploadImpl provides an asynchronous upload implementation.
type AsyncUploadImpl struct {
	conn *connection
	done chan bool
	err  chan error

	bytes atomic.Int64
}

var _ AsyncUpload = (*AsyncUploadImpl)(nil)

// NewAsyncUpload returns an initialized AsyncUploadImpl struct that wraps the
// underlying SFTP connection.
func NewAsyncUpload(conn *connection) AsyncUploadImpl {
	return AsyncUploadImpl{
		conn: conn,
		done: make(chan bool, 1),
		err:  make(chan error, 1),
	}
}

// Bytes returns the current number of bytes uploaded.
func (u *AsyncUploadImpl) Bytes() int64 {
	return int64(u.bytes.Load())
}

// Close closes SFTP connection used for the upload. Close must be called
// when the upload is complete to prevent memory leaks.
func (u *AsyncUploadImpl) Close() error {
	return u.conn.Close()
}

// Done returns a done channel that receives a true value when the upload is
// complete.
func (u *AsyncUploadImpl) Done() chan bool {
	return u.done
}

// Error returns an error channel that receives an error if the upload
// encounters an error.
func (u *AsyncUploadImpl) Err() chan error {
	return u.err
}

// Write adds the length of p to the total number of bytes written on the
// connection.
//
// Write implements the io.Writer interface.
func (u *AsyncUploadImpl) Write(p []byte) (int, error) {
	n := len(p)
	u.bytes.Add(int64(n))
	return n, nil
}
