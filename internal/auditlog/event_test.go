package auditlog

import (
	"testing"

	"gotest.tools/v3/assert"
)

func TestEvent(t *testing.T) {
	t.Parallel()

	ev := Event{
		Msg:      "SIP created",
		Type:     "sip.created",
		ObjectID: "sip-123",
		UserID:   "user-456",
	}

	got := ev.Args()
	want := []any{
		"type", "sip.created",
		"objectID", "sip-123",
		"userID", "user-456",
	}
	assert.DeepEqual(t, got, want)
}
