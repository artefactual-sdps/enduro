package workflow

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/artefactual-sdps/temporal-activities/archiveextract"
	"github.com/artefactual-sdps/temporal-activities/bagvalidate"
	"github.com/artefactual-sdps/temporal-activities/bucketdownload"
	"github.com/artefactual-sdps/temporal-activities/bucketupload"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"go.artefactual.dev/tools/ref"

	"github.com/artefactual-sdps/enduro/internal/a3m"
	"github.com/artefactual-sdps/enduro/internal/am"
	"github.com/artefactual-sdps/enduro/internal/config"
	"github.com/artefactual-sdps/enduro/internal/datatypes"
	"github.com/artefactual-sdps/enduro/internal/enums"
	"github.com/artefactual-sdps/enduro/internal/ingest"
	"github.com/artefactual-sdps/enduro/internal/poststorage"
	"github.com/artefactual-sdps/enduro/internal/premis"
	"github.com/artefactual-sdps/enduro/internal/preprocessing"
	"github.com/artefactual-sdps/enduro/internal/pres"
	"github.com/artefactual-sdps/enduro/internal/storage"
	"github.com/artefactual-sdps/enduro/internal/temporal"
	"github.com/artefactual-sdps/enduro/internal/workflow/activities"
	"github.com/artefactual-sdps/enduro/internal/workflow/localact"
)

func TestProcessingWorkflow(t *testing.T) {
	suite.Run(t, new(ProcessingWorkflowTestSuite))
}

