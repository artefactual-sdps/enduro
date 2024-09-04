package activities

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	securejoin "github.com/cyphar/filepath-securejoin"
	"github.com/otiai10/copy"
	temporal_tools "go.artefactual.dev/tools/temporal"

	"github.com/artefactual-sdps/enduro/internal/bagit"
	"github.com/artefactual-sdps/enduro/internal/bundler"
	"github.com/artefactual-sdps/enduro/internal/fsutil"
)

const (
	ModeDir  = 0o750
	ModeFile = 0o640
)

type BundleActivity struct{}

func NewBundleActivity() *BundleActivity {
	return &BundleActivity{}
}

type BundleActivityParams struct {
	// SourcePath is the path of the transfer file or directory.
	SourcePath string

	// TransferDir is the target directory for the bundled package.
	TransferDir string

	// IsDir indicates that the transfer is a local directory when true. If true
	// the transfer will be copied to TransferDir without modification.
	IsDir bool
}

type BundleActivityResult struct {
	FullPath string // Full path to the transfer in the worker running the session.
}

func (a *BundleActivity) Execute(ctx context.Context, params *BundleActivityParams) (*BundleActivityResult, error) {
	var (
		res = &BundleActivityResult{}
		err error
	)

	logger := temporal_tools.GetLogger(ctx)
	logger.V(1).Info("Executing BundleActivity",
		"SourcePath", params.SourcePath,
		"TransferDir", params.TransferDir,
		"IsDir", params.IsDir,
	)

	if params.TransferDir == "" {
		params.TransferDir, err = os.MkdirTemp("", "*-enduro-transfer")
		if err != nil {
			return nil, err
		}
	}

	if params.IsDir {
		res.FullPath, err = a.Bundle(ctx, params.SourcePath, params.TransferDir)
		if err != nil {
			err = fmt.Errorf("bundle dir: %v", err)
		}
	} else {
		res.FullPath, err = a.SingleFile(ctx, params.SourcePath, params.TransferDir)
		if err != nil {
			err = fmt.Errorf("bundle single file: %v", err)
		}
	}
	if err != nil {
		return nil, temporal_tools.NewNonRetryableError(err)
	}

	err = unbag(res.FullPath)
	if err != nil {
		return nil, temporal_tools.NewNonRetryableError(
			fmt.Errorf("bundle: unbag: %v", err),
		)
	}

	if err = fsutil.SetFileModes(res.FullPath, ModeDir, ModeFile); err != nil {
		return nil, temporal_tools.NewNonRetryableError(
			fmt.Errorf("bundle: set permissions: %v", err),
		)
	}

	return res, nil
}

// SingleFile bundles a transfer with the downloaded blob in it.
//
// TODO: Write metadata.csv and checksum files to the metadata dir.
func (a *BundleActivity) SingleFile(
	ctx context.Context,
	sourcePath string,
	transferDir string,
) (string, error) {
	b, err := bundler.NewBundlerWithTempDir(transferDir)
	if err != nil {
		return "", fmt.Errorf("create bundler: %v", err)
	}

	src, err := os.Open(sourcePath) // #nosec G304 -- trusted file path.
	if err != nil {
		return "", fmt.Errorf("open source file: %v", err)
	}
	defer src.Close()

	err = b.Write(filepath.Join("objects", filepath.Base(sourcePath)), src)
	if err != nil {
		return "", fmt.Errorf("write file: %v", err)
	}

	if err := b.Bundle(); err != nil {
		return "", fmt.Errorf("write bundle: %v", err)
	}

	return b.FullBaseFsPath(), nil
}

// Bundle a transfer with the contents found in the archive.
func (a *BundleActivity) Bundle(
	ctx context.Context,
	sourcePath string,
	transferDir string,
) (string, error) {
	tempDir, err := a.Copy(ctx, sourcePath, transferDir)
	if err != nil {
		return "", fmt.Errorf("bundle: %v", err)
	}

	// Delete the archive. We still have a copy in the watched source.
	_ = os.Remove(sourcePath)

	return tempDir, nil
}

