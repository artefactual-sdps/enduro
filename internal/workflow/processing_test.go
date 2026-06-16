package workflow

import (
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/artefactual-sdps/temporal-activities/archiveextract"
	"github.com/artefactual-sdps/temporal-activities/archivezip"
	"github.com/artefactual-sdps/temporal-activities/bucketcopy"
	"github.com/artefactual-sdps/temporal-activities/bucketdelete"
	"github.com/artefactual-sdps/temporal-activities/bucketdownload"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	temporalsdk_workflow "go.temporal.io/sdk/workflow"

	"github.com/artefactual-sdps/enduro/internal/a3m"
	"github.com/artefactual-sdps/enduro/internal/am"
	"github.com/artefactual-sdps/enduro/internal/config"
	"github.com/artefactual-sdps/enduro/internal/datatypes"
	"github.com/artefactual-sdps/enduro/internal/enums"
	"github.com/artefactual-sdps/enduro/internal/ingest"
	"github.com/artefactual-sdps/enduro/internal/premis"
	"github.com/artefactual-sdps/enduro/internal/pres"
	"github.com/artefactual-sdps/enduro/internal/temporal"
	"github.com/artefactual-sdps/enduro/internal/workflow/activities"
	"github.com/artefactual-sdps/enduro/internal/workflow/localact"
	"github.com/artefactual-sdps/enduro/pkg/childwf"
)

func TestProcessingWorkflow(t *testing.T) {
	suite.Run(t, new(ProcessingWorkflowTestSuite))
}

// TestConfirmation tests:
// - a3m as preservation system.
// - The "create and review AIP" workflow type.
// - The user accepting the AIP in the review step.
// - Watched bucket download.
// - Watched bucket retention period.
func (s *ProcessingWorkflowTestSuite) TestConfirmation() {
	s.SetupWorkflowTest(config.Configuration{
		A3m:          a3m.Config{ShareDir: s.CreateTransferDir()},
		Preservation: pres.Config{TaskQueue: temporal.A3mWorkerTaskQueue},
		Ingest:       ingest.Config{Storage: ingest.StorageConfig{DefaultPermanentLocationID: locationID}},
	}, nil)

	// Signal handler that mimics SIP/AIP confirmation.
	s.env.RegisterDelayedCallback(
		func() {
			s.env.SignalWorkflow(
				ingest.ReviewPerformedSignalName,
				ingest.ReviewPerformedSignal{Accepted: true, LocationID: &locationID},
			)
		},
		0,
	)

	params := defaultParams()
	params.workflowType = enums.WorkflowTypeCreateAndReviewAip
	downloadExpectations(s, params)
	calcChecksumExpectations(s, params)
	checkDuplicateSIPExpectations(s, params)
	expectations["archiveExtract"](s, params)
	expectations["classifySIP"](s, params)
	countSIPFilesExpectations(s, params)
	expectations["saveFileCount"](s, params)
	reviewA3mExpectations(s, params)
	params.updateTaskParams(reviewAIPTaskID, enums.TaskStatusDone, "", "Reviewed and accepted")
	expectations["completeTask"](s, params)
	params.updateTaskParams(moveAIPTaskID, enums.TaskStatusInProgress, "Move AIP", "Moving to permanent storage")
	expectations["createTask"](s, params)
	expectations["moveAIP"](s, params)
	expectations["pollMoveAIP"](s, params)
	params.updateTaskParams(
		moveAIPTaskID,
		enums.TaskStatusDone,
		"",
		"Moved to location f2cc963f-c14d-4eaa-b950-bd207189a1f1",
	)
	expectations["completeTask"](s, params)
	cleanupExpectations(s, params)

	s.ExecuteAndValidateWorkflow(&ingest.ProcessingWorkflowRequest{
		Key:             key,
		WatcherName:     watcherName,
		RetentionPeriod: retentionPeriod,
		Type:            enums.WorkflowTypeCreateAndReviewAip,
		SIPUUID:         sipUUID,
		SIPName:         sipName,
	}, &ingest.ProcessingWorkflowResult{}, false)
}

// TestRejection tests:
// - a3m as preservation system.
// - The "create and review AIP" workflow type.
// - The user rejecting the AIP in the review step.
// - Watched bucket download.
// - Watched bucket retention period.
func (s *ProcessingWorkflowTestSuite) TestRejection() {
	s.SetupWorkflowTest(config.Configuration{
		A3m:          a3m.Config{ShareDir: s.CreateTransferDir()},
		Preservation: pres.Config{TaskQueue: temporal.A3mWorkerTaskQueue},
		Ingest:       ingest.Config{Storage: ingest.StorageConfig{DefaultPermanentLocationID: locationID}},
	}, nil)

	// Signal handler that mimics SIP/AIP rejection.
	s.env.RegisterDelayedCallback(
		func() {
			s.env.SignalWorkflow(
				ingest.ReviewPerformedSignalName,
				ingest.ReviewPerformedSignal{Accepted: false},
			)
		},
		0,
	)

	params := defaultParams()
	params.workflowType = enums.WorkflowTypeCreateAndReviewAip
	downloadExpectations(s, params)
	calcChecksumExpectations(s, params)
	checkDuplicateSIPExpectations(s, params)
	expectations["archiveExtract"](s, params)
	expectations["classifySIP"](s, params)
	countSIPFilesExpectations(s, params)
	expectations["saveFileCount"](s, params)
	reviewA3mExpectations(s, params)
	params.updateTaskParams(reviewAIPTaskID, enums.TaskStatusDone, "", "Reviewed and rejected")
	expectations["completeTask"](s, params)

	s.env.OnActivity(
		activities.RejectSIPActivityName,
		sessionCtx,
		&activities.RejectSIPActivityParams{AIPID: aipUUID.String()},
	).Return(nil, nil)

	cleanupExpectations(s, params)

	s.ExecuteAndValidateWorkflow(&ingest.ProcessingWorkflowRequest{
		Key:             key,
		WatcherName:     watcherName,
		RetentionPeriod: retentionPeriod,
		Type:            enums.WorkflowTypeCreateAndReviewAip,
		SIPUUID:         sipUUID,
		SIPName:         sipName,
	}, &ingest.ProcessingWorkflowResult{}, false)
}

