// Code generated by goa v3.15.2, DO NOT EDIT.
//
// enduro HTTP client CLI support package
//
// Command:
// $ goa gen github.com/artefactual-sdps/enduro/internal/api/design -o
// internal/api

package cli

import (
	"flag"
	"fmt"
	"net/http"
	"os"

	ingestc "github.com/artefactual-sdps/enduro/internal/api/gen/http/ingest/client"
	storagec "github.com/artefactual-sdps/enduro/internal/api/gen/http/storage/client"
	goahttp "goa.design/goa/v3/http"
	goa "goa.design/goa/v3/pkg"
)

// UsageCommands returns the set of commands and sub-commands using the format
//
//	command (subcommand1|subcommand2|...)
func UsageCommands() string {
	return `ingest (monitor-request|monitor|list-sips|show-sip|list-sip-workflows|confirm-sip|reject-sip|move-sip|move-sip-status|upload-sip)
storage (list-aips|create-aip|submit-aip|update-aip|download-aip|move-aip|move-aip-status|reject-aip|show-aip|list-locations|create-location|show-location|list-location-aips)
`
}

// UsageExamples produces an example of a valid invocation of the CLI tool.
func UsageExamples() string {
	return os.Args[0] + ` ingest monitor-request --token "abc123"` + "\n" +
		os.Args[0] + ` storage list-aips --name "abc123" --earliest-created-time "1970-01-01T00:00:01Z" --latest-created-time "1970-01-01T00:00:01Z" --status "in_review" --limit 1 --offset 1 --token "abc123"` + "\n" +
		""
}

