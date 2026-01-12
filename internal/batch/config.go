package batch

type Config struct {
	// Poststorage contains the configuration for an optional post-storage
	// workflow that runs after all SIPs in a batch have been ingested and
	// stored. If nil, no post-storage workflow will be executed.
	Poststorage *PostStorageConfig
}

type PostStorageConfig struct {
	// Namespace is the Temporal namespace used by the batch post-storage
	// workflow. A unique Namespace will isolating the post-storage workflow
	// from other Temporal workflows running on the same server. If no Namespace
	// is specified, the default namespace will be used.
	Namespace string

	// TaskQueue is the Temporal task queue that will be used to schedule
	// batch post-storage workflow execution. A Temporal worker must be
	// listening on the specified task queue to execute the workflow.
	TaskQueue string

	// WorkflowName is the name of the batch post-storage Temporal workflow.
	WorkflowName string
}
