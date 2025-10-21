package report

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/pdfcpu/pdfcpu/pkg/api"

	pres_cfg "github.com/artefactual-sdps/enduro/internal/pres"
	"github.com/artefactual-sdps/enduro/internal/storage"
	"github.com/artefactual-sdps/enduro/internal/temporal"
	"github.com/artefactual-sdps/enduro/internal/version"
)

type AIPDeletion struct {
	storageSvc   storage.Service
	templatePath string
	presCfg      pres_cfg.Config
}

func NewAIPDeletion(presCfg pres_cfg.Config, storageSvc storage.Service, templatePath string) *AIPDeletion {
	return &AIPDeletion{
		storageSvc:   storageSvc,
		templatePath: templatePath,
		presCfg:      presCfg,
	}
}

type PDFData struct {
	AIPName            string
	AIPUUID            uuid.UUID
	Requester          string
	RequestedAt        time.Time
	Reason             string
	Reviewer           string
	ReviewedAt         time.Time
	Status             string
	WorkflowID         uuid.UUID
	DeletedAt          time.Time
	EnduroVersion      string
	PreservationSystem string
	StorageSystem      string
}

type PDFHeader struct {
	Creation string `json:"creation"`
	Creator  string `json:"creator"`
}

func (h *PDFHeader) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Creation string `json:"creation"`
		Creator  string `json:"creator"`
	}{
		Creation: time.Now().UTC().Format(time.RFC3339),
		Creator:  "Enduro AIP Deletion Report Generator",
	})
}

type PDFDocument struct {
	Header PDFHeader `json:"header"`
	Forms  []PDFForm `json:"forms"`

	Data *PDFData
}

type PDFForm struct {
	Textfield []PDFFormField `json:"textfield"`
}

func (d *PDFDocument) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Header PDFHeader `json:"header"`
		Forms  []PDFForm `json:"forms"`
	}{
		Header: d.Header,
		Forms: []PDFForm{
			{
				Textfield: []PDFFormField{
					{Name: "report_timestamp", Value: time.Now().Format(time.RFC3339)},
					{Name: "aip_name", Value: d.Data.AIPName},
					{Name: "aip_uuid", Value: d.Data.AIPUUID.String()},
					{Name: "requester", Value: d.Data.Requester},
					{Name: "requested_at", Value: d.Data.RequestedAt.Format(time.RFC3339)},
					{Name: "reason", Value: d.Data.Reason},
					{Name: "reviewer", Value: d.Data.Reviewer},
					{Name: "reviewed_at", Value: d.Data.ReviewedAt.Format(time.RFC3339)},
					{Name: "status", Value: d.Data.Status},
					{Name: "workflow_id", Value: d.Data.WorkflowID.String()},
					{Name: "deleted_at", Value: d.Data.DeletedAt.Format(time.RFC3339)},
					{Name: "enduro_version", Value: d.Data.EnduroVersion},
					{Name: "preservation_system", Value: d.Data.PreservationSystem},
					{Name: "storage_system", Value: d.Data.StorageSystem},
				},
			},
		},
	})
}

type PDFFormField struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

func (f *PDFFormField) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Name   string `json:"name"`
		Value  string `json:"value"`
		Locked bool   `json:"locked"`
	}{
		Name:   f.Name,
		Value:  f.Value,
		Locked: true,
	})
}

func (dr *AIPDeletion) Write(ctx context.Context, data *PDFData, reportPath string) error {
	r, err := os.OpenFile(dr.templatePath, os.O_RDWR, 0644)
	if err != nil {
		fmt.Println("Error opening template file:", err)
		return err
	}
	defer r.Close()

	w, err := os.Create(reportPath)
	if err != nil {
		fmt.Println("Error creating report file:", err)
		return err
	}
	defer w.Close()

	jsonData, err := json.Marshal(PDFDocument{Data: data})
	if err != nil {
		return err
	}

	api.FillForm(r, bytes.NewReader(jsonData), w, nil)

	return nil
}

func (r *AIPDeletion) LoadFormData(ctx context.Context, drID uuid.UUID) (*PDFData, error) {
	dr, err := r.storageSvc.ReadDeletionRequest(ctx, drID)
	if err != nil {
		return nil, err
	}

	aip, err := r.storageSvc.ReadAip(ctx, dr.AIPUUID)
	if err != nil {
		return nil, err
	}

	d := PDFData{
		AIPName:            aip.Name,
		AIPUUID:            aip.UUID,
		Requester:          dr.Requester,
		RequestedAt:        dr.RequestedAt,
		Reason:             dr.Reason,
		Reviewer:           dr.Reviewer,
		ReviewedAt:         dr.ReviewedAt,
		Status:             dr.Status.String(),
		EnduroVersion:      version.Long,
		PreservationSystem: "a3m",
		StorageSystem:      "Enduro Storage Service",
	}

	if r.presCfg.TaskQueue == temporal.AmWorkerTaskQueue {
		d.PreservationSystem = "Archivematica"
		d.StorageSystem = "Archivematica Storage Service"
	}

	return &d, nil
}
