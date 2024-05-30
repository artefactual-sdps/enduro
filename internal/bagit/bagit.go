package bagit

import (
	go_bagit "github.com/nyudlts/go-bagit"
)

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

// Valid tests whether the bag at path is complete and the file checksums are
// valid.
func Valid(path string) error {
	bag, err := go_bagit.GetExistingBag(path)
	if err != nil {
		return err
	}

	return bag.ValidateBag(false, false)
}