// ParseEndpoint returns the endpoint and payload as specified on the command
// line.
func ParseEndpoint(
	scheme, host string,
	doer goahttp.Doer,
	enc func(*http.Request) goahttp.Encoder,
	dec func(*http.Response) goahttp.Decoder,
	restore bool,
	dialer goahttp.Dialer,
	ingestConfigurer *ingestc.ConnConfigurer,
) (goa.Endpoint, any, error) {
	var (
		ingestFlags = flag.NewFlagSet("ingest", flag.ContinueOnError)

		ingestMonitorRequestFlags     = flag.NewFlagSet("monitor-request", flag.ExitOnError)
		ingestMonitorRequestTokenFlag = ingestMonitorRequestFlags.String("token", "", "")

		ingestMonitorFlags      = flag.NewFlagSet("monitor", flag.ExitOnError)
		ingestMonitorTicketFlag = ingestMonitorFlags.String("ticket", "", "")

		ingestListSipsFlags                   = flag.NewFlagSet("list-sips", flag.ExitOnError)
		ingestListSipsNameFlag                = ingestListSipsFlags.String("name", "", "")
		ingestListSipsAipIDFlag               = ingestListSipsFlags.String("aip-id", "", "")
		ingestListSipsEarliestCreatedTimeFlag = ingestListSipsFlags.String("earliest-created-time", "", "")
		ingestListSipsLatestCreatedTimeFlag   = ingestListSipsFlags.String("latest-created-time", "", "")
		ingestListSipsStatusFlag              = ingestListSipsFlags.String("status", "", "")
		ingestListSipsLimitFlag               = ingestListSipsFlags.String("limit", "", "")
		ingestListSipsOffsetFlag              = ingestListSipsFlags.String("offset", "", "")
		ingestListSipsTokenFlag               = ingestListSipsFlags.String("token", "", "")

		ingestShowSipFlags     = flag.NewFlagSet("show-sip", flag.ExitOnError)
		ingestShowSipIDFlag    = ingestShowSipFlags.String("id", "REQUIRED", "Identifier of SIP to show")
		ingestShowSipTokenFlag = ingestShowSipFlags.String("token", "", "")

		ingestListSipWorkflowsFlags     = flag.NewFlagSet("list-sip-workflows", flag.ExitOnError)
		ingestListSipWorkflowsIDFlag    = ingestListSipWorkflowsFlags.String("id", "REQUIRED", "Identifier of SIP to look up")
		ingestListSipWorkflowsTokenFlag = ingestListSipWorkflowsFlags.String("token", "", "")

		ingestConfirmSipFlags     = flag.NewFlagSet("confirm-sip", flag.ExitOnError)
		ingestConfirmSipBodyFlag  = ingestConfirmSipFlags.String("body", "REQUIRED", "")
		ingestConfirmSipIDFlag    = ingestConfirmSipFlags.String("id", "REQUIRED", "Identifier of SIP to look up")
		ingestConfirmSipTokenFlag = ingestConfirmSipFlags.String("token", "", "")

		ingestRejectSipFlags     = flag.NewFlagSet("reject-sip", flag.ExitOnError)
		ingestRejectSipIDFlag    = ingestRejectSipFlags.String("id", "REQUIRED", "Identifier of SIP to look up")
		ingestRejectSipTokenFlag = ingestRejectSipFlags.String("token", "", "")

		ingestMoveSipFlags     = flag.NewFlagSet("move-sip", flag.ExitOnError)
		ingestMoveSipBodyFlag  = ingestMoveSipFlags.String("body", "REQUIRED", "")
		ingestMoveSipIDFlag    = ingestMoveSipFlags.String("id", "REQUIRED", "Identifier of SIP to move")
		ingestMoveSipTokenFlag = ingestMoveSipFlags.String("token", "", "")

		ingestMoveSipStatusFlags     = flag.NewFlagSet("move-sip-status", flag.ExitOnError)
		ingestMoveSipStatusIDFlag    = ingestMoveSipStatusFlags.String("id", "REQUIRED", "Identifier of SIP to move")
		ingestMoveSipStatusTokenFlag = ingestMoveSipStatusFlags.String("token", "", "")

		ingestUploadSipFlags           = flag.NewFlagSet("upload-sip", flag.ExitOnError)
		ingestUploadSipContentTypeFlag = ingestUploadSipFlags.String("content-type", "multipart/form-data; boundary=goa", "")
		ingestUploadSipTokenFlag       = ingestUploadSipFlags.String("token", "", "")
		ingestUploadSipStreamFlag      = ingestUploadSipFlags.String("stream", "REQUIRED", "path to file containing the streamed request body")

		storageFlags = flag.NewFlagSet("storage", flag.ContinueOnError)

		storageListAipsFlags                   = flag.NewFlagSet("list-aips", flag.ExitOnError)
		storageListAipsNameFlag                = storageListAipsFlags.String("name", "", "")
		storageListAipsEarliestCreatedTimeFlag = storageListAipsFlags.String("earliest-created-time", "", "")
		storageListAipsLatestCreatedTimeFlag   = storageListAipsFlags.String("latest-created-time", "", "")
		storageListAipsStatusFlag              = storageListAipsFlags.String("status", "", "")
		storageListAipsLimitFlag               = storageListAipsFlags.String("limit", "", "")
		storageListAipsOffsetFlag              = storageListAipsFlags.String("offset", "", "")
		storageListAipsTokenFlag               = storageListAipsFlags.String("token", "", "")

		storageCreateAipFlags     = flag.NewFlagSet("create-aip", flag.ExitOnError)
		storageCreateAipBodyFlag  = storageCreateAipFlags.String("body", "REQUIRED", "")
		storageCreateAipTokenFlag = storageCreateAipFlags.String("token", "", "")

		storageSubmitAipFlags     = flag.NewFlagSet("submit-aip", flag.ExitOnError)
		storageSubmitAipBodyFlag  = storageSubmitAipFlags.String("body", "REQUIRED", "")
		storageSubmitAipUUIDFlag  = storageSubmitAipFlags.String("uuid", "REQUIRED", "Identifier of AIP")
		storageSubmitAipTokenFlag = storageSubmitAipFlags.String("token", "", "")

		storageUpdateAipFlags     = flag.NewFlagSet("update-aip", flag.ExitOnError)
		storageUpdateAipUUIDFlag  = storageUpdateAipFlags.String("uuid", "REQUIRED", "Identifier of AIP")
		storageUpdateAipTokenFlag = storageUpdateAipFlags.String("token", "", "")

		storageDownloadAipFlags     = flag.NewFlagSet("download-aip", flag.ExitOnError)
		storageDownloadAipUUIDFlag  = storageDownloadAipFlags.String("uuid", "REQUIRED", "Identifier of AIP")
		storageDownloadAipTokenFlag = storageDownloadAipFlags.String("token", "", "")

		storageMoveAipFlags     = flag.NewFlagSet("move-aip", flag.ExitOnError)
		storageMoveAipBodyFlag  = storageMoveAipFlags.String("body", "REQUIRED", "")
		storageMoveAipUUIDFlag  = storageMoveAipFlags.String("uuid", "REQUIRED", "Identifier of AIP")
		storageMoveAipTokenFlag = storageMoveAipFlags.String("token", "", "")

		storageMoveAipStatusFlags     = flag.NewFlagSet("move-aip-status", flag.ExitOnError)
		storageMoveAipStatusUUIDFlag  = storageMoveAipStatusFlags.String("uuid", "REQUIRED", "Identifier of AIP")
		storageMoveAipStatusTokenFlag = storageMoveAipStatusFlags.String("token", "", "")

		storageRejectAipFlags     = flag.NewFlagSet("reject-aip", flag.ExitOnError)
		storageRejectAipUUIDFlag  = storageRejectAipFlags.String("uuid", "REQUIRED", "Identifier of AIP")
		storageRejectAipTokenFlag = storageRejectAipFlags.String("token", "", "")

		storageShowAipFlags     = flag.NewFlagSet("show-aip", flag.ExitOnError)
		storageShowAipUUIDFlag  = storageShowAipFlags.String("uuid", "REQUIRED", "Identifier of AIP")
		storageShowAipTokenFlag = storageShowAipFlags.String("token", "", "")

		storageListLocationsFlags     = flag.NewFlagSet("list-locations", flag.ExitOnError)
		storageListLocationsTokenFlag = storageListLocationsFlags.String("token", "", "")

		storageCreateLocationFlags     = flag.NewFlagSet("create-location", flag.ExitOnError)
		storageCreateLocationBodyFlag  = storageCreateLocationFlags.String("body", "REQUIRED", "")
		storageCreateLocationTokenFlag = storageCreateLocationFlags.String("token", "", "")

		storageShowLocationFlags     = flag.NewFlagSet("show-location", flag.ExitOnError)
		storageShowLocationUUIDFlag  = storageShowLocationFlags.String("uuid", "REQUIRED", "Identifier of location")
		storageShowLocationTokenFlag = storageShowLocationFlags.String("token", "", "")

		storageListLocationAipsFlags     = flag.NewFlagSet("list-location-aips", flag.ExitOnError)
		storageListLocationAipsUUIDFlag  = storageListLocationAipsFlags.String("uuid", "REQUIRED", "Identifier of location")
		storageListLocationAipsTokenFlag = storageListLocationAipsFlags.String("token", "", "")
	)
	ingestFlags.Usage = ingestUsage
	ingestMonitorRequestFlags.Usage = ingestMonitorRequestUsage
	ingestMonitorFlags.Usage = ingestMonitorUsage
	ingestListSipsFlags.Usage = ingestListSipsUsage
	ingestShowSipFlags.Usage = ingestShowSipUsage
	ingestListSipWorkflowsFlags.Usage = ingestListSipWorkflowsUsage
	ingestConfirmSipFlags.Usage = ingestConfirmSipUsage
	ingestRejectSipFlags.Usage = ingestRejectSipUsage
	ingestMoveSipFlags.Usage = ingestMoveSipUsage
	ingestMoveSipStatusFlags.Usage = ingestMoveSipStatusUsage
	ingestUploadSipFlags.Usage = ingestUploadSipUsage

	storageFlags.Usage = storageUsage
	storageListAipsFlags.Usage = storageListAipsUsage
	storageCreateAipFlags.Usage = storageCreateAipUsage
	storageSubmitAipFlags.Usage = storageSubmitAipUsage
	storageUpdateAipFlags.Usage = storageUpdateAipUsage
	storageDownloadAipFlags.Usage = storageDownloadAipUsage
	storageMoveAipFlags.Usage = storageMoveAipUsage
	storageMoveAipStatusFlags.Usage = storageMoveAipStatusUsage
	storageRejectAipFlags.Usage = storageRejectAipUsage
	storageShowAipFlags.Usage = storageShowAipUsage
	storageListLocationsFlags.Usage = storageListLocationsUsage
	storageCreateLocationFlags.Usage = storageCreateLocationUsage
	storageShowLocationFlags.Usage = storageShowLocationUsage
	storageListLocationAipsFlags.Usage = storageListLocationAipsUsage

	if err := flag.CommandLine.Parse(os.Args[1:]); err != nil {
		return nil, nil, err
	}

	if flag.NArg() < 2 { // two non flag args are required: SERVICE and ENDPOINT (aka COMMAND)
		return nil, nil, fmt.Errorf("not enough arguments")
	}

	var (
		svcn string
		svcf *flag.FlagSet
	)
	{
		svcn = flag.Arg(0)
		switch svcn {
		case "ingest":
			svcf = ingestFlags
		case "storage":
			svcf = storageFlags
		default:
			return nil, nil, fmt.Errorf("unknown service %q", svcn)
		}
	}
	if err := svcf.Parse(flag.Args()[1:]); err != nil {
		return nil, nil, err
	}

	var (
		epn string
		epf *flag.FlagSet
	)
	{
		epn = svcf.Arg(0)
		switch svcn {
		case "ingest":
			switch epn {
			case "monitor-request":
				epf = ingestMonitorRequestFlags

			case "monitor":
				epf = ingestMonitorFlags

			case "list-sips":
				epf = ingestListSipsFlags

			case "show-sip":
				epf = ingestShowSipFlags

			case "list-sip-workflows":
				epf = ingestListSipWorkflowsFlags

			case "confirm-sip":
				epf = ingestConfirmSipFlags

			case "reject-sip":
				epf = ingestRejectSipFlags

			case "move-sip":
				epf = ingestMoveSipFlags

			case "move-sip-status":
				epf = ingestMoveSipStatusFlags

			case "upload-sip":
				epf = ingestUploadSipFlags

			}

		case "storage":
			switch epn {
			case "list-aips":
				epf = storageListAipsFlags

			case "create-aip":
				epf = storageCreateAipFlags

			case "submit-aip":
				epf = storageSubmitAipFlags

			case "update-aip":
				epf = storageUpdateAipFlags

			case "download-aip":
				epf = storageDownloadAipFlags

			case "move-aip":
				epf = storageMoveAipFlags

			case "move-aip-status":
				epf = storageMoveAipStatusFlags

			case "reject-aip":
				epf = storageRejectAipFlags

			case "show-aip":
				epf = storageShowAipFlags

			case "list-locations":
				epf = storageListLocationsFlags

			case "create-location":
				epf = storageCreateLocationFlags

			case "show-location":
				epf = storageShowLocationFlags

			case "list-location-aips":
				epf = storageListLocationAipsFlags

			}

		}
	}
	if epf == nil {
		return nil, nil, fmt.Errorf("unknown %q endpoint %q", svcn, epn)
	}

	// Parse endpoint flags if any
	if svcf.NArg() > 1 {
		if err := epf.Parse(svcf.Args()[1:]); err != nil {
			return nil, nil, err
		}
	}

	var (
		data     any
		endpoint goa.Endpoint
		err      error
	)
	{
		switch svcn {
		case "ingest":
			c := ingestc.NewClient(scheme, host, doer, enc, dec, restore, dialer, ingestConfigurer)
			switch epn {
			case "monitor-request":
				endpoint = c.MonitorRequest()
				data, err = ingestc.BuildMonitorRequestPayload(*ingestMonitorRequestTokenFlag)
			case "monitor":
				endpoint = c.Monitor()
				data, err = ingestc.BuildMonitorPayload(*ingestMonitorTicketFlag)
			case "list-sips":
				endpoint = c.ListSips()
				data, err = ingestc.BuildListSipsPayload(*ingestListSipsNameFlag, *ingestListSipsAipIDFlag, *ingestListSipsEarliestCreatedTimeFlag, *ingestListSipsLatestCreatedTimeFlag, *ingestListSipsStatusFlag, *ingestListSipsLimitFlag, *ingestListSipsOffsetFlag, *ingestListSipsTokenFlag)
			case "show-sip":
				endpoint = c.ShowSip()
				data, err = ingestc.BuildShowSipPayload(*ingestShowSipIDFlag, *ingestShowSipTokenFlag)
			case "list-sip-workflows":
				endpoint = c.ListSipWorkflows()
				data, err = ingestc.BuildListSipWorkflowsPayload(*ingestListSipWorkflowsIDFlag, *ingestListSipWorkflowsTokenFlag)
			case "confirm-sip":
				endpoint = c.ConfirmSip()
				data, err = ingestc.BuildConfirmSipPayload(*ingestConfirmSipBodyFlag, *ingestConfirmSipIDFlag, *ingestConfirmSipTokenFlag)
			case "reject-sip":
				endpoint = c.RejectSip()
				data, err = ingestc.BuildRejectSipPayload(*ingestRejectSipIDFlag, *ingestRejectSipTokenFlag)
			case "move-sip":
				endpoint = c.MoveSip()
				data, err = ingestc.BuildMoveSipPayload(*ingestMoveSipBodyFlag, *ingestMoveSipIDFlag, *ingestMoveSipTokenFlag)
			case "move-sip-status":
				endpoint = c.MoveSipStatus()
				data, err = ingestc.BuildMoveSipStatusPayload(*ingestMoveSipStatusIDFlag, *ingestMoveSipStatusTokenFlag)
			case "upload-sip":
				endpoint = c.UploadSip()
				data, err = ingestc.BuildUploadSipPayload(*ingestUploadSipContentTypeFlag, *ingestUploadSipTokenFlag)
				if err == nil {
					data, err = ingestc.BuildUploadSipStreamPayload(data, *ingestUploadSipStreamFlag)
				}
			}
		case "storage":
			c := storagec.NewClient(scheme, host, doer, enc, dec, restore)
			switch epn {
			case "list-aips":
				endpoint = c.ListAips()
				data, err = storagec.BuildListAipsPayload(*storageListAipsNameFlag, *storageListAipsEarliestCreatedTimeFlag, *storageListAipsLatestCreatedTimeFlag, *storageListAipsStatusFlag, *storageListAipsLimitFlag, *storageListAipsOffsetFlag, *storageListAipsTokenFlag)
			case "create-aip":
				endpoint = c.CreateAip()
				data, err = storagec.BuildCreateAipPayload(*storageCreateAipBodyFlag, *storageCreateAipTokenFlag)
			case "submit-aip":
				endpoint = c.SubmitAip()
				data, err = storagec.BuildSubmitAipPayload(*storageSubmitAipBodyFlag, *storageSubmitAipUUIDFlag, *storageSubmitAipTokenFlag)
			case "update-aip":
				endpoint = c.UpdateAip()
				data, err = storagec.BuildUpdateAipPayload(*storageUpdateAipUUIDFlag, *storageUpdateAipTokenFlag)
			case "download-aip":
				endpoint = c.DownloadAip()
				data, err = storagec.BuildDownloadAipPayload(*storageDownloadAipUUIDFlag, *storageDownloadAipTokenFlag)
			case "move-aip":
				endpoint = c.MoveAip()
				data, err = storagec.BuildMoveAipPayload(*storageMoveAipBodyFlag, *storageMoveAipUUIDFlag, *storageMoveAipTokenFlag)
			case "move-aip-status":
				endpoint = c.MoveAipStatus()
				data, err = storagec.BuildMoveAipStatusPayload(*storageMoveAipStatusUUIDFlag, *storageMoveAipStatusTokenFlag)
			case "reject-aip":
				endpoint = c.RejectAip()
				data, err = storagec.BuildRejectAipPayload(*storageRejectAipUUIDFlag, *storageRejectAipTokenFlag)
			case "show-aip":
				endpoint = c.ShowAip()
				data, err = storagec.BuildShowAipPayload(*storageShowAipUUIDFlag, *storageShowAipTokenFlag)
			case "list-locations":
				endpoint = c.ListLocations()
				data, err = storagec.BuildListLocationsPayload(*storageListLocationsTokenFlag)
			case "create-location":
				endpoint = c.CreateLocation()
				data, err = storagec.BuildCreateLocationPayload(*storageCreateLocationBodyFlag, *storageCreateLocationTokenFlag)
			case "show-location":
				endpoint = c.ShowLocation()
				data, err = storagec.BuildShowLocationPayload(*storageShowLocationUUIDFlag, *storageShowLocationTokenFlag)
			case "list-location-aips":
				endpoint = c.ListLocationAips()
				data, err = storagec.BuildListLocationAipsPayload(*storageListLocationAipsUUIDFlag, *storageListLocationAipsTokenFlag)
			}
		}
	}
	if err != nil {
		return nil, nil, err
	}

	return endpoint, data, nil
}

