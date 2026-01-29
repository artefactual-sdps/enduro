package workflow

import (
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

	"github.com/artefactual-sdps/enduro/internal/a3m"
	"github.com/artefactual-sdps/enduro/internal/am"
	"github.com/artefactual-sdps/enduro/internal/childwf"
	"github.com/artefactual-sdps/enduro/internal/config"
	"github.com/artefactual-sdps/enduro/internal/enums"
	"github.com/artefactual-sdps/enduro/internal/ingest"
	"github.com/artefactual-sdps/enduro/internal/premis"
	"github.com/artefactual-sdps/enduro/internal/pres"
	"github.com/artefactual-sdps/enduro/internal/storage"
	"github.com/artefactual-sdps/enduro/internal/temporal"
	"github.com/artefactual-sdps/enduro/internal/workflow/activities"
	"github.com/artefactual-sdps/enduro/internal/workflow/localact"
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
		Storage:      storage.Config{DefaultPermanentLocationID: locationID},
	})

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
	expectations["archiveExtract"](s, params)
	expectations["classifySIP"](s, params)
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
	}, false)
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
		Storage:      storage.Config{DefaultPermanentLocationID: locationID},
	})

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
	expectations["archiveExtract"](s, params)
	expectations["classifySIP"](s, params)
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
	}, false)
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
		Storage:      storage.Config{DefaultPermanentLocationID: locationID},
	})

	params := defaultParams()
	downloadExpectations(s, params)
	expectations["archiveExtract"](s, params)
	expectations["classifySIP"](s, params)
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
	}, false)
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
		Storage:        storage.Config{DefaultPermanentLocationID: amssLocationID},
		ValidatePREMIS: premis.Config{Enabled: true, XSDPath: "/home/enduro/premis.xsd"},
	})

	params := defaultParams()
	downloadExpectations(s, params)
	expectations["archiveExtract"](s, params)
	expectations["classifySIP"](s, params)
	expectations["createBag"](s, params)
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
	}, false)
}

