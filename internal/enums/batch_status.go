package enums

/*
ENUM(
queued       // Awaiting resource allocation.
processing   // Undergoing work.
pending      // Awaiting user decision.
ingested     // Successfully ingested.
canceled     // Ingest canceled by user.
)
*/
type BatchStatus string
