package sipsource

import "context"

type SIPSource interface {
	// ListItems returns a paged list of items in the SIP source.
	ListItems(ctx context.Context, token []byte, limit int) (*Page, error)

	// Close releases resources associated with the SIP source.
	Close() error
}
