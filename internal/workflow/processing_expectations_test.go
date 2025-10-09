package workflow

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/artefactual-sdps/temporal-activities/archiveextract"
	"github.com/artefactual-sdps/temporal-activities/archivezip"
	"github.com/artefactual-sdps/temporal-activities/bagcreate"
	"github.com/artefactual-sdps/temporal-activities/bagvalidate"
	"github.com/artefactual-sdps/temporal-activities/bucketupload"
	"github.com/artefactual-sdps/temporal-activities/removepaths"
	"github.com/artefactual-sdps/temporal-activities/xmlvalidate"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"

	"github.com/artefactual-sdps/enduro/internal/a3m"
	"github.com/artefactual-sdps/enduro/internal/datatypes"
	"github.com/artefactual-sdps/enduro/internal/enums"
	"github.com/artefactual-sdps/enduro/internal/workflow/activities"
)

const (
	tempPath         = "/tmp/enduro123456"
	extractPath      = "/tmp/enduro123456/extract"
	transferPath     = "/home/a3m/.local/share/a3m/share/enduro2985726865"
	prepSharedPath   = "/home/enduro/preprocessing/"
	prepDownloadPath = "/home/enduro/preprocessing/enduro123456"
	prepExtractPath  = "/home/enduro/preprocessing/enduro123456/extract"
	failedSIPKey     = "Failed_SIP_name-e2ace0da-8697-453d-9ea1-4c9b62309e54.zip"
	failedPIPKey     = "Failed_PIP_name-e2ace0da-8697-453d-9ea1-4c9b62309e54.zip"
	sipID            = 1
	workflowID       = 1
	copySIPTaskID    = 100
	valBagTaskID     = 101
	valPREMISTaskID  = 102
	uploadTaskID     = 103
	reviewAIPTaskID  = 104
	moveAIPTaskID    = 105
	deleteSIPTaskID  = 106
	sipName          = "name.zip"
	key              = "transfer.zip"
	watcherName      = "watcher"
)

var (
	ctx             = mock.AnythingOfType("*context.valueCtx")
	sessionCtx      = mock.AnythingOfType("*context.timerCtx")
	internalCtx     = mock.AnythingOfType("*internal.valueCtx")
	sipUUID         = uuid.MustParse("e2ace0da-8697-453d-9ea1-4c9b62309e54")
	locationID      = uuid.MustParse("f2cc963f-c14d-4eaa-b950-bd207189a1f1")
	amssLocationID  = uuid.MustParse("e0ed8b2a-8ae2-4546-b5d8-f0090919df04")
	transferID      = uuid.MustParse("65233405-771e-4f7e-b2d9-b08439570ba2")
	aipUUID         = uuid.MustParse("9e8161cc-2815-4d6f-8a75-f003c41b257b")
	workflowUUID    = uuid.MustParse("8fdfaea1-06ed-4cf6-8bdf-d15d80420f35")
	startTime       = time.Date(2024, 7, 9, 16, 55, 13, 50, time.UTC)
	retentionPeriod = time.Second
)

// expectationParams holds configurable parameters for setting up activity expectations.
type expectationParams struct {
	// SIP type used for classification.
	sipType enums.SIPType

	// SIP status for status updates.
	sipStatus enums.SIPStatus

	// Workflow type for workflow creation.
	workflowType enums.WorkflowType

	// Workflow status for workflow status updates.
	workflowStatus enums.WorkflowStatus

	// Task identifier for task operations.
	taskID int

	// Task name for task operations.
	taskName string

	// Task note for task operations.
	taskNote string

	// Task status for task operations.
	taskStatus enums.TaskStatus

	// Failure type for failed SIP updates.
	failedAs enums.SIPFailedAs

	// Failure key for failed updates and bucket uploads.
	failedKey string

	// Cleanup paths for removal operations.
	removePaths []string

	// Download path for download operations.
	downloadPath string

	// Download destination path for download operations.
	downloadDestPath string

	// Extract path for multiple file operations.
	extractPath string

	// Failed path for failed uploads.
	failedPath string

	// PREMIS XML file path for validation.
	premisXMLPath string

	// Retention period for original SIP deletion.
	retentionPeriod time.Duration
}