// TestAutoApprovedAIP tests:
// - a3m as preservation system.
// - The "create AIP" workflow type.
// - Watched bucket download.
// - Watched bucket negative retention period.
func (s *ProcessingWorkflowTestSuite) TestAutoApprovedAIP() {
	s.SetupWorkflowTest(config.Configuration{
		A3m:          a3m.Config{ShareDir: s.CreateTransferDir()},
		Preservation: pres.Config{TaskQueue: temporal.A3mWorkerTaskQueue},
		Ingest:       ingest.Config{Storage: ingest.StorageConfig{DefaultPermanentLocationID: locationID}},
	}, nil)

	params := defaultParams()
	downloadExpectations(s, params)
	calcChecksumExpectations(s, params)
	checkDuplicateSIPExpectations(s, params)
	expectations["archiveExtract"](s, params)
	expectations["classifySIP"](s, params)
	countSIPFilesExpectations(s, params)
	expectations["saveFileCount"](s, params)
	autoApproveA3mExpectations(s, params)
	params.retentionPeriod = -1 * time.Second
	cleanupExpectations(s, params)

	s.ExecuteAndValidateWorkflow(&ingest.ProcessingWorkflowRequest{
		Key:             key,
		WatcherName:     watcherName,
		RetentionPeriod: params.retentionPeriod,
		Type:            enums.WorkflowTypeCreateAip,
		SIPUUID:         sipUUID,
		SIPName:         sipName,
	}, &ingest.ProcessingWorkflowResult{}, false)
}

// TestAMWorkflow tests:
// - Archivematica as preservation system.
// - The "create AIP" workflow type.
// - Bag creation.
// - PREMIS validation in the AM branch.
// - Watched bucket download.
// - Watched bucket retention period.
// - Batch signal handling (continue).
func (s *ProcessingWorkflowTestSuite) TestAMWorkflow() {
	s.SetupWorkflowTest(config.Configuration{
		AM:             am.Config{ZipPIP: true, TransferDeadline: time.Second},
		Preservation:   pres.Config{TaskQueue: temporal.AmWorkerTaskQueue},
		Ingest:         ingest.Config{Storage: ingest.StorageConfig{DefaultPermanentLocationID: amssLocationID}},
		ValidatePREMIS: premis.Config{Enabled: true, XSDPath: "/home/enduro/premis.xsd"},
	}, nil)

	params := defaultParams()
	downloadExpectations(s, params)
	calcChecksumExpectations(s, params)
	checkDuplicateSIPExpectations(s, params)
	expectations["archiveExtract"](s, params)
	expectations["classifySIP"](s, params)
	expectations["createBag"](s, params)
	countSIPFilesExpectations(s, params)
	expectations["saveFileCount"](s, params)
	params.updateTaskParams(valPREMISTaskID, enums.TaskStatusInProgress, "Validate PREMIS", "")
	expectations["createTask"](s, params)
	params.premisXMLPath = filepath.Join(extractPath, "data", "metadata", "premis.xml")
	expectations["validatePREMIS"](s, params)
	params.updateTaskParams(valPREMISTaskID, enums.TaskStatusDone, "", "PREMIS is valid")
	expectations["completeTask"](s, params)
	expectations["zipArchive"](s, params)

	// Batch signal handler and expectations.
	s.env.RegisterDelayedCallback(func() {
		s.env.SignalWorkflow(ingest.BatchSignalName, ingest.BatchSignal{Continue: true})
	}, 0)
	params.sipStatus = enums.SIPStatusValidated
	expectations["setStatus"](s, params)
	params.updateTaskParams(
		batchWaitTaskID,
		enums.TaskStatusInProgress,
		"Waiting for other SIPs in Batch",
		"The SIP has been validated and is waiting for other SIPs in the Batch to be validated.",
	)
	expectations["createTask"](s, params)
	params.updateTaskParams(batchWaitTaskID, enums.TaskStatusDone, "", "All SIPs in the Batch have been validated.")
	expectations["completeTask"](s, params)
	params.sipStatus = enums.SIPStatusProcessing
	expectations["setStatus"](s, params)

	// Archivematica specific expectations.
	baseName := filepath.Base(extractPath)
	s.env.OnActivity(
		am.UploadTransferActivityName,
		sessionCtx,
		&am.UploadTransferActivityParams{SourcePath: extractPath + "/"},
	).Return(&am.UploadTransferActivityResult{RemoteRelativePath: baseName}, nil)
	s.env.OnActivity(
		am.StartTransferActivityName,
		sessionCtx,
		&am.StartTransferActivityParams{
			Name:         sipName,
			RelativePath: filepath.Join(baseName, key),
			ZipPIP:       true,
		},
	).Return(&am.StartTransferActivityResult{TransferID: transferID.String()}, nil)
	s.env.OnActivity(
		am.PollTransferActivityName,
		sessionCtx,
		&am.PollTransferActivityParams{
			TransferID:   transferID.String(),
			WorkflowUUID: workflowUUID,
		},
	).Return(&am.PollTransferActivityResult{SIPID: aipUUID.String()}, nil)
	s.env.OnActivity(
		am.PollIngestActivityName,
		sessionCtx,
		&am.PollIngestActivityParams{
			SIPID:        aipUUID.String(),
			WorkflowUUID: workflowUUID,
		},
	).Return(&am.PollIngestActivityResult{Status: "COMPLETE"}, nil)
	s.env.OnActivity(
		activities.CreateStorageAIPActivityName,
		sessionCtx,
		&activities.CreateStorageAIPActivityParams{
			Name:       sipName,
			AIPID:      aipUUID.String(),
			ObjectKey:  aipUUID.String(),
			Status:     "stored",
			LocationID: &amssLocationID,
		},
	).Return(&activities.CreateStorageAIPActivityResult{}, nil)
	s.env.OnActivity(
		am.DeleteTransferActivityName,
		sessionCtx,
		&am.DeleteTransferActivityParams{Destination: baseName},
	).Return(nil, nil)

	expectations["updateSIPProcessing"](s, params)
	params.removePaths = []string{tempPath, extractPath + "/transfer.zip"}
	cleanupExpectations(s, params)

	s.ExecuteAndValidateWorkflow(&ingest.ProcessingWorkflowRequest{
		WatcherName:     watcherName,
		RetentionPeriod: retentionPeriod,
		Type:            enums.WorkflowTypeCreateAip,
		Key:             key,
		SIPUUID:         sipUUID,
		SIPName:         sipName,
		BatchUUID:       batchUUID,
	}, &ingest.ProcessingWorkflowResult{}, false)
}

