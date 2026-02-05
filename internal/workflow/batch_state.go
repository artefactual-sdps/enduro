package workflow

import (
	"github.com/google/uuid"
	temporalsdk_log "go.temporal.io/sdk/log"
	temporalsdk_workflow "go.temporal.io/sdk/workflow"

	"github.com/artefactual-sdps/enduro/internal/childwf"
	"github.com/artefactual-sdps/enduro/internal/datatypes"
	"github.com/artefactual-sdps/enduro/internal/ingest"
)

// batchWorkflowState maintains the state of a batch workflow execution.
// It tracks the batch itself and the details of all SIPs being processed.
type batchWorkflowState struct {
	// logger is used for logging workflow execution details.
	logger temporalsdk_log.Logger

	// batch represents the batch being processed.
	batch datatypes.Batch

	// sipDetails contains details for each SIP in the batch.
	sipDetails []*sipDetails
}

// sipDetails holds information about a single SIP within a batch workflow.
// It includes references to its child workflow for tracking completion.
type sipDetails struct {
	// sip represents the SIP being processed.
	sip datatypes.SIP

	// workflowFuture is used to wait for workflow completion.
	workflowFuture temporalsdk_workflow.ChildWorkflowFuture

	// workflowExecution contains the execution details of the child workflow.
	workflowExecution temporalsdk_workflow.Execution
}

// newBatchWorkflowState initializes a new batchWorkflowState with the given
// workflow context and batch workflow request.
func newBatchWorkflowState(ctx temporalsdk_workflow.Context, req *ingest.BatchWorkflowRequest) *batchWorkflowState {
	return &batchWorkflowState{
		logger:     temporalsdk_workflow.GetLogger(ctx),
		batch:      req.Batch,
		sipDetails: make([]*sipDetails, len(req.Keys)),
	}
}

func (s *batchWorkflowState) PostbatchParams() *childwf.PostbatchParams {
	return &childwf.PostbatchParams{
		Batch: &childwf.PostbatchBatch{
			UUID:      s.batch.UUID,
			SIPSCount: s.batch.SIPSCount,
		},
		SIPs: s.postbatchSIPs(),
	}
}

func (s *batchWorkflowState) SIPs() []datatypes.SIP {
	sips := make([]datatypes.SIP, len(s.sipDetails))
	for i, sd := range s.sipDetails {
		sips[i] = sd.sip
	}
	return sips
}

// addSIPDetails adds a new sipDetails entry to the batchWorkflowState at the specified index.
func (s *batchWorkflowState) addSIPDetails(
	index int,
	sip datatypes.SIP,
	wf temporalsdk_workflow.ChildWorkflowFuture,
	we temporalsdk_workflow.Execution,
) {
	s.sipDetails[index] = &sipDetails{
		sip:               sip,
		workflowFuture:    wf,
		workflowExecution: we,
	}
}

func (s *batchWorkflowState) postbatchSIPs() []*childwf.PostbatchSIP {
	sips := s.SIPs()
	pbs := make([]*childwf.PostbatchSIP, len(sips))

	for i, sip := range sips {
		s := &childwf.PostbatchSIP{
			UUID: sip.UUID,
			Name: sip.Name,
		}
		if sip.AIPID.Valid {
			s.AIPID = &sip.AIPID.UUID
		}
		pbs[i] = s
	}

	return pbs
}

// updateAIPIDs updates the AIP IDs of the SIPs in the batch workflow state
// based on the provided map of SIP UUIDs to AIP UUIDs.
func (state *batchWorkflowState) updateAIPIDs(aipIDs map[uuid.UUID]uuid.UUID) {
	for _, sd := range state.sipDetails {
		aipID, ok := aipIDs[sd.sip.UUID]
		if !ok {
			continue
		}

		sd.sip.AIPID = uuid.NullUUID{UUID: aipID, Valid: true}
	}
}
