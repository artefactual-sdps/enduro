// Code generated by goa v3.8.3, DO NOT EDIT.
//
// HTTP request path constructors for the package service.
//
// Command:
// $ goa-v3.8.3 gen github.com/artefactual-sdps/enduro/internal/api/design -o
// internal/api

package server

import (
	"fmt"
)

// MonitorPackagePath returns the URL path to the package service monitor HTTP endpoint.
func MonitorPackagePath() string {
	return "/package/monitor"
}

// ListPackagePath returns the URL path to the package service list HTTP endpoint.
func ListPackagePath() string {
	return "/package"
}

// ShowPackagePath returns the URL path to the package service show HTTP endpoint.
func ShowPackagePath(id uint) string {
	return fmt.Sprintf("/package/%v", id)
}

// PreservationActionsPackagePath returns the URL path to the package service preservation-actions HTTP endpoint.
func PreservationActionsPackagePath(id uint) string {
	return fmt.Sprintf("/package/%v/preservation-actions", id)
}

// ConfirmPackagePath returns the URL path to the package service confirm HTTP endpoint.
func ConfirmPackagePath(id uint) string {
	return fmt.Sprintf("/package/%v/confirm", id)
}

// RejectPackagePath returns the URL path to the package service reject HTTP endpoint.
func RejectPackagePath(id uint) string {
	return fmt.Sprintf("/package/%v/reject", id)
}

// MovePackagePath returns the URL path to the package service move HTTP endpoint.
func MovePackagePath(id uint) string {
	return fmt.Sprintf("/package/%v/move", id)
}

// MoveStatusPackagePath returns the URL path to the package service move_status HTTP endpoint.
func MoveStatusPackagePath(id uint) string {
	return fmt.Sprintf("/package/%v/move", id)
}