// TestChildWorkflows tests:
// - a3m as preservation system.
// - The "create AIP" workflow type.
// - preprocessing child workflow.
// - custom metadata from preprocessing to poststorage.
// - poststorage child workflows.
// - user context propagation to child workflows.
// - Bag validation.
// - Watched bucket download.
// - Watched bucket custom retention period.
func (s *ProcessingWorkflowTestSuite) TestChildWorkflows() {
	user := &childwf.User{Email: "nobody@example.com"}
	s.SetupWorkflowTest(config.Configuration{
		A3m:          a3m.Config{ShareDir: s.CreateTransferDir()},
		Preservation: pres.Config{TaskQueue: temporal.A3mWorkerTaskQueue},
		Ingest:       ingest.Config{Storage: ingest.StorageConfig{DefaultPermanentLocationID: locationID}},
		ChildWorkflows: config.ChildWorkflowConfigs{
			{
				Type:         childwf.WorkflowTypePreprocessing,
				Namespace:    "default",
				TaskQueue:    "preprocessing",
				WorkflowName: "preprocessing",
				Extract:      true,
				SharedPath:   prepSharedPath,
			},
			{
				Type:         childwf.WorkflowTypePoststorage,
				Namespace:    "default",
				TaskQueue:    "poststorage",
				WorkflowName: "poststorage",
			},
		},
	}, nil)

	params := defaultParams()
	params.downloadDestPath = prepSharedPath
	params.downloadPath = prepDownloadPath + "/" + key
	params.extractPath = prepExtractPath
	downloadExpectations(s, params)
	calcChecksumExpectations(s, params)
	checkDuplicateSIPExpectations(s, params)
	preprocessingMetadata := childwf.CustomMetadata{
		"external_id": json.RawMessage(`"12345"`),
		"flags":       json.RawMessage(`{"validated":true}`),
	}
	poststorageMetadata := childwf.CustomMetadata{
		"package": json.RawMessage(`{"type":"aip","identifier":"AIP-67890"}`),
		"storage": json.RawMessage(`{"notified":true,"location":"permanent"}`),
	}

	s.env.OnWorkflow(
		"preprocessing",
		internalCtx,
		&childwf.PreprocessingParams{
			User:         user,
			RelativePath: strings.TrimPrefix(prepDownloadPath+"/"+key, prepSharedPath),
			SIPID:        sipUUID,
			SIPName:      sipName,
		},
	).Return(
		&childwf.PreprocessingResult{
			Outcome:        childwf.OutcomeSuccess,
			CustomMetadata: preprocessingMetadata,
			RelativePath:   strings.TrimPrefix(prepExtractPath, prepSharedPath),
			Tasks: []*childwf.Task{
				{
					Name:        "Identify SIP structure",
					Message:     "SIP structure identified: VecteurAIP",
					Outcome:     childwf.TaskOutcomeSuccess,
					StartedAt:   time.Date(2024, 6, 14, 10, 5, 32, 0, time.UTC),
					CompletedAt: time.Date(2024, 6, 14, 10, 5, 33, 0, time.UTC),
				},
			},
		},
		nil,
	)

	s.env.OnActivity(
		localact.SaveChildwfTasksActivity,
		ctx,
		localact.SaveChildwfTasksActivityParams{
			Ingestsvc:    s.workflow.ingestsvc,
			RNG:          s.workflow.rng,
			WorkflowUUID: workflowUUID,
			Tasks: []*childwf.Task{
				{
					Name:        "Identify SIP structure",
					Message:     "SIP structure identified: VecteurAIP",
					Outcome:     childwf.TaskOutcomeSuccess,
					StartedAt:   time.Date(2024, 6, 14, 10, 5, 32, 0, time.UTC),
					CompletedAt: time.Date(2024, 6, 14, 10, 5, 33, 0, time.UTC),
				},
			},
		},
	).Return(&localact.SaveChildwfTasksActivityResult{Count: 1}, nil)

	params.sipType = enums.SIPTypeBagIt
	expectations["classifySIP"](s, params)
	params.updateTaskParams(valBagTaskID, enums.TaskStatusInProgress, "Validate Bag", "")
	expectations["createTask"](s, params)
	expectations["validateBag"](s, params)
	params.updateTaskParams(valBagTaskID, enums.TaskStatusDone, "", "Bag successfully validated")
	expectations["completeTask"](s, params)
	countSIPFilesExpectations(s, params)
	expectations["saveFileCount"](s, params)
	autoApproveA3mExpectations(s, params)

	s.env.OnWorkflow(
		"poststorage",
		internalCtx,
		&childwf.PostStorageParams{
			User:           user,
			AIPUUID:        aipUUID.String(),
			CustomMetadata: preprocessingMetadata,
		},
	).Return(
		&childwf.PostStorageResult{
			Outcome:        childwf.OutcomeSuccess,
			CustomMetadata: poststorageMetadata,
			Tasks: []*childwf.Task{
				{
					Name:        "Notify external system",
					Message:     "External system notified.",
					Outcome:     childwf.TaskOutcomeSuccess,
					StartedAt:   time.Date(2024, 6, 14, 10, 6, 32, 0, time.UTC),
					CompletedAt: time.Date(2024, 6, 14, 10, 6, 33, 0, time.UTC),
				},
			},
		},
		nil,
	)

	s.env.OnActivity(
		localact.SaveChildwfTasksActivity,
		ctx,
		localact.SaveChildwfTasksActivityParams{
			Ingestsvc:    s.workflow.ingestsvc,
			RNG:          s.workflow.rng,
			WorkflowUUID: workflowUUID,
			Tasks: []*childwf.Task{
				{
					Name:        "Notify external system",
					Message:     "External system notified.",
					Outcome:     childwf.TaskOutcomeSuccess,
					StartedAt:   time.Date(2024, 6, 14, 10, 6, 32, 0, time.UTC),
					CompletedAt: time.Date(2024, 6, 14, 10, 6, 33, 0, time.UTC),
				},
			},
		},
	).Return(&localact.SaveChildwfTasksActivityResult{Count: 1}, nil)

	params.removePaths = []string{prepDownloadPath, transferPath}
	params.retentionPeriod = 48 * time.Hour
	cleanupExpectations(s, params)

	s.ExecuteAndValidateWorkflow(&ingest.ProcessingWorkflowRequest{
		User:            user,
		Key:             key,
		WatcherName:     watcherName,
		RetentionPeriod: params.retentionPeriod,
		Type:            enums.WorkflowTypeCreateAip,
		SIPUUID:         sipUUID,
		SIPName:         sipName,
	}, &ingest.ProcessingWorkflowResult{
		CustomMetadata: childwf.CustomMetadata{
			"external_id": json.RawMessage(`"12345"`),
			"flags":       json.RawMessage(`{"validated":true}`),
			"package":     json.RawMessage(`{"type":"aip","identifier":"AIP-67890"}`),
			"storage":     json.RawMessage(`{"notified":true,"location":"permanent"}`),
		},
	}, false)
}

