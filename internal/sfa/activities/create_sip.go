package activities

import (
	"context"
	"errors"
	"os/exec"
)

const SipCreationName = "sip-creation"

type SipCreationActivity struct{}

func NewSipCreationActivity() *SipCreationActivity {
	return &SipCreationActivity{}
}

type SipCreationParams struct {
	SipPath string
}

type SipCreationResult struct {
	Out        string
	NewSipPath string
}

func (sc *SipCreationActivity) Execute(ctx context.Context, params *SipCreationParams) (*SipCreationResult, error) {
	res := &SipCreationResult{}
	e, err := exec.Command("python3", "repackage_sip.py", params.SipPath).CombinedOutput() // #nosec G204
	if err != nil {
		return nil, err
	}
	res.Out = string(e)
	if res.Out != params.SipPath+"_bag\n" {
		return nil, errors.New("Failed to repackage sip correctly: " + res.Out)
	}

	res.NewSipPath = params.SipPath + "_bag"
	return res, nil
}
