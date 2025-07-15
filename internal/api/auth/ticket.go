package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"time"
)

const TicketTTL = time.Second * 5

// TicketProvider issues WebSocket authentication tickets.
type TicketProvider struct {
	// Internal store used to persist tickets. When nil, the provider is no-op.
	store TicketStore

	// Tickets will be considered expired when ttl is exceeded.
	ttl time.Duration

	// The source of randomness used in the ticket generator.
	rander io.Reader
}

// NewTicketProvider creates a new TicketProvider. The provider is no-op when
// the store is nil.
func NewTicketProvider(ctx context.Context, store TicketStore, rander io.Reader) *TicketProvider {
	if store == nil {
		return &TicketProvider{}
	}

	if rander == nil {
		rander = rand.Reader
	}

	return &TicketProvider{
		store:  store,
		ttl:    TicketTTL,
		rander: rander,
	}
}

// Request a new ticket saving the key/value pair in the store.
func (t *TicketProvider) Request(ctx context.Context, value any) (string, error) {
	if t.store == nil {
		return "", nil
	}

	ticket, err := t.ticket()
	if err != nil {
		return "", fmt.Errorf("error creating ticket: %v", err)
	}

	err = t.store.SetEx(ctx, ticket, value, t.ttl)
	if err != nil {
		return "", fmt.Errorf("error storing ticket: %v", err)
	}

	return ticket, nil
}

// Check that a ticket is known to the provider and scan its value,
// not including tickets that exceeded the time-to-live attribute.
func (t *TicketProvider) Check(ctx context.Context, ticket *string, value any) error {
	if t.store == nil {
		return nil
	}

	if ticket == nil {
		return fmt.Errorf("missing ticket to retrieve")
	}

	err := t.store.GetDel(ctx, *ticket, value)
	if err != nil {
		return fmt.Errorf("error retrieving ticket: %v", err)
	}

	return nil
}

func (t TicketProvider) ticket() (string, error) {
	b := make([]byte, 32)
	_, err := t.rander.Read(b)
	if err != nil {
		return "", err
	}

	return base64.RawURLEncoding.EncodeToString(b), nil
}

// Close closes the provider, releasing resources associated to the store.
func (t *TicketProvider) Close() error {
	if t.store == nil {
		return nil
	}

	return t.store.Close()
}
