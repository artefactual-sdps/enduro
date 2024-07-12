package activities

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"

	"github.com/artefactual-sdps/enduro/internal/bagit"
	"github.com/artefactual-sdps/enduro/internal/enums"
)

const ClassifyPackageActivityName = "classify-package-activity"

type (
	ClassifyPackageActivity struct {
		logger logr.Logger
	}
	ClassifyPackageActivityParams struct {
		// Path is the full path of the package.
		Path string
	}
	ClassifyPackageActivityResult struct {
		// Type of the package.
		Type enums.PackageType
	}
)

func NewClassifyPackageActivity(logger logr.Logger) *ClassifyPackageActivity {
	return &ClassifyPackageActivity{logger: logger}
}

func (a *ClassifyPackageActivity) Execute(
	ctx context.Context,
	params ClassifyPackageActivityParams,
) (*ClassifyPackageActivityResult, error) {
	a.logger.V(1).Info(
		fmt.Sprintf("Executing %s", ClassifyPackageActivityName),
		"Path", params.Path,
	)

	r := ClassifyPackageActivityResult{Type: enums.PackageTypeUnknown}
	if bagit.Is(params.Path) {
		r.Type = enums.PackageTypeBagIt
	}

	return &r, nil
}
