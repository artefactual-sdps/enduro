package storage

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/go-logr/logr"
	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	"gotest.tools/v3/assert"

	goastorage "github.com/artefactual-sdps/enduro/internal/api/gen/storage"
	"github.com/artefactual-sdps/enduro/internal/auth"
	"github.com/artefactual-sdps/enduro/internal/event"
)

// mockMonitorServerStream implements goastorage.MonitorServerStream for testing.
type mockMonitorServerStream struct {
	events     []*goastorage.StorageEvent
	closed     bool
	failOnSend int
}

func (m *mockMonitorServerStream) Send(event *goastorage.StorageEvent) error {
	return m.SendWithContext(context.Background(), event)
}

func (m *mockMonitorServerStream) SendWithContext(ctx context.Context, event *goastorage.StorageEvent) error {
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
	allEvents := []*goastorage.StorageEvent{
		{Value: NewEventValue(&goastorage.LocationCreatedEvent{UUID: testUUID})},
		{Value: NewEventValue(&goastorage.AIPCreatedEvent{UUID: testUUID})},
		{Value: NewEventValue(&goastorage.AIPUpdatedEvent{UUID: testUUID})},
		{Value: NewEventValue(&goastorage.AIPStatusUpdatedEvent{UUID: testUUID})},
		{Value: NewEventValue(&goastorage.AIPLocationUpdatedEvent{UUID: testUUID})},
		{Value: NewEventValue(&goastorage.AIPWorkflowCreatedEvent{UUID: testUUID})},
		{Value: NewEventValue(&goastorage.AIPWorkflowUpdatedEvent{UUID: testUUID})},
		{Value: NewEventValue(&goastorage.AIPTaskCreatedEvent{UUID: testUUID})},
		{Value: NewEventValue(&goastorage.AIPTaskUpdatedEvent{UUID: testUUID})},
	}
	allWantEvents := []*goastorage.StorageEvent{
		{Value: NewEventValue(&goastorage.StoragePingEvent{Message: new("Hello")})},
		{Value: NewEventValue(&goastorage.LocationCreatedEvent{UUID: testUUID})},
		{Value: NewEventValue(&goastorage.AIPCreatedEvent{UUID: testUUID})},
		{Value: NewEventValue(&goastorage.AIPUpdatedEvent{UUID: testUUID})},
		{Value: NewEventValue(&goastorage.AIPStatusUpdatedEvent{UUID: testUUID})},
		{Value: NewEventValue(&goastorage.AIPLocationUpdatedEvent{UUID: testUUID})},
		{Value: NewEventValue(&goastorage.AIPWorkflowCreatedEvent{UUID: testUUID})},
		{Value: NewEventValue(&goastorage.AIPWorkflowUpdatedEvent{UUID: testUUID})},
		{Value: NewEventValue(&goastorage.AIPTaskCreatedEvent{UUID: testUUID})},
		{Value: NewEventValue(&goastorage.AIPTaskUpdatedEvent{UUID: testUUID})},
	}

	for _, tt := range []struct {
		name       string
		claims     *auth.Claims
		events     []*goastorage.StorageEvent
		wantEvents []*goastorage.StorageEvent
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
			wantEvents: []*goastorage.StorageEvent{
				{Value: NewEventValue(&goastorage.StoragePingEvent{Message: new("Hello")})},
			},
		},
		{
			name: "Filters events based on permissions",
			claims: &auth.Claims{
				Email:         "test@example.com",
				EmailVerified: true,
				Attributes:    []string{auth.StorageLocationsListAttr, auth.StorageAIPSReadAttr},
			},
			events: allEvents,
			wantEvents: []*goastorage.StorageEvent{
				{Value: NewEventValue(&goastorage.StoragePingEvent{Message: new("Hello")})},
				{Value: NewEventValue(&goastorage.LocationCreatedEvent{UUID: testUUID})},
				{Value: NewEventValue(&goastorage.AIPUpdatedEvent{UUID: testUUID})},
				{Value: NewEventValue(&goastorage.AIPStatusUpdatedEvent{UUID: testUUID})},
				{Value: NewEventValue(&goastorage.AIPLocationUpdatedEvent{UUID: testUUID})},
			},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			evsvc := event.NewServiceInMem[*goastorage.StorageEvent]()
			stream := &mockMonitorServerStream{}

			svc := &serviceImpl{
				logger: logr.Discard(),
				evsvc:  evsvc,
			}

			// Create a context that will be cancelled to stop the monitor.
			ctx, cancel := context.WithTimeout(t.Context(), 100*time.Millisecond)
			defer cancel()
			ctx = auth.WithUserClaims(ctx, tt.claims)

			// Start monitor in a goroutine.
			errCh := make(chan error, 1)
			go func() {
				errCh <- svc.Monitor(ctx, &goastorage.MonitorPayload{}, stream)
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

			assert.DeepEqual(t, stream.events, tt.wantEvents, cmp.AllowUnexported(goastorage.Value{}))
		})
	}
}

func TestMonitorReturnsNilOnStreamSendError(t *testing.T) {
	t.Parallel()

	testUUID := uuid.New()

	for _, tt := range []struct {
		name       string
		failOnSend int
		publish    func(context.Context, event.Service[*goastorage.StorageEvent])
		wantEvents []*goastorage.StorageEvent
	}{
		{
			name:       "Hello",
			failOnSend: 1,
		},
		{
			name:       "Subscribed event",
			failOnSend: 2,
			publish: func(ctx context.Context, evsvc event.Service[*goastorage.StorageEvent]) {
				evsvc.PublishEvent(ctx, &goastorage.StorageEvent{
					Value: NewEventValue(&goastorage.LocationCreatedEvent{UUID: testUUID}),
				})
			},
			wantEvents: []*goastorage.StorageEvent{
				{Value: NewEventValue(&goastorage.StoragePingEvent{Message: new("Hello")})},
			},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			evsvc := event.NewServiceInMem[*goastorage.StorageEvent]()
			stream := &mockMonitorServerStream{failOnSend: tt.failOnSend}
			svc := &serviceImpl{
				logger: logr.Discard(),
				evsvc:  evsvc,
			}

			ctx, cancel := context.WithTimeout(t.Context(), 100*time.Millisecond)
			defer cancel()
			ctx = auth.WithUserClaims(ctx, &auth.Claims{Attributes: []string{"*"}})

			errCh := make(chan error, 1)
			go func() {
				errCh <- svc.Monitor(ctx, &goastorage.MonitorPayload{}, stream)
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
			assert.DeepEqual(t, stream.events, tt.wantEvents, cmp.AllowUnexported(goastorage.Value{}))
		})
	}
}
