package sipsource

import (
	"context"
	"errors"
	"strings"
	"time"
)

const defaultLimit = 100

var (
	ErrInvalidSource = errors.New("invalid SIP source")
	ErrInvalidToken  = errors.New("invalid token")
)

// SIPSource defines the interface for a SIP source location.
type SIPSource interface {
	// ListObjects returns a paged list of items in the SIP source.
	ListObjects(context.Context, ListOptions) (*Page, error)

	// Close releases resources associated with the SIP source.
	Close() error

	// RetentionPeriod returns the duration for which SIPs should be retained
	// after a successful ingest. If negative, SIPs will be retained indefinitely.
	RetentionPeriod() time.Duration
}

// ListOptions specifies options for listing SIP source objects.
type ListOptions struct {
	// Token is used to retrieve a specific page of objects. If Token is nil,
	// the first page of objects will be returned.
	Token []byte

	// Limit is the maximum number of objects to return. If Limit is less than
	// or equal to zero, the page limit will be set to defaultLimit.
	Limit int

	// Sort specifies the key and direction by which to sort the objects. If
	// Sort is nil, the objects will be returned in the default order provided
	// by the underlying implementation.
	Sort *Sort
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

// Sort specifies the attribute and order by which to sort SIP source objects.
type Sort struct {
	// Attr is the object attribute to sort by.
	attr string
	// Desc indicates whether to sort in descending order.
	desc bool
}

// SortByKey returns a Sort that sorts by the object key.
func SortByKey() *Sort {
	return &Sort{attr: "key"}
}

// SortByModTime returns a Sort that sorts by the object modification time.
func SortByModTime() *Sort {
	return &Sort{attr: "modtime"}
}

// Desc sets the sort order to descending (Z-A).
func (s *Sort) Desc() *Sort {
	s.desc = true
	return s
}

// Asc sets the sort order to ascending (A-Z).
func (s *Sort) Asc() *Sort {
	s.desc = false
	return s
}

// Compare two objects based on the Sort configuration. Compare returns -1 if
// `a` sorts before `b`, 0 if `a` and `b` are equal in sort order, and 1 if `a`
// sorts after `b`. If Sort is nil or has no attribute set, Compare returns 0.
func (s *Sort) Compare(a, b *Object) int {
	if s == nil || s.attr == "" {
		return 0
	}

	var r int
	switch s.attr {
	case "key":
		r = strings.Compare(a.Key, b.Key)
	case "modtime":
		r = a.ModTime.Compare(b.ModTime)
	default:
		return 0
	}

	if s.desc {
		return -r
	}
	return r
}