// TestChildWorkflows tests:
// - a3m as preservation system.
// - The "create AIP" workflow type.
// - preprocessing child workflow.
// - poststorage child workflows.
// - Bag validation.
// - Watched bucket download.
// - Watched bucket custom retention period.
func (s *ProcessingWorkflowTestSuite) TestChildWorkflows() {
	s.SetupWorkflowTest(config.Configuration{
		A3m:          a3m.Config{ShareDir: s.CreateTransferDir()},
		Preservation: pres.Config{TaskQueue: temporal.A3mWorkerTaskQueue},
		Storage:      storage.Config{DefaultPermanentLocationID: locationID},
		ChildWorkflows: childwf.Configs{
			{
				Type:         enums.ChildWorkflowTypePreprocessing,
				Namespace:    "default",
				TaskQueue:    "preprocessing",
				WorkflowName: "preprocessing",
				Extract:      true,
				SharedPath:   prepSharedPath,
			},
			{
				Type:         enums.ChildWorkflowTypePoststorage,
				Namespace:    "default",
				TaskQueue:    "poststorage",
				WorkflowName: "poststorage",
			},
		},
	})

	params := defaultParams()
	params.downloadDestPath = prepSharedPath
	params.downloadPath = prepDownloadPath + "/" + key
	params.extractPath = prepExtractPath
	downloadExpectations(s, params)

	s.env.OnWorkflow(
		preprocessingChildWorkflow,
		internalCtx,
		&childwf.PreprocessingParams{
			RelativePath: strings.TrimPrefix(prepDownloadPath+"/"+key, prepSharedPath),
		},
	).Return(
		&childwf.PreprocessingResult{
			Outcome:      childwf.OutcomeSuccess,
			RelativePath: strings.TrimPrefix(prepExtractPath, prepSharedPath),
			PreservationTasks: []childwf.Task{
				{
					Name:        "Identify SIP structure",
					Message:     "SIP structure identified: VecteurAIP",
					Outcome:     enums.PreprocessingTaskOutcomeSuccess,
					StartedAt:   time.Date(2024, 6, 14, 10, 5, 32, 0, time.UTC),
					CompletedAt: time.Date(2024, 6, 14, 10, 5, 33, 0, time.UTC),
				},
			},
		},
		nil,
	)

	s.env.OnActivity(
		localact.SavePreprocessingTasksActivity,
		ctx,
		localact.SavePreprocessingTasksActivityParams{
			Ingestsvc:    s.workflow.ingestsvc,
			RNG:          s.workflow.rng,
			WorkflowUUID: workflowUUID,
			Tasks: []childwf.Task{
				{
					Name:        "Identify SIP structure",
					Message:     "SIP structure identified: VecteurAIP",
					Outcome:     enums.PreprocessingTaskOutcomeSuccess,
					StartedAt:   time.Date(2024, 6, 14, 10, 5, 32, 0, time.UTC),
					CompletedAt: time.Date(2024, 6, 14, 10, 5, 33, 0, time.UTC),
				},
			},
		},
	).Return(&localact.SavePreprocessingTasksActivityResult{Count: 1}, nil)

	params.sipType = enums.SIPTypeBagIt
	expectations["classifySIP"](s, params)
	params.updateTaskParams(valBagTaskID, enums.TaskStatusInProgress, "Validate Bag", "")
	expectations["createTask"](s, params)
	expectations["validateBag"](s, params)
	params.updateTaskParams(valBagTaskID, enums.TaskStatusDone, "", "Bag successfully validated")
	expectations["completeTask"](s, params)
	autoApproveA3mExpectations(s, params)

	s.env.OnWorkflow(
		"poststorage",
		internalCtx,
		&childwf.PostStorageParams{AIPUUID: aipUUID.String()},
	).Return(nil, nil)

	params.removePaths = []string{prepDownloadPath, transferPath}
	params.retentionPeriod = 48 * time.Hour
	cleanupExpectations(s, params)

	s.ExecuteAndValidateWorkflow(&ingest.ProcessingWorkflowRequest{
		Key:             key,
		WatcherName:     watcherName,
		RetentionPeriod: params.retentionPeriod,
		Type:            enums.WorkflowTypeCreateAip,
		SIPUUID:         sipUUID,
		SIPName:         sipName,
	}, false)
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

		Storage: storage.Config{DefaultPermanentLocationID: locationID},
		ChildWorkflows: childwf.Configs{
			{
				Type:         enums.ChildWorkflowTypePreprocessing,
				Namespace:    "default",
				TaskQueue:    "preprocessing",
				WorkflowName: "preprocessing",
				Extract:      true,
				SharedPath:   prepSharedPath,
			},
		},
	})

	params := defaultParams()
	params.downloadDestPath = prepSharedPath
	params.downloadPath = prepDownloadPath + "/" + key
	downloadExpectations(s, params)

	// Fail the workflow on preprocessing.
	s.env.OnWorkflow(
		preprocessingChildWorkflow,
		internalCtx,
		&childwf.PreprocessingParams{
			RelativePath: strings.TrimPrefix(prepDownloadPath+"/"+key, prepSharedPath),
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
	}, true)
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
		Storage:        storage.Config{DefaultPermanentLocationID: locationID},
		ValidatePREMIS: premis.Config{Enabled: true, XSDPath: "/home/enduro/premis.xsd"},
	})

	params := defaultParams()
	downloadExpectations(s, params)
	expectations["archiveExtract"](s, params)
	params.sipType = enums.SIPTypeBagIt
	expectations["classifySIP"](s, params)
	params.updateTaskParams(valBagTaskID, enums.TaskStatusInProgress, "Validate Bag", "")
	expectations["createTask"](s, params)
	expectations["validateBag"](s, params)
	params.updateTaskParams(valBagTaskID, enums.TaskStatusDone, "", "Bag successfully validated")
	expectations["completeTask"](s, params)
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
	}, true)
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
		Storage:      storage.Config{DefaultPermanentLocationID: amssLocationID},
	})

	params := defaultParams()
	downloadExpectations(s, params)
	expectations["archiveExtract"](s, params)
	expectations["classifySIP"](s, params)
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
	}, true)
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
		Storage:      storage.Config{DefaultPermanentLocationID: locationID},
	})

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
	expectations["archiveExtract"](s, params)
	expectations["classifySIP"](s, params)
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
	}, false)
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
		Storage:      storage.Config{DefaultPermanentLocationID: locationID},
		ChildWorkflows: childwf.Configs{
			{
				Type:         enums.ChildWorkflowTypePreprocessing,
				WorkflowName: "preprocessing",
				SharedPath:   prepSharedPath,
			},
		},
	})

	downloadDir := filepath.Join(prepSharedPath, sipUUID.String())
	params := defaultParams()
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
	).Return(&bucketdownload.Result{FilePath: downloadDir + "/" + key}, nil)

	params.updateTaskParams(copySIPTaskID, enums.TaskStatusDone, "", "SIP successfully copied")
	expectations["completeTask"](s, params)

	// Fail the workflow on extraction.
	s.env.OnActivity(
		archiveextract.Name,
		sessionCtx,
		&archiveextract.Params{SourcePath: downloadDir + "/" + key},
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
	}, true)
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
		Storage:      storage.Config{DefaultPermanentLocationID: locationID},
	})

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
	expectations["getSIPExtension"](s, params)
	expectations["archiveExtract"](s, params)
	expectations["classifySIP"](s, params)
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
	}, false)
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
		Storage:      storage.Config{DefaultPermanentLocationID: locationID},
	})

	params := defaultParams()
	downloadExpectations(s, params)
	expectations["archiveExtract"](s, params)
	expectations["classifySIP"](s, params)
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
	}, false)
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
		Storage:      storage.Config{DefaultPermanentLocationID: locationID},
	})

	params := defaultParams()
	downloadExpectations(s, params)
	expectations["archiveExtract"](s, params)
	expectations["classifySIP"](s, params)
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
	}, true)
}
