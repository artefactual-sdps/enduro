package activities

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/google/uuid"
	"github.com/jonboulle/clockwork"
	"go.artefactual.dev/tools/ref"

	"github.com/artefactual-sdps/enduro/internal/storage"
	"github.com/artefactual-sdps/enduro/internal/storage/enums"
	"github.com/artefactual-sdps/enduro/internal/storage/pdfs"
	"github.com/artefactual-sdps/enduro/internal/storage/persistence"
	"github.com/artefactual-sdps/enduro/internal/storage/types"
	"github.com/artefactual-sdps/enduro/internal/version"
)

const AIPDeletionReportActivityName = "aip-deletion-report-activity"

type AIPDeletionReportActivity struct {
	cfg        storage.AIPDeletionConfig
	clock      clockwork.Clock
	storageSvc storage.Service
	formFiller pdfs.FormFiller
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
	ff pdfs.FormFiller,
) *AIPDeletionReportActivity {
	return &AIPDeletionReportActivity{
		cfg:        cfg,
		clock:      clock,
		storageSvc: svc,
		formFiller: ff,
	}
}

func (a *AIPDeletionReportActivity) Execute(
	ctx context.Context,
	params *AIPDeletionReportActivityParams,
) (*AIPDeletionReportActivityResult, error) {
	if a.cfg.ReportTemplatePath == "" {
		return nil, fmt.Errorf("AIP deletion report: template path is not configured")
	}
	if _, err := os.Stat(a.cfg.ReportTemplatePath); err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("AIP deletion report: template file not found: %s", a.cfg.ReportTemplatePath)
		}
		return nil, fmt.Errorf("AIP deletion report: template file: %v", err)
	}

	key := fmt.Sprintf("%saip_deletion_report_%s.pdf", storage.ReportPrefix, params.AIPID.String())

	data, err := a.loadData(ctx, params.AIPID, params.LocationSource)
	if err != nil {
		return nil, err
	}

	if err := a.write(ctx, data, key); err != nil {
		return nil, fmt.Errorf("AIP deletion report: %v", err)
	}

	// Persist report key.
	_, err = a.storageSvc.UpdateDeletionRequest(
		ctx,
		data.DeletionRequestDBID,
		func(d *types.DeletionRequest) (*types.DeletionRequest, error) {
			d.ReportKey = key
			return d, nil
		},
	)
	if err != nil {
		return nil, fmt.Errorf("AIP deletion report: update deletion request: %v", err)
	}

	return &AIPDeletionReportActivityResult{Key: key}, nil
}

func (a *AIPDeletionReportActivity) loadData(
	ctx context.Context,
	aipID uuid.UUID,
	locationSource enums.LocationSource,
) (*types.DeletionReportData, error) {
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

	d := types.DeletionReportData{
		DeletionRequestDBID: drs[0].DBID,
		AIPName:             aip.Name,
		AIPUUID:             aip.UUID,
		DeletedAt:           wf.CompletedAt,
		EnduroVersion:       version.Long,
		PreservationSystem:  "a3m",
		Reason:              drs[0].Reason,
		RequestedAt:         drs[0].RequestedAt,
		Requester:           drs[0].Requester,
		ReviewedAt:          drs[0].ReviewedAt,
		Reviewer:            drs[0].Reviewer,
		Status:              drs[0].Status.String(),
		StorageLocation:     aip.LocationUUID.String(),
		StorageSystem:       "Enduro Storage Service",
	}

	if locationSource == enums.LocationSourceAmss {
		d.PreservationSystem = "Archivematica"
		d.StorageSystem = "Archivematica Storage Service"
	}

	return &d, nil
}

func (a *AIPDeletionReportActivity) bucketWriter(ctx context.Context, key string) (io.WriteCloser, error) {
	loc, err := a.storageSvc.Location(ctx, uuid.Nil)
	if err != nil {
		return nil, fmt.Errorf("get storage location: %v", err)
	}

	b, err := loc.OpenBucket(ctx)
	if err != nil {
		return nil, fmt.Errorf("open storage bucket: %v", err)
	}
	defer b.Close()

	w, err := b.NewWriter(ctx, key, nil)
	if err != nil {
		return nil, fmt.Errorf("open bucket writer: %v", err)
	}

	return w, nil
}

func (a *AIPDeletionReportActivity) write(ctx context.Context, data *types.DeletionReportData, key string) error {
	if data == nil {
		return errors.New("data is nil")
	}

	src, err := os.Open(a.cfg.ReportTemplatePath)
	if err != nil {
		return fmt.Errorf("open template: %v", err)
	}
	defer src.Close()

	w, err := a.bucketWriter(ctx, key)
	if err != nil {
		return err
	}
	defer w.Close()

	// Set the report generation timestamp to now.
	data.ReportTimestamp = a.clock.Now().UTC()

	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("marshal JSON: %v", err)
	}

	if err := a.formFiller.FillForm(src, bytes.NewReader(jsonData), w); err != nil {
		return fmt.Errorf("fill form: %v", err)
	}

	return nil
}
