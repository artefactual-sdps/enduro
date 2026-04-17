package activities

import (
	"context"
	"fmt"
	"slices"
	"time"

	"github.com/google/uuid"
	"go.artefactual.dev/tools/ref"
	temporal_tools "go.artefactual.dev/tools/temporal"

	goaingest "github.com/artefactual-sdps/enduro/internal/api/gen/ingest"
	"github.com/artefactual-sdps/enduro/internal/datatypes"
	"github.com/artefactual-sdps/enduro/internal/entfilter"
	"github.com/artefactual-sdps/enduro/internal/enums"
	"github.com/artefactual-sdps/enduro/internal/ingest"
)

// finalStatuses lists all SIP statuses that are considered terminal. Any SIP
// reporting one of these statuses is treated as final by the polling logic,
// except the caller's ExpectedStatus, which is checked separately for success.
var finalStatuses = []enums.SIPStatus{
	enums.SIPStatusIngested,
	enums.SIPStatusFailed,
	enums.SIPStatusError,
	enums.SIPStatusCanceled,
}

const PollSIPStatusesActivityName = "poll-sip-statuses-activity"

type PollSIPStatusesActivityParams struct {
	BatchUUID        uuid.UUID
	ExpectedSIPCount int
	ExpectedStatus   enums.SIPStatus
}

type PollSIPStatusesActivityResult struct {
	// AllExpectedStatus is true if all SIPs in the batch have the
	// ExpectedStatus.
	AllExpectedStatus bool

	// SIPs is a sparse list of the SIPs in this batch, containing only the
	// AIPID and FileCount fields, so they can be passed to the postbatch
	// workflow without needing to query the ingest service again.
	SIPs map[uuid.UUID]datatypes.SIP
}

// PollSIPStatusesActivity polls the ingest service until all SIPs in a batch
// reach a final state, failing immediately if the SIP count differs from
// ExpectedSIPCount or any SIP reports an invalid final status.
type PollSIPStatusesActivity struct {
	ingestsvc    ingest.Service
	pollInterval time.Duration
}

func NewPollSIPStatusesActivity(ingestsvc ingest.Service, pollInterval time.Duration) *PollSIPStatusesActivity {
	return &PollSIPStatusesActivity{
		ingestsvc:    ingestsvc,
		pollInterval: pollInterval,
	}
}

func (a *PollSIPStatusesActivity) Execute(
	ctx context.Context,
	params *PollSIPStatusesActivityParams,
) (*PollSIPStatusesActivityResult, error) {
	h := temporal_tools.StartAutoHeartbeat(ctx)
	defer h.Stop()

	ticker := time.NewTicker(a.pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-ticker.C:
			result, err := a.checkSIPStatuses(ctx, params.BatchUUID, params.ExpectedSIPCount, params.ExpectedStatus)
			if err != nil {
				return nil, fmt.Errorf("check SIP statuses: %v", err)
			}
			if result.done {
				return &PollSIPStatusesActivityResult{
					AllExpectedStatus: result.allExpected,
					SIPs:              result.sips,
				}, nil
			}
		}
	}
}

type checkResult struct {
	done        bool
	allExpected bool
	sips        map[uuid.UUID]datatypes.SIP
}

func (a *PollSIPStatusesActivity) checkSIPStatuses(
	ctx context.Context,
	batchUUID uuid.UUID,
	expectedSIPCount int,
	expectedStatus enums.SIPStatus,
) (*checkResult, error) {
	// Query all SIPs for this batch. Limit to the maximum page size for now,
	// we may switch to a stats-based or aggregated query approach in the future.
	result, err := a.ingestsvc.ListSips(ctx, &goaingest.ListSipsPayload{
		BatchUUID: new(batchUUID.String()),
		Limit:     ref.New(entfilter.MaxPageSize),
	})
	if err != nil {
		return nil, fmt.Errorf("list SIPs: %v", err)
	}

	// Fail if we don't have the expected number of SIPs.
	if len(result.Items) != expectedSIPCount {
		return nil, fmt.Errorf("expected %d SIPs but found %d", expectedSIPCount, len(result.Items))
	}

	expectedStatusCount := 0
	sips := make(map[uuid.UUID]datatypes.SIP, len(result.Items))
	for _, item := range result.Items {
		status, err := enums.ParseSIPStatus(item.Status)
		if err != nil {
			return nil, fmt.Errorf("invalid SIP status: %s", item.Status)
		}

		// Check if this SIP has the expected status.
		// If not and it's not in a final status, keep polling.
		if status == expectedStatus {
			expectedStatusCount++
		} else if !slices.Contains(finalStatuses, status) {
			return &checkResult{done: false}, nil
		}

		sips[item.UUID] = goaSIPtoSIP(item)
	}

	res := &checkResult{
		done:        true,
		allExpected: expectedStatusCount == expectedSIPCount,
		sips:        sips,
	}

	return res, nil
}

func goaSIPtoSIP(goaSIP *goaingest.SIP) datatypes.SIP {
	sip := datatypes.SIP{
		UUID:      goaSIP.UUID,
		FileCount: ref.DerefZero(goaSIP.FileCount),
	}

	if goaSIP.AipUUID != nil {
		id, err := uuid.Parse(*goaSIP.AipUUID)
		if err == nil {
			sip.AIPID = uuid.NullUUID{UUID: id, Valid: true}
		}
	}

	return sip
}