// defaultParams returns a new expectationParams instance with sensible defaults.
// All fields that have meaningful defaults are set. Fields that must be provided
// on each expectation (like taskID, taskName, etc.) are left as zero values.
func defaultParams() expectationParams {
	return expectationParams{
		sipType:         enums.SIPTypeUnknown,
		workflowType:    enums.WorkflowTypeCreateAip,
		workflowStatus:  enums.WorkflowStatusPending,
		removePaths:     []string{tempPath, transferPath},
		extractPath:     extractPath,
		downloadPath:    tempPath + "/" + key,
		failedPath:      tempPath + "/" + key,
		premisXMLPath:   filepath.Join(transferPath, "metadata", "premis.xml"),
		retentionPeriod: retentionPeriod,
	}
}

// updateTaskParams updates task related fields in expectationParams.
func (p *expectationParams) updateTaskParams(id int, status enums.TaskStatus, name, note string) {
	p.taskID = id
	p.taskStatus = status
	p.taskName = name
	p.taskNote = note
}

// expectationFunc represents a function that sets up activity expectations for testing.
// Functions configure mock activities using s.env.OnActivity() calls with the provided
// test suite and parameters.
type expectationFunc func(s *ProcessingWorkflowTestSuite, params expectationParams)

// expectations is a map of activity names to their expectation functions.
// Each key represents a logical activity name that corresponds to a Temporal activity.
var expectations = map[string]expectationFunc{
	"createSIP": func(s *ProcessingWorkflowTestSuite, params expectationParams) {
		s.env.OnActivity(
			createSIPLocalActivity,
			ctx,
			s.workflow.ingestsvc,
			&createSIPLocalActivityParams{
				UUID:   sipUUID,
				Name:   sipName,
				Status: enums.SIPStatusQueued,
			},
		).Return(sipID, nil)
	},
	"setStatusInProgress": func(s *ProcessingWorkflowTestSuite, params expectationParams) {
		s.env.OnActivity(
			setStatusInProgressLocalActivity,
			ctx,
			s.workflow.ingestsvc,
			sipUUID,
			startTime,
		).Return(&setStatusInProgressLocalActivityResult{}, nil)
	},
	"setStatus": func(s *ProcessingWorkflowTestSuite, params expectationParams) {
		s.env.OnActivity(
			setStatusLocalActivity,
			ctx,
			s.workflow.ingestsvc,
			sipUUID,
			params.sipStatus,
		).Return(&setStatusLocalActivityResult{}, nil)
	},
	"createWorkflow": func(s *ProcessingWorkflowTestSuite, params expectationParams) {
		s.env.OnActivity(
			createWorkflowLocalActivity,
			ctx,
			s.workflow.ingestsvc,
			&createWorkflowLocalActivityParams{
				RNG:        s.workflow.rng,
				TemporalID: "default-test-workflow-id",
				Type:       params.workflowType,
				Status:     enums.WorkflowStatusInProgress,
				StartedAt:  startTime,
				SIPUUID:    sipUUID,
			},
		).Return(&createWorkflowLocalActivityResult{ID: workflowID, UUID: workflowUUID}, nil)
	},
	"setWorkflowStatus": func(s *ProcessingWorkflowTestSuite, params expectationParams) {
		s.env.OnActivity(
			setWorkflowStatusLocalActivity,
			ctx,
			s.workflow.ingestsvc,
			workflowID,
			params.workflowStatus,
		).Return(nil, nil)
	},
	"completeWorkflow": func(s *ProcessingWorkflowTestSuite, params expectationParams) {
		s.env.OnActivity(
			completeWorkflowLocalActivity,
			ctx,
			s.workflow.ingestsvc,
			mock.AnythingOfType("*workflow.completeWorkflowLocalActivityParams"),
		).Return(nil, nil)
	},
	"createTask": func(s *ProcessingWorkflowTestSuite, params expectationParams) {
		s.env.OnActivity(
			createTaskLocalActivity,
			ctx,
			&createTaskLocalActivityParams{
				Ingestsvc: s.workflow.ingestsvc,
				RNG:       s.workflow.rng,
				Task: &datatypes.Task{
					Name:         params.taskName,
					Status:       params.taskStatus,
					WorkflowUUID: workflowUUID,
					Note:         params.taskNote,
				},
			},
		).Return(params.taskID, nil)
	},
	"completeTask": func(s *ProcessingWorkflowTestSuite, params expectationParams) {
		var taskNote *string
		if params.taskNote != "" {
			taskNote = &params.taskNote
		}
		s.env.OnActivity(
			completeTaskLocalActivity,
			ctx,
			s.workflow.ingestsvc,
			&completeTaskLocalActivityParams{
				ID:     params.taskID,
				Status: params.taskStatus,
				Note:   taskNote,
			},
		).Return(&completeTaskLocalActivityResult{}, nil)
	},
	"download": func(s *ProcessingWorkflowTestSuite, params expectationParams) {
		s.env.OnActivity(
			activities.DownloadActivityName,
			sessionCtx,
			&activities.DownloadActivityParams{
				Key:             key,
				WatcherName:     watcherName,
				DestinationPath: params.downloadDestPath,
			},
		).Return(&activities.DownloadActivityResult{Path: params.downloadPath}, nil)
	},
	"getSIPExtension": func(s *ProcessingWorkflowTestSuite, params expectationParams) {
		s.env.OnActivity(
			activities.GetSIPExtensionActivityName,
			sessionCtx,
			&activities.GetSIPExtensionActivityParams{Path: params.downloadPath},
		).Return(&activities.GetSIPExtensionActivityResult{Extension: ".zip"}, nil)
	},
	"archiveExtract": func(s *ProcessingWorkflowTestSuite, params expectationParams) {
		s.env.OnActivity(
			archiveextract.Name,
			sessionCtx,
			&archiveextract.Params{SourcePath: params.downloadPath},
		).Return(&archiveextract.Result{ExtractPath: params.extractPath}, nil)
	},
	"classifySIP": func(s *ProcessingWorkflowTestSuite, params expectationParams) {
		s.env.OnActivity(
			activities.ClassifySIPActivityName,
			sessionCtx,
			activities.ClassifySIPActivityParams{Path: params.extractPath},
		).Return(&activities.ClassifySIPActivityResult{Type: params.sipType}, nil)
	},
	"updateSIPProcessing": func(s *ProcessingWorkflowTestSuite, params expectationParams) {
		s.env.OnActivity(
			updateSIPLocalActivity,
			ctx,
			s.workflow.ingestsvc,
			&updateSIPLocalActivityParams{
				UUID:    sipUUID,
				Name:    sipName,
				AIPUUID: aipUUID.String(),
				Status:  enums.SIPStatusProcessing,
			},
		).Return(nil, nil)
	},
	"updateSIPIngested": func(s *ProcessingWorkflowTestSuite, params expectationParams) {
		s.env.OnActivity(
			updateSIPLocalActivity,
			ctx,
			s.workflow.ingestsvc,
			mock.MatchedBy(func(updateParams *updateSIPLocalActivityParams) bool {
				return updateParams.UUID == sipUUID &&
					updateParams.Name == sipName &&
					updateParams.AIPUUID == aipUUID.String() &&
					updateParams.Status == enums.SIPStatusIngested &&
					updateParams.FailedAs == "" &&
					updateParams.FailedKey == "" &&
					!updateParams.CompletedAt.IsZero()
			}),
		).Return(nil, nil)
	},
	"updateSIPFailed": func(s *ProcessingWorkflowTestSuite, params expectationParams) {
		s.env.OnActivity(
			updateSIPLocalActivity,
			ctx,
			s.workflow.ingestsvc,
			mock.MatchedBy(func(updateParams *updateSIPLocalActivityParams) bool {
				return updateParams.UUID == sipUUID &&
					updateParams.Name == sipName &&
					updateParams.AIPUUID == "" &&
					updateParams.Status == params.sipStatus &&
					updateParams.FailedAs == params.failedAs &&
					updateParams.FailedKey == params.failedKey &&
					!updateParams.CompletedAt.IsZero()
			}),
		).Return(nil, nil)
	},
	"removePaths": func(s *ProcessingWorkflowTestSuite, params expectationParams) {
		s.env.OnActivity(
			removepaths.Name,
			sessionCtx,
			&removepaths.Params{Paths: params.removePaths},
		).Return(&removepaths.Result{}, nil)
	},
	"deleteOriginal": func(s *ProcessingWorkflowTestSuite, params expectationParams) {
		s.env.OnActivity(
			activities.DeleteOriginalActivityName,
			sessionCtx,
			watcherName,
			key,
		).Return(nil, nil)
	},
	"validateBag": func(s *ProcessingWorkflowTestSuite, params expectationParams) {
		s.env.OnActivity(
			bagvalidate.Name,
			sessionCtx,
			&bagvalidate.Params{Path: params.extractPath},
		).Return(&bagvalidate.Result{Valid: true}, nil)
	},
	"validatePREMIS": func(s *ProcessingWorkflowTestSuite, params expectationParams) {
		s.env.OnActivity(
			xmlvalidate.Name,
			sessionCtx,
			&xmlvalidate.Params{
				XMLPath: params.premisXMLPath,
				XSDPath: "/home/enduro/premis.xsd",
			},
		).Return(&xmlvalidate.Result{Failures: []string{}}, nil)
	},
	"createBag": func(s *ProcessingWorkflowTestSuite, params expectationParams) {
		s.env.OnActivity(
			bagcreate.Name,
			sessionCtx,
			&bagcreate.Params{SourcePath: params.extractPath},
		).Return(&bagcreate.Result{BagPath: params.extractPath}, nil)
	},
	"zipArchive": func(s *ProcessingWorkflowTestSuite, params expectationParams) {
		s.env.OnActivity(
			archivezip.Name,
			sessionCtx,
			&archivezip.Params{SourceDir: params.extractPath},
		).Return(&archivezip.Result{Path: params.extractPath + "/" + key}, nil)
	},
	"uploadToFailed": func(s *ProcessingWorkflowTestSuite, params expectationParams) {
		s.env.OnActivity(
			bucketupload.Name,
			sessionCtx,
			&bucketupload.Params{
				Path:       params.failedPath,
				Key:        params.failedKey,
				BufferSize: 100_000_000,
			},
		).Return(&bucketupload.Result{}, nil)
	},
	"bundle": func(s *ProcessingWorkflowTestSuite, params expectationParams) {
		s.env.OnActivity(
			activities.BundleActivityName,
			sessionCtx,
			&activities.BundleActivityParams{
				SourcePath:  params.extractPath,
				TransferDir: s.transferDir,
				IsDir:       true,
			},
		).Return(&activities.BundleActivityResult{FullPath: transferPath}, nil)
	},
	"createAIP": func(s *ProcessingWorkflowTestSuite, params expectationParams) {
		s.env.OnActivity(
			a3m.CreateAIPActivityName,
			sessionCtx,
			&a3m.CreateAIPActivityParams{
				Name:         sipName,
				Path:         transferPath,
				WorkflowUUID: workflowUUID,
			},
		).Return(&a3m.CreateAIPActivityResult{UUID: aipUUID.String()}, nil)
	},
	"uploadAIP": func(s *ProcessingWorkflowTestSuite, params expectationParams) {
		s.env.OnActivity(
			activities.UploadActivityName,
			sessionCtx,
			&activities.UploadActivityParams{
				Name:  sipName,
				AIPID: aipUUID.String(),
			},
		).Return(nil, nil)
	},
	"moveAIP": func(s *ProcessingWorkflowTestSuite, params expectationParams) {
		s.env.OnActivity(
			activities.MoveToPermanentStorageActivityName,
			sessionCtx,
			&activities.MoveToPermanentStorageActivityParams{
				AIPID:      aipUUID.String(),
				LocationID: locationID,
			},
		).Return(nil, nil)
	},
	"pollMoveAIP": func(s *ProcessingWorkflowTestSuite, params expectationParams) {
		s.env.OnActivity(
			activities.PollMoveToPermanentStorageActivityName,
			sessionCtx,
			&activities.PollMoveToPermanentStorageActivityParams{AIPID: aipUUID.String()},
		).Return(nil, nil)
	},
}

