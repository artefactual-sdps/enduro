package enums

/*
ENUM(
unspecified  // Status is indeterminate.
in progress  // Work is ongoing.
done         // Completed successfully.
error        // Halted due to a system error.
queued       // Awaiting resource allocation.
pending      // Awaiting user decision.
failed       // Halted due to a policy violation.
)
*/
type TaskStatus uint
