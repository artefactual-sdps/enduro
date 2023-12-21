// Code generated by goa v3.14.1, DO NOT EDIT.
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

// SubmitStoragePath returns the URL path to the storage service submit HTTP endpoint.
func SubmitStoragePath(aipID string) string {
	return fmt.Sprintf("/storage/package/%v/submit", aipID)
}

// UpdateStoragePath returns the URL path to the storage service update HTTP endpoint.
func UpdateStoragePath(aipID string) string {
	return fmt.Sprintf("/storage/package/%v/update", aipID)
}

// DownloadStoragePath returns the URL path to the storage service download HTTP endpoint.
func DownloadStoragePath(aipID string) string {
	return fmt.Sprintf("/storage/package/%v/download", aipID)
}

// LocationsStoragePath returns the URL path to the storage service locations HTTP endpoint.
func LocationsStoragePath() string {
	return "/storage/location"
}

// AddLocationStoragePath returns the URL path to the storage service add_location HTTP endpoint.
func AddLocationStoragePath() string {
	return "/storage/location"
}

// MoveStoragePath returns the URL path to the storage service move HTTP endpoint.
func MoveStoragePath(aipID string) string {
	return fmt.Sprintf("/storage/package/%v/store", aipID)
}

// MoveStatusStoragePath returns the URL path to the storage service move_status HTTP endpoint.
func MoveStatusStoragePath(aipID string) string {
	return fmt.Sprintf("/storage/package/%v/store", aipID)
}

// RejectStoragePath returns the URL path to the storage service reject HTTP endpoint.
func RejectStoragePath(aipID string) string {
	return fmt.Sprintf("/storage/package/%v/reject", aipID)
}

// ShowStoragePath returns the URL path to the storage service show HTTP endpoint.
func ShowStoragePath(aipID string) string {
	return fmt.Sprintf("/storage/package/%v", aipID)
}

// ShowLocationStoragePath returns the URL path to the storage service show_location HTTP endpoint.
func ShowLocationStoragePath(uuid string) string {
	return fmt.Sprintf("/storage/location/%v", uuid)
}

// LocationPackagesStoragePath returns the URL path to the storage service location_packages HTTP endpoint.
func LocationPackagesStoragePath(uuid string) string {
	return fmt.Sprintf("/storage/location/%v/packages", uuid)
}
