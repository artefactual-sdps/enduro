package ingest_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	"gotest.tools/v3/assert"

	goaingest "github.com/artefactual-sdps/enduro/internal/api/gen/ingest"
	"github.com/artefactual-sdps/enduro/internal/auth"
	"github.com/artefactual-sdps/enduro/internal/event"
	"github.com/artefactual-sdps/enduro/internal/ingest"
)

// mockMonitorServerStream implements goaingest.MonitorServerStream for testing.
type mockMonitorServerStream struct {
	events     []*goaingest.IngestEvent
	closed     bool
	failOnSend int
}

func (m *mockMonitorServerStream) Send(event *goaingest.IngestEvent) error {
	return m.SendWithContext(context.Background(), event)
}

func (m *mockMonitorServerStream) SendWithContext(ctx context.Context, event *goaingest.IngestEvent) error {
	if m.closed {
		return fmt.Errorf("stream closed")
	}
	if m.failOnSend > 0 && m.failOnSend == len(m.events)+1 {
		return fmt.Errorf("send failed")
	}
	m.events = append(m.events, event)
	return nil
}

func (m *mockMonitorServerStream) Close() error {
	m.closed = true
	return nil
}

func TestMonitor(t *testing.T) {
	t.Parallel()

	testUUID := uuid.New()
	allEvents := []*goaingest.IngestEvent{
		{Value: ingest.NewEventValue(&goaingest.SIPCreatedEvent{UUID: testUUID})},
		{Value: ingest.NewEventValue(&goaingest.SIPUpdatedEvent{UUID: testUUID})},
		{Value: ingest.NewEventValue(&goaingest.SIPStatusUpdatedEvent{UUID: testUUID})},
		{Value: ingest.NewEventValue(&goaingest.SIPWorkflowCreatedEvent{UUID: testUUID})},
		{Value: ingest.NewEventValue(&goaingest.SIPWorkflowUpdatedEvent{UUID: testUUID})},
		{Value: ingest.NewEventValue(&goaingest.SIPTaskCreatedEvent{UUID: testUUID})},
		{Value: ingest.NewEventValue(&goaingest.SIPTaskUpdatedEvent{UUID: testUUID})},
		{Value: ingest.NewEventValue(&goaingest.BatchCreatedEvent{UUID: testUUID})},
		{Value: ingest.NewEventValue(&goaingest.BatchUpdatedEvent{UUID: testUUID})},
	}
	allWantEvents := []*goaingest.IngestEvent{
		{Value: ingest.NewEventValue(&goaingest.IngestPingEvent{Message: new("Hello")})},
		{Value: ingest.NewEventValue(&goaingest.SIPCreatedEvent{UUID: testUUID})},
		{Value: ingest.NewEventValue(&goaingest.SIPUpdatedEvent{UUID: testUUID})},
		{Value: ingest.NewEventValue(&goaingest.SIPStatusUpdatedEvent{UUID: testUUID})},
		{Value: ingest.NewEventValue(&goaingest.SIPWorkflowCreatedEvent{UUID: testUUID})},
		{Value: ingest.NewEventValue(&goaingest.SIPWorkflowUpdatedEvent{UUID: testUUID})},
		{Value: ingest.NewEventValue(&goaingest.SIPTaskCreatedEvent{UUID: testUUID})},
		{Value: ingest.NewEventValue(&goaingest.SIPTaskUpdatedEvent{UUID: testUUID})},
		{Value: ingest.NewEventValue(&goaingest.BatchCreatedEvent{UUID: testUUID})},
		{Value: ingest.NewEventValue(&goaingest.BatchUpdatedEvent{UUID: testUUID})},
	}

	for _, tt := range []struct {
		name       string
		claims     *auth.Claims
		events     []*goaingest.IngestEvent
		wantEvents []*goaingest.IngestEvent
	}{
		{
			name: "Sends all events for a user with all permissions",
			claims: &auth.Claims{
				Email:         "test@example.com",
				EmailVerified: true,
				Attributes:    []string{"*"},
			},
			events:     allEvents,
			wantEvents: allWantEvents,
		},
		{
			name:       "Sends all events when authentication and/or ABAC is disabled",
			claims:     &auth.Claims{},
			events:     allEvents,
			wantEvents: allWantEvents,
		},
		{
			name: "Filters all events for a user without permissions",
			claims: &auth.Claims{
				Email:         "test@example.com",
				EmailVerified: true,
				Attributes:    []string{},
			},
			events: allEvents,
			wantEvents: []*goaingest.IngestEvent{
				{Value: ingest.NewEventValue(&goaingest.IngestPingEvent{Message: new("Hello")})},
			},
		},
		{
			name: "Filters events based on permissions",
			claims: &auth.Claims{
				Email:         "test@example.com",
				EmailVerified: true,
				Attributes:    []string{auth.IngestSIPSReadAttr},
			},
			events: allEvents,
			wantEvents: []*goaingest.IngestEvent{
				{Value: ingest.NewEventValue(&goaingest.IngestPingEvent{Message: new("Hello")})},
				{Value: ingest.NewEventValue(&goaingest.SIPUpdatedEvent{UUID: testUUID})},
				{Value: ingest.NewEventValue(&goaingest.SIPStatusUpdatedEvent{UUID: testUUID})},
			},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			evsvc := event.NewServiceInMem[*goaingest.IngestEvent]()
			stream := &mockMonitorServerStream{}

			svc := ingest.NewService(ingest.ServiceParams{
				EventService: evsvc,
			})

			// Create a context that will be cancelled to stop the monitor.
			ctx, cancel := context.WithTimeout(t.Context(), 100*time.Millisecond)
			defer cancel()
			ctx = auth.WithUserClaims(ctx, tt.claims)

			// Start monitor in a goroutine.
			errCh := make(chan error, 1)
			go func() {
				errCh <- svc.Monitor(ctx, &goaingest.MonitorPayload{}, stream)
			}()

			// Send test events after a short delay.
			time.Sleep(10 * time.Millisecond)
			for _, event := range tt.events {
				evsvc.PublishEvent(t.Context(), event)
			}

			// Wait for the monitor to finish.
			select {
			case err := <-errCh:
				assert.NilError(t, err)
			case <-time.After(200 * time.Millisecond):
				t.Fatal("Monitor did not complete in expected time")
			}

			assert.DeepEqual(t, stream.events, tt.wantEvents, cmp.AllowUnexported(goaingest.Value{}))
		})
	}
}