// ingestUsage displays the usage of the ingest command and its subcommands.
func ingestUsage() {
	fmt.Fprintf(os.Stderr, `The ingest service manages ingested SIPs.
Usage:
    %[1]s [globalflags] ingest COMMAND [flags]

COMMAND:
    monitor-request: Request access to the /monitor WebSocket
    monitor: Obtain access to the /monitor WebSocket
    list-sips: List all ingested SIPs
    show-sip: Show SIP by ID
    list-sip-workflows: List all workflows for a SIP
    confirm-sip: Signal the SIP has been reviewed and accepted
    reject-sip: Signal the SIP has been reviewed and rejected
    move-sip: Move a SIP to a permanent storage location
    move-sip-status: Retrieve the status of a permanent storage location move of the SIP
    upload-sip: Upload a SIP to trigger an ingest workflow

Additional help:
    %[1]s ingest COMMAND --help
`, os.Args[0])
}
func ingestMonitorRequestUsage() {
	fmt.Fprintf(os.Stderr, `%[1]s [flags] ingest monitor-request -token STRING

Request access to the /monitor WebSocket
    -token STRING: 

Example:
    %[1]s ingest monitor-request --token "abc123"
`, os.Args[0])
}

func ingestMonitorUsage() {
	fmt.Fprintf(os.Stderr, `%[1]s [flags] ingest monitor -ticket STRING

Obtain access to the /monitor WebSocket
    -ticket STRING: 

Example:
    %[1]s ingest monitor --ticket "abc123"
`, os.Args[0])
}

