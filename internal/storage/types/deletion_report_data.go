package types

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type DeletionReportData struct {
	AIPName            string
	AIPUUID            uuid.UUID
	DeletedAt          time.Time
	EnduroVersion      string
	PreservationSystem string
	Reason             string
	ReportTimestamp    time.Time
	RequestedAt        time.Time
	Requester          string
	ReviewedAt         time.Time
	Reviewer           string
	Status             string
	StorageLocation    string
	StorageSystem      string
}

func (d *DeletionReportData) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Header aipDeletionHeader `json:"header"`
		Forms  []aipDeletionForm `json:"forms"`
	}{
		Header: aipDeletionHeader{
			Creation: d.ReportTimestamp.Format(time.RFC3339),
			Producer: "Enduro",
		},
		Forms: []aipDeletionForm{
			{
				Textfield: []aipDeletionField{
					{Name: "aip_name", Value: d.AIPName},
					{Name: "aip_uuid", Value: d.AIPUUID.String()},
					{Name: "deleted_at", Value: d.DeletedAt.Format(time.RFC3339)},
					{Name: "enduro_version", Value: d.EnduroVersion},
					{Name: "preservation_system", Value: d.PreservationSystem},
					{Name: "reason", Value: d.Reason},
					{Name: "report_timestamp", Value: d.ReportTimestamp.Format(time.RFC3339)},
					{Name: "requester", Value: d.Requester},
					{Name: "requested_at", Value: d.RequestedAt.Format(time.RFC3339)},
					{Name: "reviewer", Value: d.Reviewer},
					{Name: "reviewed_at", Value: d.ReviewedAt.Format(time.RFC3339)},
					{Name: "status", Value: d.Status},
					{Name: "storage_location", Value: d.StorageLocation},
					{Name: "storage_system", Value: d.StorageSystem},
				},
			},
		},
	})
}

type aipDeletionHeader struct {
	Creation string `json:"creation"`
	Producer string `json:"producer"`
}

type aipDeletionForm struct {
	Textfield []aipDeletionField `json:"textfield"`
}

type aipDeletionField struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

func (f *aipDeletionField) MarshalJSON() ([]byte, error) {
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
