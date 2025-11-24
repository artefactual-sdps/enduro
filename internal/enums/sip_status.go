package enums

/*
ENUM(
error        // Failed due to a system error.
failed       // Failed due to invalid contents.
queued       // Awaiting resource allocation.
processing   // Undergoing work.
pending      // Awaiting user decision.
ingested     // Successfully ingested.
validated	 // Passed validation, waiting for other SIPs in the Batch.
canceled     // Canceled as part of a Batch that failed.
)
*/
type SIPStatus string
