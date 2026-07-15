package auth

import (
	"encoding/json"
	"fmt"
)

// TicketPurpose identifies the operation authorized by a ticket.
type TicketPurpose string

const (
	TicketPurposeIngestSIPDownload        TicketPurpose = "ingest.sip.download"
	TicketPurposeStorageAIPDownload       TicketPurpose = "storage.aip.download"
	TicketPurposeStorageAIPDeletionReport TicketPurpose = "storage.aip.deletion_report"
)

// TicketGrant binds a ticket to one operation and resource.
type TicketGrant struct {
	Purpose    TicketPurpose `json:"purpose"`
	ResourceID string        `json:"resource_id"`
}

// NewTicketGrant creates a grant for an operation and resource.
func NewTicketGrant(purpose TicketPurpose, resourceID string) TicketGrant {
	return TicketGrant{
		Purpose:    purpose,
		ResourceID: resourceID,
	}
}

// Validate checks that the grant authorizes the expected operation and resource.
func (g TicketGrant) Validate(purpose TicketPurpose, resourceID string) error {
	if g.Purpose != purpose {
		return fmt.Errorf("ticket purpose mismatch: got %q, want %q", g.Purpose, purpose)
	}
	if g.ResourceID != resourceID {
		return fmt.Errorf("ticket resource mismatch: got %q, want %q", g.ResourceID, resourceID)
	}
	return nil
}

// MarshalBinary implements encoding.BinaryMarshaler for Redis compatibility.
func (g TicketGrant) MarshalBinary() ([]byte, error) {
	return json.Marshal(g)
}

// UnmarshalBinary implements encoding.BinaryUnmarshaler for Redis compatibility.
func (g *TicketGrant) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, g)
}
