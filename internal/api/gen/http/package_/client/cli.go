// Code generated by goa v3.8.4, DO NOT EDIT.
//
// package HTTP client CLI support package
//
// Command:
// $ goa-v3.8.4 gen github.com/artefactual-sdps/enduro/internal/api/design -o
// internal/api

package client

import (
	"encoding/json"
	"fmt"
	"strconv"

	package_ "github.com/artefactual-sdps/enduro/internal/api/gen/package_"
	goa "goa.design/goa/v3/pkg"
)

// BuildListPayload builds the payload for the package list endpoint from CLI
// flags.
func BuildListPayload(package_ListName string, package_ListAipID string, package_ListEarliestCreatedTime string, package_ListLatestCreatedTime string, package_ListLocationID string, package_ListStatus string, package_ListCursor string) (*package_.ListPayload, error) {
	var err error
	var name *string
	{
		if package_ListName != "" {
			name = &package_ListName
		}
	}
	var aipID *string
	{
		if package_ListAipID != "" {
			aipID = &package_ListAipID
			err = goa.MergeErrors(err, goa.ValidateFormat("aipID", *aipID, goa.FormatUUID))
			if err != nil {
				return nil, err
			}
		}
	}
	var earliestCreatedTime *string
	{
		if package_ListEarliestCreatedTime != "" {
			earliestCreatedTime = &package_ListEarliestCreatedTime
			err = goa.MergeErrors(err, goa.ValidateFormat("earliestCreatedTime", *earliestCreatedTime, goa.FormatDateTime))
			if err != nil {
				return nil, err
			}
		}
	}
	var latestCreatedTime *string
	{
		if package_ListLatestCreatedTime != "" {
			latestCreatedTime = &package_ListLatestCreatedTime
			err = goa.MergeErrors(err, goa.ValidateFormat("latestCreatedTime", *latestCreatedTime, goa.FormatDateTime))
			if err != nil {
				return nil, err
			}
		}
	}
	var locationID *string
	{
		if package_ListLocationID != "" {
			locationID = &package_ListLocationID
			err = goa.MergeErrors(err, goa.ValidateFormat("locationID", *locationID, goa.FormatUUID))
			if err != nil {
				return nil, err
			}
		}
	}
	var status *string
	{
		if package_ListStatus != "" {
			status = &package_ListStatus
			if !(*status == "new" || *status == "in progress" || *status == "done" || *status == "error" || *status == "unknown" || *status == "queued" || *status == "pending" || *status == "abandoned") {
				err = goa.MergeErrors(err, goa.InvalidEnumValueError("status", *status, []interface{}{"new", "in progress", "done", "error", "unknown", "queued", "pending", "abandoned"}))
			}
			if err != nil {
				return nil, err
			}
		}
	}
	var cursor *string
	{
		if package_ListCursor != "" {
			cursor = &package_ListCursor
		}
	}
	v := &package_.ListPayload{}
	v.Name = name
	v.AipID = aipID
	v.EarliestCreatedTime = earliestCreatedTime
	v.LatestCreatedTime = latestCreatedTime
	v.LocationID = locationID
	v.Status = status
	v.Cursor = cursor

	return v, nil
}

// BuildShowPayload builds the payload for the package show endpoint from CLI
// flags.
func BuildShowPayload(package_ShowID string) (*package_.ShowPayload, error) {
	var err error
	var id uint
	{
		var v uint64
		v, err = strconv.ParseUint(package_ShowID, 10, strconv.IntSize)
		id = uint(v)
		if err != nil {
			return nil, fmt.Errorf("invalid value for id, must be UINT")
		}
	}
	v := &package_.ShowPayload{}
	v.ID = id

	return v, nil
}

// BuildPreservationActionsPayload builds the payload for the package
// preservation_actions endpoint from CLI flags.
func BuildPreservationActionsPayload(package_PreservationActionsID string) (*package_.PreservationActionsPayload, error) {
	var err error
	var id uint
	{
		var v uint64
		v, err = strconv.ParseUint(package_PreservationActionsID, 10, strconv.IntSize)
		id = uint(v)
		if err != nil {
			return nil, fmt.Errorf("invalid value for id, must be UINT")
		}
	}
	v := &package_.PreservationActionsPayload{}
	v.ID = id

	return v, nil
}

// BuildConfirmPayload builds the payload for the package confirm endpoint from
// CLI flags.
func BuildConfirmPayload(package_ConfirmBody string, package_ConfirmID string) (*package_.ConfirmPayload, error) {
	var err error
	var body ConfirmRequestBody
	{
		err = json.Unmarshal([]byte(package_ConfirmBody), &body)
		if err != nil {
			return nil, fmt.Errorf("invalid JSON for body, \nerror: %s, \nexample of valid JSON:\n%s", err, "'{\n      \"location_id\": \"Ut quas.\"\n   }'")
		}
	}
	var id uint
	{
		var v uint64
		v, err = strconv.ParseUint(package_ConfirmID, 10, strconv.IntSize)
		id = uint(v)
		if err != nil {
			return nil, fmt.Errorf("invalid value for id, must be UINT")
		}
	}
	v := &package_.ConfirmPayload{
		LocationID: body.LocationID,
	}
	v.ID = id

	return v, nil
}

// BuildRejectPayload builds the payload for the package reject endpoint from
// CLI flags.
func BuildRejectPayload(package_RejectID string) (*package_.RejectPayload, error) {
	var err error
	var id uint
	{
		var v uint64
		v, err = strconv.ParseUint(package_RejectID, 10, strconv.IntSize)
		id = uint(v)
		if err != nil {
			return nil, fmt.Errorf("invalid value for id, must be UINT")
		}
	}
	v := &package_.RejectPayload{}
	v.ID = id

	return v, nil
}

// BuildMovePayload builds the payload for the package move endpoint from CLI
// flags.
func BuildMovePayload(package_MoveBody string, package_MoveID string) (*package_.MovePayload, error) {
	var err error
	var body MoveRequestBody
	{
		err = json.Unmarshal([]byte(package_MoveBody), &body)
		if err != nil {
			return nil, fmt.Errorf("invalid JSON for body, \nerror: %s, \nexample of valid JSON:\n%s", err, "'{\n      \"location_id\": \"Temporibus iusto et.\"\n   }'")
		}
	}
	var id uint
	{
		var v uint64
		v, err = strconv.ParseUint(package_MoveID, 10, strconv.IntSize)
		id = uint(v)
		if err != nil {
			return nil, fmt.Errorf("invalid value for id, must be UINT")
		}
	}
	v := &package_.MovePayload{
		LocationID: body.LocationID,
	}
	v.ID = id

	return v, nil
}

// BuildMoveStatusPayload builds the payload for the package move_status
// endpoint from CLI flags.
func BuildMoveStatusPayload(package_MoveStatusID string) (*package_.MoveStatusPayload, error) {
	var err error
	var id uint
	{
		var v uint64
		v, err = strconv.ParseUint(package_MoveStatusID, 10, strconv.IntSize)
		id = uint(v)
		if err != nil {
			return nil, fmt.Errorf("invalid value for id, must be UINT")
		}
	}
	v := &package_.MoveStatusPayload{}
	v.ID = id

	return v, nil
}
