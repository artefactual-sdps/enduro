package activities

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"go.artefactual.dev/tools/fsutil"
)

const CountBagFilesActivityName = "count-bag-files-activity"

type CountBagFilesActivity struct{}

type CountBagFilesActivityParams struct {
	// Path is the Bag filepath.
	Path string
}

type CountBagFilesActivityResult struct {
	// Count is the number of preservation files in the Bag.
	Count int
}

func NewCountBagFilesActivity() *CountBagFilesActivity {
	return &CountBagFilesActivity{}
}

// Execute counts the number of preservation files in the Bag. The Bag
// specification requires all preservation files, and only preservations files,
// to be in the data directory or a subdirectory.
func (a *CountBagFilesActivity) Execute(
	ctx context.Context,
	params *CountBagFilesActivityParams,
) (*CountBagFilesActivityResult, error) {
	dataDir := filepath.Join(params.Path, "data")
	found, err := fsutil.Exists(dataDir)
	if err != nil {
		return nil, fmt.Errorf("count bag files: data dir: %v", err)
	}
	if !found {
		return nil, fmt.Errorf("count bag files: missing data dir: %s", dataDir)
	}

	var count int
	err = filepath.WalkDir(dataDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			count++
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("count bag files: walk data dir: %v", err)
	}

	return &CountBagFilesActivityResult{Count: count}, nil
}
