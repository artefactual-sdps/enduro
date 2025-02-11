package activities

import (
	"context"
	"fmt"

	"go.artefactual.dev/tools/temporal"

	"github.com/artefactual-sdps/enduro/internal/bagit"
	"github.com/artefactual-sdps/enduro/internal/enums"
)

const ClassifyPackageActivityName = "classify-package-activity"

type (
	ClassifyPackageActivity       struct{}
	ClassifyPackageActivityParams struct {
		// Path is the full path of the package.
		Path string
	}
	ClassifyPackageActivityResult struct {
		// Type of the package.
		Type enums.SIPType
	}
)

func NewClassifyPackageActivity() *ClassifyPackageActivity {
	return &ClassifyPackageActivity{}
}

func (a *ClassifyPackageActivity) Execute(
	ctx context.Context,
	params ClassifyPackageActivityParams,
) (*ClassifyPackageActivityResult, error) {
	logger := temporal.GetLogger(ctx)
	logger.V(1).Info(
		fmt.Sprintf("Executing %s", ClassifyPackageActivityName),
		"Path", params.Path,
	)

	r := ClassifyPackageActivityResult{Type: enums.SIPTypeUnknown}
	if bagit.Is(params.Path) {
		r.Type = enums.SIPTypeBagIt
	}

	return &r, nil
}
