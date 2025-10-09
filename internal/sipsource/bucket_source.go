package sipsource

import (
	"context"
	"fmt"
	"io"
	"slices"
	"time"

	"github.com/google/uuid"
	"go.artefactual.dev/tools/bucket"
	"gocloud.dev/blob"
)

// BucketSource represents a SIP source that uses a cloud storage bucket
// to store SIPs. It implements the SIPSource interface.
type BucketSource struct {
	// ID is the unique identifier for the SIP source.
	ID uuid.UUID
	// Bucket is the bucket where SIPs are stored.
	Bucket *blob.Bucket
	// Name is the human-readable name of the SIP source.
	Name string
	// retentionPeriod is the duration for which SIPs should be retained after
	// a successful ingest. If negative, SIPs will be retained indefinitely.
	retentionPeriod time.Duration
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
		ID:              cfg.ID,
		Bucket:          bucket,
		Name:            cfg.Name,
		retentionPeriod: cfg.RetentionPeriod,
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

// ListObjects returns a paged list of items in the SIP source bucket with the
// provided options.
//
// See the sipsource.ListOptions comments for details on how to use the options.
//
// If the source bucket is not configured properly or can not be accessed,
// ListObjects returns an ErrInvalidSource error. If an invalid page token is
// provided, ListObjects returns an ErrInvalidToken error. If the query returns
// no items (e.g. there are no more results) ListObjects returns a nil Page.
//
// The current implementation retrieves all objects from the bucket, sorts them
// in memory, and then paginates the results. This solution will not scale for
// buckets with a very large number of objects, but it is sufficient for the
// current use cases.
func (s *BucketSource) ListObjects(ctx context.Context, opts ListOptions) (*Page, error) {
	if s.Bucket == nil {
		return nil, ErrInvalidSource
	}
	if opts.Limit <= 0 {
		opts.Limit = defaultLimit
	}

	// Get all the objects in the bucket.
	var objects []*Object
	iter := s.Bucket.List(nil)
	for {
		i, err := iter.Next(ctx)
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("SIP bucket source: list objects: %w", err)
		}
		objects = append(objects, &Object{
			Key:     i.Key,
			ModTime: i.ModTime,
			Size:    i.Size,
			IsDir:   i.IsDir,
		})
	}

	// Sort the objects if a sort is specified.
	if opts.Sort != nil {
		slices.SortFunc(objects, opts.Sort.Compare)
	}

	// Find the index of the first object after the token object.
	first := 0
	if opts.Token != nil {
		index := slices.IndexFunc(objects, func(o *Object) bool {
			return o.Key == string(opts.Token)
		})

		// If the token is not found return an error.
		if index == -1 {
			return nil, ErrInvalidToken
		}

		// If the token is the last object, return a nil page (no more results).
		if index == len(objects)-1 {
			return nil, nil
		}

		first = index + 1
	}

	// Limit the results to the max page size.
	var page []*Object
	if first+opts.Limit > len(objects) {
		page = objects[first:]
	} else {
		page = objects[first : first+opts.Limit]
	}

	// The next token is the key of the last object of the page.
	next := []byte(page[len(page)-1].Key)

	return &Page{Objects: page, Limit: opts.Limit, NextToken: next}, nil
}

func (s *BucketSource) RetentionPeriod() time.Duration {
	return s.retentionPeriod
}