func ingestListSipsUsage() {
	fmt.Fprintf(os.Stderr, `%[1]s [flags] ingest list-sips -name STRING -aip-id STRING -earliest-created-time STRING -latest-created-time STRING -status STRING -limit INT -offset INT -token STRING

List all ingested SIPs
    -name STRING: 
    -aip-id STRING: 
    -earliest-created-time STRING: 
    -latest-created-time STRING: 
    -status STRING: 
    -limit INT: 
    -offset INT: 
    -token STRING: 

Example:
    %[1]s ingest list-sips --name "abc123" --aip-id "d1845cb6-a5ea-474a-9ab8-26f9bcd919f5" --earliest-created-time "1970-01-01T00:00:01Z" --latest-created-time "1970-01-01T00:00:01Z" --status "in progress" --limit 1 --offset 1 --token "abc123"
`, os.Args[0])
}

func ingestShowSipUsage() {
	fmt.Fprintf(os.Stderr, `%[1]s [flags] ingest show-sip -id UINT -token STRING

Show SIP by ID
    -id UINT: Identifier of SIP to show
    -token STRING: 

Example:
    %[1]s ingest show-sip --id 1 --token "abc123"
`, os.Args[0])
}

func ingestListSipWorkflowsUsage() {
	fmt.Fprintf(os.Stderr, `%[1]s [flags] ingest list-sip-workflows -id UINT -token STRING

List all workflows for a SIP
    -id UINT: Identifier of SIP to look up
    -token STRING: 

Example:
    %[1]s ingest list-sip-workflows --id 1 --token "abc123"
`, os.Args[0])
}

