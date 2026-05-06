package childwf

const (
	// DecisionRequestSignalName is sent by a child workflow to its parent when
	// execution requires a user decision.
	DecisionRequestSignalName = "decision-request-signal"

	// DecisionResponseSignalName is sent by the parent workflow back to the
	// child workflow once a decision has been made.
	DecisionResponseSignalName = "decision-response-signal"
)

type DecisionRequest struct {
	// Message is a human-readable description of the decision point.
	Message string

	// Options is the list of allowed decision values.
	Options []string
}

type DecisionResponse struct {
	// Option is the selected decision value.
	Option string
}
