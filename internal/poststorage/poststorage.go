package poststorage

type Config struct {
	Namespace    string
	TaskQueue    string
	WorkflowName string
}

type WorkflowParams struct {
	AIPUUID         string
	PreprocessingID string
}
