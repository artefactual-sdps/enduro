package sipsource

import (
	"context"
	"time"
)

// SIPSource defines the interface for a SIP source location.
type SIPSource interface {
	// ListObjects returns a paged list of items in the SIP source.
	ListObjects(ctx context.Context, token []byte, limit int) (*Page, error)

	// Close releases resources associated with the SIP source.
	Close() error

	// RetentionPeriod returns the duration for which SIPs should be retained
	// after a successful ingest. If negative, SIPs will be retained indefinitely.
	RetentionPeriod() time.Duration
}

// Page is a paginated list of SIP source objects.
type Page struct {
	// Objects is the current page of SIP source objects.
	Objects []*Object

	// Limit is the maximum number of objects returned per page.
	Limit int

	// NextToken is used retrieve the next page of objects. If NextToken is nil
	// there are no more objects to list.
	NextToken []byte
}

// Object represents a single object in the SIP source.
type Object struct {
	// Key is the unique identifier for the object.
	Key string

	// ModTime is the last modification time of the object.
	ModTime time.Time

	// Size is the size of the object in bytes.
	Size int64

	// IsDir indicates whether the object is a directory.
	IsDir bool
}
