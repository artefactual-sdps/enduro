package am

import (
	context "context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"github.com/jonboulle/clockwork"
	"go.artefactual.dev/amclient"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	"github.com/artefactual-sdps/enduro/internal/datatypes"
	"github.com/artefactual-sdps/enduro/internal/enums"
	"github.com/artefactual-sdps/enduro/internal/ingest"
)

var keepJobs = map[string]struct{}{
	// Transfer jobs.
	"3229e01f-adf3-4294-85f7-4acb01b3fbcf": {}, // Extract zipped bag transfer
	"154dd501-a344-45a9-97e3-b30093da35f5": {}, // Rename with transfer UUID
	"3c526a07-c3b8-4e53-801b-7f3d0c4857a5": {}, // Assign file UUIDs to objects
	"c77fee8c-7c4e-4871-a72e-94d499994869": {}, // Assign checksums and file sizes to objects
	"5415c813-3637-49ab-afec-9b435c2e4d2c": {}, // Assign UUIDs to directories
	"1c2550f1-3fc0-45d8-8bc4-4c06d720283b": {}, // Scan for viruses in directories
	"56eebd45-5600-4768-a8c2-ec0114555a3d": {}, // Generate transfer structure report
	"2584b25c-8d98-44b7-beca-2b3ea2ea2505": {}, // Change object and directory filenames
	"a329d39b-4711-4231-b54e-b5958934dccb": {}, // Change transfer name
	"2522d680-c7d9-4d06-8b11-a28d8bd8a71f": {}, // Identify file format
	"1a136608-ae7b-42b4-bf2f-de0e514cfd47": {}, // Load rights
	"303a65f6-a16f-4a06-807b-cb3425a30201": {}, // Characterize and extract metadata
	"d51a7ed8-9cf4-424b-9671-85fd8b5b95aa": {}, // Load PREMIS events from metadata/premis.xml
	"a536828c-be65-4088-80bd-eb511a0a063d": {}, // Validate formats
	"307edcde-ad10-401c-92c4-652917c993ed": {}, // Generate METS.xml document
	"675acd22-828d-4949-adc7-1888240f5e3d": {}, // Parse external METS
	// Ingest jobs.
	"15a2df8a-7b45-4c11-b6fa-884c9b7e5c67": {}, // Identify manually normalized files
	"5b0042a2-2244-475c-85d5-41e4b11e65d6": {}, // Validate preservation derivatives
	"78b7adff-861d-4450-b6dd-3fabe96a849e": {}, // Check for manual normalized files
	"47dd6ea6-1ee7-4462-8b84-3fc4c1eeeb7f": {}, // Check for submission documentation
	"2a62f025-83ec-4f23-adb4-11d5da7ad8c2": {}, // Assign checksums and file sizes to submissionDocumentation
	"11033dbd-e4d4-4dd6-8bcf-48c424e222e3": {}, // Change file and directory names in submission documentation
	"1ba589db-88d1-48cf-bb1a-a5f9d2b17378": {}, // Scan for viruses in submission documentation
	"1dce8e21-7263-4cc4-aa59-968d9793b5f2": {}, // Identify file format
	"33d7ac55-291c-43ae-bb42-f599ef428325": {}, // Characterize and extract metadata on submission documentation
	"b0ffcd90-eb26-4caf-8fab-58572d205f04": {}, // Process JSON metadata
	"e4b0c713-988a-4606-82ea-4b565936d9a7": {}, // Move metadata to objects directory
	"dc9d4991-aefa-4d7e-b7b5-84e3c4336e74": {}, // Assign file UUIDs to metadata
	"b6b0fe37-aa26-40bd-8be8-d3acebf3ccf8": {}, // Assign checksums and file sizes to metadata
	"b21018df-f67d-469a-9ceb-ac92ac68654e": {}, // Change file and directory names in metadata
	"8bc92801-4308-4e3b-885b-1a89fdcd3014": {}, // Scan for viruses in metadata
	"b2444a6e-c626-4487-9abc-1556dd89a8ae": {}, // Identify file format of metadata files
	"04493ab2-6cad-400d-8832-06941f121a96": {}, // Characterize and extract metadata on metadata files
	"873b428f-2c86-42b6-b463-aeda925bf559": {}, // Load persistent identifiers from external file
	"ccf8ec5c-3a9a-404a-a7e7-8f567d3b36a0": {}, // Generate METS.xml document
	"523c97cc-b267-4cfb-8209-d99e523bf4b3": {}, // Add README file
	"3e25bda6-5314-4bb4-aa1e-90900dce887d": {}, // Prepare AIP
	"d55b42c8-c7c5-4a40-b626-d248d2bd883f": {}, // Compress AIP
	"3f543585-fa4f-4099-9153-dd6d53572f5c": {}, // Verify AIP
	"20515483-25ed-4133-b23e-5bb14cab8e22": {}, // Store the AIP
	"48703fad-dc44-4c8e-8f47-933df3ef6179": {}, // Index AIP
	"b7cf0d9a-504f-4f4e-9930-befa817d67ff": {}, // Clean up after storing AIP
}

