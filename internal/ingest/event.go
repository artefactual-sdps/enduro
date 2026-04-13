package ingest

import (
	"context"
	"encoding/json"
	"fmt"

	ingestclient "github.com/artefactual-sdps/enduro/internal/api/gen/http/ingest/client"
	ingestserver "github.com/artefactual-sdps/enduro/internal/api/gen/http/ingest/server"
	goaingest "github.com/artefactual-sdps/enduro/internal/api/gen/ingest"
	"github.com/artefactual-sdps/enduro/internal/event"
)

// EventSerializer handles serialization/deserialization of ingest events.
type EventSerializer struct{}

var _ event.Serializer[*goaingest.IngestEvent] = (*EventSerializer)(nil)

func (s *EventSerializer) Marshal(event *goaingest.IngestEvent) ([]byte, error) {
	return json.Marshal(ingestserver.NewMonitorResponseBody(event))
}

func (s *EventSerializer) Unmarshal(data []byte) (*goaingest.IngestEvent, error) {
	payload := ingestclient.MonitorResponseBody{}
	if err := json.Unmarshal(data, &payload); err != nil {
		return nil, err
	}
	if err := ingestclient.ValidateMonitorResponseBody(&payload); err != nil {
		return nil, err
	}
	return ingestclient.NewMonitorIngestEventOK(&payload), nil
}

// Event is a type constraint for all ingest events.
type Event interface {
	*goaingest.IngestPingEvent |
		*goaingest.SIPCreatedEvent |
		*goaingest.SIPUpdatedEvent |
		*goaingest.SIPStatusUpdatedEvent |
		*goaingest.SIPWorkflowCreatedEvent |
		*goaingest.SIPWorkflowUpdatedEvent |
		*goaingest.SIPTaskCreatedEvent |
		*goaingest.SIPTaskUpdatedEvent |
		*goaingest.BatchCreatedEvent |
		*goaingest.BatchUpdatedEvent
}

// PublishEvent publishes an ingest event with type safety.
func PublishEvent[E Event](ctx context.Context, svc event.Service[*goaingest.IngestEvent], event E) {
	svc.PublishEvent(ctx, &goaingest.IngestEvent{Value: NewEventValue(event)})
}

func NewEventValue[E Event](event E) goaingest.Value {
	switch e := any(event).(type) {
	case *goaingest.IngestPingEvent:
		return goaingest.NewValueIngestPingEvent(e)
	case *goaingest.SIPCreatedEvent:
		return goaingest.NewValueSipCreatedEvent(e)
	case *goaingest.SIPUpdatedEvent:
		return goaingest.NewValueSipUpdatedEvent(e)
	case *goaingest.SIPStatusUpdatedEvent:
		return goaingest.NewValueSipStatusUpdatedEvent(e)
	case *goaingest.SIPWorkflowCreatedEvent:
		return goaingest.NewValueSipWorkflowCreatedEvent(e)
	case *goaingest.SIPWorkflowUpdatedEvent:
		return goaingest.NewValueSipWorkflowUpdatedEvent(e)
	case *goaingest.SIPTaskCreatedEvent:
		return goaingest.NewValueSipTaskCreatedEvent(e)
	case *goaingest.SIPTaskUpdatedEvent:
		return goaingest.NewValueSipTaskUpdatedEvent(e)
	case *goaingest.BatchCreatedEvent:
		return goaingest.NewValueBatchCreatedEvent(e)
	case *goaingest.BatchUpdatedEvent:
		return goaingest.NewValueBatchUpdatedEvent(e)
	default:
		panic(fmt.Sprintf("unsupported ingest event type %T", event))
	}
}
