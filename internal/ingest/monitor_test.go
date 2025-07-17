package ingest

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
	goaingest "github.com/artefactual-sdps/enduro/internal/api/gen/ingest"
	"github.com/artefactual-sdps/enduro/internal/event"
)

func TestMonitorRequest(t *testing.T) {
	t.Parallel()

	for _, tt := range []struct {
		name    string
		claims  *auth.Claims
		mock    func(*authfake.MockTicketProvider, context.Context, *auth.Claims)
		want    *goaingest.MonitorRequestResult
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
			want: &goaingest.MonitorRequestResult{Ticket: ref.New("ticket")},
		},
		{
			name: "Returns empty result when no ticket is provided",
			mock: func(tp *authfake.MockTicketProvider, ctx context.Context, claims *auth.Claims) {
				tp.EXPECT().
					Request(ctx, claims).
					Return("", nil)
			},
			want: &goaingest.MonitorRequestResult{},
		},
		{
			name: "Fails when ticket request fails",
			mock: func(tp *authfake.MockTicketProvider, ctx context.Context, claims *auth.Claims) {
				tp.EXPECT().
					Request(ctx, claims).
					Return("", fmt.Errorf("error"))
			},
			wantErr: "cannot perform operation",
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			tpMock := authfake.NewMockTicketProvider(gomock.NewController(t))
			gw := &goaWrapper{
				ingestImpl: &ingestImpl{
					logger:         logr.Discard(),
					ticketProvider: tpMock,
				},
			}

			ctx := auth.WithUserClaims(t.Context(), tt.claims)
			tt.mock(tpMock, ctx, tt.claims)

			res, err := gw.MonitorRequest(ctx, &goaingest.MonitorRequestPayload{})
			if tt.wantErr != "" {
				assert.ErrorContains(t, err, tt.wantErr)
				return
			}
			assert.NilError(t, err)
			assert.DeepEqual(t, res, tt.want)
		})
	}
}

// mockMonitorServerStream implements goaingest.MonitorServerStream for testing.
type mockMonitorServerStream struct {
	events []any
	closed bool
}

func (m *mockMonitorServerStream) Send(event *goaingest.MonitorEvent) error {
	if m.closed {
		return fmt.Errorf("stream closed")
	}
	m.events = append(m.events, event.Event)
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
	allEvents := []*goaingest.MonitorEvent{
		{Event: &goaingest.SIPCreatedEvent{UUID: testUUID}},
		{Event: &goaingest.SIPUpdatedEvent{UUID: testUUID}},
		{Event: &goaingest.SIPStatusUpdatedEvent{UUID: testUUID}},
		{Event: &goaingest.SIPWorkflowCreatedEvent{UUID: testUUID}},
		{Event: &goaingest.SIPWorkflowUpdatedEvent{UUID: testUUID}},
		{Event: &goaingest.SIPTaskCreatedEvent{UUID: testUUID}},
		{Event: &goaingest.SIPTaskUpdatedEvent{UUID: testUUID}},
	}
	allWantEvents := []any{
		&goaingest.MonitorPingEvent{Message: ref.New("Hello")},
		&goaingest.SIPCreatedEvent{UUID: testUUID},
		&goaingest.SIPUpdatedEvent{UUID: testUUID},
		&goaingest.SIPStatusUpdatedEvent{UUID: testUUID},
		&goaingest.SIPWorkflowCreatedEvent{UUID: testUUID},
		&goaingest.SIPWorkflowUpdatedEvent{UUID: testUUID},
		&goaingest.SIPTaskCreatedEvent{UUID: testUUID},
		&goaingest.SIPTaskUpdatedEvent{UUID: testUUID},
	}

	for _, tt := range []struct {
		name       string
		claims     *auth.Claims
		mock       func(*authfake.MockTicketProvider, context.Context, *string, *auth.Claims)
		events     []*goaingest.MonitorEvent
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
				&goaingest.MonitorPingEvent{Message: ref.New("Hello")},
			},
		},
		{
			name: "Filters events based on permissions",
			claims: &auth.Claims{
				Email:         "test@example.com",
				EmailVerified: true,
				Attributes:    []string{auth.IngestSIPSReadAttr},
			},
			mock:   successMock,
			events: allEvents,
			wantEvents: []any{
				&goaingest.MonitorPingEvent{Message: ref.New("Hello")},
				&goaingest.SIPUpdatedEvent{UUID: testUUID},
				&goaingest.SIPStatusUpdatedEvent{UUID: testUUID},
			},
		},
		{
			name: "Fails when ticket check fails",
			mock: func(tp *authfake.MockTicketProvider, ctx context.Context, ticket *string, claims *auth.Claims) {
				tp.EXPECT().
					Check(ctx, ticket, &auth.Claims{}).
					Return(fmt.Errorf("invalid ticket"))
			},
			wantErr: "cannot perform operation",
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			tpMock := authfake.NewMockTicketProvider(gomock.NewController(t))
			evsvc := event.NewEventServiceInMemImpl()
			stream := &mockMonitorServerStream{}

			gw := &goaWrapper{
				ingestImpl: &ingestImpl{
					logger:         logr.Discard(),
					evsvc:          evsvc,
					ticketProvider: tpMock,
				},
			}

			// Create a context that will be cancelled to stop the monitor.
			ctx, cancel := context.WithTimeout(t.Context(), 100*time.Millisecond)
			defer cancel()

			tt.mock(tpMock, ctx, ticket, tt.claims)

			// Start monitor in a goroutine.
			errCh := make(chan error, 1)
			go func() {
				errCh <- gw.Monitor(ctx, &goaingest.MonitorPayload{Ticket: ticket}, stream)
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
