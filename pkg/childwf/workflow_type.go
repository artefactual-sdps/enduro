package childwf

/*
ENUM(
preprocessing     // preprocessing runs before SIP processing.
poststorage       // poststorage runs after AIP storage.
postbatch         // postbatch runs after ingesting a batch of SIPs.
)
*/
type WorkflowType string
