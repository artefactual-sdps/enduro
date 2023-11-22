package temporal

const (
	// There are task queues used by our workflow and activity workers. It may
	// be convenient to make these configurable in the future .
	GlobalTaskQueue    = "global"
	A3mWorkerTaskQueue = "a3m"
	AmWorkerTaskQueue  = "am"
)
