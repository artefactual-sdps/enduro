package localact

import (
	"context"

	"github.com/artefactual-sdps/enduro/internal/bagit"
	"github.com/artefactual-sdps/enduro/internal/enums"
)

type (
	ClassifyPackageActivityParams struct {
		// Path is the full path of the package.
		Path string
	}

	ClassifyPackageActivityResult struct {
		// Type of the package.
		Type enums.PackageType
	}
)

func ClassifyPackageActivity(
	ctx context.Context,
	params ClassifyPackageActivityParams,
) (*ClassifyPackageActivityResult, error) {
	r := ClassifyPackageActivityResult{Type: enums.PackageTypeUnknown}
	if bagit.IsABag(params.Path) {
		r.Type = enums.PackageTypeBagIt
	}

	return &r, nil
}
