package preprocessing

type Config struct {
	// Extract SIP in preprocessing.
	Extract bool

	// SharedPath is a filesystem path shared between Enduro and the
	// preprocessing child workflow.
	SharedPath string
}
