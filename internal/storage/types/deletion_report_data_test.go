package types_test

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"gotest.tools/v3/assert"

	"github.com/artefactual-sdps/enduro/internal/storage/types"
)

func TestDeletionReportData_MarshalJSON(t *testing.T) {
	data := &types.DeletionReportData{
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