func ingestConfirmSipUsage() {
	fmt.Fprintf(os.Stderr, `%[1]s [flags] ingest confirm-sip -body JSON -id UINT -token STRING

Signal the SIP has been reviewed and accepted
    -body JSON: 
    -id UINT: Identifier of SIP to look up
    -token STRING: 

Example:
    %[1]s ingest confirm-sip --body '{
      "location_id": "d1845cb6-a5ea-474a-9ab8-26f9bcd919f5"
   }' --id 1 --token "abc123"
`, os.Args[0])
}

func ingestRejectSipUsage() {
	fmt.Fprintf(os.Stderr, `%[1]s [flags] ingest reject-sip -id UINT -token STRING

Signal the SIP has been reviewed and rejected
    -id UINT: Identifier of SIP to look up
    -token STRING: 

Example:
    %[1]s ingest reject-sip --id 1 --token "abc123"
`, os.Args[0])
}

func ingestMoveSipUsage() {
	fmt.Fprintf(os.Stderr, `%[1]s [flags] ingest move-sip -body JSON -id UINT -token STRING

Move a SIP to a permanent storage location
    -body JSON: 
    -id UINT: Identifier of SIP to move
    -token STRING: 

Example:
    %[1]s ingest move-sip --body '{
      "location_id": "d1845cb6-a5ea-474a-9ab8-26f9bcd919f5"
   }' --id 1 --token "abc123"
`, os.Args[0])
}

