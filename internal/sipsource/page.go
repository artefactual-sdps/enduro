package sipsource

import (
	"time"
)

// Page is a paginated list of SIP source items.
type Page struct {
	// Items is the current page of SIP source items.
	Items []*Item

	// Limit is the maximum number of items returned per page.
	Limit int

	// NextToken is used retrieve the next page of items. If NextToken is nil
	// there are no more items to list.
	NextToken []byte
}

type Item struct {
	// Key is the unique identifier for the item.
	Key string

	// ModTime is the last modification time of the item.
	ModTime time.Time

	// Size is the size of the item in bytes.
	Size int64

	// IsDir indicates whether the item is a directory.
	IsDir bool
}