func downloadExpectations(s *ProcessingWorkflowTestSuite, params expectationParams) {
	expectations["createSIP"](s, params)
	expectations["setStatusInProgress"](s, params)
	expectations["createWorkflow"](s, params)
	params.updateTaskParams(copySIPTaskID, enums.TaskStatusInProgress, "Copy SIP to workspace", "")
	expectations["createTask"](s, params)
	expectations["download"](s, params)
	params.updateTaskParams(copySIPTaskID, enums.TaskStatusDone, "", "SIP successfully copied")
	expectations["completeTask"](s, params)
	expectations["getSIPExtension"](s, params)
}

func reviewA3mExpectations(s *ProcessingWorkflowTestSuite, params expectationParams) {
	expectations["bundle"](s, params)
	expectations["createAIP"](s, params)
	expectations["updateSIPProcessing"](s, params)
	params.updateTaskParams(uploadTaskID, enums.TaskStatusInProgress, "Move AIP", "Moving to review bucket")
	expectations["createTask"](s, params)
	expectations["uploadAIP"](s, params)
	params.updateTaskParams(uploadTaskID, enums.TaskStatusDone, "", "Moved to review bucket")
	expectations["completeTask"](s, params)
	params.sipStatus = enums.SIPStatusPending
	expectations["setStatus"](s, params)
	params.workflowStatus = enums.WorkflowStatusPending
	expectations["setWorkflowStatus"](s, params)
	params.updateTaskParams(reviewAIPTaskID, enums.TaskStatusPending, "Review AIP", "Awaiting user decision")
	expectations["createTask"](s, params)
	params.sipStatus = enums.SIPStatusProcessing
	expectations["setStatus"](s, params)
	params.workflowStatus = enums.WorkflowStatusInProgress
	expectations["setWorkflowStatus"](s, params)
}

func autoApproveA3mExpectations(s *ProcessingWorkflowTestSuite, params expectationParams) {
	expectations["bundle"](s, params)
	expectations["createAIP"](s, params)
	expectations["updateSIPProcessing"](s, params)
	expectations["uploadAIP"](s, params)
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
}

func cleanupExpectations(s *ProcessingWorkflowTestSuite, params expectationParams) {
	expectations["removePaths"](s, params)
	if params.retentionPeriod >= 0 {
		params.updateTaskParams(
			deleteSIPTaskID,
			enums.TaskStatusInProgress,
			"Delete original SIP",
			fmt.Sprintf("The original SIP will be deleted in %s", params.retentionPeriod),
		)
		expectations["createTask"](s, params)
		expectations["deleteOriginal"](s, params)
		params.updateTaskParams(deleteSIPTaskID, enums.TaskStatusDone, "", "SIP successfully deleted")
		expectations["completeTask"](s, params)
	}
	expectations["updateSIPIngested"](s, params)
	expectations["completeWorkflow"](s, params)
}
