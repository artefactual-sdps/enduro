package storage

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/go-logr/logr"
	"github.com/google/uuid"
	"go.artefactual.dev/tools/ref"
	"go.uber.org/mock/gomock"
	"gotest.tools/v3/assert"

	"github.com/artefactual-sdps/enduro/internal/api/auth"
	authfake "github.com/artefactual-sdps/enduro/internal/api/auth/fake"
	goastorage "github.com/artefactual-sdps/enduro/internal/api/gen/storage"
	"github.com/artefactual-sdps/enduro/internal/event"
)

func TestMonitorRequest(t *testing.T) {
	t.Parallel()

	for _, tt := range []struct {
		name    string
		claims  *auth.Claims
		mock    func(*authfake.MockTicketProvider, context.Context, *auth.Claims)
		want    *goastorage.MonitorRequestResult
		wantErr string
	}{
		{
			name: "Returns ticket when available",
			claims: &auth.Claims{
				Email:         "test@example.com",
				EmailVerified: true,
				Attributes:    []string{"*"},
			},
			mock: func(tp *authfake.MockTicketProvider, ctx context.Context, claims *auth.Claims) {
				tp.EXPECT().
					Request(ctx, claims).
					Return("ticket", nil)
			},
			want: &goastorage.MonitorRequestResult{Ticket: ref.New("ticket")},
		},
		{
			name: "Returns empty result when no ticket is provided",
			mock: func(tp *authfake.MockTicketProvider, ctx context.Context, claims *auth.Claims) {
				tp.EXPECT().
					Request(ctx, claims).
					Return("", nil)
			},
			want: &goastorage.MonitorRequestResult{},
		},
		{
			name: "Fails when ticket request fails",
			mock: func(tp *authfake.MockTicketProvider, ctx context.Context, claims *auth.Claims) {
				tp.EXPECT().
					Request(ctx, claims).
					Return("", fmt.Errorf("error"))
			},
			wantErr: "internal error",
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			tpMock := authfake.NewMockTicketProvider(gomock.NewController(t))
			svc := &serviceImpl{
				logger:         logr.Discard(),
				ticketProvider: tpMock,
			}

			ctx := auth.WithUserClaims(t.Context(), tt.claims)
			tt.mock(tpMock, ctx, tt.claims)

			res, err := svc.MonitorRequest(ctx, &goastorage.MonitorRequestPayload{})
			if tt.wantErr != "" {
				assert.ErrorContains(t, err, tt.wantErr)
				return
			}
			assert.NilError(t, err)
			assert.DeepEqual(t, res, tt.want)
		})
	}
}

// mockMonitorServerStream implements goastorage.MonitorServerStream for testing.
type mockMonitorServerStream struct {
	events []any
	closed bool
}

func (m *mockMonitorServerStream) Send(event *goastorage.StorageEvent) error {
	if m.closed {
		return fmt.Errorf("stream closed")
	}
	m.events = append(m.events, event.StorageValue)
	return nil
}

func (m *mockMonitorServerStream) Close() error {
	m.closed = true
	return nil
}

