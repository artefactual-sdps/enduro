// Code generated by goa v3.15.2, DO NOT EDIT.
//
// HTTP request path constructors for the ingest service.
//
// Command:
// $ goa gen github.com/artefactual-sdps/enduro/internal/api/design -o
// internal/api

package client

import (
	"fmt"
)

// MonitorRequestIngestPath returns the URL path to the ingest service monitor_request HTTP endpoint.
func MonitorRequestIngestPath() string {
	return "/ingest/monitor"
}

// MonitorIngestPath returns the URL path to the ingest service monitor HTTP endpoint.
func MonitorIngestPath() string {
	return "/ingest/monitor"
}

// ListSipsIngestPath returns the URL path to the ingest service list_sips HTTP endpoint.
func ListSipsIngestPath() string {
	return "/ingest/sips"
}

// ShowSipIngestPath returns the URL path to the ingest service show_sip HTTP endpoint.
func ShowSipIngestPath(id uint) string {
	return fmt.Sprintf("/ingest/sips/%v", id)
}

// ListSipWorkflowsIngestPath returns the URL path to the ingest service list_sip_workflows HTTP endpoint.
func ListSipWorkflowsIngestPath(id uint) string {
	return fmt.Sprintf("/ingest/sips/%v/workflows", id)
}

// ConfirmSipIngestPath returns the URL path to the ingest service confirm_sip HTTP endpoint.
func ConfirmSipIngestPath(id uint) string {
	return fmt.Sprintf("/ingest/sips/%v/confirm", id)
}

// RejectSipIngestPath returns the URL path to the ingest service reject_sip HTTP endpoint.
func RejectSipIngestPath(id uint) string {
	return fmt.Sprintf("/ingest/sips/%v/reject", id)
}

// UploadSipIngestPath returns the URL path to the ingest service upload_sip HTTP endpoint.
func UploadSipIngestPath() string {
	return "/ingest/sips/upload"
}
