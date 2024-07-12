package bagit

import (
	"path/filepath"

	go_bagit "github.com/nyudlts/go-bagit"

	"github.com/artefactual-sdps/enduro/internal/fsutil"
)

// Is returns true when dir is a BagIt bag.
func Is(dir string) bool {
	return fsutil.FileExists(filepath.Join(dir, "bagit.txt"))
}

// Complete tests whether the bag at path has the expected number of files and
// total size on disk indicated by the Payload-Oxum, but doesn't do checksum
// validation.
func Complete(path string) error {
	bag, err := go_bagit.GetExistingBag(path)
	if err != nil {
		return err
	}

	return bag.ValidateBag(false, true)
}
