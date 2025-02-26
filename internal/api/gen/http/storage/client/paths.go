// Code generated by goa v3.15.2, DO NOT EDIT.
//
// HTTP request path constructors for the storage service.
//
// Command:
// $ goa gen github.com/artefactual-sdps/enduro/internal/api/design -o
// internal/api

package client

import (
	"fmt"
)

// CreateAipStoragePath returns the URL path to the storage service create_aip HTTP endpoint.
func CreateAipStoragePath() string {
	return "/storage/aips"
}

// SubmitAipStoragePath returns the URL path to the storage service submit_aip HTTP endpoint.
func SubmitAipStoragePath(uuid string) string {
	return fmt.Sprintf("/storage/aips/%v/submit", uuid)
}

// UpdateAipStoragePath returns the URL path to the storage service update_aip HTTP endpoint.
func UpdateAipStoragePath(uuid string) string {
	return fmt.Sprintf("/storage/aips/%v/update", uuid)
}

// DownloadAipStoragePath returns the URL path to the storage service download_aip HTTP endpoint.
func DownloadAipStoragePath(uuid string) string {
	return fmt.Sprintf("/storage/aips/%v/download", uuid)
}

// MoveAipStoragePath returns the URL path to the storage service move_aip HTTP endpoint.
func MoveAipStoragePath(uuid string) string {
	return fmt.Sprintf("/storage/aips/%v/store", uuid)
}

// MoveAipStatusStoragePath returns the URL path to the storage service move_aip_status HTTP endpoint.
func MoveAipStatusStoragePath(uuid string) string {
	return fmt.Sprintf("/storage/aips/%v/store", uuid)
}

// RejectAipStoragePath returns the URL path to the storage service reject_aip HTTP endpoint.
func RejectAipStoragePath(uuid string) string {
	return fmt.Sprintf("/storage/aips/%v/reject", uuid)
}

// ShowAipStoragePath returns the URL path to the storage service show_aip HTTP endpoint.
func ShowAipStoragePath(uuid string) string {
	return fmt.Sprintf("/storage/aips/%v", uuid)
}

// ListLocationsStoragePath returns the URL path to the storage service list_locations HTTP endpoint.
func ListLocationsStoragePath() string {
	return "/storage/locations"
}

// CreateLocationStoragePath returns the URL path to the storage service create_location HTTP endpoint.
func CreateLocationStoragePath() string {
	return "/storage/locations"
}

// ShowLocationStoragePath returns the URL path to the storage service show_location HTTP endpoint.
func ShowLocationStoragePath(uuid string) string {
	return fmt.Sprintf("/storage/locations/%v", uuid)
}

// ListLocationAipsStoragePath returns the URL path to the storage service list_location_aips HTTP endpoint.
func ListLocationAipsStoragePath(uuid string) string {
	return fmt.Sprintf("/storage/locations/%v/aips", uuid)
}