func TestMonitorReturnsNilOnStreamSendError(t *testing.T) {
	t.Parallel()

	testUUID := uuid.New()

	for _, tt := range []struct {
		name       string
		failOnSend int
		publish    func(context.Context, event.Service[*goaingest.IngestEvent])
		wantEvents []*goaingest.IngestEvent
	}{
		{
			name:       "Hello",
			failOnSend: 1,
		},
		{
			name:       "Subscribed event",
			failOnSend: 2,
			publish: func(ctx context.Context, evsvc event.Service[*goaingest.IngestEvent]) {
				evsvc.PublishEvent(ctx, &goaingest.IngestEvent{
					Value: ingest.NewEventValue(&goaingest.SIPCreatedEvent{UUID: testUUID}),
				})
			},
			wantEvents: []*goaingest.IngestEvent{
				{Value: ingest.NewEventValue(&goaingest.IngestPingEvent{Message: new("Hello")})},
			},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			evsvc := event.NewServiceInMem[*goaingest.IngestEvent]()
			stream := &mockMonitorServerStream{failOnSend: tt.failOnSend}
			svc := ingest.NewService(ingest.ServiceParams{
				EventService: evsvc,
			})

			ctx, cancel := context.WithTimeout(t.Context(), 100*time.Millisecond)
			defer cancel()
			ctx = auth.WithUserClaims(ctx, &auth.Claims{Attributes: []string{"*"}})

			errCh := make(chan error, 1)
			go func() {
				errCh <- svc.Monitor(ctx, &goaingest.MonitorPayload{}, stream)
			}()

			if tt.publish != nil {
				time.Sleep(10 * time.Millisecond)
				tt.publish(t.Context(), evsvc)
			}

			select {
			case err := <-errCh:
				assert.NilError(t, err)
			case <-time.After(200 * time.Millisecond):
				t.Fatal("Monitor did not complete in expected time")
			}

			assert.Assert(t, stream.closed)
			assert.DeepEqual(t, stream.events, tt.wantEvents, cmp.AllowUnexported(goaingest.Value{}))
		})
	}
}
