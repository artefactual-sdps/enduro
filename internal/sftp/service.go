package sftp

import (
	"io"
)

type Service interface {
	Upload(src io.Reader, dest string) (int64, error)
}