func (s *ProcessingWorkflowTestSuite) TestConfirmation() {
	s.SetupWorkflowTest(config.Configuration{
		A3m:          a3m.Config{ShareDir: s.CreateTransferDir()},
		Preservation: pres.Config{TaskQueue: temporal.A3mWorkerTaskQueue},
		Storage: storage.Config{
			DefaultPermanentLocationID: locationID,
		},
		ValidatePREMIS: premis.Config{
			Enabled: true,
			XSDPath: "/home/enduro/premis.xsd",
		},
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

	expectations["createSIP"](s, expectationParams{})
	expectations["setStatusInProgress"](s, expectationParams{})
	expectations["createWorkflow"](s, expectationParams{
		WorkflowType: enums.WorkflowTypeCreateAndReviewAip,
	})
	expectations["createTask"](s, expectationParams{
		TaskID:   copySIPTaskID,
		TaskName: "Copy SIP to workspace",
	})
	expectations["download"](s, expectationParams{})
	expectations["completeTask"](s, expectationParams{
		TaskID:   copySIPTaskID,
		TaskNote: "SIP successfully copied",
	})
	expectations["getSIPExtension"](s, expectationParams{})
	expectations["archiveExtract"](s, expectationParams{})
	expectations["bundleActivity"](s, expectationParams{})
	expectations["createTask"](s, expectationParams{
		TaskID:   valPREMISTaskID,
		TaskName: "Validate PREMIS",
	})
	expectations["validatePREMIS"](s, expectationParams{})
	expectations["completeTask"](s, expectationParams{
		TaskID:   valPREMISTaskID,
		TaskNote: "PREMIS is valid",
	})
	expectations["createAIP"](s, expectationParams{})
	expectations["updateSIPProcessing"](s, expectationParams{})

	s.env.OnActivity(
		createTaskLocalActivity,
		mock.Anything,
		mock.Anything,
		mock.Anything,
	).Return(0, nil)
	s.env.OnActivity(
		activities.UploadActivityName,
		mock.Anything,
		mock.Anything,
	).Return(nil, nil)
	s.env.OnActivity(
		setStatusLocalActivity,
		mock.Anything,
		mock.Anything,
		mock.Anything,
		mock.Anything,
	).Return(nil, nil)
	s.env.OnActivity(
		setWorkflowStatusLocalActivity,
		mock.Anything,
		mock.Anything,
		mock.Anything,
		mock.Anything,
	).Return(nil, nil)
	s.env.OnActivity(
		completeTaskLocalActivity,
		mock.Anything,
		mock.Anything,
		mock.Anything,
	).Return(nil, nil)
	s.env.OnActivity(
		activities.MoveToPermanentStorageActivityName,
		mock.Anything,
		mock.Anything,
	).Return(nil, nil)
	s.env.OnActivity(
		activities.PollMoveToPermanentStorageActivityName,
		mock.Anything,
		mock.Anything,
	).Return(nil, nil)

	expectations["completeWorkflow"](s, expectationParams{})
	expectations["removePaths"](s, expectationParams{})
	expectations["deleteOriginal"](s, expectationParams{})
	expectations["updateSIPIngested"](s, expectationParams{})

	s.ExecuteAndValidateWorkflow(&ingest.ProcessingWorkflowRequest{
		Key:             key,
		WatcherName:     watcherName,
		RetentionPeriod: retentionPeriod,
		Type:            enums.WorkflowTypeCreateAndReviewAip,
		SIPUUID:         sipUUID,
		SIPName:         sipName,
	}, false)
}

func (s *ProcessingWorkflowTestSuite) TestRejection() {
	s.SetupWorkflowTest(config.Configuration{
		A3m:          a3m.Config{ShareDir: s.CreateTransferDir()},
		Preservation: pres.Config{TaskQueue: temporal.A3mWorkerTaskQueue},
		Storage: storage.Config{
			DefaultPermanentLocationID: locationID,
		},
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

	expectations["createSIP"](s, expectationParams{})
	expectations["setStatusInProgress"](s, expectationParams{})
	expectations["createWorkflow"](s, expectationParams{
		WorkflowType: enums.WorkflowTypeCreateAndReviewAip,
	})
	expectations["createTask"](s, expectationParams{
		TaskID:   copySIPTaskID,
		TaskName: "Copy SIP to workspace",
	})
	expectations["download"](s, expectationParams{})
	expectations["completeTask"](s, expectationParams{
		TaskID:   copySIPTaskID,
		TaskNote: "SIP successfully copied",
	})
	expectations["getSIPExtension"](s, expectationParams{})
	expectations["archiveExtract"](s, expectationParams{})
	expectations["classifySIP"](s, expectationParams{})
	expectations["bundleActivity"](s, expectationParams{})

	s.env.OnActivity(
		a3m.CreateAIPActivityName,
		mock.Anything,
		mock.Anything,
	).Return(&a3m.CreateAIPActivityResult{UUID: aipUUID.String()}, nil)

	expectations["updateSIPProcessing"](s, expectationParams{})

	s.env.OnActivity(
		completeTaskLocalActivity,
		mock.Anything,
		mock.Anything,
		mock.Anything,
	).Return(nil, nil)
	s.env.OnActivity(
		activities.UploadActivityName,
		mock.Anything,
		mock.Anything,
	).Return(nil, nil)
	s.env.OnActivity(
		setStatusLocalActivity,
		mock.Anything,
		mock.Anything,
		mock.Anything,
		mock.Anything,
	).Return(nil, nil)
	s.env.OnActivity(
		setWorkflowStatusLocalActivity,
		mock.Anything,
		mock.Anything,
		mock.Anything,
		mock.Anything,
	).Return(nil, nil)
	s.env.OnActivity(
		createTaskLocalActivity,
		mock.Anything,
		mock.Anything,
		mock.Anything,
	).Return(0, nil)
	s.env.OnActivity(
		activities.RejectSIPActivityName,
		mock.Anything,
		mock.Anything,
		mock.Anything,
		mock.Anything,
	).Return(nil, nil)

	expectations["completeWorkflow"](s, expectationParams{})
	expectations["removePaths"](s, expectationParams{})
	expectations["deleteOriginal"](s, expectationParams{})
	expectations["updateSIPIngested"](s, expectationParams{})

	s.ExecuteAndValidateWorkflow(&ingest.ProcessingWorkflowRequest{
		Key:             key,
		WatcherName:     watcherName,
		RetentionPeriod: retentionPeriod,
		Type:            enums.WorkflowTypeCreateAndReviewAip,
		SIPUUID:         sipUUID,
		SIPName:         sipName,
	}, false)
}

func (s *ProcessingWorkflowTestSuite) TestAutoApprovedAIP() {
	s.SetupWorkflowTest(config.Configuration{
		A3m:          a3m.Config{ShareDir: s.CreateTransferDir()},
		Preservation: pres.Config{TaskQueue: temporal.A3mWorkerTaskQueue},
		Storage: storage.Config{
			DefaultPermanentLocationID: locationID,
		},
	})

	expectations["createSIP"](s, expectationParams{})
	expectations["setStatusInProgress"](s, expectationParams{})
	expectations["createWorkflow"](s, expectationParams{})
	expectations["createTask"](s, expectationParams{
		TaskID:   copySIPTaskID,
		TaskName: "Copy SIP to workspace",
	})
	expectations["download"](s, expectationParams{})
	expectations["completeTask"](s, expectationParams{
		TaskID:   copySIPTaskID,
		TaskNote: "SIP successfully copied",
	})
	expectations["getSIPExtension"](s, expectationParams{})
	expectations["archiveExtract"](s, expectationParams{})
	expectations["classifySIP"](s, expectationParams{SIPType: enums.SIPTypeBagIt})
	expectations["createTask"](s, expectationParams{
		TaskID:   valBagTaskID,
		TaskName: "Validate Bag",
	})
	expectations["validateBag"](s, expectationParams{})
	expectations["completeTask"](s, expectationParams{
		TaskID:   valBagTaskID,
		TaskNote: "Bag successfully validated",
	})
	expectations["bundleActivity"](s, expectationParams{})

	s.env.OnActivity(
		a3m.CreateAIPActivityName,
		sessionCtx,
		mock.AnythingOfType("*a3m.CreateAIPActivityParams"),
	).Return(&a3m.CreateAIPActivityResult{UUID: aipUUID.String()}, nil)

	expectations["updateSIPProcessing"](s, expectationParams{})
	expectations["createTask"](s, expectationParams{
		TaskID:   moveAIPTaskID,
		TaskName: "Move AIP",
		TaskNote: "Moving to permanent storage",
	})

	s.env.OnActivity(
		activities.UploadActivityName,
		sessionCtx,
		mock.AnythingOfType("*activities.UploadActivityParams"),
	).Return(nil, nil)

	expectations["completeTask"](s, expectationParams{
		TaskID:   moveAIPTaskID,
		TaskNote: "Moved to location f2cc963f-c14d-4eaa-b950-bd207189a1f1",
	})

	s.env.OnActivity(
		activities.MoveToPermanentStorageActivityName,
		sessionCtx,
		mock.AnythingOfType("*activities.MoveToPermanentStorageActivityParams"),
	).Return(nil, nil)
	s.env.OnActivity(
		activities.PollMoveToPermanentStorageActivityName,
		sessionCtx,
		mock.AnythingOfType("*activities.PollMoveToPermanentStorageActivityParams"),
	).Return(nil, nil)

	expectations["updateSIPIngested"](s, expectationParams{})
	expectations["completeWorkflow"](s, expectationParams{})
	expectations["removePaths"](s, expectationParams{})
	expectations["deleteOriginal"](s, expectationParams{})

	s.ExecuteAndValidateWorkflow(&ingest.ProcessingWorkflowRequest{
		Key:             key,
		WatcherName:     watcherName,
		RetentionPeriod: retentionPeriod,
		Type:            enums.WorkflowTypeCreateAip,
		SIPUUID:         sipUUID,
		SIPName:         sipName,
	}, false)
}

func (s *ProcessingWorkflowTestSuite) TestAMWorkflow() {
	s.SetupWorkflowTest(config.Configuration{
		A3m: a3m.Config{ShareDir: s.CreateTransferDir()},
		AM: am.Config{
			ZipPIP:           true,
			TransferDeadline: time.Second,
		},
		Preservation: pres.Config{TaskQueue: temporal.AmWorkerTaskQueue},
		Storage:      storage.Config{DefaultPermanentLocationID: amssLocationID},
		ValidatePREMIS: premis.Config{
			Enabled: true,
			XSDPath: "/home/enduro/premis.xsd",
		},
	})

	expectations["createSIP"](s, expectationParams{})
	expectations["setStatusInProgress"](s, expectationParams{})
	expectations["createWorkflow"](s, expectationParams{})
	expectations["createTask"](s, expectationParams{
		TaskID:   copySIPTaskID,
		TaskName: "Copy SIP to workspace",
	})
	expectations["download"](s, expectationParams{})
	expectations["completeTask"](s, expectationParams{
		TaskID:   copySIPTaskID,
		TaskNote: "SIP successfully copied",
	})
	expectations["getSIPExtension"](s, expectationParams{})
	expectations["archiveExtract"](s, expectationParams{})
	expectations["classifySIP"](s, expectationParams{})
	expectations["createBag"](s, expectationParams{})
	expectations["createTask"](s, expectationParams{
		TaskID:   valPREMISTaskID,
		TaskName: "Validate PREMIS",
	})
	expectations["validatePREMIS"](s, expectationParams{
		PremisXMLPath: filepath.Join(extractPath, "data", "metadata", "premis.xml"),
	})
	expectations["completeTask"](s, expectationParams{
		TaskID:   valPREMISTaskID,
		TaskNote: "PREMIS is valid",
	})
	expectations["zipArchive"](s, expectationParams{})
	expectations["uploadTransferAM"](s, expectationParams{})
	expectations["startTransferAM"](s, expectationParams{})
	expectations["pollTransferAM"](s, expectationParams{})
	expectations["pollIngestAM"](s, expectationParams{})
	expectations["createStorageAIP"](s, expectationParams{})
	expectations["deleteTransferAM"](s, expectationParams{})
	expectations["updateSIPProcessing"](s, expectationParams{})
	expectations["updateSIPIngested"](s, expectationParams{})
	expectations["completeWorkflow"](s, expectationParams{})
	expectations["removePaths"](s, expectationParams{
		RemovePaths: []string{tempPath, extractPath + "/transfer.zip"},
	})
	expectations["deleteOriginal"](s, expectationParams{})

	s.ExecuteAndValidateWorkflow(&ingest.ProcessingWorkflowRequest{
		WatcherName:     watcherName,
		RetentionPeriod: retentionPeriod,
		Type:            enums.WorkflowTypeCreateAip,
		Key:             key,
		SIPUUID:         sipUUID,
		SIPName:         sipName,
	}, false)
}

func (s *ProcessingWorkflowTestSuite) TestChildWorkflows() {
	s.SetupWorkflowTest(config.Configuration{
		A3m:          a3m.Config{ShareDir: s.CreateTransferDir()},
		Preservation: pres.Config{TaskQueue: temporal.A3mWorkerTaskQueue},
		Preprocessing: preprocessing.Config{
			Enabled:    true,
			Extract:    true,
			SharedPath: prepSharedPath,
			Temporal: preprocessing.Temporal{
				Namespace:    "default",
				TaskQueue:    "preprocessing",
				WorkflowName: "preprocessing",
			},
		},
		Poststorage: []poststorage.Config{
			{
				Namespace:    "default",
				TaskQueue:    "poststorage",
				WorkflowName: "poststorage_1",
			},
			{
				Namespace:    "default",
				TaskQueue:    "poststorage",
				WorkflowName: "poststorage_2",
			},
		},
		Storage: storage.Config{
			DefaultPermanentLocationID: locationID,
		},
	})

	expectations["createSIP"](s, expectationParams{})
	expectations["setStatusInProgress"](s, expectationParams{})
	expectations["createWorkflow"](s, expectationParams{})
	expectations["createTask"](s, expectationParams{
		TaskID:   copySIPTaskID,
		TaskName: "Copy SIP to workspace",
	})

	s.env.OnActivity(
		activities.DownloadActivityName,
		sessionCtx,
		&activities.DownloadActivityParams{
			Key:             key,
			WatcherName:     watcherName,
			DestinationPath: prepSharedPath,
		},
	).Return(&activities.DownloadActivityResult{Path: prepDownloadPath + "/" + key}, nil)

	expectations["completeTask"](s, expectationParams{
		TaskID:   copySIPTaskID,
		TaskNote: "SIP successfully copied",
	})

	s.env.OnActivity(
		activities.GetSIPExtensionActivityName,
		sessionCtx,
		&activities.GetSIPExtensionActivityParams{Path: prepDownloadPath + "/" + key},
	).Return(&activities.GetSIPExtensionActivityResult{Extension: ".zip"}, nil)

	s.env.OnWorkflow(
		preprocessingChildWorkflow,
		mock.AnythingOfType("*internal.valueCtx"),
		&preprocessing.WorkflowParams{
			RelativePath: strings.TrimPrefix(prepDownloadPath+"/"+key, prepSharedPath),
		},
	).Return(
		&preprocessing.WorkflowResult{
			Outcome:      preprocessing.OutcomeSuccess,
			RelativePath: strings.TrimPrefix(prepExtractPath, prepSharedPath),
			PreservationTasks: []preprocessing.Task{
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
			WorkflowUUID: wUUID,
			Tasks: []preprocessing.Task{
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

	s.env.OnActivity(
		activities.ClassifySIPActivityName,
		sessionCtx,
		activities.ClassifySIPActivityParams{Path: prepExtractPath},
	).Return(&activities.ClassifySIPActivityResult{Type: enums.SIPTypeBagIt}, nil)

	s.env.OnActivity(
		createTaskLocalActivity,
		ctx,
		&createTaskLocalActivityParams{
			Ingestsvc: s.workflow.ingestsvc,
			RNG:       s.workflow.rng,
			Task: &datatypes.Task{
				Name:         "Validate Bag",
				Status:       enums.TaskStatusInProgress,
				WorkflowUUID: wUUID,
			},
		},
	).Return(valBagTaskID, nil)

	s.env.OnActivity(
		bagvalidate.Name,
		sessionCtx,
		&bagvalidate.Params{Path: prepExtractPath},
	).Return(&bagvalidate.Result{Valid: true}, nil)

	s.env.OnActivity(
		completeTaskLocalActivity,
		ctx,
		s.workflow.ingestsvc,
		&completeTaskLocalActivityParams{
			ID:          valBagTaskID,
			Status:      enums.TaskStatusDone,
			CompletedAt: startTime,
			Note:        ref.New("Bag successfully validated"),
		},
	).Return(&completeTaskLocalActivityResult{}, nil)

	expectations["bundleActivity"](s, expectationParams{BundlePath: prepExtractPath})

	s.env.OnActivity(
		a3m.CreateAIPActivityName,
		sessionCtx,
		mock.AnythingOfType("*a3m.CreateAIPActivityParams"),
	).Return(&a3m.CreateAIPActivityResult{UUID: aipUUID.String()}, nil)

	expectations["updateSIPProcessing"](s, expectationParams{})
	expectations["createTask"](s, expectationParams{
		TaskID:   moveAIPTaskID,
		TaskName: "Move AIP",
		TaskNote: "Moving to permanent storage",
	})

	s.env.OnActivity(
		activities.UploadActivityName,
		sessionCtx,
		mock.AnythingOfType("*activities.UploadActivityParams"),
	).Return(nil, nil)

	expectations["completeTask"](s, expectationParams{
		TaskID:   moveAIPTaskID,
		TaskNote: "Moved to location f2cc963f-c14d-4eaa-b950-bd207189a1f1",
	})

	s.env.OnActivity(
		activities.UploadActivityName,
		sessionCtx,
		mock.AnythingOfType("*activities.UploadActivityParams"),
	).Return(nil, nil)

	s.env.OnActivity(
		activities.MoveToPermanentStorageActivityName,
		sessionCtx,
		mock.AnythingOfType("*activities.MoveToPermanentStorageActivityParams"),
	).Return(nil, nil)

	s.env.OnActivity(
		activities.PollMoveToPermanentStorageActivityName,
		sessionCtx,
		mock.AnythingOfType("*activities.PollMoveToPermanentStorageActivityParams"),
	).Return(nil, nil)

	s.env.OnWorkflow(
		"poststorage_1",
		mock.AnythingOfType("*internal.valueCtx"),
		&poststorage.WorkflowParams{AIPUUID: aipUUID.String()},
	).Return(nil, nil)

	s.env.OnWorkflow(
		"poststorage_2",
		mock.AnythingOfType("*internal.valueCtx"),
		&poststorage.WorkflowParams{AIPUUID: aipUUID.String()},
	).Return(nil, nil)

	expectations["completeWorkflow"](s, expectationParams{})
	expectations["removePaths"](s, expectationParams{
		RemovePaths: []string{prepDownloadPath, transferPath},
	})
	expectations["deleteOriginal"](s, expectationParams{})
	expectations["updateSIPIngested"](s, expectationParams{})

	s.ExecuteAndValidateWorkflow(&ingest.ProcessingWorkflowRequest{
		Key:             key,
		WatcherName:     watcherName,
		RetentionPeriod: retentionPeriod,
		Type:            enums.WorkflowTypeCreateAip,
		SIPUUID:         sipUUID,
		SIPName:         sipName,
	}, false)
}

func (s *ProcessingWorkflowTestSuite) TestFailedSIP() {
	s.SetupWorkflowTest(config.Configuration{
		A3m:          a3m.Config{ShareDir: s.CreateTransferDir()},
		Preservation: pres.Config{TaskQueue: temporal.A3mWorkerTaskQueue},
		Preprocessing: preprocessing.Config{
			Enabled:    true,
			Extract:    true,
			SharedPath: prepSharedPath,
			Temporal: preprocessing.Temporal{
				Namespace:    "default",
				TaskQueue:    "preprocessing",
				WorkflowName: "preprocessing",
			},
		},
		Storage: storage.Config{DefaultPermanentLocationID: locationID},
	})

	expectations["createSIP"](s, expectationParams{})
	expectations["setStatusInProgress"](s, expectationParams{})
	expectations["createWorkflow"](s, expectationParams{})
	expectations["createTask"](s, expectationParams{
		TaskID:   copySIPTaskID,
		TaskName: "Copy SIP to workspace",
	})

	s.env.OnActivity(
		activities.DownloadActivityName,
		sessionCtx,
		&activities.DownloadActivityParams{
			Key:             key,
			WatcherName:     watcherName,
			DestinationPath: prepSharedPath,
		},
	).Return(&activities.DownloadActivityResult{Path: prepDownloadPath + "/" + key}, nil)

	expectations["completeTask"](s, expectationParams{
		TaskID:   copySIPTaskID,
		TaskNote: "SIP successfully copied",
	})

	s.env.OnActivity(
		activities.GetSIPExtensionActivityName,
		sessionCtx,
		&activities.GetSIPExtensionActivityParams{Path: prepDownloadPath + "/" + key},
	).Return(&activities.GetSIPExtensionActivityResult{Extension: ".zip"}, nil)

	s.env.OnWorkflow(
		preprocessingChildWorkflow,
		mock.AnythingOfType("*internal.valueCtx"),
		&preprocessing.WorkflowParams{
			RelativePath: strings.TrimPrefix(prepDownloadPath+"/"+key, prepSharedPath),
		},
	).Return(
		&preprocessing.WorkflowResult{
			Outcome:      preprocessing.OutcomeContentError,
			RelativePath: strings.TrimPrefix(prepExtractPath, prepSharedPath),
		},
		nil,
	)

	s.env.OnActivity(
		bucketupload.Name,
		sessionCtx,
		&bucketupload.Params{
			Path:       prepDownloadPath + "/" + key,
			Key:        failedSIPKey,
			BufferSize: 100_000_000,
		},
	).Return(&bucketupload.Result{}, nil)

	expectations["updateSIPFailed"](s, expectationParams{
		Status:    enums.SIPStatusFailed,
		FailedAs:  enums.SIPFailedAsSIP,
		FailedKey: failedSIPKey,
	})
	expectations["completeWorkflow"](s, expectationParams{})
	expectations["removePaths"](s, expectationParams{
		RemovePaths: []string{prepDownloadPath},
	})

	s.ExecuteAndValidateWorkflow(&ingest.ProcessingWorkflowRequest{
		Key:             key,
		WatcherName:     watcherName,
		RetentionPeriod: retentionPeriod,
		Type:            enums.WorkflowTypeCreateAip,
		SIPUUID:         sipUUID,
		SIPName:         sipName,
	}, true)
}

func (s *ProcessingWorkflowTestSuite) TestFailedPIPA3m() {
	s.SetupWorkflowTest(config.Configuration{
		A3m:          a3m.Config{ShareDir: s.CreateTransferDir()},
		Preservation: pres.Config{TaskQueue: temporal.A3mWorkerTaskQueue},
		Storage:      storage.Config{DefaultPermanentLocationID: locationID},
	})

	expectations["createSIP"](s, expectationParams{})
	expectations["setStatusInProgress"](s, expectationParams{})
	expectations["createWorkflow"](s, expectationParams{})
	expectations["createTask"](s, expectationParams{
		TaskID:   copySIPTaskID,
		TaskName: "Copy SIP to workspace",
	})
	expectations["download"](s, expectationParams{})
	expectations["completeTask"](s, expectationParams{
		TaskID:   copySIPTaskID,
		TaskNote: "SIP successfully copied",
	})
	expectations["getSIPExtension"](s, expectationParams{})
	expectations["archiveExtract"](s, expectationParams{})
	expectations["classifySIP"](s, expectationParams{SIPType: enums.SIPTypeBagIt})
	expectations["createTask"](s, expectationParams{
		TaskID:   valBagTaskID,
		TaskName: "Validate Bag",
	})
	expectations["validateBag"](s, expectationParams{})
	expectations["completeTask"](s, expectationParams{
		TaskID:   valBagTaskID,
		TaskNote: "Bag successfully validated",
	})
	expectations["bundleActivity"](s, expectationParams{})

	s.env.OnActivity(
		a3m.CreateAIPActivityName,
		sessionCtx,
		mock.AnythingOfType("*a3m.CreateAIPActivityParams"),
	).Return(nil, fmt.Errorf("a3m error"))

	expectations["zipArchive"](s, expectationParams{ZipSourceDir: transferPath})
	expectations["uploadToFailed"](s, expectationParams{
		UploadPath: transferPath + ".zip",
		FailedKey:  failedPIPKey,
	})
	expectations["updateSIPFailed"](s, expectationParams{
		Status:    enums.SIPStatusError,
		FailedAs:  enums.SIPFailedAsPIP,
		FailedKey: failedPIPKey,
	})
	expectations["completeWorkflow"](s, expectationParams{})
	expectations["removePaths"](s, expectationParams{})

	s.ExecuteAndValidateWorkflow(&ingest.ProcessingWorkflowRequest{
		Key:             key,
		WatcherName:     watcherName,
		RetentionPeriod: retentionPeriod,
		Type:            enums.WorkflowTypeCreateAip,
		SIPUUID:         sipUUID,
		SIPName:         sipName,
	}, true)
}

func (s *ProcessingWorkflowTestSuite) TestFailedPIPAM() {
	s.SetupWorkflowTest(config.Configuration{
		AM:           am.Config{ZipPIP: true},
		Preservation: pres.Config{TaskQueue: temporal.AmWorkerTaskQueue},
		Storage:      storage.Config{DefaultPermanentLocationID: amssLocationID},
	})

	expectations["createSIP"](s, expectationParams{})
	expectations["setStatusInProgress"](s, expectationParams{})
	expectations["createWorkflow"](s, expectationParams{})
	expectations["createTask"](s, expectationParams{
		TaskID:   copySIPTaskID,
		TaskName: "Copy SIP to workspace",
	})
	expectations["download"](s, expectationParams{})
	expectations["completeTask"](s, expectationParams{
		TaskID:   copySIPTaskID,
		TaskNote: "SIP successfully copied",
	})
	expectations["getSIPExtension"](s, expectationParams{})
	expectations["archiveExtract"](s, expectationParams{})
	expectations["classifySIP"](s, expectationParams{})
	expectations["createBag"](s, expectationParams{})
	expectations["zipArchive"](s, expectationParams{})

	s.env.OnActivity(
		am.UploadTransferActivityName,
		sessionCtx,
		&am.UploadTransferActivityParams{SourcePath: extractPath + "/transfer.zip"},
	).Return(nil, fmt.Errorf("AM error"))

	s.env.OnActivity(
		bucketupload.Name,
		sessionCtx,
		&bucketupload.Params{
			Path:       extractPath + "/transfer.zip",
			Key:        failedPIPKey,
			BufferSize: 100_000_000,
		},
	).Return(&bucketupload.Result{}, nil)

	expectations["updateSIPFailed"](s, expectationParams{
		Status:    enums.SIPStatusError,
		FailedAs:  enums.SIPFailedAsPIP,
		FailedKey: failedPIPKey,
	})
	expectations["completeWorkflow"](s, expectationParams{})
	expectations["removePaths"](s, expectationParams{
		RemovePaths: []string{tempPath, extractPath + "/transfer.zip"},
	})

	s.ExecuteAndValidateWorkflow(&ingest.ProcessingWorkflowRequest{
		WatcherName:     watcherName,
		RetentionPeriod: retentionPeriod,
		Type:            enums.WorkflowTypeCreateAip,
		Key:             key,
		SIPUUID:         sipUUID,
		SIPName:         sipName,
	}, true)
}

func (s *ProcessingWorkflowTestSuite) TestInternalUpload() {
	s.SetupWorkflowTest(config.Configuration{
		A3m:           a3m.Config{ShareDir: s.CreateTransferDir()},
		Preservation:  pres.Config{TaskQueue: temporal.A3mWorkerTaskQueue},
		Storage:       storage.Config{DefaultPermanentLocationID: locationID},
		Preprocessing: preprocessing.Config{Enabled: true, SharedPath: prepSharedPath},
	})

	downloadDir := filepath.Join(prepSharedPath, sipUUID.String())

	expectations["setStatusInProgress"](s, expectationParams{})
	expectations["createWorkflow"](s, expectationParams{})
	expectations["createTask"](s, expectationParams{
		TaskID:   copySIPTaskID,
		TaskName: "Copy SIP to workspace",
	})

	s.env.OnActivity(
		activities.DownloadFromInternalBucketActivityName,
		sessionCtx,
		&bucketdownload.Params{
			Key:     key,
			DirPath: downloadDir,
		},
	).Return(&bucketdownload.Result{FilePath: downloadDir + "/" + key}, nil)

	expectations["completeTask"](s, expectationParams{
		TaskID:   copySIPTaskID,
		TaskNote: "SIP successfully copied",
	})

	// Fail the workflow on extraction.
	s.env.OnActivity(
		archiveextract.Name,
		sessionCtx,
		&archiveextract.Params{SourcePath: downloadDir + "/" + key},
	).Return(nil, errors.New("extract error"))

	expectations["copyInBucket"](s, expectationParams{FailedKey: failedSIPKey})
	expectations["deleteFromBucket"](s, expectationParams{})
	expectations["updateSIPFailed"](s, expectationParams{
		Status:    enums.SIPStatusError,
		FailedAs:  enums.SIPFailedAsSIP,
		FailedKey: failedSIPKey,
	})
	expectations["completeWorkflow"](s, expectationParams{})
	expectations["removePaths"](s, expectationParams{RemovePaths: []string{downloadDir}})

	s.ExecuteAndValidateWorkflow(&ingest.ProcessingWorkflowRequest{
		Key:       key,
		Type:      enums.WorkflowTypeCreateAip,
		SIPUUID:   sipUUID,
		SIPName:   sipName,
		Extension: ".zip",
	}, true)
}
