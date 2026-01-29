package childwf

type PreprocessingParams struct {
	// Relative path to the shared path.
	RelativePath string
}

type PreprocessingResult struct {
	// Outcome is an integer indicating if the workflow completed successfully,
	// or with errors.
	Outcome Outcome

	// Relative path to the shared path.
	RelativePath string

	// PreservationTasks is a log of the tasks performed by preprocessing.
	PreservationTasks []Task
}
