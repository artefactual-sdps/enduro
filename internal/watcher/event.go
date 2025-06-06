package watcher

import (
	"fmt"
	"time"

	"github.com/artefactual-sdps/enduro/internal/enums"
)

// BlobEvent is a serializable event that describes a blob.
//
// BlobEvent can be sent over the wire, i.e. they're serializable. Receivers,
// typicially Temporal activities, can download the blob via the service
// implementation in this package.
//
// TODO: use signed URLs to simplify access to buckets?
type BlobEvent struct {
	// Name of the watcher that received this blob.
	WatcherName string

	// Retention period for this blob.
	RetentionPeriod *time.Duration

	// Directory where the transfer is moved to once processing has completed
	// successfully.
	CompletedDir string

	// Key of the blob.
	Key string

	// Whether the blob is a directory (fs watcher)
	IsDir bool

	// Bucket where the blob lives.
	Bucket string `json:"Bucket,omitempty"`

	// Type of workflow to execute.
	WorkflowType enums.WorkflowType
}

func NewBlobEvent(w Watcher, key string, isDir bool) *BlobEvent {
	return &BlobEvent{
		WatcherName:     w.String(),
		RetentionPeriod: w.RetentionPeriod(),
		CompletedDir:    w.CompletedDir(),
		WorkflowType:    w.WorkflowType(),
		Key:             key,
		IsDir:           isDir,
	}
}

func NewBlobEventWithBucket(w Watcher, bucket, key string) *BlobEvent {
	e := NewBlobEvent(w, key, false)
	e.Bucket = bucket
	return e
}

func (e BlobEvent) String() string {
	key := e.Key

	if e.Bucket != "" {
		key = fmt.Sprintf("%s:%s", e.Bucket, key)
	}

	return fmt.Sprintf("%q:%q", e.WatcherName, key)
}
