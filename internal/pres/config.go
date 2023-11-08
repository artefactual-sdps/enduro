package pres

type Config struct {
	// TaskQueue sets the Temporal task queue to use for the processing workflow
	// (e.g. ` temporal.A3mWorkerTaskQueue`, ` temporal.AMWorkerTaskQueue`).
	// The task queue determines which processing worker will run the processing
	// workflow - a3m or Archivematica, and is used to branch the processing
	// workflow logic.
	TaskQueue string
}
