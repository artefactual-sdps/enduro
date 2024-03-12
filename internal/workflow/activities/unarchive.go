package activities

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/go-logr/logr"
	"github.com/mholt/archiver/v3"

	"github.com/artefactual-sdps/enduro/internal/fsutil"
)

const UnarchiveActivityName = "unarchive-activity"

// UnarchiveActivity extracts transfer files from an archive (e.g. zip, tgz).
type UnarchiveActivity struct {
	logger logr.Logger
}

type UnarchiveActivityParams struct {
	// SourcePath is the path of the archive file to be extracted.
	SourcePath string

	// StripTopLevelDir indicates whether the top-level "container" directory of
	// the archive should be removed from extract directory.
	StripTopLevelDir bool
}

type UnarchiveActivityResult struct {
	// DestPath is the path to the extracted archive contents.
	DestPath string

	// IsDir is true if DestPath is a directory.
	IsDir bool
}

func NewUnarchiveActivity(logger logr.Logger) *UnarchiveActivity {
	return &UnarchiveActivity{logger: logger}
}

// Execute attempts to unarchive the archive file at SourcePath to a temporary
// directory to DestPath. If SourcePath points to a non-archive file then the
// returned DestPath will be equal to SourcePath.
func (a *UnarchiveActivity) Execute(ctx context.Context, params *UnarchiveActivityParams) (*UnarchiveActivityResult, error) {
	a.logger.V(1).Info("Executing UnarchiveActivity",
		"SourcePath", params.SourcePath,
		"StripTopLevelDir", params.StripTopLevelDir,
	)

	u, err := unarchiver(params.SourcePath)
	if err != nil {
		return nil, fmt.Errorf("unarchive: unarchiver: %v", err)
	}
	if u == nil {
		// Couldn't find an unarchiver, so this is probably a regular file.
		// Return the source path unaltered, and IsDir as false.
		a.logger.V(2).Info("Unarchive: not an archive, skipping.")
		return &UnarchiveActivityResult{DestPath: params.SourcePath}, nil
	}

	dest := filepath.Join(filepath.Dir(params.SourcePath), "extract")
	if err := u.Unarchive(params.SourcePath, dest); err != nil {
		return nil, fmt.Errorf("unarchive: unarchive: %v", err)
	}

	if params.StripTopLevelDir {
		if err = stripDirContainer(dest); err != nil {
			return nil, fmt.Errorf("unarchive: strip top-level dir: %v", err)
		}
	}

	if err := fsutil.SetFileModes(dest, ModeDir, ModeFile); err != nil {
		return nil, fmt.Errorf("unarchive: %v", err)
	}

	if err := os.Remove(params.SourcePath); err != nil {
		a.logger.Error(err, "Unarchive: couldn't delete source archive.")
	}

	return &UnarchiveActivityResult{DestPath: dest, IsDir: true}, err
}

// Unarchiver returns the unarchiver suited for the archival format.
func unarchiver(filename string) (archiver.Unarchiver, error) {
	if iface, err := archiver.ByExtension(filename); err == nil {
		if u, ok := iface.(archiver.Unarchiver); ok {
			return u, nil
		}
	}

	file, err := os.Open(filepath.Clean(filename))
	if err != nil {
		return nil, fmt.Errorf("open file: %v", err)
	}
	defer file.Close()

	if u, err := archiver.ByHeader(file); err == nil {
		return u, nil
	}

	return nil, nil
}

// stripDirContainer strips the top-level directory of a transfer.
func stripDirContainer(dir string) error {
	tld, err := topLevelDir(dir)
	if err != nil {
		return fmt.Errorf("get top-level dir: %v", err)
	}

	tmpDir, err := os.MkdirTemp("", "enduro-")
	if err != nil {
		return fmt.Errorf("make temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Move the top-level directory and contents to tmpPath.
	tmpPath := filepath.Join(tmpDir, tld)
	if err := os.Rename(filepath.Join(dir, tld), tmpPath); err != nil {
		return fmt.Errorf("move to temp dir: %v", err)
	}

	// Move the TLD contents back to the original path.
	err = filepath.WalkDir(tmpPath, func(path string, d fs.DirEntry, err error) error {
		if path == tmpPath {
			return nil
		}

		if err := os.Rename(path, filepath.Join(dir, d.Name())); err != nil {
			return fmt.Errorf("move to original dir: %v", err)
		}

		// Don't descend into sub-directories.
		if d.IsDir() {
			return filepath.SkipDir
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("walk dir: %v", err)
	}

	return nil
}

// topLevelDir reads the directory at path and returns the name of it's
// immediate sub-directory. If path contains anything other than a single
// sub-directory then topLevelDirectory will return an error.
func topLevelDir(path string) (string, error) {
	r, err := os.Open(filepath.Clean(path))
	if err != nil {
		return "", fmt.Errorf("open path: %v", err)
	}
	defer r.Close()

	fis, err := r.Readdir(2)
	if err != nil {
		return "", fmt.Errorf("read dir: %v", err)
	}
	if len(fis) == 0 {
		return "", fmt.Errorf("directory %q is empty", path)
	}
	if len(fis) > 1 {
		return "", fmt.Errorf("directory %q has more than one child", path)
	}
	if !fis[0].IsDir() {
		return "", fmt.Errorf("%q is not a directory", path+fis[0].Name())
	}

	return fis[0].Name(), nil
}
