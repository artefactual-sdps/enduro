package report_test

import (
	"context"
	"encoding/json"
	"path/filepath"
	"testing"
	"time"

	"github.com/artefactual-sdps/enduro/internal/storage/fake"
	"github.com/artefactual-sdps/enduro/internal/storage/report"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

const templatePath = "../../../assets/templates/aip_deletion_report.tmpl.pdf"

func TestFormData_MarshalJSON(t *testing.T) {
	fd := &report.PDFData{
		AIPName:     "Test AIP",
		AIPUUID:     uuid.New(),
		Requester:   "requester-123",
		RequestedAt: time.Now().Add(-48 * time.Hour),
		Reason:      "Test reason for deletion",
		Reviewer:    "reviewer-456",
		ReviewedAt:  time.Now().Add(-24 * time.Hour),
		Status:      "Approved",
		WorkflowID:  uuid.New(),
		DeletedAt:   time.Now(),
	}

	data, err := json.Marshal(fd)
	if err != nil {
		t.Fatalf("couldn't marshal FormData = %v", err)
	}

	aipName := `"aip_name":"Test AIP"`
	assert.Contains(t, string(data), aipName, "Missing %s", aipName)

	requesterID := `"requester_id":"requester-123"`
	assert.Contains(t, string(data), requesterID, "Missing %s", requesterID)
}

func TestAIPDeletionReport_Write(t *testing.T) {
	ctx := context.Background()
	ctrl := gomock.NewController(t)

	dr := report.NewAIPDeletion(
		fake.NewMockService(ctrl),
		templatePath,
	)

	dr.Write(
		ctx,
		&report.PDFData{
			AIPName:     "Test AIP",
			AIPUUID:     uuid.New(),
			Requester:   "requester-123",
			RequestedAt: time.Now().Add(-48 * time.Hour),
			Reason:      "Test reason for deletion",
			Reviewer:    "reviewer-456",
			ReviewedAt:  time.Now().Add(-24 * time.Hour),
			Status:      "Approved",
			WorkflowID:  uuid.New(),
			DeletedAt:   time.Now(),
		},
		filepath.Join(t.TempDir(), "test_aip_deletion_report.pdf"),
	)
}