// TestPreprocessingDecisionFlow tests:
// - a3m as preservation system.
// - The "create AIP" workflow type.
// - preprocessing child workflow requesting a human decision from the parent.
// - parent signaling the selected option back to the child.
func (s *ProcessingWorkflowTestSuite) TestPreprocessingDecisionFlow() {
	const preprocessingDecisionTaskID = 109

	s.SetupWorkflowTest(config.Configuration{
		A3m:          a3m.Config{ShareDir: s.CreateTransferDir()},
		Preservation: pres.Config{TaskQueue: temporal.A3mWorkerTaskQueue},
		Ingest:       ingest.Config{Storage: ingest.StorageConfig{DefaultPermanentLocationID: locationID}},
		ChildWorkflows: config.ChildWorkflowConfigs{
			{
				Type:         childwf.WorkflowTypePreprocessing,
				Namespace:    "default",
				TaskQueue:    "preprocessing",
				WorkflowName: "preprocessing",
				Extract:      true,
				SharedPath:   prepSharedPath,
			},
		},
	}, func(
		ctx temporalsdk_workflow.Context,
		params *childwf.PreprocessingParams,
	) (*childwf.PreprocessingResult, error) {
		parent := temporalsdk_workflow.GetInfo(ctx).ParentWorkflowExecution
		err := temporalsdk_workflow.SignalExternalWorkflow(
			ctx,
			parent.ID,
			parent.RunID,
			childwf.DecisionRequestSignalName,
			childwf.DecisionRequest{
				Message: "Preprocessing requires human decision.",
				Options: []string{"Cancel", "Continue"},
			},
		).Get(ctx, nil)
		if err != nil {
			return nil, err
		}

		var decision childwf.DecisionResponse
		_ = temporalsdk_workflow.GetSignalChannel(ctx, childwf.DecisionResponseSignalName).Receive(ctx, &decision)
		s.Equal(decision.Option, "Continue")

		return &childwf.PreprocessingResult{
			Outcome:      childwf.OutcomeSuccess,
			RelativePath: strings.TrimPrefix(prepExtractPath, prepSharedPath),
		}, nil
	})

	params := defaultParams()
	params.downloadDestPath = prepSharedPath
	params.downloadPath = prepDownloadPath + "/" + key
	params.extractPath = prepExtractPath
	downloadExpectations(s, params)
	calcChecksumExpectations(s, params)
	checkDuplicateSIPExpectations(s, params)

	params.updateTaskParams(
		preprocessingDecisionTaskID,
		enums.TaskStatusPending,
		"Preprocessing workflow is waiting for user decision",
		"Preprocessing requires human decision.\n\nAvailable options:\n- Cancel\n- Continue",
	)
	expectations["createTask"](s, params)
	params.sipStatus = enums.SIPStatusPending
	s.env.OnActivity(
		setStatusLocalActivity,
		ctx,
		s.workflow.ingestsvc,
		sipUUID,
		params.sipStatus,
	).Run(func(args mock.Arguments) {
		// The update handler is registered from workflow start, but rejects until
		// the child decision request is recorded. Hook this to the pending-status
		// activity so the test submits the decision at the same point a user sees it.
		s.env.UpdateWorkflowNoRejection(
			ingest.ChildDecisionUpdateName,
			"preprocessing-decision",
			s.T(),
			childwf.DecisionResponse{Option: "Continue"},
		)
	}).Return(&setStatusLocalActivityResult{}, nil)
	params.workflowStatus = enums.WorkflowStatusPending
	expectations["setWorkflowStatus"](s, params)
	params.sipStatus = enums.SIPStatusProcessing
	expectations["setStatus"](s, params)
	params.workflowStatus = enums.WorkflowStatusInProgress
	expectations["setWorkflowStatus"](s, params)
	params.updateTaskParams(
		preprocessingDecisionTaskID,
		enums.TaskStatusDone,
		"",
		"Preprocessing requires human decision.\n\nAvailable options:\n- Cancel\n- Continue\n\nUser selected option: Continue",
	)
	expectations["completeTask"](s, params)

	params.sipType = enums.SIPTypeBagIt
	expectations["classifySIP"](s, params)
	params.updateTaskParams(valBagTaskID, enums.TaskStatusInProgress, "Validate Bag", "")
	expectations["createTask"](s, params)
	expectations["validateBag"](s, params)
	params.updateTaskParams(valBagTaskID, enums.TaskStatusDone, "", "Bag successfully validated")
	expectations["completeTask"](s, params)
	countSIPFilesExpectations(s, params)
	s.env.OnActivity(
		updateSIPLocalActivity,
		ctx,
		s.workflow.ingestsvc,
		&updateSIPLocalActivityParams{
			UUID:      sipUUID,
			FileCount: fileCount,
		},
	).Return(nil, nil)
	autoApproveA3mExpectations(s, params)
	params.removePaths = []string{prepDownloadPath, transferPath}
	cleanupExpectations(s, params)

	s.ExecuteAndValidateWorkflow(&ingest.ProcessingWorkflowRequest{
		WatcherName:     watcherName,
		RetentionPeriod: retentionPeriod,
		Type:            enums.WorkflowTypeCreateAip,
		Key:             key,
		SIPUUID:         sipUUID,
		SIPName:         sipName,
	}, &ingest.ProcessingWorkflowResult{}, false)
}

// TestFailedSIP tests:
// - a3m as preservation system.
// - The "create AIP" workflow type.
// - preprocessing child workflow error.
// - Move to failed SIP.
// - Watched bucket download.
func (s *ProcessingWorkflowTestSuite) TestFailedSIP() {
	s.SetupWorkflowTest(config.Configuration{
		A3m:          a3m.Config{ShareDir: s.CreateTransferDir()},
		Preservation: pres.Config{TaskQueue: temporal.A3mWorkerTaskQueue},

		Ingest: ingest.Config{Storage: ingest.StorageConfig{DefaultPermanentLocationID: locationID}},
		ChildWorkflows: config.ChildWorkflowConfigs{
			{
				Type:         childwf.WorkflowTypePreprocessing,
				Namespace:    "default",
				TaskQueue:    "preprocessing",
				WorkflowName: "preprocessing",
				Extract:      true,
				SharedPath:   prepSharedPath,
			},
		},
	}, nil)

	params := defaultParams()
	params.downloadDestPath = prepSharedPath
	params.downloadPath = prepDownloadPath + "/" + key
	downloadExpectations(s, params)
	calcChecksumExpectations(s, params)
	checkDuplicateSIPExpectations(s, params)

	// Fail the workflow on preprocessing.
	s.env.OnWorkflow(
		"preprocessing",
		internalCtx,
		&childwf.PreprocessingParams{
			RelativePath: strings.TrimPrefix(prepDownloadPath+"/"+key, prepSharedPath),
			SIPID:        sipUUID,
			SIPName:      sipName,
		},
	).Return(
		&childwf.PreprocessingResult{
			Outcome:      childwf.OutcomeContentError,
			RelativePath: strings.TrimPrefix(prepExtractPath, prepSharedPath),
		},
		nil,
	)

	params.sipStatus = enums.SIPStatusFailed
	params.failedAs = enums.SIPFailedAsSIP
	params.failedKey = failedSIPKey
	params.failedPath = prepDownloadPath + "/" + key
	params.removePaths = []string{prepDownloadPath}
	expectations["uploadToFailed"](s, params)
	expectations["removePaths"](s, params)
	expectations["updateSIPFailed"](s, params)
	expectations["completeWorkflow"](s, params)

	s.ExecuteAndValidateWorkflow(&ingest.ProcessingWorkflowRequest{
		Key:         key,
		WatcherName: watcherName,
		Type:        enums.WorkflowTypeCreateAip,
		SIPUUID:     sipUUID,
		SIPName:     sipName,
	}, nil, true)
}

