package sipsource

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"go.artefactual.dev/tools/bucket"
	"gocloud.dev/blob"
)

const defaultLimit = 100

// BucketSource represents a SIP source that uses a cloud storage bucket
// to store SIPs. It implements the SIPSource interface.
type BucketSource struct {
	// ID is the unique identifier for the SIP source.
	ID uuid.UUID
	// Bucket is the bucket where SIPs are stored.
	Bucket *blob.Bucket
	// Name is the human-readable name of the SIP source.
	Name string
}

var _ SIPSource = (*BucketSource)(nil)

// NewBucketSource creates a new BucketSource from the provided configuration.
func NewBucketSource(ctx context.Context, cfg *Config) (*BucketSource, error) {
	if cfg.IsEmpty() {
		// Return an empty BucketSource if the configuration is empty.
		return &BucketSource{}, nil
	}

	bucket, err := bucket.NewWithConfig(ctx, cfg.Bucket)
	if err != nil {
		return nil, fmt.Errorf("SIP source: new bucket source: %w", err)
	}

	return &BucketSource{
		ID:     cfg.ID,
		Bucket: bucket,
		Name:   cfg.Name,
	}, nil
}

// Close closes the underlying bucket. It should be called when the
// BucketSource is no longer needed to release resources.
func (s *BucketSource) Close() error {
	if s.Bucket != nil {
		if err := s.Bucket.Close(); err != nil {
			return fmt.Errorf("SIP bucket source: close bucket: %w", err)
		}
	}
	return nil
}

// ListObjects returns a paged list of items in the SIP source bucket.
//
// If the token parameter is nil, the first page of items will be returned.
// Subsequent calls to ListObjects should provide the NextToken from the previous
// response to retrieve the next page.
//
// The limit parameter specifies the maximum number of items to return. If limit
// is less than or equal to zero, the page limit will be set to defaultLimit.
//
// If the paged query returns no items (e.g. the items were deleted) ListObjects
// returns a nil Page.
//
// If the source bucket is not configured properly or can not be accessed,
// ListObjects returns an ErrInvalidBucket error.
func (s *BucketSource) ListObjects(ctx context.Context, token []byte, limit int) (*Page, error) {
	if s.Bucket == nil {
		return nil, ErrMissingBucket
	}
	if token == nil {
		token = blob.FirstPageToken
	}
	if limit <= 0 {
		limit = defaultLimit
	}

	r, next, err := s.Bucket.ListPage(ctx, token, limit, nil)
	if err != nil {
		return nil, fmt.Errorf("SIP bucket source: list items: %w", err)
	}

	objects := make([]*Object, len(r))
	for i, obj := range r {
		objects[i] = &Object{
			Key:     obj.Key,
			ModTime: obj.ModTime,
			Size:    obj.Size,
			IsDir:   obj.IsDir,
		}
	}

	return &Page{Objects: objects, Limit: limit, NextToken: next}, nil
}
