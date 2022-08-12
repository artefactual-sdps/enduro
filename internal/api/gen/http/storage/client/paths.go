// Code generated by goa v3.8.1, DO NOT EDIT.
//
// HTTP request path constructors for the storage service.
//
// Command:
// $ goa-v3.8.1 gen github.com/artefactual-sdps/enduro/internal/api/design -o
// internal/api

package client

import (
	"fmt"
)

// SubmitStoragePath returns the URL path to the storage service submit HTTP endpoint.
func SubmitStoragePath(aipID string) string {
	return fmt.Sprintf("/storage/%v/submit", aipID)
}

// UpdateStoragePath returns the URL path to the storage service update HTTP endpoint.
func UpdateStoragePath(aipID string) string {
	return fmt.Sprintf("/storage/%v/update", aipID)
}

// DownloadStoragePath returns the URL path to the storage service download HTTP endpoint.
func DownloadStoragePath(aipID string) string {
	return fmt.Sprintf("/storage/%v/download", aipID)
}

// LocationsStoragePath returns the URL path to the storage service locations HTTP endpoint.
func LocationsStoragePath() string {
	return "/storage/location"
}

// MoveStoragePath returns the URL path to the storage service move HTTP endpoint.
func MoveStoragePath(aipID string) string {
	return fmt.Sprintf("/storage/%v/store", aipID)
}

// MoveStatusStoragePath returns the URL path to the storage service move_status HTTP endpoint.
func MoveStatusStoragePath(aipID string) string {
	return fmt.Sprintf("/storage/%v/store", aipID)
}

// RejectStoragePath returns the URL path to the storage service reject HTTP endpoint.
func RejectStoragePath(aipID string) string {
	return fmt.Sprintf("/storage/%v/reject", aipID)
}

// ShowStoragePath returns the URL path to the storage service show HTTP endpoint.
func ShowStoragePath(aipID string) string {
	return fmt.Sprintf("/storage/%v", aipID)
}