func ingestMoveSipStatusUsage() {
	fmt.Fprintf(os.Stderr, `%[1]s [flags] ingest move-sip-status -id UINT -token STRING

Retrieve the status of a permanent storage location move of the SIP
    -id UINT: Identifier of SIP to move
    -token STRING: 

Example:
    %[1]s ingest move-sip-status --id 1 --token "abc123"
`, os.Args[0])
}

func ingestUploadSipUsage() {
	fmt.Fprintf(os.Stderr, `%[1]s [flags] ingest upload-sip -content-type STRING -token STRING -stream STRING

Upload a SIP to trigger an ingest workflow
    -content-type STRING: 
    -token STRING: 
    -stream STRING: path to file containing the streamed request body

Example:
    %[1]s ingest upload-sip --content-type "multipart/form-data; boundary=goa" --token "abc123" --stream "goa.png"
`, os.Args[0])
}

// storageUsage displays the usage of the storage command and its subcommands.
func storageUsage() {
	fmt.Fprintf(os.Stderr, `The storage service manages locations and AIPs.
Usage:
    %[1]s [globalflags] storage COMMAND [flags]

COMMAND:
    list-aips: List all AIPs
    create-aip: Create a new AIP
    submit-aip: Start the submission of an AIP
    update-aip: Signal that an AIP submission is complete
    download-aip: Download AIP by AIPID
    move-aip: Move an AIP to a permanent storage location
    move-aip-status: Retrieve the status of a permanent storage location move of the AIP
    reject-aip: Reject an AIP
    show-aip: Show AIP by AIPID
    list-locations: List locations
    create-location: Create a storage location
    show-location: Show location by UUID
    list-location-aips: List all the AIPs stored in the location with UUID

Additional help:
    %[1]s storage COMMAND --help
`, os.Args[0])
}
func storageListAipsUsage() {
	fmt.Fprintf(os.Stderr, `%[1]s [flags] storage list-aips -name STRING -earliest-created-time STRING -latest-created-time STRING -status STRING -limit INT -offset INT -token STRING

List all AIPs
    -name STRING: 
    -earliest-created-time STRING: 
    -latest-created-time STRING: 
    -status STRING: 
    -limit INT: 
    -offset INT: 
    -token STRING: 

Example:
    %[1]s storage list-aips --name "abc123" --earliest-created-time "1970-01-01T00:00:01Z" --latest-created-time "1970-01-01T00:00:01Z" --status "in_review" --limit 1 --offset 1 --token "abc123"
`, os.Args[0])
}

