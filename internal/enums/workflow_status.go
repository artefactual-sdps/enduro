package enums

/*
ENUM(
unspecified  // Status is indeterminate.
in progress  // Work is ongoing.
done         // Work has completed successfully.
error        // Halted due to a system error.
queued       // Awaiting resource allocation.
pending      // Awaiting user decision.
failed       // Halted due to a policy violation.
canceled     // Canceled by Batch workflow.
)
*/
type WorkflowStatus uint
