package watcher

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/go-logr/logr"
	"go.opentelemetry.io/otel/trace"
	"gocloud.dev/blob"
)

var (
	ErrWatchTimeout   = errors.New("watcher timed out")
	ErrBucketMismatch = errors.New("bucket mismatch")
)

type Cleanup func(ctx context.Context) error

func noopCleanup(ctx context.Context) error {
	return nil
}

type Watcher interface {
	// Watch waits until a blob is dispatched.
	// After the event is successfully processed by the receiver, the returned
	// Cleanup function must be executed to mark the event as processed.
	// Implementors must not return a nil function.
	Watch(ctx context.Context) (*BlobEvent, Cleanup, error)

	// Download copies the file or directory identified by key to dest.
	Download(ctx context.Context, dest, key string) error

	// OpenBucket returns the bucket where the blobs can be found.
	OpenBucket(ctx context.Context) (*blob.Bucket, error)

	RetentionPeriod() *time.Duration
	CompletedDir() string
	StripTopLevelDir() bool

	// Full path of the watched bucket when available, empty string otherwise.
	Path() string

	fmt.Stringer // It should return the name of the watcher.
}

type commonWatcherImpl struct {
	name             string
	retentionPeriod  *time.Duration
	completedDir     string
	stripTopLevelDir bool
}

func (w *commonWatcherImpl) String() string {
	return w.name
}

func (w *commonWatcherImpl) RetentionPeriod() *time.Duration {
	return w.retentionPeriod
}

func (w *commonWatcherImpl) CompletedDir() string {
	return w.completedDir
}

func (w *commonWatcherImpl) StripTopLevelDir() bool {
	return w.stripTopLevelDir
}

type Service interface {
	// Watchers return all known watchers.
	Watchers() []Watcher

	// Return a watcher given its name.
	ByName(name string) (Watcher, error)

	// Download copies the watcherName file or directory identified by key to
	// dest.
	Download(ctx context.Context, dest, watcherName, key string) error

	// Delete blob given an event.
	Delete(ctx context.Context, watcherName, key string) error

	// Dipose blob into the completedDir directory.
	Dispose(ctx context.Context, watcherName, key string) error
}

type serviceImpl struct {
	watchers map[string]Watcher
	mu       sync.RWMutex
}

var _ Service = (*serviceImpl)(nil)

func New(ctx context.Context, tp trace.TracerProvider, logger logr.Logger, c *Config) (*serviceImpl, error) {
	watchers := map[string]Watcher{}
	minioConfigs := append(c.Minio, c.Embedded)

	for _, item := range minioConfigs {
		w, err := NewMinioWatcher(ctx, tp, logger, item)
		if err != nil {
			return nil, err
		}

		watchers[item.Name] = w
	}

	for _, item := range c.Filesystem {
		w, err := NewFilesystemWatcher(ctx, item)
		if err != nil {
			return nil, err
		}

		watchers[item.Name] = w
	}

	if len(watchers) == 0 {
		return nil, errors.New("there are not watchers configured")
	}

	return &serviceImpl{watchers: watchers}, nil
}

func (svc *serviceImpl) Watchers() []Watcher {
	svc.mu.RLock()
	defer svc.mu.RUnlock()

	ww := []Watcher{}
	for _, item := range svc.watchers {
		ww = append(ww, item)
	}

	return ww
}

func (svc *serviceImpl) watcher(name string) (Watcher, error) {
	svc.mu.RLock()
	defer svc.mu.RUnlock()

	w, ok := svc.watchers[name]
	if !ok {
		return nil, fmt.Errorf("error loading watcher: unknown watcher %s", name)
	}

	return w, nil
}

func (svc *serviceImpl) ByName(name string) (Watcher, error) {
	return svc.watcher(name)
}

func (svc *serviceImpl) Download(ctx context.Context, dest, watcherName, key string) error {
	w, err := svc.watcher(watcherName)
	if err != nil {
		return err
	}

	return w.Download(ctx, dest, key)
}

func (svc *serviceImpl) Delete(ctx context.Context, watcherName, key string) error {
	w, err := svc.watcher(watcherName)
	if err != nil {
		return err
	}

	bucket, err := w.OpenBucket(ctx)
	if err != nil {
		return fmt.Errorf("error opening bucket: %w", err)
	}
	defer bucket.Close()

	// Exceptionally, a filesystem-based watcher may be dealing with a
	// directory instead of a regular fileblob.
	var fi os.FileInfo
	if bucket.As(&fi) && fi.IsDir() {
		fw, ok := w.(*filesystemWatcher)
		if !ok {
			return fmt.Errorf("error removing directory: %s", err)
		}
		return fw.RemoveAll(key)
	}

	return bucket.Delete(ctx, key)
}

func (svc *serviceImpl) Dispose(ctx context.Context, watcherName, key string) error {
	w, err := svc.watcher(watcherName)
	if err != nil {
		return err
	}

	fw, ok := w.(*filesystemWatcher)
	if !ok {
		return fmt.Errorf("not available in this type of watcher: %s", err)
	}

	return fw.Dispose(key)
}
