package activities

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/artefactual-sdps/enduro/internal/datatypes"
	"github.com/artefactual-sdps/enduro/internal/ingest"
)

const CheckDuplicateSIPActivityName = "check-duplicate-sip-activity"

type (
	CheckDuplicateSIPActivity struct {
		ingestSvc ingest.Service
	}
	CheckDuplicateSIPActivityParams struct {
		SIPID    uuid.UUID
		Checksum datatypes.Checksum
	}
	CheckDuplicateSIPActivityResult struct {
		Duplicate *datatypes.SIP
	}
)

func NewCheckDuplicateSIPActivity(svc ingest.Service) *CheckDuplicateSIPActivity {
	return &CheckDuplicateSIPActivity{ingestSvc: svc}
}

func (a *CheckDuplicateSIPActivity) Execute(
	ctx context.Context,
	params CheckDuplicateSIPActivityParams,
) (*CheckDuplicateSIPActivityResult, error) {
	duplicate, err := a.ingestSvc.FindDuplicateSIP(ctx, params.SIPID, params.Checksum)
	if err != nil {
		return nil, fmt.Errorf("check duplicate SIP: %w", err)
	}

	return &CheckDuplicateSIPActivityResult{Duplicate: duplicate}, nil
}
