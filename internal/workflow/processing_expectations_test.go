package workflow

import (
	"path/filepath"
	"time"

	"github.com/artefactual-sdps/temporal-activities/archiveextract"
	"github.com/artefactual-sdps/temporal-activities/archivezip"
	"github.com/artefactual-sdps/temporal-activities/bagcreate"
	"github.com/artefactual-sdps/temporal-activities/bagvalidate"
	"github.com/artefactual-sdps/temporal-activities/bucketcopy"
	"github.com/artefactual-sdps/temporal-activities/bucketdelete"
	"github.com/artefactual-sdps/temporal-activities/bucketdownload"
	"github.com/artefactual-sdps/temporal-activities/bucketupload"
	"github.com/artefactual-sdps/temporal-activities/removepaths"
	"github.com/artefactual-sdps/temporal-activities/xmlvalidate"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"go.artefactual.dev/tools/ref"

	"github.com/artefactual-sdps/enduro/internal/a3m"
	"github.com/artefactual-sdps/enduro/internal/am"
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
	moveAIPTaskID    = 103
	sipName          = "name.zip"
	key              = "transfer.zip"
	watcherName      = "watcher"
)

var (
	ctx             = mock.AnythingOfType("*context.valueCtx")
	sessionCtx      = mock.AnythingOfType("*context.timerCtx")
	sipUUID         = uuid.MustParse("e2ace0da-8697-453d-9ea1-4c9b62309e54")
	locationID      = uuid.MustParse("f2cc963f-c14d-4eaa-b950-bd207189a1f1")
	amssLocationID  = uuid.MustParse("e0ed8b2a-8ae2-4546-b5d8-f0090919df04")
	transferID      = uuid.MustParse("65233405-771e-4f7e-b2d9-b08439570ba2")
	aipUUID         = uuid.MustParse("9e8161cc-2815-4d6f-8a75-f003c41b257b")
	wUUID           = uuid.MustParse("8fdfaea1-06ed-4cf6-8bdf-d15d80420f35")
	startTime       = time.Date(2024, 7, 9, 16, 55, 13, 50, time.UTC)
	retentionPeriod = ref.New(time.Second)
)

// expectationParams holds configurable parameters for setting up activity expectations.
type expectationParams struct {
	WorkflowType  enums.WorkflowType // Workflow type (defaults to enums.WorkflowTypeCreateAip).
	SIPType       enums.SIPType      // SIP type for classification (defaults to enums.SIPTypeUnknown).
	TaskID        int                // Task ID for create/complete task activities.
	TaskName      string             // Task name for create/complete task activities.
	TaskNote      string             // Task note for create/complete task activities.
	Status        enums.SIPStatus    // SIP status for failed update.
	FailedAs      enums.SIPFailedAs  // Failure type for failed update.
	FailedKey     string             // Failed key for failed update and bucket uploads.
	RemovePaths   []string           // Cleanup paths (defaults to tempPath, transferPath).
	UploadPath    string             // Upload path (defaults to tempPath + "/" + key).
	ZipSourceDir  string             // Zip archive source directory (defaults to extractPath).
	BundlePath    string             // Bundle path (defaults to extractPath).
	PremisXMLPath string             // PREMIS XML path (defaults to transferPath + "/metadata/premis.xml").
}

// expectationFunc represents a function that sets up activity expectations.
type expectationFunc func(s *ProcessingWorkflowTestSuite, params expectationParams)