func storageCreateAipUsage() {
	fmt.Fprintf(os.Stderr, `%[1]s [flags] storage create-aip -body JSON -token STRING

Create a new AIP
    -body JSON: 
    -token STRING: 

Example:
    %[1]s storage create-aip --body '{
      "location_id": "d1845cb6-a5ea-474a-9ab8-26f9bcd919f5",
      "name": "abc123",
      "object_key": "d1845cb6-a5ea-474a-9ab8-26f9bcd919f5",
      "status": "in_review",
      "uuid": "d1845cb6-a5ea-474a-9ab8-26f9bcd919f5"
   }' --token "abc123"
`, os.Args[0])
}

func storageSubmitAipUsage() {
	fmt.Fprintf(os.Stderr, `%[1]s [flags] storage submit-aip -body JSON -uuid STRING -token STRING

Start the submission of an AIP
    -body JSON: 
    -uuid STRING: Identifier of AIP
    -token STRING: 

Example:
    %[1]s storage submit-aip --body '{
      "name": "abc123"
   }' --uuid "d1845cb6-a5ea-474a-9ab8-26f9bcd919f5" --token "abc123"
`, os.Args[0])
}

func storageUpdateAipUsage() {
	fmt.Fprintf(os.Stderr, `%[1]s [flags] storage update-aip -uuid STRING -token STRING

Signal that an AIP submission is complete
    -uuid STRING: Identifier of AIP
    -token STRING: 

Example:
    %[1]s storage update-aip --uuid "d1845cb6-a5ea-474a-9ab8-26f9bcd919f5" --token "abc123"
`, os.Args[0])
}

