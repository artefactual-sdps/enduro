package activities

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"go.artefactual.dev/tools/ref"
	"go.artefactual.dev/tools/temporal"

	goaingest "github.com/artefactual-sdps/enduro/internal/api/gen/ingest"
	goastorage "github.com/artefactual-sdps/enduro/internal/api/gen/storage"
	"github.com/artefactual-sdps/enduro/internal/enums"
	"github.com/artefactual-sdps/enduro/internal/ingest"
	storage_enums "github.com/artefactual-sdps/enduro/internal/storage/enums"
)

const ClearIngestedSIPsActivityName = "clear-ingested-sips-activity"

type ClearIngestedSIPsActivity struct {
	ingestsvc     ingest.Service
	storageClient ingest.StorageClient
	pollInterval  time.Duration
}

type ClearIngestedSIPsActivityParams struct {
	BatchUUID uuid.UUID
}

type ClearIngestedSIPsActivityResult struct{}

func NewClearIngestedSIPsActivity(
	ingestsvc ingest.Service,
	storageClient ingest.StorageClient,
	pollInterval time.Duration,
) *ClearIngestedSIPsActivity {
	return &ClearIngestedSIPsActivity{
		ingestsvc:     ingestsvc,
		storageClient: storageClient,
		pollInterval:  pollInterval,
	}
}

func (a *ClearIngestedSIPsActivity) Execute(
	ctx context.Context,
	params *ClearIngestedSIPsActivityParams,
) (*ClearIngestedSIPsActivityResult, error) {
	h := temporal.StartAutoHeartbeat(ctx)
	defer h.Stop()

	logger := temporal.GetLogger(ctx)
	logger.V(1).Info("Executing ClearIngestedSIPsActivity", "params", params)

	// Paginate through all SIPs in the batch.
	var sipCount int
	var errs error
	var aipIDs []string
	for {
		result, err := a.ingestsvc.ListSips(ctx, &goaingest.ListSipsPayload{
			BatchUUID: ref.New(params.BatchUUID.String()),
			Limit:     ref.New(1000),
			Offset:    ref.New(sipCount),
		})
		if err != nil {
			return nil, fmt.Errorf("list SIPs: %v", err)
		}

		for _, sip := range result.Items {
			status, err := enums.ParseSIPStatus(sip.Status)
			if err != nil {
				errs = errors.Join(errs, fmt.Errorf("parse SIP %q status: %v", sip.UUID.String(), err))
				continue
			}

			// If the SIP was ingested, mark it as canceled.
			if status == enums.SIPStatusIngested {
				if err := a.ingestsvc.SetStatus(ctx, sip.UUID, enums.SIPStatusCanceled); err != nil {
					errs = errors.Join(
						errs, fmt.Errorf("set SIP %q status to canceled: %v", sip.UUID.String(), err),
					)
					continue
				}
			}

			// Ignore SIPs that are not ingested or canceled, or if they don't have an AIP UUID.
			if status != enums.SIPStatusIngested && status != enums.SIPStatusCanceled {
				continue
			}
			if sip.AipUUID == nil || *sip.AipUUID == "" {
				continue
			}

			// Request deletion of the AIP.
			if err := a.deleteAIP(ctx, params.BatchUUID, *sip.AipUUID); err != nil {
				errs = errors.Join(errs, err)
				continue
			}

			// Keep track of the AIP IDs for polling.
			aipIDs = append(aipIDs, *sip.AipUUID)
		}

		sipCount += len(result.Items)
		if sipCount == 0 || result.Page == nil || sipCount >= result.Page.Total {
			break
		}
	}

	if len(aipIDs) == 0 {
		return &ClearIngestedSIPsActivityResult{}, errs
	}

	// Poll the status of the AIPs until they are all deleted.
	if err := a.waitForAIPsDeleted(ctx, aipIDs); err != nil {
		errs = errors.Join(errs, err)
	}

	return &ClearIngestedSIPsActivityResult{}, errs
}

// deleteAIP requests deletion of the AIP with the given ID.
// If the AIP is already deleted, it returns nil.
func (a *ClearIngestedSIPsActivity) deleteAIP(ctx context.Context, batchUUID uuid.UUID, aipID string) error {
	err := a.storageClient.AipDeletionAuto(ctx, &goastorage.AipDeletionAutoPayload{
		UUID:       aipID,
		Reason:     fmt.Sprintf("Batch %s canceled", batchUUID),
		SkipReport: ref.New(true),
	})
	if err != nil {
		aip, showErr := a.storageClient.ShowAip(ctx, &goastorage.ShowAipPayload{UUID: aipID})
		if showErr != nil {
			return fmt.Errorf("request AIP %q deletion: %v", aipID, errors.Join(err, showErr))
		}
		if aip.Status == storage_enums.AIPStatusDeleted.String() {
			return nil
		}
		return fmt.Errorf("request AIP %q deletion: %v", aipID, err)
	}

	return nil
}

// waitForAIPsDeleted polls the status of the AIPs with the given IDs
// until they are all deleted or the context is canceled.
func (a *ClearIngestedSIPsActivity) waitForAIPsDeleted(ctx context.Context, aipIDs []string) error {
	ticker := time.NewTicker(a.pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			done, err := a.checkAIPStatuses(ctx, aipIDs)
			if err != nil {
				return err
			}
			if done {
				return nil
			}
		}
	}
}

// checkAIPStatuses checks the status of the AIPs with the given IDs.
// It returns true if all AIPs are deleted, false otherwise.
func (a *ClearIngestedSIPsActivity) checkAIPStatuses(ctx context.Context, aipIDs []string) (bool, error) {
	var errs error
	for _, aipID := range aipIDs {
		aip, err := a.storageClient.ShowAip(ctx, &goastorage.ShowAipPayload{UUID: aipID})
		if err != nil {
			errs = errors.Join(errs, fmt.Errorf("show AIP %q: %v", aipID, err))
			continue
		}

		if aip.Status == storage_enums.AIPStatusStored.String() {
			errs = errors.Join(errs, fmt.Errorf("AIP %q could not be deleted", aipID))
			continue
		}

		if aip.Status != storage_enums.AIPStatusDeleted.String() {
			return false, nil
		}
	}

	return true, errs
}