// expectations is a map of activity names to their expectation functions.
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
	"createWorkflow": func(s *ProcessingWorkflowTestSuite, params expectationParams) {
		workflowType := enums.WorkflowTypeCreateAip
		if params.WorkflowType.IsValid() {
			workflowType = params.WorkflowType
		}
		s.env.OnActivity(
			createWorkflowLocalActivity,
			ctx,
			s.workflow.ingestsvc,
			&createWorkflowLocalActivityParams{
				RNG:        s.workflow.rng,
				TemporalID: "default-test-workflow-id",
				Type:       workflowType,
				Status:     enums.WorkflowStatusInProgress,
				StartedAt:  startTime,
				SIPUUID:    sipUUID,
			},
		).Return(&createWorkflowLocalActivityResult{ID: workflowID, UUID: wUUID}, nil)
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
					Name:         params.TaskName,
					Status:       enums.TaskStatusInProgress,
					WorkflowUUID: wUUID,
					Note:         params.TaskNote,
				},
			},
		).Return(params.TaskID, nil)
	},
	"completeTask": func(s *ProcessingWorkflowTestSuite, params expectationParams) {
		var taskNote *string
		if params.TaskNote != "" {
			taskNote = &params.TaskNote
		}
		s.env.OnActivity(
			completeTaskLocalActivity,
			ctx,
			s.workflow.ingestsvc,
			&completeTaskLocalActivityParams{
				ID:          params.TaskID,
				Status:      enums.TaskStatusDone,
				CompletedAt: startTime,
				Note:        taskNote,
			},
		).Return(&completeTaskLocalActivityResult{}, nil)
	},
	"download": func(s *ProcessingWorkflowTestSuite, params expectationParams) {
		s.env.OnActivity(
			activities.DownloadActivityName,
			sessionCtx,
			&activities.DownloadActivityParams{
				Key:         key,
				WatcherName: watcherName,
			},
		).Return(&activities.DownloadActivityResult{Path: tempPath + "/" + key}, nil)
	},
	"getSIPExtension": func(s *ProcessingWorkflowTestSuite, params expectationParams) {
		s.env.OnActivity(
			activities.GetSIPExtensionActivityName,
			sessionCtx,
			&activities.GetSIPExtensionActivityParams{Path: tempPath + "/" + key},
		).Return(&activities.GetSIPExtensionActivityResult{Extension: ".zip"}, nil)
	},
	"archiveExtract": func(s *ProcessingWorkflowTestSuite, params expectationParams) {
		s.env.OnActivity(
			archiveextract.Name,
			sessionCtx,
			&archiveextract.Params{SourcePath: tempPath + "/" + key},
		).Return(&archiveextract.Result{ExtractPath: extractPath}, nil)
	},
	"classifySIP": func(s *ProcessingWorkflowTestSuite, params expectationParams) {
		sipType := enums.SIPTypeUnknown
		if params.SIPType.IsValid() {
			sipType = params.SIPType
		}
		s.env.OnActivity(
			activities.ClassifySIPActivityName,
			sessionCtx,
			activities.ClassifySIPActivityParams{Path: extractPath},
		).Return(&activities.ClassifySIPActivityResult{Type: sipType}, nil)
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
					updateParams.Status == params.Status &&
					updateParams.FailedAs == params.FailedAs &&
					updateParams.FailedKey == params.FailedKey &&
					!updateParams.CompletedAt.IsZero()
			}),
		).Return(nil, nil)
	},
	"removePaths": func(s *ProcessingWorkflowTestSuite, params expectationParams) {
		paths := params.RemovePaths
		if len(paths) == 0 {
			paths = []string{tempPath, transferPath}
		}
		s.env.OnActivity(
			removepaths.Name,
			sessionCtx,
			&removepaths.Params{Paths: paths},
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
			&bagvalidate.Params{Path: extractPath},
		).Return(&bagvalidate.Result{Valid: true}, nil)
	},
	"validatePREMIS": func(s *ProcessingWorkflowTestSuite, params expectationParams) {
		xmlPath := params.PremisXMLPath
		if xmlPath == "" {
			xmlPath = filepath.Join(transferPath, "metadata", "premis.xml")
		}
		s.env.OnActivity(
			xmlvalidate.Name,
			sessionCtx,
			&xmlvalidate.Params{
				XMLPath: xmlPath,
				XSDPath: "/home/enduro/premis.xsd",
			},
		).Return(&xmlvalidate.Result{Failures: []string{}}, nil)
	},
	"createBag": func(s *ProcessingWorkflowTestSuite, params expectationParams) {
		s.env.OnActivity(
			bagcreate.Name,
			sessionCtx,
			&bagcreate.Params{SourcePath: extractPath},
		).Return(&bagcreate.Result{BagPath: extractPath}, nil)
	},
	"zipArchive": func(s *ProcessingWorkflowTestSuite, params expectationParams) {
		sourceDir := params.ZipSourceDir
		resultPath := sourceDir + ".zip"
		if sourceDir == "" {
			sourceDir = extractPath
			resultPath = extractPath + "/" + key
		}
		s.env.OnActivity(
			archivezip.Name,
			sessionCtx,
			&archivezip.Params{SourceDir: sourceDir},
		).Return(&archivezip.Result{Path: resultPath}, nil)
	},
	"uploadTransferAM": func(s *ProcessingWorkflowTestSuite, params expectationParams) {
		s.env.OnActivity(
			am.UploadTransferActivityName,
			sessionCtx,
			&am.UploadTransferActivityParams{SourcePath: extractPath + "/" + key},
		).Return(&am.UploadTransferActivityResult{RemoteRelativePath: key}, nil)
	},
	"startTransferAM": func(s *ProcessingWorkflowTestSuite, params expectationParams) {
		s.env.OnActivity(
			am.StartTransferActivityName,
			sessionCtx,
			&am.StartTransferActivityParams{
				Name:         sipName,
				RelativePath: key,
				ZipPIP:       true,
			},
		).Return(&am.StartTransferActivityResult{TransferID: transferID.String()}, nil)
	},
	"pollTransferAM": func(s *ProcessingWorkflowTestSuite, params expectationParams) {
		s.env.OnActivity(
			am.PollTransferActivityName,
			sessionCtx,
			&am.PollTransferActivityParams{
				TransferID:   transferID.String(),
				WorkflowUUID: wUUID,
			},
		).Return(&am.PollTransferActivityResult{SIPID: aipUUID.String()}, nil)
	},
	"pollIngestAM": func(s *ProcessingWorkflowTestSuite, params expectationParams) {
		s.env.OnActivity(
			am.PollIngestActivityName,
			sessionCtx,
			&am.PollIngestActivityParams{
				SIPID:        aipUUID.String(),
				WorkflowUUID: wUUID,
			},
		).Return(&am.PollIngestActivityResult{Status: "COMPLETE"}, nil)
	},
	"createStorageAIP": func(s *ProcessingWorkflowTestSuite, params expectationParams) {
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
	},
	"deleteTransferAM": func(s *ProcessingWorkflowTestSuite, params expectationParams) {
		s.env.OnActivity(
			am.DeleteTransferActivityName,
			sessionCtx,
			&am.DeleteTransferActivityParams{Destination: key},
		).Return(nil, nil)
	},
	"uploadToFailed": func(s *ProcessingWorkflowTestSuite, params expectationParams) {
		sourcePath := params.UploadPath
		if sourcePath == "" {
			sourcePath = tempPath + "/" + key
		}
		s.env.OnActivity(
			bucketupload.Name,
			sessionCtx,
			&bucketupload.Params{
				Path:       sourcePath,
				Key:        params.FailedKey,
				BufferSize: 100_000_000,
			},
		).Return(&bucketupload.Result{}, nil)
	},
	"downloadFromInternalBucket": func(s *ProcessingWorkflowTestSuite, params expectationParams) {
		s.env.OnActivity(
			activities.DownloadFromInternalBucketActivityName,
			sessionCtx,
			&bucketdownload.Params{
				Key:     key,
				DirPath: tempPath,
			},
		).Return(&bucketdownload.Result{FilePath: tempPath + "/" + key}, nil)
	},
	"copyInBucket": func(s *ProcessingWorkflowTestSuite, params expectationParams) {
		s.env.OnActivity(
			bucketcopy.Name,
			sessionCtx,
			&bucketcopy.Params{
				SourceKey: key,
				DestKey:   params.FailedKey,
			},
		).Return(&bucketcopy.Result{}, nil)
	},
	"deleteFromBucket": func(s *ProcessingWorkflowTestSuite, params expectationParams) {
		s.env.OnActivity(
			bucketdelete.Name,
			sessionCtx,
			&bucketdelete.Params{Key: key},
		).Return(&bucketdelete.Result{}, nil)
	},
	"bundleActivity": func(s *ProcessingWorkflowTestSuite, params expectationParams) {
		sourcePath := params.BundlePath
		if sourcePath == "" {
			sourcePath = extractPath
		}
		s.env.OnActivity(
			activities.BundleActivityName,
			sessionCtx,
			&activities.BundleActivityParams{
				SourcePath:  sourcePath,
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
				WorkflowUUID: wUUID,
			},
		).Return(&a3m.CreateAIPActivityResult{UUID: aipUUID.String()}, nil)
	},
}
