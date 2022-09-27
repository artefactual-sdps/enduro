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

	// Tickets are prefixed to allow for sharing a space with other instances.
	prefix string

	// Tickets will be considered expired when ttl is exceeded.
	ttl time.Duration

	// The source of randomness used in the ticket generator.
	rander io.Reader
}

// NewTicketProvider creates a new TicketProvider.
func NewTicketProvider(ctx context.Context, store TicketStore, prefix string, rander io.Reader) *TicketProvider {
	if store == nil {
		return &TicketProvider{}
	}

	if rander == nil {
		rander = rand.Reader
	}

	return &TicketProvider{
		store:  store,
		prefix: prefix,
		ttl:    TicketTTL,
		rander: rander,
	}
}

// Request a new ticket.
func (t *TicketProvider) Request(ctx context.Context) (string, error) {
	if t.store == nil {
		return "", nil
	}

	ticket, err := t.ticket()
	if err != nil {
		return "", fmt.Errorf("error creating ticket: %v", err)
	}

	err = t.store.SetEX(ctx, t.storeKey(ticket), t.ttl)
	if err != nil {
		return "", fmt.Errorf("error storing ticket: %v", err)
	}

	return ticket, nil
}

// Check that a ticket is known to the provider, not including tickets that
// exceeded the time-to-live attribute.
func (t *TicketProvider) Check(ctx context.Context, ticket string) error {
	if t.store == nil {
		return nil
	}

	err := t.store.GetDel(ctx, t.storeKey(ticket))
	if err != nil {
		return fmt.Errorf("error retrieving ticket: %v", err)
	}

	return nil
}

func (t TicketProvider) storeKey(ticket string) string {
	return t.prefix + ":session:" + ticket
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