func TestMonitor(t *testing.T) {
	t.Parallel()

	testUUID := uuid.New()
	ticket := ref.New("ticket")
	successMock := func(tp *authfake.MockTicketProvider, ctx context.Context, ticket *string, claims *auth.Claims) {
		tp.EXPECT().
			Check(ctx, ticket, &auth.Claims{}).
			DoAndReturn(func(ctx context.Context, ticket *string, dst any) error {
				if c, ok := dst.(*auth.Claims); ok {
					*c = *claims
				}
				return nil
			})
	}
	allEvents := []*goastorage.StorageEvent{
		{StorageValue: &goastorage.LocationCreatedEvent{UUID: testUUID}},
		{StorageValue: &goastorage.AIPCreatedEvent{UUID: testUUID}},
		{StorageValue: &goastorage.AIPStatusUpdatedEvent{UUID: testUUID}},
		{StorageValue: &goastorage.AIPLocationUpdatedEvent{UUID: testUUID}},
		{StorageValue: &goastorage.AIPWorkflowCreatedEvent{UUID: testUUID}},
		{StorageValue: &goastorage.AIPWorkflowUpdatedEvent{UUID: testUUID}},
		{StorageValue: &goastorage.AIPTaskCreatedEvent{UUID: testUUID}},
		{StorageValue: &goastorage.AIPTaskUpdatedEvent{UUID: testUUID}},
	}
	allWantEvents := []any{
		&goastorage.StoragePingEvent{Message: ref.New("Hello")},
		&goastorage.LocationCreatedEvent{UUID: testUUID},
		&goastorage.AIPCreatedEvent{UUID: testUUID},
		&goastorage.AIPStatusUpdatedEvent{UUID: testUUID},
		&goastorage.AIPLocationUpdatedEvent{UUID: testUUID},
		&goastorage.AIPWorkflowCreatedEvent{UUID: testUUID},
		&goastorage.AIPWorkflowUpdatedEvent{UUID: testUUID},
		&goastorage.AIPTaskCreatedEvent{UUID: testUUID},
		&goastorage.AIPTaskUpdatedEvent{UUID: testUUID},
	}

	for _, tt := range []struct {
		name       string
		claims     *auth.Claims
		mock       func(*authfake.MockTicketProvider, context.Context, *string, *auth.Claims)
		events     []*goastorage.StorageEvent
		wantEvents []any
		wantErr    string
	}{
		{
			name: "Sends all events for a user with all permissions",
			claims: &auth.Claims{
				Email:         "test@example.com",
				EmailVerified: true,
				Attributes:    []string{"*"},
			},
			mock:       successMock,
			events:     allEvents,
			wantEvents: allWantEvents,
		},
		{
			name:       "Sends all events when authentication and/or ABAC is disabled",
			claims:     &auth.Claims{},
			mock:       successMock,
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
			mock:   successMock,
			events: allEvents,
			wantEvents: []any{
				&goastorage.StoragePingEvent{Message: ref.New("Hello")},
			},
		},
		{
			name: "Filters events based on permissions",
			claims: &auth.Claims{
				Email:         "test@example.com",
				EmailVerified: true,
				Attributes:    []string{auth.StorageLocationsListAttr, auth.StorageAIPSReadAttr},
			},
			mock:   successMock,
			events: allEvents,
			wantEvents: []any{
				&goastorage.StoragePingEvent{Message: ref.New("Hello")},
				&goastorage.LocationCreatedEvent{UUID: testUUID},
				&goastorage.AIPStatusUpdatedEvent{UUID: testUUID},
				&goastorage.AIPLocationUpdatedEvent{UUID: testUUID},
			},
		},
		{
			name: "Fails when ticket check fails",
			mock: func(tp *authfake.MockTicketProvider, ctx context.Context, ticket *string, claims *auth.Claims) {
				tp.EXPECT().
					Check(ctx, ticket, &auth.Claims{}).
					Return(fmt.Errorf("invalid ticket"))
			},
			wantErr: "internal error",
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			tpMock := authfake.NewMockTicketProvider(gomock.NewController(t))
			evsvc := event.NewServiceInMem[*goastorage.StorageEvent]()
			stream := &mockMonitorServerStream{}

			svc := &serviceImpl{
				logger:         logr.Discard(),
				evsvc:          evsvc,
				ticketProvider: tpMock,
			}

			// Create a context that will be cancelled to stop the monitor.
			ctx, cancel := context.WithTimeout(t.Context(), 100*time.Millisecond)
			defer cancel()

			tt.mock(tpMock, ctx, ticket, tt.claims)

			// Start monitor in a goroutine.
			errCh := make(chan error, 1)
			go func() {
				errCh <- svc.Monitor(ctx, &goastorage.MonitorPayload{Ticket: ticket}, stream)
			}()

			// Send test events after a short delay.
			time.Sleep(10 * time.Millisecond)
			for _, event := range tt.events {
				evsvc.PublishEvent(t.Context(), event)
			}

			// Wait for the monitor to finish.
			select {
			case err := <-errCh:
				if tt.wantErr != "" {
					assert.ErrorContains(t, err, tt.wantErr)
					return
				}
				assert.NilError(t, err)
			case <-time.After(200 * time.Millisecond):
				t.Fatal("Monitor did not complete in expected time")
			}

			assert.DeepEqual(t, stream.events, tt.wantEvents)
		})
	}
}
