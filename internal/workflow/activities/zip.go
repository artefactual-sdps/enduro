package activities

import (
	"archive/zip"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/go-logr/logr"
)

const ZipActivityName = "ZipActivity"

type ZipActivityParams struct {
	SourceDir string
	DestPath  string
}

type ZipActivityResult struct {
	Path string
}

type zipActivity struct {
	logger logr.Logger
}

func NewZipActivity(logger logr.Logger) *zipActivity {
	return &zipActivity{logger: logger}
}

// Execute creates a Zip archive at params.DestPath from the contents of
// params.SourceDir. If params.DestPath is not specified then params.SourceDir
// + ".zip" will be used.
func (a *zipActivity) Execute(ctx context.Context, params *ZipActivityParams) (*ZipActivityResult, error) {
	a.logger.V(1).Info("Executing ZipActivity", "sourceDir", params.SourceDir, "DestPath", params.DestPath)

	if params.SourceDir == "" {
		return nil, errors.New("zip: missing source directory")
	}

	var dest string
	if params.DestPath == "" {
		dest = params.SourceDir + ".zip"
		a.logger.V(1).Info("ZipActivity dest changed", "dest", params.DestPath)
	}

	w, err := os.Create(dest)
	if err != nil {
		return nil, fmt.Errorf("zip: couldn't create file: %v", err)
	}
	defer w.Close()

	z := zip.NewWriter(w)
	defer z.Close()

	err = filepath.WalkDir(params.SourceDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		f, err := z.Create(path)
		if err != nil {
			return err
		}

		r, err := os.Open(path)
		if err != nil {
			return err
		}
		defer r.Close()

		if _, err := io.Copy(f, r); err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("zip: %v", err)
	}

	return &ZipActivityResult{Path: dest}, nil
}
