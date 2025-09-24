package activities

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/mholt/archives"
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

	format, _, err := archives.Identify(ctx, params.Path, f)
	if err != nil {
		if errors.Is(err, archives.NoMatch) {
			return nil, ErrInvalidArchive
		}
		return nil, fmt.Errorf("%s: identify SIP format: %v", GetSIPExtensionActivityName, err)
	}

	return &GetSIPExtensionActivityResult{Extension: format.Extension()}, nil
}
