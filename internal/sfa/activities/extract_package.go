package activities

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/mholt/archiver/v3"
	"go.artefactual.dev/tools/temporal"
)

const ExtractPackageName = "extract-package"

type ExtractPackageParams struct {
	Path string
	Key  string
}

type ExtractPackageResult struct {
	Path string
}

type ExtractPackage struct{}

func NewExtractPackage() *ExtractPackage {
	return &ExtractPackage{}
}

func (a *ExtractPackage) Execute(ctx context.Context, params *ExtractPackageParams) (*ExtractPackageResult, error) {
	iface, err := archiver.ByExtension(params.Key)
	if err != nil {
		return nil, temporal.NewNonRetryableError(fmt.Errorf("couldn't find a decompressor for %s: %v", params.Key, err))
	}

	unar, ok := iface.(archiver.Unarchiver)
	if !ok {
		return nil, temporal.NewNonRetryableError(fmt.Errorf("couldn't find a decompressor for %s", params.Path))
	}

	tempDir, err := os.MkdirTemp(filepath.Dir(params.Path), "package-*")
	if err != nil {
		return nil, temporal.NewNonRetryableError(fmt.Errorf("error creating temporary directory: %v", err))
	}

	if err := unar.Unarchive(params.Path, tempDir); err != nil {
		return nil, temporal.NewNonRetryableError(fmt.Errorf("error extracting file %s: %v", params.Path, err))
	}

	entries, err := os.ReadDir(tempDir)
	if err != nil {
		return nil, temporal.NewNonRetryableError(fmt.Errorf("error reading extracted package directory: %v", err))
	}

	if len(entries) == 0 {
		return nil, temporal.NewNonRetryableError(fmt.Errorf("no entry found in extracted package directory"))
	}

	path := ""
	if len(entries) > 1 {
		path = tempDir
		// return nil, temporal.NewNonRetryableError(fmt.Errorf("more than one entry found in extracted package directory"))
	} else {
		path = filepath.Join(tempDir, entries[0].Name())
	}

	return &ExtractPackageResult{Path: path}, nil
}
