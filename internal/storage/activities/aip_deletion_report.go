package activities

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jonboulle/clockwork"
	"go.artefactual.dev/tools/ref"

	"github.com/artefactual-sdps/enduro/internal/storage"
	"github.com/artefactual-sdps/enduro/internal/storage/enums"
	"github.com/artefactual-sdps/enduro/internal/storage/persistence"
	"github.com/artefactual-sdps/enduro/internal/storage/reports"
	"github.com/artefactual-sdps/enduro/internal/version"
)

const AIPDeletionReportActivityName = "aip-deletion-report-activity"

type AIPDeletionReportActivity struct {
	cfg        storage.AIPDeletionConfig
	clock      clockwork.Clock
	storageSvc storage.Service
}

type AIPDeletionReportActivityParams struct {
	AIPID          uuid.UUID
	LocationSource enums.LocationSource
}

type AIPDeletionReportActivityResult struct {
	Key string
}

func NewAIPDeletionReportActivity(
	clock clockwork.Clock,
	cfg storage.AIPDeletionConfig,
	svc storage.Service,
) *AIPDeletionReportActivity {
	return &AIPDeletionReportActivity{
		cfg:        cfg,
		clock:      clock,
		storageSvc: svc,
	}
}

func (a *AIPDeletionReportActivity) Execute(
	ctx context.Context,
	params *AIPDeletionReportActivityParams,
) (*AIPDeletionReportActivityResult, error) {
	if a.cfg.ReportTemplatePath == "" {
		return nil, fmt.Errorf("AIP deletion report: template path is not configured")
	}

	key := fmt.Sprintf("%saip_deletion_report_%s.pdf", storage.ReportPrefix, params.AIPID.String())

	dr, err := reports.NewAIPDeletion(a.clock, a.cfg.ReportTemplatePath)
	if err != nil {
		return nil, err
	}

	data, err := a.loadData(ctx, params.AIPID, params.LocationSource)
	if err != nil {
		return nil, err
	}

	loc, err := a.storageSvc.Location(ctx, uuid.Nil)
	if err != nil {
		return nil, fmt.Errorf("AIP deletion report: get storage location: %w", err)
	}

	b, err := loc.OpenBucket(ctx)
	if err != nil {
		return nil, fmt.Errorf("AIP deletion report: open storage bucket: %w", err)
	}
	defer b.Close()

	w, err := b.NewWriter(ctx, key, nil)
	if err != nil {
		return nil, fmt.Errorf("AIP deletion report: open bucket writer: %w", err)
	}
	defer w.Close()

	if err := dr.Write(ctx, data, w); err != nil {
		return nil, fmt.Errorf("AIP deletion report: write report to bucket: %w", err)
	}

	return &AIPDeletionReportActivityResult{Key: key}, nil
}

func (a *AIPDeletionReportActivity) loadData(
	ctx context.Context,
	aipID uuid.UUID,
	locationSource enums.LocationSource,
) (*reports.AIPDeletionData, error) {
	aip, err := a.storageSvc.ReadAip(ctx, aipID)
	if err != nil {
		return nil, fmt.Errorf("AIP deletion report: load data: ReadAip: %v", err)
	}

	drs, err := a.storageSvc.ListDeletionRequests(ctx, &persistence.DeletionRequestFilter{
		AIPUUID: &aipID,
		Status:  ref.New(enums.DeletionRequestStatusApproved),
	})
	if err != nil {
		return nil, fmt.Errorf("AIP deletion report: load data: ListDeletionRequests: %v", err)
	}
	if len(drs) == 0 {
		return nil, fmt.Errorf("AIP deletion report: no approved deletion request found for AIP %s", aip.UUID)
	}

	wf, err := a.storageSvc.ReadWorkflow(ctx, drs[0].WorkflowDBID)
	if err != nil {
		return nil, fmt.Errorf("AIP deletion report: load data: ReadWorkflow: %v", err)
	}

	d := reports.AIPDeletionData{
		AIPName:            aip.Name,
		AIPUUID:            aip.UUID,
		DeletedAt:          wf.CompletedAt,
		EnduroVersion:      version.Long,
		PreservationSystem: "a3m",
		Reason:             drs[0].Reason,
		RequestedAt:        drs[0].RequestedAt,
		Requester:          drs[0].Requester,
		ReviewedAt:         drs[0].ReviewedAt,
		Reviewer:           drs[0].Reviewer,
		Status:             drs[0].Status.String(),
		StorageLocation:    aip.LocationUUID.String(),
		StorageSystem:      "Enduro Storage Service",
	}

	if locationSource == enums.LocationSourceAmss {
		d.PreservationSystem = "Archivematica"
		d.StorageSystem = "Archivematica Storage Service"
	}

	return &d, nil
}
