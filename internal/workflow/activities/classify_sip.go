package activities

import (
	"context"
	"fmt"

	"go.artefactual.dev/tools/temporal"

	"github.com/artefactual-sdps/enduro/internal/bagit"
	"github.com/artefactual-sdps/enduro/internal/enums"
)

const ClassifySIPActivityName = "classify-sip-activity"

type (
	ClassifySIPActivity       struct{}
	ClassifySIPActivityParams struct {
		// Path is the full path of the SIP.
		Path string
	}
	ClassifySIPActivityResult struct {
		// Type of the SIP.
		Type enums.SIPType
	}
)

func NewClassifySIPActivity() *ClassifySIPActivity {
	return &ClassifySIPActivity{}
}

func (a *ClassifySIPActivity) Execute(
	ctx context.Context,
	params ClassifySIPActivityParams,
) (*ClassifySIPActivityResult, error) {
	logger := temporal.GetLogger(ctx)
	logger.V(1).Info(
		fmt.Sprintf("Executing %s", ClassifySIPActivityName),
		"Path", params.Path,
	)

	r := ClassifySIPActivityResult{Type: enums.SIPTypeUnknown}
	if bagit.Is(params.Path) {
		r.Type = enums.SIPTypeBagIt
	}

	return &r, nil
}
