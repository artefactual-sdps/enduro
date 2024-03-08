package activities

import (
	"archive/zip"
	"context"
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
	a.logger.V(1).Info("Executing ZipActivity",
		"SourceDir", params.SourceDir,
		"DestPath", params.DestPath,
	)

	if params.SourceDir == "" {
		return &ZipActivityResult{}, fmt.Errorf("%s: missing source dir", ZipActivityName)
	}

	dest := params.DestPath
	if params.DestPath == "" {
		dest = params.SourceDir + ".zip"
		a.logger.V(1).Info(ZipActivityName+": dest changed", "dest", dest)
	}

	w, err := os.Create(dest) // #nosec G304 -- trusted path
	if err != nil {
		return &ZipActivityResult{}, fmt.Errorf("%s: create: %v", ZipActivityName, err)
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

		// Include SourceDir in the zip paths, but not its parent dirs.
		p, err := filepath.Rel(filepath.Dir(params.SourceDir), path)
		if err != nil {
			return err
		}

		f, err := z.Create(p)
		if err != nil {
			return err
		}

		r, err := os.Open(path) // #nosec G304 -- trusted path
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
		return &ZipActivityResult{}, fmt.Errorf("%s: add files: %v", ZipActivityName, err)
	}

	return &ZipActivityResult{Path: dest}, nil
}
