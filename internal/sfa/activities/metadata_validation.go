package activities

import (
	"context"
	"errors"
	"os/exec"
	"path/filepath"
)

const MetadataValidationName = "metadata-validation"

type MetadataValidationActivity struct{}

func NewMetadataValidationActivity() *MetadataValidationActivity {
	return &MetadataValidationActivity{}
}

type MetadataValidationParams struct {
	SipPath string
}

type MetadataValidationResult struct {
	Out string
}

func (md *MetadataValidationActivity) Execute(ctx context.Context, params *MetadataValidationParams) (*MetadataValidationResult, error) {
	res := &MetadataValidationResult{}
	e, err := exec.Command("python3",
		"xsdval.py",
		// Arguments.
		filepath.Join(params.SipPath, "/header/metadata.xml"),
		"arelda.xsd").CombinedOutput() // #nosec G204
	if err != nil {
		return nil, err
	}

	res.Out = string(e)
	if res.Out != "Is metadata.xml valid:  True\n" {
		return nil, errors.New("Failed to validate metadata files: " + res.Out)
	}
	return res, nil
}