func storageDownloadAipUsage() {
	fmt.Fprintf(os.Stderr, `%[1]s [flags] storage download-aip -uuid STRING -token STRING

Download AIP by AIPID
    -uuid STRING: Identifier of AIP
    -token STRING: 

Example:
    %[1]s storage download-aip --uuid "d1845cb6-a5ea-474a-9ab8-26f9bcd919f5" --token "abc123"
`, os.Args[0])
}

func storageMoveAipUsage() {
	fmt.Fprintf(os.Stderr, `%[1]s [flags] storage move-aip -body JSON -uuid STRING -token STRING

Move an AIP to a permanent storage location
    -body JSON: 
    -uuid STRING: Identifier of AIP
    -token STRING: 

Example:
    %[1]s storage move-aip --body '{
      "location_id": "d1845cb6-a5ea-474a-9ab8-26f9bcd919f5"
   }' --uuid "d1845cb6-a5ea-474a-9ab8-26f9bcd919f5" --token "abc123"
`, os.Args[0])
}

func storageMoveAipStatusUsage() {
	fmt.Fprintf(os.Stderr, `%[1]s [flags] storage move-aip-status -uuid STRING -token STRING

Retrieve the status of a permanent storage location move of the AIP
    -uuid STRING: Identifier of AIP
    -token STRING: 

Example:
    %[1]s storage move-aip-status --uuid "d1845cb6-a5ea-474a-9ab8-26f9bcd919f5" --token "abc123"
`, os.Args[0])
}

func storageRejectAipUsage() {
	fmt.Fprintf(os.Stderr, `%[1]s [flags] storage reject-aip -uuid STRING -token STRING

Reject an AIP
    -uuid STRING: Identifier of AIP
    -token STRING: 

Example:
    %[1]s storage reject-aip --uuid "d1845cb6-a5ea-474a-9ab8-26f9bcd919f5" --token "abc123"
`, os.Args[0])
}

func storageShowAipUsage() {
	fmt.Fprintf(os.Stderr, `%[1]s [flags] storage show-aip -uuid STRING -token STRING

Show AIP by AIPID
    -uuid STRING: Identifier of AIP
    -token STRING: 

Example:
    %[1]s storage show-aip --uuid "d1845cb6-a5ea-474a-9ab8-26f9bcd919f5" --token "abc123"
`, os.Args[0])
}

func storageListLocationsUsage() {
	fmt.Fprintf(os.Stderr, `%[1]s [flags] storage list-locations -token STRING

List locations
    -token STRING: 

Example:
    %[1]s storage list-locations --token "abc123"
`, os.Args[0])
}

func storageCreateLocationUsage() {
	fmt.Fprintf(os.Stderr, `%[1]s [flags] storage create-location -body JSON -token STRING

Create a storage location
    -body JSON: 
    -token STRING: 

Example:
    %[1]s storage create-location --body '{
      "config": {
         "Type": "s3",
         "Value": "{\"bucket\":\"abc123\",\"endpoint\":\"abc123\",\"key\":\"abc123\",\"path_style\":false,\"profile\":\"abc123\",\"region\":\"abc123\",\"secret\":\"abc123\",\"token\":\"abc123\"}"
      },
      "description": "abc123",
      "name": "abc123",
      "purpose": "aip_store",
      "source": "minio"
   }' --token "abc123"
`, os.Args[0])
}

func storageShowLocationUsage() {
	fmt.Fprintf(os.Stderr, `%[1]s [flags] storage show-location -uuid STRING -token STRING

Show location by UUID
    -uuid STRING: Identifier of location
    -token STRING: 

Example:
    %[1]s storage show-location --uuid "d1845cb6-a5ea-474a-9ab8-26f9bcd919f5" --token "abc123"
`, os.Args[0])
}

func storageListLocationAipsUsage() {
	fmt.Fprintf(os.Stderr, `%[1]s [flags] storage list-location-aips -uuid STRING -token STRING

List all the AIPs stored in the location with UUID
    -uuid STRING: Identifier of location
    -token STRING: 

Example:
    %[1]s storage list-location-aips --uuid "d1845cb6-a5ea-474a-9ab8-26f9bcd919f5" --token "abc123"
`, os.Args[0])
}
