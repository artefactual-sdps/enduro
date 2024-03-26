package am

import (
	context "context"
	"database/sql"

	"github.com/jonboulle/clockwork"
	"go.artefactual.dev/amclient"

	"github.com/artefactual-sdps/enduro/internal/package_"
)

var jobStatusToPreservationTaskStatus = map[amclient.JobStatus]package_.PreservationTaskStatus{
	amclient.JobStatusUnknown:    package_.TaskStatusUnspecified,
	amclient.JobStatusComplete:   package_.TaskStatusDone,
	amclient.JobStatusProcessing: package_.TaskStatusInProgress,
	amclient.JobStatusFailed:     package_.TaskStatusError,
}

type JobTracker struct {
	// clock is a service that provides clock time.
	clock  clockwork.Clock
	jobSvc amclient.JobsService
	pkgSvc package_.Service

	// presActionID is the PreservationAction ID that will be the parent ID for
	// all saved preservation tasks.
	presActionID uint

	// savedIDs caches the ID of jobs that have already been saved so we don't
	// create duplicates.
	savedIDs map[string]struct{}
}

func NewJobTracker(
	clock clockwork.Clock,
	jobSvc amclient.JobsService,
	pkgSvc package_.Service,
	presActionID uint,
) *JobTracker {
	return &JobTracker{
		clock:  clock,
		jobSvc: jobSvc,
		pkgSvc: pkgSvc,

		presActionID: presActionID,
		savedIDs:     make(map[string]struct{}),
	}
}

// SavePreservationTasks queries the Archivematica jobs list endpoint to get a
// list of completed jobs related to the transfer or ingest identified by
// unitID, then saves any new jobs as preservation tasks.
func (jt *JobTracker) SavePreservationTasks(ctx context.Context, unitID string) (int, error) {
	jobs, err := jt.list(ctx, unitID)
	if err != nil {
		return 0, err
	}

	count, err := jt.savePreservationTasks(ctx, jobs)
	if err != nil {
		return 0, err
	}

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

// savePreservationTasks persists Archivematica jobs data as preservation tasks.
func (jt *JobTracker) savePreservationTasks(ctx context.Context, jobs []amclient.Job) (int, error) {
	var count int
	jobs = filterSavedJobs(jobs, jt.savedIDs)
	for _, job := range jobs {
		// Wait until a job is complete (or failed) before saving it.
		if job.Status == amclient.JobStatusProcessing {
			continue
		}

		pt := ConvertJobToPreservationTask(job)
		pt.PreservationActionID = jt.presActionID

		err := jt.pkgSvc.CreatePreservationTask(ctx, &pt)
		if err != nil {
			return 0, err
		}

		// Add this job ID to the list of savedIDs.
		jt.savedIDs[job.ID] = struct{}{}
		count++
	}

	return count, nil
}

// filterSavedJobs filters out jobs that have an ID in saved then returns the
// filtered job list.
func filterSavedJobs(jobs []amclient.Job, saved map[string]struct{}) []amclient.Job {
	var unsaved []amclient.Job
	for _, job := range jobs {
		if _, ok := saved[job.ID]; !ok {
			unsaved = append(unsaved, job)
		}
	}
	return unsaved
}

// ConvertJobToPreservationTask converts an amclient.Job to a
// package_.PreservationTask.
func ConvertJobToPreservationTask(job amclient.Job) package_.PreservationTask {
	st, co := jobTimeRange(job)
	return package_.PreservationTask{
		TaskID:      job.ID,
		Name:        job.Name,
		Status:      jobStatusToPreservationTaskStatus[job.Status],
		StartedAt:   st,
		CompletedAt: co,
	}
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