// TestFailedSIP tests:
// - a3m as preservation system.
// - The "create AIP" workflow type.
// - PREMIS validation in the a3m branch.
// - a3m error.
// - Move to failed PIP.
// - Watched bucket download.
func (s *ProcessingWorkflowTestSuite) TestFailedPIPA3m() {
	s.SetupWorkflowTest(config.Configuration{
		A3m:            a3m.Config{ShareDir: s.CreateTransferDir()},
		Preservation:   pres.Config{TaskQueue: temporal.A3mWorkerTaskQueue},
		Ingest:         ingest.Config{Storage: ingest.StorageConfig{DefaultPermanentLocationID: locationID}},
		ValidatePREMIS: premis.Config{Enabled: true, XSDPath: "/home/enduro/premis.xsd"},
	}, nil)

	params := defaultParams()
	downloadExpectations(s, params)
	calcChecksumExpectations(s, params)
	checkDuplicateSIPExpectations(s, params)
	expectations["archiveExtract"](s, params)
	params.sipType = enums.SIPTypeBagIt
	expectations["classifySIP"](s, params)
	params.updateTaskParams(valBagTaskID, enums.TaskStatusInProgress, "Validate Bag", "")
	expectations["createTask"](s, params)
	expectations["validateBag"](s, params)
	params.updateTaskParams(valBagTaskID, enums.TaskStatusDone, "", "Bag successfully validated")
	expectations["completeTask"](s, params)
	countSIPFilesExpectations(s, params)
	expectations["saveFileCount"](s, params)
	expectations["bundle"](s, params)
	params.updateTaskParams(valPREMISTaskID, enums.TaskStatusInProgress, "Validate PREMIS", "")
	expectations["createTask"](s, params)
	expectations["validatePREMIS"](s, params)
	params.updateTaskParams(valPREMISTaskID, enums.TaskStatusDone, "", "PREMIS is valid")
	expectations["completeTask"](s, params)

	// Fail the workflow on a3m AIP creation.
	s.env.OnActivity(
		a3m.CreateAIPActivityName,
		sessionCtx,
		&a3m.CreateAIPActivityParams{
			Name:         sipName,
			Path:         transferPath,
			WorkflowUUID: workflowUUID,
		},
	).Return(nil, fmt.Errorf("a3m error"))

	s.env.OnActivity(
		archivezip.Name,
		sessionCtx,
		&archivezip.Params{SourceDir: transferPath},
	).Return(&archivezip.Result{Path: transferPath + ".zip"}, nil)

	params.sipStatus = enums.SIPStatusError
	params.failedAs = enums.SIPFailedAsPIP
	params.failedKey = failedPIPKey
	params.failedPath = transferPath + ".zip"
	expectations["uploadToFailed"](s, params)
	expectations["removePaths"](s, params)
	expectations["updateSIPFailed"](s, params)
	expectations["completeWorkflow"](s, params)

	s.ExecuteAndValidateWorkflow(&ingest.ProcessingWorkflowRequest{
		Key:         key,
		WatcherName: watcherName,
		Type:        enums.WorkflowTypeCreateAip,
		SIPUUID:     sipUUID,
		SIPName:     sipName,
	}, nil, true)
}

// TestFailedPIPAM tests:
// - Archivematica as preservation system.
// - The "create AIP" workflow type.
// - Archivematica error.
// - Move to failed PIP.
// - Watched bucket download.
func (s *ProcessingWorkflowTestSuite) TestFailedPIPAM() {
	s.SetupWorkflowTest(config.Configuration{
		AM:           am.Config{ZipPIP: true},
		Preservation: pres.Config{TaskQueue: temporal.AmWorkerTaskQueue},
		Ingest:       ingest.Config{Storage: ingest.StorageConfig{DefaultPermanentLocationID: amssLocationID}},
	}, nil)

	params := defaultParams()
	downloadExpectations(s, params)
	calcChecksumExpectations(s, params)
	checkDuplicateSIPExpectations(s, params)
	expectations["archiveExtract"](s, params)
	expectations["classifySIP"](s, params)
	countSIPFilesExpectations(s, params)
	expectations["saveFileCount"](s, params)
	expectations["createBag"](s, params)
	expectations["zipArchive"](s, params)

	// Fail the workflow on AM upload.
	s.env.OnActivity(
		am.UploadTransferActivityName,
		sessionCtx,
		&am.UploadTransferActivityParams{SourcePath: extractPath + "/"},
	).Return(nil, fmt.Errorf("AM error"))

	params.removePaths = []string{tempPath, extractPath + "/transfer.zip"}
	params.sipStatus = enums.SIPStatusError
	params.failedAs = enums.SIPFailedAsPIP
	params.failedKey = failedPIPKey
	params.failedPath = extractPath + "/transfer.zip"
	expectations["uploadToFailed"](s, params)
	expectations["removePaths"](s, params)
	expectations["updateSIPFailed"](s, params)
	expectations["completeWorkflow"](s, params)

	s.ExecuteAndValidateWorkflow(&ingest.ProcessingWorkflowRequest{
		WatcherName: watcherName,
		Type:        enums.WorkflowTypeCreateAip,
		Key:         key,
		SIPUUID:     sipUUID,
		SIPName:     sipName,
	}, nil, true)
}

