// Code generated by goa v3.15.2, DO NOT EDIT.
//
// ingest HTTP client CLI support package
//
// Command:
// $ goa gen github.com/artefactual-sdps/enduro/internal/api/design -o
// internal/api

package client

import (
	"encoding/json"
	"fmt"
	"strconv"

	ingest "github.com/artefactual-sdps/enduro/internal/api/gen/ingest"
	goa "goa.design/goa/v3/pkg"
)

// BuildMonitorRequestPayload builds the payload for the ingest monitor_request
// endpoint from CLI flags.
func BuildMonitorRequestPayload(ingestMonitorRequestToken string) (*ingest.MonitorRequestPayload, error) {
	var token *string
	{
		if ingestMonitorRequestToken != "" {
			token = &ingestMonitorRequestToken
		}
	}
	v := &ingest.MonitorRequestPayload{}
	v.Token = token

	return v, nil
}

// BuildMonitorPayload builds the payload for the ingest monitor endpoint from
// CLI flags.
func BuildMonitorPayload(ingestMonitorTicket string) (*ingest.MonitorPayload, error) {
	var ticket *string
	{
		if ingestMonitorTicket != "" {
			ticket = &ingestMonitorTicket
		}
	}
	v := &ingest.MonitorPayload{}
	v.Ticket = ticket

	return v, nil
}

// BuildListSipsPayload builds the payload for the ingest list_sips endpoint
// from CLI flags.
func BuildListSipsPayload(ingestListSipsName string, ingestListSipsAipID string, ingestListSipsEarliestCreatedTime string, ingestListSipsLatestCreatedTime string, ingestListSipsStatus string, ingestListSipsLimit string, ingestListSipsOffset string, ingestListSipsToken string) (*ingest.ListSipsPayload, error) {
	var err error
	var name *string
	{
		if ingestListSipsName != "" {
			name = &ingestListSipsName
		}
	}
	var aipID *string
	{
		if ingestListSipsAipID != "" {
			aipID = &ingestListSipsAipID
			err = goa.MergeErrors(err, goa.ValidateFormat("aip_id", *aipID, goa.FormatUUID))
			if err != nil {
				return nil, err
			}
		}
	}
	var earliestCreatedTime *string
	{
		if ingestListSipsEarliestCreatedTime != "" {
			earliestCreatedTime = &ingestListSipsEarliestCreatedTime
			err = goa.MergeErrors(err, goa.ValidateFormat("earliest_created_time", *earliestCreatedTime, goa.FormatDateTime))
			if err != nil {
				return nil, err
			}
		}
	}
	var latestCreatedTime *string
	{
		if ingestListSipsLatestCreatedTime != "" {
			latestCreatedTime = &ingestListSipsLatestCreatedTime
			err = goa.MergeErrors(err, goa.ValidateFormat("latest_created_time", *latestCreatedTime, goa.FormatDateTime))
			if err != nil {
				return nil, err
			}
		}
	}
	var status *string
	{
		if ingestListSipsStatus != "" {
			status = &ingestListSipsStatus
			if !(*status == "new" || *status == "in progress" || *status == "done" || *status == "error" || *status == "unknown" || *status == "queued" || *status == "abandoned" || *status == "pending") {
				err = goa.MergeErrors(err, goa.InvalidEnumValueError("status", *status, []any{"new", "in progress", "done", "error", "unknown", "queued", "abandoned", "pending"}))
			}
			if err != nil {
				return nil, err
			}
		}
	}
	var limit *int
	{
		if ingestListSipsLimit != "" {
			var v int64
			v, err = strconv.ParseInt(ingestListSipsLimit, 10, strconv.IntSize)
			val := int(v)
			limit = &val
			if err != nil {
				return nil, fmt.Errorf("invalid value for limit, must be INT")
			}
		}
	}
	var offset *int
	{
		if ingestListSipsOffset != "" {
			var v int64
			v, err = strconv.ParseInt(ingestListSipsOffset, 10, strconv.IntSize)
			val := int(v)
			offset = &val
			if err != nil {
				return nil, fmt.Errorf("invalid value for offset, must be INT")
			}
		}
	}
	var token *string
	{
		if ingestListSipsToken != "" {
			token = &ingestListSipsToken
		}
	}
	v := &ingest.ListSipsPayload{}
	v.Name = name
	v.AipID = aipID
	v.EarliestCreatedTime = earliestCreatedTime
	v.LatestCreatedTime = latestCreatedTime
	v.Status = status
	v.Limit = limit
	v.Offset = offset
	v.Token = token

	return v, nil
}

// BuildShowSipPayload builds the payload for the ingest show_sip endpoint from
// CLI flags.
func BuildShowSipPayload(ingestShowSipID string, ingestShowSipToken string) (*ingest.ShowSipPayload, error) {
	var err error
	var id uint
	{
		var v uint64
		v, err = strconv.ParseUint(ingestShowSipID, 10, strconv.IntSize)
		id = uint(v)
		if err != nil {
			return nil, fmt.Errorf("invalid value for id, must be UINT")
		}
	}
	var token *string
	{
		if ingestShowSipToken != "" {
			token = &ingestShowSipToken
		}
	}
	v := &ingest.ShowSipPayload{}
	v.ID = id
	v.Token = token

	return v, nil
}

