package childwf

/*
ENUM(
unspecified         // Unused!
success             // Completed successfully.
system failure      // Failed due to a system error.
validation failure  // Failed due to a policy violation.
)
*/
type TaskOutcome string