// TestInternalUpload tests:
// - a3m as preservation system.
// - The "create AIP" workflow type.
// - Internal bucket download.
// - Internal upload retention period.
func (s *ProcessingWorkflowTestSuite) TestInternalUpload() {
	s.SetupWorkflowTest(config.Configuration{
		A3m:          a3m.Config{ShareDir: s.CreateTransferDir()},
		Preservation: pres.Config{TaskQueue: temporal.A3mWorkerTaskQueue},
		Ingest:       ingest.Config{Storage: ingest.StorageConfig{DefaultPermanentLocationID: locationID}},
	}, nil)

	params := defaultParams()
	expectations["setStatusInProgress"](s, params)
	expectations["createWorkflow"](s, params)
	params.updateTaskParams(copySIPTaskID, enums.TaskStatusInProgress, "Copy SIP to workspace", "")
	expectations["createTask"](s, params)

	s.env.OnActivity(
		activities.DownloadFromInternalBucketActivityName,
		sessionCtx,
		&bucketdownload.Params{Key: key},
	).Return(&bucketdownload.Result{FilePath: tempPath + "/" + key}, nil)

	params.updateTaskParams(copySIPTaskID, enums.TaskStatusDone, "", "SIP successfully copied")
	expectations["completeTask"](s, params)
	calcChecksumExpectations(s, params)
	checkDuplicateSIPExpectations(s, params)
	expectations["archiveExtract"](s, params)
	expectations["classifySIP"](s, params)
	countSIPFilesExpectations(s, params)
	expectations["saveFileCount"](s, params)
	autoApproveA3mExpectations(s, params)

	expectations["removePaths"](s, params)
	params.updateTaskParams(
		deleteSIPTaskID,
		enums.TaskStatusInProgress,
		"Delete original SIP",
		fmt.Sprintf("The original SIP will be deleted in %s", retentionPeriod),
	)
	expectations["createTask"](s, params)

	s.env.OnActivity(
		activities.DeleteOriginalFromInternalBucketActivityName,
		sessionCtx,
		&bucketdelete.Params{Key: key},
	).Return(&bucketdelete.Result{}, nil)

	params.updateTaskParams(deleteSIPTaskID, enums.TaskStatusDone, "", "SIP successfully deleted")
	expectations["completeTask"](s, params)
	expectations["updateSIPIngested"](s, params)
	expectations["completeWorkflow"](s, params)

	s.ExecuteAndValidateWorkflow(&ingest.ProcessingWorkflowRequest{
		Key:             key,
		Type:            enums.WorkflowTypeCreateAip,
		RetentionPeriod: retentionPeriod,
		SIPUUID:         sipUUID,
		SIPName:         sipName,
		Extension:       ".zip",
	}, &ingest.ProcessingWorkflowResult{}, false)
}

// TestInternalUploadError tests:
// - a3m as preservation system.
// - The "create AIP" workflow type.
// - Extraction error.
// - Move to failed SIP (internal bucket).
// - Internal bucket download (with preprocessing).
func (s *ProcessingWorkflowTestSuite) TestInternalUploadError() {
	s.SetupWorkflowTest(config.Configuration{
		A3m:          a3m.Config{ShareDir: s.CreateTransferDir()},
		Preservation: pres.Config{TaskQueue: temporal.A3mWorkerTaskQueue},
		Ingest:       ingest.Config{Storage: ingest.StorageConfig{DefaultPermanentLocationID: locationID}},
		ChildWorkflows: config.ChildWorkflowConfigs{
			{
				Type:         childwf.WorkflowTypePreprocessing,
				WorkflowName: "preprocessing",
				SharedPath:   prepSharedPath,
			},
		},
	}, nil)

	downloadDir := filepath.Join(prepSharedPath, sipUUID.String())
	params := defaultParams()
	params.downloadPath = filepath.Join(downloadDir, key)
	expectations["setStatusInProgress"](s, params)
	expectations["createWorkflow"](s, params)
	params.updateTaskParams(copySIPTaskID, enums.TaskStatusInProgress, "Copy SIP to workspace", "")
	expectations["createTask"](s, params)

	s.env.OnActivity(
		activities.DownloadFromInternalBucketActivityName,
		sessionCtx,
		&bucketdownload.Params{
			Key:     key,
			DirPath: downloadDir,
		},
	).Return(&bucketdownload.Result{FilePath: params.downloadPath}, nil)

	params.updateTaskParams(copySIPTaskID, enums.TaskStatusDone, "", "SIP successfully copied")
	expectations["completeTask"](s, params)
	calcChecksumExpectations(s, params)
	checkDuplicateSIPExpectations(s, params)

	// Fail the workflow on extraction.
	s.env.OnActivity(
		archiveextract.Name,
		sessionCtx,
		&archiveextract.Params{SourcePath: params.downloadPath},
	).Return(nil, errors.New("extract error"))
	s.env.OnActivity(
		bucketcopy.Name,
		sessionCtx,
		&bucketcopy.Params{
			SourceKey: key,
			DestKey:   failedSIPKey,
		},
	).Return(&bucketcopy.Result{}, nil)
	s.env.OnActivity(
		bucketdelete.Name,
		sessionCtx,
		&bucketdelete.Params{Key: key},
	).Return(&bucketdelete.Result{}, nil)

	params.removePaths = []string{downloadDir}
	params.sipStatus = enums.SIPStatusError
	params.failedAs = enums.SIPFailedAsSIP
	params.failedKey = failedSIPKey
	expectations["removePaths"](s, params)
	expectations["updateSIPFailed"](s, params)
	expectations["completeWorkflow"](s, params)

	s.ExecuteAndValidateWorkflow(&ingest.ProcessingWorkflowRequest{
		Key:       key,
		Type:      enums.WorkflowTypeCreateAip,
		SIPUUID:   sipUUID,
		SIPName:   sipName,
		Extension: ".zip",
	}, nil, true)
}

