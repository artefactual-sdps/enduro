package activities

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"go.artefactual.dev/tools/fsutil"

	"github.com/artefactual-sdps/enduro/internal/enums"
)

const CountSIPFilesActivityName = "count-files-activity"

type (
	CountSIPFilesActivity struct{}

	CountSIPFilesActivityParams struct {
		// Path is the Bag filepath.
		Path string

		// SIPType is the SIP type (e.g. BagIt) which may be used to determine
		// where the preservation files are located in the SIP.
		SIPType enums.SIPType
	}

	CountSIPFilesActivityResult struct {
		// Count is the number of preservation files in the Bag.
		Count int
	}
)

func NewCountSIPFilesActivity() *CountSIPFilesActivity {
	return &CountSIPFilesActivity{}
}

// Execute counts the number of preservation files in the SIP at params.Path.
//
// If the SIP is a BagIt Bag, Execute counts the files in the "data" directory.
// If the SIP type is unknown, Execute counts all the files in the SIP.
func (a *CountSIPFilesActivity) Execute(
	ctx context.Context,
	params *CountSIPFilesActivityParams,
) (*CountSIPFilesActivityResult, error) {
	path := params.Path
	if params.SIPType == enums.SIPTypeBagIt {
		// For BagIt Bags only count the files in the "data" directory.
		path = filepath.Join(params.Path, "data")
	}

	found, err := fsutil.Exists(path)
	if err != nil {
		return nil, fmt.Errorf("count SIP files: %v", err)
	}
	if !found {
		return nil, fmt.Errorf("count SIP files: directory not found: %s", path)
	}

	var count int
	err = filepath.WalkDir(path, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			count++
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("count SIP files: walk dir: %v", err)
	}

	return &CountSIPFilesActivityResult{Count: count}, nil
}
