package sipsource

import (
	"context"
	"fmt"
	"io"

	"github.com/google/uuid"
	"go.artefactual.dev/tools/bucket"
	"gocloud.dev/blob"
)

const defaultLimit = 100

type SIPSource struct {
	// ID is the unique identifier for the SIP source.
	ID uuid.UUID
	// Bucket is the bucket where SIPs are stored.
	Bucket *blob.Bucket
	// Name is the human-readable name of the SIP source.
	Name string
}

// NewWithConfig creates a new SIPSource from the provided configuration.
func NewWithConfig(ctx context.Context, cfg *Config) (*SIPSource, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	bucket, err := bucket.NewWithConfig(ctx, cfg.Bucket)
	if err != nil {
		return nil, fmt.Errorf("SIP source: create bucket: %w", err)
	}

	return &SIPSource{
		ID:     cfg.ID,
		Bucket: bucket,
		Name:   cfg.Name,
	}, nil
}

// Close closes the underlying SIP Source bucket. It should be called when the
// SIP source is no longer needed to release resources.
func (s *SIPSource) Close() error {
	if s.Bucket != nil {
		if err := s.Bucket.Close(); err != nil {
			return fmt.Errorf("SIP source: close bucket: %w", err)
		}
	}
	return nil
}

// ListItems returns a paged list of items in the SIP source bucket.
//
// If the token parameter is nil, the first page of items will be returned.
// Subsequent calls to ListItems should provide the NextToken from the previous
// response to retrieve the next page.
//
// The limit parameter specifies the maximum number of items to return. If limit
// is less than or equal to zero, the page limit will be set to defaultLimit.
//
// If the paged query returns no items (e.g. the items were deleted) ListItems
// returns a nil Page.
//
// If the source bucket is not configured properly or can not be accessed,
// ListItems returns an ErrInvalidBucket error.
func (s *SIPSource) ListItems(ctx context.Context, token []byte, limit int) (*Page, error) {
	if s.Bucket == nil {
		return nil, ErrInvalidBucket
	}
	if token == nil {
		token = blob.FirstPageToken
	}
	if limit <= 0 {
		limit = defaultLimit
	}

	items, next, err := s.Bucket.ListPage(ctx, token, limit, nil)
	if err != nil {
		if err == io.EOF {
			return nil, nil // No more items to list
		}
		return nil, fmt.Errorf("SIP source: list items: %w", err)
	}

	return &Page{Items: items, Limit: limit, NextToken: next}, nil
}

// Page is a paginated list of SIP source items.
type Page struct {
	// Items is the current page of SIP source items.
	Items []*blob.ListObject

	// Limit is the maximum number of items returned per page.
	Limit int

	// NextToken is used retrieve the next page of items. If NextToken is nil
	// there are no more items to list.
	NextToken []byte
}