// TestSIPSourceUpload tests:
// - a3m as preservation system.
// - The "create AIP" workflow type.
// - SIP source bucket download.
// - SIP source retention period.
func (s *ProcessingWorkflowTestSuite) TestSIPSourceUpload() {
	s.SetupWorkflowTest(config.Configuration{
		A3m:          a3m.Config{ShareDir: s.CreateTransferDir()},
		Preservation: pres.Config{TaskQueue: temporal.A3mWorkerTaskQueue},
		Ingest:       ingest.Config{Storage: ingest.StorageConfig{DefaultPermanentLocationID: locationID}},
	}, nil)

	params := defaultParams()
	expectations["setStatusInProgress"](s, params)
	expectations["createWorkflow"](s, params)
	params.updateTaskParams(copySIPTaskID, enums.TaskStatusInProgress, "Copy SIP to workspace", "")
	expectations["createTask"](s, params)

	s.env.OnActivity(
		activities.DownloadFromSIPSourceActivityName,
		sessionCtx,
		&bucketdownload.Params{Key: key},
	).Return(&bucketdownload.Result{FilePath: tempPath + "/" + key}, nil)

	params.updateTaskParams(copySIPTaskID, enums.TaskStatusDone, "", "SIP successfully copied")
	expectations["completeTask"](s, params)
	calcChecksumExpectations(s, params)
	checkDuplicateSIPExpectations(s, params)
	expectations["getSIPExtension"](s, params)
	expectations["archiveExtract"](s, params)
	expectations["classifySIP"](s, params)
	countSIPFilesExpectations(s, params)
	expectations["saveFileCount"](s, params)
	autoApproveA3mExpectations(s, params)

	expectations["removePaths"](s, params)
	params.updateTaskParams(
		deleteSIPTaskID,
		enums.TaskStatusInProgress,
		"Delete original SIP",
		fmt.Sprintf("The original SIP will be deleted in %s", retentionPeriod),
	)
	expectations["createTask"](s, params)

	s.env.OnActivity(
		activities.DeleteOriginalFromSIPSourceActivityName,
		sessionCtx,
		&bucketdelete.Params{Key: key},
	).Return(&bucketdelete.Result{}, nil)

	params.updateTaskParams(deleteSIPTaskID, enums.TaskStatusDone, "", "SIP successfully deleted")
	expectations["completeTask"](s, params)
	expectations["updateSIPIngested"](s, params)
	expectations["completeWorkflow"](s, params)

	s.ExecuteAndValidateWorkflow(&ingest.ProcessingWorkflowRequest{
		Key:             key,
		Type:            enums.WorkflowTypeCreateAip,
		RetentionPeriod: retentionPeriod,
		SIPUUID:         sipUUID,
		SIPName:         sipName,
		SIPSourceID:     uuid.New(),
	}, &ingest.ProcessingWorkflowResult{}, false)
}

// TestSIPDeletionError tests:
// - a3m as preservation system.
// - The "create AIP" workflow type.
// - Watched bucket download.
// - Watched bucket original SIP deletion error.
func (s *ProcessingWorkflowTestSuite) TestSIPDeletionError() {
	s.SetupWorkflowTest(config.Configuration{
		A3m:          a3m.Config{ShareDir: s.CreateTransferDir()},
		Preservation: pres.Config{TaskQueue: temporal.A3mWorkerTaskQueue},
		Ingest:       ingest.Config{Storage: ingest.StorageConfig{DefaultPermanentLocationID: locationID}},
	}, nil)

	params := defaultParams()
	downloadExpectations(s, params)
	calcChecksumExpectations(s, params)
	checkDuplicateSIPExpectations(s, params)
	expectations["archiveExtract"](s, params)
	expectations["classifySIP"](s, params)
	countSIPFilesExpectations(s, params)
	expectations["saveFileCount"](s, params)
	autoApproveA3mExpectations(s, params)
	expectations["removePaths"](s, params)
	params.updateTaskParams(
		deleteSIPTaskID,
		enums.TaskStatusInProgress,
		"Delete original SIP",
		fmt.Sprintf("The original SIP will be deleted in %s", retentionPeriod),
	)
	expectations["createTask"](s, params)

	// Fail the deletion of the original SIP.
	s.env.OnActivity(
		activities.DeleteOriginalActivityName,
		sessionCtx,
		watcherName,
		key,
	).Return(nil, errors.New("deletion error"))

	params.updateTaskParams(
		deleteSIPTaskID,
		enums.TaskStatusError,
		"",
		"System error: Original SIP deletion has failed.\n\nAn error has occurred while attempting to delete the original SIP.",
	)
	expectations["completeTask"](s, params)
	expectations["updateSIPIngested"](s, params)
	expectations["completeWorkflow"](s, params)

	s.ExecuteAndValidateWorkflow(&ingest.ProcessingWorkflowRequest{
		Key:             key,
		WatcherName:     watcherName,
		RetentionPeriod: retentionPeriod,
		Type:            enums.WorkflowTypeCreateAip,
		SIPUUID:         sipUUID,
		SIPName:         sipName,
	}, &ingest.ProcessingWorkflowResult{}, false)
}

// TestBatchSignalDoNotContinue tests:
// - a3m as preservation system.
// - The "create AIP" workflow type.
// - Watched bucket download.
// - Batch signal handling (do not continue).
func (s *ProcessingWorkflowTestSuite) TestBatchSignalDoNotContinue() {
	s.SetupWorkflowTest(config.Configuration{
		A3m:          a3m.Config{ShareDir: s.CreateTransferDir()},
		Preservation: pres.Config{TaskQueue: temporal.A3mWorkerTaskQueue},
		Ingest:       ingest.Config{Storage: ingest.StorageConfig{DefaultPermanentLocationID: locationID}},
	}, nil)

	params := defaultParams()
	downloadExpectations(s, params)
	calcChecksumExpectations(s, params)
	checkDuplicateSIPExpectations(s, params)
	expectations["archiveExtract"](s, params)
	expectations["classifySIP"](s, params)
	countSIPFilesExpectations(s, params)
	expectations["saveFileCount"](s, params)
	expectations["bundle"](s, params)

	// Batch signal handler and expectations.
	s.env.RegisterDelayedCallback(func() {
		s.env.SignalWorkflow(ingest.BatchSignalName, ingest.BatchSignal{Continue: false})
	}, 0)
	params.sipStatus = enums.SIPStatusValidated
	expectations["setStatus"](s, params)
	params.updateTaskParams(
		batchWaitTaskID,
		enums.TaskStatusInProgress,
		"Waiting for other SIPs in Batch",
		"The SIP has been validated and is waiting for other SIPs in the Batch to be validated.",
	)
	expectations["createTask"](s, params)
	params.updateTaskParams(batchWaitTaskID, enums.TaskStatusDone, "", "Some SIPs in the Batch have failed validation.")
	expectations["completeTask"](s, params)

	// Cleanup expectations for canceled SIP.
	expectations["removePaths"](s, params)
	s.env.OnActivity(
		updateSIPLocalActivity,
		ctx,
		s.workflow.ingestsvc,
		mock.MatchedBy(func(updateParams *updateSIPLocalActivityParams) bool {
			return updateParams.UUID == sipUUID &&
				updateParams.Name == sipName &&
				updateParams.AIPUUID == "" &&
				updateParams.Status == enums.SIPStatusCanceled &&
				updateParams.FailedAs == "" &&
				updateParams.FailedKey == "" &&
				!updateParams.CompletedAt.IsZero()
		}),
	).Return(nil, nil)
	expectations["completeWorkflow"](s, params)

	s.ExecuteAndValidateWorkflow(&ingest.ProcessingWorkflowRequest{
		Key:             key,
		WatcherName:     watcherName,
		RetentionPeriod: params.retentionPeriod,
		Type:            enums.WorkflowTypeCreateAip,
		SIPUUID:         sipUUID,
		SIPName:         sipName,
		BatchUUID:       batchUUID,
	}, nil, true)
}

