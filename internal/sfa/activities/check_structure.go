package activities

import (
	"context"
	"fmt"

	"github.com/artefactual-sdps/enduro/internal/sfa/sip"
)

const CheckSipStructureName = "check-sip-structure"

type CheckSipStructureActivity struct{}

func NewCheckSipStructure() *CheckSipStructureActivity {
	return &CheckSipStructureActivity{}
}

type CheckSipStructureParams struct {
	SipPath string
}

type CheckSipStructureResult struct {
	Ok     bool
	SIP    *sip.SFASip
	Errors []string
}

func (md *CheckSipStructureActivity) Execute(ctx context.Context, params *CheckSipStructureParams) (*CheckSipStructureResult, error) {
	res := &CheckSipStructureResult{}
	s, err := sip.NewSFASip(params.SipPath)
	if err != nil {
		return nil, err
	}

	if s.Content == nil {
		res.Errors = append(res.Errors, "content folder is missing")
	}
	if s.Header == nil {
		res.Errors = append(res.Errors, "header folder is missing")
	}
	if !s.MetadataPresent {
		res.Errors = append(res.Errors, "metadata.xml is missing")
	}
	if !s.XSDPresent {
		res.Errors = append(res.Errors, "XSD folder is missing")
	}
	for _, f := range s.Unexpected {
		res.Errors = append(res.Errors, fmt.Sprintf("unexpected file or folder: %s", f))
	}

	if len(res.Errors) == 0 {
		res.Ok = true
	}

	res.SIP = s
	return res, nil
}