// Copy a transfer in the given destination using an intermediate temp. directory.
func (a *BundleActivity) Copy(ctx context.Context, src, dst string) (string, error) {
	const prefix = "enduro"
	tempDir, err := os.MkdirTemp(dst, prefix)
	if err != nil {
		return "", fmt.Errorf("error creating temporary directory: %s", err)
	}

	if err := copy.Copy(src, tempDir); err != nil {
		return "", fmt.Errorf("error copying transfer: %v", err)
	}

	return tempDir, nil
}

// unbag converts a bagged transfer into a standard Archivematica transfer.
// It returns a nil error if a bag is not identified, and non-nil errors when
// the bag seems invalid, without verifying the actual file contents.
func unbag(path string) error {
	fi, err := os.Stat(path)
	if err != nil {
		return err
	}
	if !fi.IsDir() {
		return errors.New("not a directory")
	}

	// Only continue if we have a bag.
	securePath, _ := securejoin.SecureJoin(path, "bagit.txt")
	if !fsutil.FileExists(securePath) {
		return nil
	}

	// Confirm completeness of the bag.
	if err := bagit.Complete(path); err != nil {
		return err
	}

	// Move files in data up one level if 'objects' folder already exists.
	// Otherwise, rename data to objects.
	dataPath, _ := securejoin.SecureJoin(path, "data")
	if fi, err := os.Stat(dataPath); !os.IsNotExist(err) && fi.IsDir() {
		items, err := os.ReadDir(dataPath)
		if err != nil {
			return err
		}
		for _, item := range items {
			src, _ := securejoin.SecureJoin(dataPath, item.Name())
			dst, _ := securejoin.SecureJoin(path, filepath.Base(src))
			if err := os.Rename(src, dst); err != nil {
				return err
			}
		}
		if err := os.RemoveAll(dataPath); err != nil {
			return err
		}
	} else {
		dst, _ := securejoin.SecureJoin(path, "objects")
		if err := os.Rename(dataPath, dst); err != nil {
			return err
		}
	}

	// Create metadata and submissionDocumentation directories.
	metadataPath, _ := securejoin.SecureJoin(path, "metadata")
	err = os.MkdirAll(metadataPath, ModeDir)
	if err != nil {
		return err
	}

	docPath, _ := securejoin.SecureJoin(metadataPath, "submissionDocumentation")
	err = os.MkdirAll(docPath, ModeDir)
	if err != nil {
		return err
	}

	// Write manifest checksums to checksum file.
	for _, item := range [][2]string{
		{"manifest-sha512.txt", "checksum.sha512"},
		{"manifest-sha256.txt", "checksum.sha256"},
		{"manifest-sha1.txt", "checksum.sha1"},
		{"manifest-md5.txt", "checksum.md5"},
	} {
		securePath, _ := securejoin.SecureJoin(path, item[0])
		file, err := os.Open(
			securePath,
		) //#nosec G304 -- Potential file inclusion not possible. item[0] is coming from controlled list.
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				continue
			}
			return err
		}
		defer file.Close()

		securePath, _ = securejoin.SecureJoin(metadataPath, item[1])
		newFile, err := os.Create(
			securePath,
		) //#nosec G304 -- Potential file inclusion not possible. item[1] is coming from controlled list.
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				continue
			}
			return err
		}
		defer newFile.Close()

		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			line := scanner.Text()
			newLine := ""
			if strings.Contains(line, "data/objects/") {
				newLine = strings.Replace(line, "data/objects/", "../objects/", 1)
			} else {
				newLine = strings.Replace(line, "data/", "../objects/", 1)
			}
			fmt.Fprintln(newFile, newLine)
		}

		break // One file is enough.
	}

	// Move bag files to submissionDocumentation.
	for _, item := range []string{
		"bag-info.txt",
		"bagit.txt",
		"manifest-md5.txt",
		"tagmanifest-md5.txt",
		"manifest-sha1.txt",
		"tagmanifest-sha1.txt",
		"manifest-sha256.txt",
		"tagmanifest-sha256.txt",
		"manifest-sha512.txt",
		"tagmanifest-sha512.txt",
	} {
		src, _ := securejoin.SecureJoin(path, item)
		dst, _ := securejoin.SecureJoin(docPath, item)
		_ = os.Rename(src, dst)
	}

	return nil
}
