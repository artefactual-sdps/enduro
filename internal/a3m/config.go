package a3m

import (
	"fmt"

	transferservice "buf.build/gen/go/artefactual/a3m/protocolbuffers/go/a3m/api/transferservice/v1beta1"
)

const (
	minCompressionLevel = 0
	maxCompressionLevel = 9
)

type Config struct {
	Name     string
	ShareDir string
	Address  string

	// Capacity sets the maximum number of worker sessions the worker can
	// handle at one time (default: 1).
	Capacity int

	Processing
}

func (c *Config) Validate() error {
	if c.AipCompressionAlgorithm < minCompressionLevel || c.AipCompressionLevel > maxCompressionLevel {
		return fmt.Errorf(
			"AipCompressionLevel: %d is outside valid range (%d to %d)",
			c.AipCompressionLevel,
			minCompressionLevel,
			maxCompressionLevel,
		)
	}

	return nil
}

// The `Processing` struct represents a configuration for processing various
// tasks in the transferservice. It mirrors the processing configuration fields
// in transferservice.ProcessingConfig.
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
	AipCompressionLevel                          int32
	AipCompressionAlgorithm                      transferservice.ProcessingConfig_AIPCompressionAlgorithm
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
	AipCompressionAlgorithm:                      transferservice.ProcessingConfig_AIP_COMPRESSION_ALGORITHM_S7_BZIP2,
}
