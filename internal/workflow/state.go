package workflow

import (
	"slices"
	"time"

	"github.com/artefactual-sdps/enduro/internal/enums"
	"github.com/artefactual-sdps/enduro/internal/ingest"
)

// workflowState is shared state that can be passed to activities.
type workflowState struct {
	// req is populated by the workflow request.
	req *ingest.ProcessingWorkflowRequest

	// status of the ingest workflow.
	status enums.WorkflowStatus

	// tempDirs is a list of temporary directories that should be deleted when
	// the workflow is complete.
	tempDirs []string

	// Identifier of the ingest workflow.
	//
	// It is populated by createWorkflowLocalActivity.
	workflowID int

	// sip, pip and aip track the state of the respective packages.
	sip *sipInfo
	pip *pipInfo
	aip *aipInfo

	// Send to failed information.
	sendToFailed sendToFailed
}

func newWorkflowState(req *ingest.ProcessingWorkflowRequest) *workflowState {
	return &workflowState{
		req:    req,
		status: enums.WorkflowStatusUnspecified,
		sip: &sipInfo{
			dbID:  req.SIPID,
			name:  req.Key,
			isDir: req.IsDir,

			// All SIPs start in queued status.
			status: enums.SIPStatusQueued,
		},

		// Initialize the PIP and AIP to empty structs to avoid nil pointer
		// dereference errors.
		pip: &pipInfo{},
		aip: &aipInfo{},
	}
}

// tempPath registers a filepath for deletion when a workflow session
// completes.
func (s *workflowState) tempPath(path string) {
	if path == "" || slices.Contains(s.tempDirs, path) {
		return
	}
	s.tempDirs = append(s.tempDirs, path)
}

// sipInfo is data about the SIP.
type sipInfo struct {
	// dbID is the database ID of the SIP. It is populated by
	// createSIPLocalActivity as one of the first steps of processing.
	dbID int

	// name is the original blob "key" (filename) of the SIP. It is used as a
	// human-readable identifier for the SIP in the database and UI.
	name string

	// path is the temporary location of the working copy of the SIP.
	path string

	// isDir indicates whether the working copy of the SIP is a directory.
	isDir bool

	// status is the ingest status of the SIP (e.g. "in progress", "done",
	// "error").
	status enums.SIPStatus
}

// pipInfo represents the PIP.
type pipInfo struct {
	// id is the UUID of the PIP. It is populated by the
	// am.PollTransferActivity.
	id string

	// path is the path to the PIP. It is populated by the BundleActivity.
	path string

	// isDir indicates whether the current working copy of the SIP is a
	// filesystem directory.
	isDir bool

	// pipType is the type of the PIP.
	pipType enums.SIPType
}

// aipInfo represents the AIP.
type aipInfo struct {
	// id is the AIP UUID.
	//
	// id is populated when the preservation system creates the AIP.
	id string

	// path to the compressed AIP generated by a3m.
	//
	// path is populated after a3m creates and stores the AIP and is used to
	// upload the AIP to the storage service.
	//
	// path is left blank when using AM as the preservation system because
	// storage of the AIP is handled by AM and the AMSS.
	path string

	// storedAt is the time when the AIP is stored.
	//
	// storedAt is set when the preservation system reports the AIP has been
	// created and stored successfully.
	storedAt time.Time
}

// sendToFailed tracks the SIP or PIP data required to send the package to
// the "failed" package location if processing fails.
type sendToFailed struct {
	// path to the SIP or PIP.
	path string

	// activityName is the name of the activity that will send the package to
	// the correct "failed" location. The value can be "send-to-failed-sips" or
	// "send-to-failed-pips".
	activityName string

	// needsZipping indicates whether the package needs to be zipped before
	// uploading it to the "failed" location.
	needsZipping bool
}
