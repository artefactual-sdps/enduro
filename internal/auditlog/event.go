package auditlog

import "log/slog"

type Event struct {
	Level    slog.Level
	Msg      string
	Type     string
	ObjectID string
	UserID   string
}

func (e Event) Args() []any {
	return []any{
		"type", e.Type,
		"objectID", e.ObjectID,
		"userID", e.UserID,
	}
}
