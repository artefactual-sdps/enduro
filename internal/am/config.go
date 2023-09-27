package am

type Config struct {
	Address string
	Name    string
	User    string
	Key     string
	Processing
	ShareDir string
}

// The `Processing` struct represents a configuration for processing various tasks.
type Processing struct {
	AssignUuidsToDirectories                     bool
	ExamineContents                              bool
	GenerateTransferStructureReport              bool
	DocumentEmptyDirectories                     bool
	ExtractPackages                              bool
	DeletePackagesAfterExtraction                bool
	IdentifyTransfer                             bool
	IdentifySubmissionAndMetadata                bool
	IdentifyBeforeNormalization                  bool
	Normalize                                    bool
	TranscribeFiles                              bool
	PerformPolicyChecksOnOriginals               bool
	PerformPolicyChecksOnPreservationDerivatives bool
	AipCompressionLevel                          int
	AipCompressionAlgorithm                      int
}

// Set the defaults for the a3m transfer service.
var ProcessingDefault = Processing{
	AssignUuidsToDirectories:                     true,
	ExamineContents:                              false,
	GenerateTransferStructureReport:              true,
	DocumentEmptyDirectories:                     true,
	ExtractPackages:                              true,
	DeletePackagesAfterExtraction:                false,
	IdentifyTransfer:                             true,
	IdentifySubmissionAndMetadata:                true,
	IdentifyBeforeNormalization:                  true,
	Normalize:                                    true,
	TranscribeFiles:                              true,
	PerformPolicyChecksOnOriginals:               true,
	PerformPolicyChecksOnPreservationDerivatives: true,
	AipCompressionLevel:                          1,
	AipCompressionAlgorithm:                      1,
}