// BuildListSipWorkflowsPayload builds the payload for the ingest
// list_sip_workflows endpoint from CLI flags.
func BuildListSipWorkflowsPayload(ingestListSipWorkflowsID string, ingestListSipWorkflowsToken string) (*ingest.ListSipWorkflowsPayload, error) {
	var err error
	var id uint
	{
		var v uint64
		v, err = strconv.ParseUint(ingestListSipWorkflowsID, 10, strconv.IntSize)
		id = uint(v)
		if err != nil {
			return nil, fmt.Errorf("invalid value for id, must be UINT")
		}
	}
	var token *string
	{
		if ingestListSipWorkflowsToken != "" {
			token = &ingestListSipWorkflowsToken
		}
	}
	v := &ingest.ListSipWorkflowsPayload{}
	v.ID = id
	v.Token = token

	return v, nil
}

// BuildConfirmSipPayload builds the payload for the ingest confirm_sip
// endpoint from CLI flags.
func BuildConfirmSipPayload(ingestConfirmSipBody string, ingestConfirmSipID string, ingestConfirmSipToken string) (*ingest.ConfirmSipPayload, error) {
	var err error
	var body ConfirmSipRequestBody
	{
		err = json.Unmarshal([]byte(ingestConfirmSipBody), &body)
		if err != nil {
			return nil, fmt.Errorf("invalid JSON for body, \nerror: %s, \nexample of valid JSON:\n%s", err, "'{\n      \"location_id\": \"d1845cb6-a5ea-474a-9ab8-26f9bcd919f5\"\n   }'")
		}
	}
	var id uint
	{
		var v uint64
		v, err = strconv.ParseUint(ingestConfirmSipID, 10, strconv.IntSize)
		id = uint(v)
		if err != nil {
			return nil, fmt.Errorf("invalid value for id, must be UINT")
		}
	}
	var token *string
	{
		if ingestConfirmSipToken != "" {
			token = &ingestConfirmSipToken
		}
	}
	v := &ingest.ConfirmSipPayload{
		LocationID: body.LocationID,
	}
	v.ID = id
	v.Token = token

	return v, nil
}

// BuildRejectSipPayload builds the payload for the ingest reject_sip endpoint
// from CLI flags.
func BuildRejectSipPayload(ingestRejectSipID string, ingestRejectSipToken string) (*ingest.RejectSipPayload, error) {
	var err error
	var id uint
	{
		var v uint64
		v, err = strconv.ParseUint(ingestRejectSipID, 10, strconv.IntSize)
		id = uint(v)
		if err != nil {
			return nil, fmt.Errorf("invalid value for id, must be UINT")
		}
	}
	var token *string
	{
		if ingestRejectSipToken != "" {
			token = &ingestRejectSipToken
		}
	}
	v := &ingest.RejectSipPayload{}
	v.ID = id
	v.Token = token

	return v, nil
}

// BuildMoveSipPayload builds the payload for the ingest move_sip endpoint from
// CLI flags.
func BuildMoveSipPayload(ingestMoveSipBody string, ingestMoveSipID string, ingestMoveSipToken string) (*ingest.MoveSipPayload, error) {
	var err error
	var body MoveSipRequestBody
	{
		err = json.Unmarshal([]byte(ingestMoveSipBody), &body)
		if err != nil {
			return nil, fmt.Errorf("invalid JSON for body, \nerror: %s, \nexample of valid JSON:\n%s", err, "'{\n      \"location_id\": \"d1845cb6-a5ea-474a-9ab8-26f9bcd919f5\"\n   }'")
		}
	}
	var id uint
	{
		var v uint64
		v, err = strconv.ParseUint(ingestMoveSipID, 10, strconv.IntSize)
		id = uint(v)
		if err != nil {
			return nil, fmt.Errorf("invalid value for id, must be UINT")
		}
	}
	var token *string
	{
		if ingestMoveSipToken != "" {
			token = &ingestMoveSipToken
		}
	}
	v := &ingest.MoveSipPayload{
		LocationID: body.LocationID,
	}
	v.ID = id
	v.Token = token

	return v, nil
}

// BuildMoveSipStatusPayload builds the payload for the ingest move_sip_status
// endpoint from CLI flags.
func BuildMoveSipStatusPayload(ingestMoveSipStatusID string, ingestMoveSipStatusToken string) (*ingest.MoveSipStatusPayload, error) {
	var err error
	var id uint
	{
		var v uint64
		v, err = strconv.ParseUint(ingestMoveSipStatusID, 10, strconv.IntSize)
		id = uint(v)
		if err != nil {
			return nil, fmt.Errorf("invalid value for id, must be UINT")
		}
	}
	var token *string
	{
		if ingestMoveSipStatusToken != "" {
			token = &ingestMoveSipStatusToken
		}
	}
	v := &ingest.MoveSipStatusPayload{}
	v.ID = id
	v.Token = token

	return v, nil
}

// BuildUploadSipPayload builds the payload for the ingest upload_sip endpoint
// from CLI flags.
func BuildUploadSipPayload(ingestUploadSipContentType string, ingestUploadSipToken string) (*ingest.UploadSipPayload, error) {
	var err error
	var contentType string
	{
		if ingestUploadSipContentType != "" {
			contentType = ingestUploadSipContentType
			err = goa.MergeErrors(err, goa.ValidatePattern("content_type", contentType, "multipart/[^;]+; boundary=.+"))
			if err != nil {
				return nil, err
			}
		}
	}
	var token *string
	{
		if ingestUploadSipToken != "" {
			token = &ingestUploadSipToken
		}
	}
	v := &ingest.UploadSipPayload{}
	v.ContentType = contentType
	v.Token = token

	return v, nil
}
