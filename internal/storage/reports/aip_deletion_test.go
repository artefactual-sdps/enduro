package reports_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jonboulle/clockwork"
	"gotest.tools/v3/assert"

	"github.com/artefactual-sdps/enduro/internal/storage"
	"github.com/artefactual-sdps/enduro/internal/storage/reports"
)

const templatePath = "../../../assets/Enduro_AIP_deletion_report_v3.tmpl.pdf"

func TestNewAIPDeletion(t *testing.T) {
	type args struct {
		clock clockwork.Clock
		cfg   storage.AIPDeletionConfig
	}

	type test struct {
		name    string
		args    args
		wantErr string
	}
	for _, tc := range []test{
		{
			name: "successful creation",
			args: args{
				clock: clockwork.NewFakeClockAt(time.Date(2025, 10, 28, 8, 20, 40, 0, time.UTC)),
				cfg:   storage.AIPDeletionConfig{ReportTemplatePath: templatePath},
			},
		},
		{
			name: "errors if template file does not exist",
			args: args{
				clock: clockwork.NewFakeClockAt(time.Date(2025, 10, 28, 8, 20, 40, 0, time.UTC)),
				cfg:   storage.AIPDeletionConfig{ReportTemplatePath: "non_existent_template.tmpl.pdf"},
			},
			wantErr: "AIP deletion report: template file does not exist: non_existent_template.tmpl.pdf",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, err := reports.NewAIPDeletion(tc.args.clock, tc.args.cfg.ReportTemplatePath)
			if tc.wantErr != "" {
				assert.ErrorContains(t, err, tc.wantErr)
				return
			}

			assert.NilError(t, err)
		})
	}
}

func TestAIPDeletion_Write(t *testing.T) {
	var buf bytes.Buffer

	dr, err := reports.NewAIPDeletion(
		clockwork.NewFakeClockAt(time.Date(2025, 10, 28, 8, 20, 40, 0, time.UTC)),
		templatePath,
	)
	if err != nil {
		t.Fatalf("couldn't create AIPDeletion: %v", err)
	}

	err = dr.Write(
		context.Background(),
		&reports.AIPDeletionData{
			AIPName:            "Test AIP",
			AIPUUID:            uuid.New(),
			DeletedAt:          time.Now(),
			EnduroVersion:      "v0.19.0",
			PreservationSystem: "Archivematica",
			Reason:             "Test reason for deletion",
			RequestedAt:        time.Now().Add(-48 * time.Hour),
			Requester:          "requester-123",
			ReviewedAt:         time.Now().Add(-24 * time.Hour),
			Reviewer:           "reviewer-456",
			Status:             "Approved",
			StorageLocation:    uuid.New().String(),
			StorageSystem:      "Archivematica Storage Service",
		},
		&buf,
	)

	assert.NilError(t, err)
	assert.Check(t, buf.Len() > 0, "expected non-empty report output")
}

func TestAIPDeletionData_MarshalJSON(t *testing.T) {
	data := &reports.AIPDeletionData{
		AIPName:            "Test AIP",
		AIPUUID:            uuid.MustParse("123e4567-e89b-12d3-a456-426614174000"),
		DeletedAt:          time.Date(2025, 10, 27, 8, 20, 43, 0, time.UTC),
		EnduroVersion:      "v0.19.0",
		PreservationSystem: "Archivematica",
		Reason:             "Test reason for deletion",
		ReportTimestamp:    time.Date(2025, 10, 28, 8, 20, 40, 0, time.UTC),
		RequestedAt:        time.Date(2025, 10, 26, 8, 20, 40, 0, time.UTC),
		Requester:          "sjones@example.com",
		ReviewedAt:         time.Date(2025, 10, 27, 8, 20, 40, 0, time.UTC),
		Reviewer:           "reviewer-456",
		Status:             "Approved",
		StorageLocation:    "storage-location-123",
		StorageSystem:      "Archivematica Storage Service",
	}

	b, err := json.Marshal(data)
	if err != nil {
		t.Fatalf("couldn't marshal AIPDeletionData: %v", err)
	}

	for _, v := range []string{
		`"creation":"2025-10-28T08:20:40Z"`,
		`"producer":"Enduro"`,
		`{"name":"aip_name","value":"Test AIP","locked":true}`,
		`{"name":"aip_uuid","value":"123e4567-e89b-12d3-a456-426614174000","locked":true}`,
		`{"name":"deleted_at","value":"2025-10-27T08:20:43Z","locked":true}`,
		`{"name":"enduro_version","value":"v0.19.0","locked":true}`,
		`{"name":"preservation_system","value":"Archivematica","locked":true}`,
		`{"name":"reason","value":"Test reason for deletion","locked":true}`,
		`{"name":"report_timestamp","value":"2025-10-28T08:20:40Z","locked":true}`,
		`{"name":"requested_at","value":"2025-10-26T08:20:40Z","locked":true}`,
		`{"name":"requester","value":"sjones@example.com","locked":true}`,
		`{"name":"reviewed_at","value":"2025-10-27T08:20:40Z","locked":true}`,
		`{"name":"reviewer","value":"reviewer-456","locked":true}`,
		`{"name":"status","value":"Approved","locked":true}`,
		`{"name":"storage_location","value":"storage-location-123","locked":true}`,
		`{"name":"storage_system","value":"Archivematica Storage Service","locked":true}`,
	} {
		assert.Check(t, strings.Contains(string(b), v), fmt.Sprintf("missing: %s, got: %s", v, string(b)))
	}
}