// TestDuplicateSIP tests:
// - Archivematica as preservation system.
// - The "create AIP" workflow type.
// - Content error due to SIP being a duplicate.
// - Move to failed SIP.
// - Watched bucket download.
func (s *ProcessingWorkflowTestSuite) TestDuplicateSIP() {
	duplicateSIPID := uuid.New()
	s.SetupWorkflowTest(config.Configuration{
		AM:           am.Config{ZipPIP: true},
		Preservation: pres.Config{TaskQueue: temporal.AmWorkerTaskQueue},
		Ingest:       ingest.Config{Storage: ingest.StorageConfig{DefaultPermanentLocationID: amssLocationID}},
	}, nil)

	params := defaultParams()
	downloadExpectations(s, params)
	calcChecksumExpectations(s, params)

	params.updateTaskParams(
		duplicateSIPTaskID,
		enums.TaskStatusInProgress,
		"Check if SIP has already been ingested",
		"",
	)
	expectations["createTask"](s, params)

	// Fail the workflow on duplicate SIP.
	s.env.OnActivity(
		activities.CheckDuplicateSIPActivityName,
		sessionCtx,
		activities.CheckDuplicateSIPActivityParams{
			SIPID: sipUUID,
			Checksum: datatypes.Checksum{
				Algorithm: datatypes.ChecksumAlgoSHA256,
				Hash:      "9f86d081884c7d659a2feaa0c55ad015a3bf4f1b2b0b822cd15d6c15b0f00a08",
			},
		},
	).Return(
		&activities.CheckDuplicateSIPActivityResult{
			Duplicate: &datatypes.SIP{
				UUID:   duplicateSIPID,
				Name:   "existing_sip.zip",
				Status: enums.SIPStatusIngested,
			},
		},
		nil,
	)

	params.updateTaskParams(
		duplicateSIPTaskID,
		enums.TaskStatusFailed,
		"",
		fmt.Sprintf(
			"Content error: SIP has already been ingested.\n\nA previously ingested SIP (UUID: %s) has the same checksum as the current SIP. Please ensure you have submitted the correct SIP and that it has not been previously submitted.",
			duplicateSIPID,
		),
	)
	expectations["completeTask"](s, params)

	params.sipStatus = enums.SIPStatusFailed
	params.failedAs = enums.SIPFailedAsSIP
	params.failedKey = failedSIPKey
	params.removePaths = []string{tempPath}
	expectations["uploadToFailed"](s, params)
	expectations["removePaths"](s, params)
	expectations["updateSIPFailed"](s, params)
	expectations["completeWorkflow"](s, params)

	s.ExecuteAndValidateWorkflow(&ingest.ProcessingWorkflowRequest{
		Key:         key,
		WatcherName: watcherName,
		Type:        enums.WorkflowTypeCreateAip,
		SIPUUID:     sipUUID,
		SIPName:     sipName,
	}, nil, true)
}

// TestAllowDuplicateSIP tests:
// - a3m as preservation system.
// - The "create AIP" workflow type.
// - Don't check for duplicate SIPs (config override).
// - Watched bucket download.
// - Watched bucket negative retention period.
func (s *ProcessingWorkflowTestSuite) TestAllowDuplicateSIP() {
	s.SetupWorkflowTest(config.Configuration{
		A3m:          a3m.Config{ShareDir: s.CreateTransferDir()},
		Preservation: pres.Config{TaskQueue: temporal.A3mWorkerTaskQueue},
		Ingest: ingest.Config{
			AllowDuplicates: true,
			Storage:         ingest.StorageConfig{DefaultPermanentLocationID: locationID},
		},
	}, nil)

	params := defaultParams()
	downloadExpectations(s, params)
	calcChecksumExpectations(s, params)
	expectations["archiveExtract"](s, params)
	expectations["classifySIP"](s, params)
	countSIPFilesExpectations(s, params)
	expectations["saveFileCount"](s, params)
	autoApproveA3mExpectations(s, params)
	params.retentionPeriod = -1 * time.Second
	cleanupExpectations(s, params)

	s.ExecuteAndValidateWorkflow(&ingest.ProcessingWorkflowRequest{
		Key:             key,
		WatcherName:     watcherName,
		RetentionPeriod: params.retentionPeriod,
		Type:            enums.WorkflowTypeCreateAip,
		SIPUUID:         sipUUID,
		SIPName:         sipName,
	}, &ingest.ProcessingWorkflowResult{}, false)
}

// TestCalculateSIPChecksumSysError tests:
// - Archivematica as preservation system.
// - The "create AIP" workflow type.
// - System error calculating checksum.
// - Move to failed SIP.
// - Watched bucket download.
func (s *ProcessingWorkflowTestSuite) TestCalculateSIPChecksumSysError() {
	s.SetupWorkflowTest(config.Configuration{
		AM:           am.Config{ZipPIP: true},
		Preservation: pres.Config{TaskQueue: temporal.AmWorkerTaskQueue},
		Ingest:       ingest.Config{Storage: ingest.StorageConfig{DefaultPermanentLocationID: amssLocationID}},
	}, nil)

	params := defaultParams()
	downloadExpectations(s, params)

	params.updateTaskParams(
		calcChecksumTaskID,
		enums.TaskStatusInProgress,
		"Calculate SIP checksum",
		"",
	)
	expectations["createTask"](s, params)

	// Return error from CalcFileChecksumActivity.
	s.env.OnActivity(
		activities.CalcFileChecksumActivityName,
		sessionCtx,
		&activities.CalcFileChecksumActivityParams{Path: params.downloadPath},
	).Return(
		nil,
		errors.New("checksum error"),
	)

	params.updateTaskParams(
		calcChecksumTaskID,
		enums.TaskStatusError,
		"",
		"System error: Calculating SIP checksum failed.\n\nAn error has occurred while calculating the SIP checksum. Please try again, or ask a system administrator to investigate.",
	)
	expectations["completeTask"](s, params)

	params.sipStatus = enums.SIPStatusError
	params.failedAs = enums.SIPFailedAsSIP
	params.failedKey = failedSIPKey
	params.removePaths = []string{tempPath}
	expectations["uploadToFailed"](s, params)
	expectations["removePaths"](s, params)
	expectations["updateSIPFailed"](s, params)
	expectations["completeWorkflow"](s, params)

	s.ExecuteAndValidateWorkflow(&ingest.ProcessingWorkflowRequest{
		Key:         key,
		WatcherName: watcherName,
		Type:        enums.WorkflowTypeCreateAip,
		SIPUUID:     sipUUID,
		SIPName:     sipName,
	}, nil, true)
}
