package fsutil

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/otiai10/copy"
)

// Used for testing.
var Renamer = os.Rename

func BaseNoExt(path string) string {
	base := filepath.Base(path)
	if before, _, ok := strings.Cut(base, "."); ok {
		return before
	}
	return base
}

// Move moves files or directories. It copies the contents when the move op
// failes because source and destination do not share the same filesystem.
func Move(src, dst string) error {
	if _, err := os.Stat(dst); err == nil {
		return errors.New("destination already exists")
	}

	// Move when possible.
	err := Renamer(src, dst)
	if err == nil {
		return nil
	}

	// Copy and delete otherwise.
	lerr, _ := err.(*os.LinkError)
	if lerr.Err.Error() == "invalid cross-device link" {
		err := copy.Copy(src, dst, copy.Options{
			Sync:        true,
			OnDirExists: func(src, dst string) copy.DirExistsAction { return copy.Untouchable },
		})
		if err != nil {
			return err
		}
		return os.RemoveAll(src)
	}

	return err
}

// SetFileModes recursively sets the file mode of root and its contents.
func SetFileModes(root string, dirMode, fileMode fs.FileMode) error {
	info, err := os.Stat(root)
	if err != nil {
		return err
	}

	if !info.IsDir() {
		if err := os.Chmod(root, fileMode); err != nil {
			return fmt.Errorf("set permissions: %v", err)
		}

		return nil
	}

	scopedRoot, err := os.OpenRoot(root)
	if err != nil {
		return err
	}
	defer scopedRoot.Close()

	return fs.WalkDir(scopedRoot.FS(), ".",
		func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}

			mode := fileMode
			if d.IsDir() {
				mode = dirMode
			}

			if err := scopedRoot.Chmod(path, mode); err != nil {
				return fmt.Errorf("set permissions: %v", err)
			}

			return nil
		},
	)
}

// FileExists returns true if a file (or directory) exists at path.  If a file
// exists but os.Stat() returns an error (e.g. insufficient permissions)
// FileExists will return false.
func FileExists(path string) bool {
	if _, err := os.Stat(path); err != nil {
		return false
	}
	return true
}