var jobStatusToTaskStatus = map[amclient.JobStatus]enums.TaskStatus{
	amclient.JobStatusUnknown:    enums.TaskStatusUnspecified,
	amclient.JobStatusComplete:   enums.TaskStatusDone,
	amclient.JobStatusProcessing: enums.TaskStatusInProgress,
	amclient.JobStatusFailed:     enums.TaskStatusError,
}

type JobTracker struct {
	// clock is a service that provides clock time.
	clock     clockwork.Clock
	jobSvc    amclient.JobsService
	ingestsvc ingest.Service

	// workflowUUID is the workflow UUID that will be the parent for
	// all saved tasks.
	workflowUUID uuid.UUID

	// savedIDs caches the ID of jobs that have already been saved so we don't
	// create duplicates.
	savedIDs map[string]struct{}

	tracer trace.Tracer
}

func NewJobTracker(
	clock clockwork.Clock,
	jobSvc amclient.JobsService,
	ingestsvc ingest.Service,
	workflowUUID uuid.UUID,
	tracer trace.Tracer,
) *JobTracker {
	return &JobTracker{
		clock:     clock,
		jobSvc:    jobSvc,
		ingestsvc: ingestsvc,

		workflowUUID: workflowUUID,
		savedIDs:     make(map[string]struct{}),
		tracer:       tracer,
	}
}

// SaveTasks queries the Archivematica jobs list endpoint to get a
// list of completed jobs related to the transfer or ingest identified by
// unitID, then saves any new jobs as tasks.
func (jt *JobTracker) SaveTasks(ctx context.Context, unitID string) (int, error) {
	ctx, span := jt.tracer.Start(ctx, "JobTracker.SaveTasks")
	defer span.End()
	span.SetAttributes(
		attribute.String("workflow.uuid", jt.workflowUUID.String()),
		attribute.String("unit.id", unitID),
	)

	jobs, err := jt.list(ctx, unitID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "list jobs")
		return 0, err
	}
	span.SetAttributes(attribute.Int("jobs.fetched", len(jobs)))

	count, err := jt.saveTasks(ctx, jobs)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "save tasks")
		return 0, err
	}
	span.SetAttributes(attribute.Int("tasks.saved", count))

	return count, nil
}

// list requests a job list for unitID from the Archivematica jobs endpoint.
func (jt *JobTracker) list(ctx context.Context, unitID string) ([]amclient.Job, error) {
	jobs, httpResp, err := jt.jobSvc.List(ctx, unitID, &amclient.JobsListRequest{
		Detailed: true,
	})
	if err != nil {
		return nil, convertAMClientError(httpResp, err)
	}

	return jobs, nil
}

// saveTasks persists Archivematica jobs data as tasks.
func (jt *JobTracker) saveTasks(ctx context.Context, jobs []amclient.Job) (int, error) {
	var count int
	jobs = jt.filterJobs(jobs)
	for _, job := range jobs {
		// Wait until a job is complete (or failed) before saving it.
		if job.Status == amclient.JobStatusProcessing {
			continue
		}

		task, err := ConvertJobToTask(job)
		if err != nil {
			return 0, err
		}
		task.WorkflowUUID = jt.workflowUUID

		err = jt.ingestsvc.CreateTask(ctx, task)
		if err != nil {
			return 0, err
		}

		// Add this job ID to the list of savedIDs.
		jt.savedIDs[job.ID] = struct{}{}
		count++
	}

	return count, nil
}

// filterJobs filters out jobs that have an ID in saved and jobs without
// LinkID in the jobs we want to keep, then returns the filtered job list.
func (jt *JobTracker) filterJobs(jobs []amclient.Job) []amclient.Job {
	var filtered []amclient.Job
	for _, job := range jobs {
		_, okSaved := jt.savedIDs[job.ID]
		_, okKeep := keepJobs[job.LinkID]
		if !okSaved && okKeep {
			filtered = append(filtered, job)
		}
	}
	return filtered
}

// ConvertJobToTask converts an amclient.Job to a datatypes.Task.
func ConvertJobToTask(job amclient.Job) (*datatypes.Task, error) {
	taskUUID, err := uuid.Parse(job.ID)
	if err != nil {
		return nil, fmt.Errorf("unable to parse task UUID from job ID: %q", job.ID)
	}
	st, co := jobTimeRange(job)
	return &datatypes.Task{
		UUID:        taskUUID,
		Name:        job.Name,
		Status:      jobStatusToTaskStatus[job.Status],
		StartedAt:   st,
		CompletedAt: co,
	}, nil
}

// jobTimeRange calculates the overall start and end times for a job.
func jobTimeRange(job amclient.Job) (sql.NullTime, sql.NullTime) {
	if len(job.Tasks) == 0 {
		return sql.NullTime{}, sql.NullTime{}
	}
	st := job.Tasks[0].StartedAt.Time
	ct := job.Tasks[0].CompletedAt.Time

	for _, t := range job.Tasks[1:] {
		// Update st to the earliest task start time.
		if st.After(t.StartedAt.Time) {
			st = t.StartedAt.Time
		}
		// Update ct to the latest task completion time.
		if ct.Before(t.CompletedAt.Time) {
			ct = t.CompletedAt.Time
		}
	}

	// Emit NULLs if we see the zero value.
	start := sql.NullTime{Time: st, Valid: !st.IsZero()}
	end := sql.NullTime{Time: ct, Valid: !ct.IsZero()}

	return start, end
}
