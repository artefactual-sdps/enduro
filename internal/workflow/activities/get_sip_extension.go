package activities

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/mholt/archiver/v4"
	"go.artefactual.dev/tools/temporal"
)

const GetSIPExtensionActivityName = "get-sip-extension-activity"

var ErrInvalidArchive = fmt.Errorf("%s: identify SIP format: %s", GetSIPExtensionActivityName, "invalid archive")

type (
	GetSIPExtensionActivity       struct{}
	GetSIPExtensionActivityParams struct {
		// Path is the full path of the SIP.
		Path string
	}
	GetSIPExtensionActivityResult struct {
		// Extension of the SIP file.
		Extension string
	}
)

func NewGetSIPExtensionActivity() *GetSIPExtensionActivity {
	return &GetSIPExtensionActivity{}
}

func (a *GetSIPExtensionActivity) Execute(
	ctx context.Context,
	params *GetSIPExtensionActivityParams,
) (*GetSIPExtensionActivityResult, error) {
	logger := temporal.GetLogger(ctx)
	logger.V(1).Info(fmt.Sprintf("Executing %s", GetSIPExtensionActivityName), "Path", params.Path)

	f, err := os.Open(params.Path) // #nosec G304 -- trusted path.
	if err != nil {
		return nil, fmt.Errorf("%s: open SIP file: %v", GetSIPExtensionActivityName, err)
	}
	defer f.Close()

	// TODO: Use github.com/mholt/archives. We still use the archived github.com/mholt/archiver/v4
	// in some activities, and using both causes a panic registering the same compressors.
	format, _, err := archiver.Identify(params.Path, f)
	if err != nil {
		if errors.Is(err, archiver.ErrNoMatch) {
			return nil, ErrInvalidArchive
		}
		return nil, fmt.Errorf("%s: identify SIP format: %v", GetSIPExtensionActivityName, err)
	}

	return &GetSIPExtensionActivityResult{Extension: format.Name()}, nil
}
