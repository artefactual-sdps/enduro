package enums

/*
ENUM(
error        // Failed due to a system error.
failed       // Failed due to invalid contents.
queued       // Awaiting resource allocation.
processing   // Undergoing work.
pending      // Awaiting user decision.
ingested     // Successfully ingested.
)
*/
type SIPStatus string
