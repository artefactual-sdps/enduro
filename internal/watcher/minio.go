package watcher

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"
	"time"

	"github.com/go-logr/logr"
	"github.com/redis/go-redis/extra/redisotel/v9"
	"github.com/redis/go-redis/v9"
	"go.artefactual.dev/tools/bucket"
	"go.opentelemetry.io/otel/trace"
	"gocloud.dev/blob"
)

// minioWatcher implements a Watcher for watching lists in Redis.
type minioWatcher struct {
	client       redis.UniversalClient
	logger       logr.Logger
	listName     string
	failedList   string
	pollInterval time.Duration
	bucketConfig *bucket.Config
	*commonWatcherImpl
}

type MinioEventSet struct {
	Event     []MinioEvent
	EventTime string
}

var _ Watcher = (*minioWatcher)(nil)

func NewMinioWatcher(
	ctx context.Context,
	tp trace.TracerProvider,
	logger logr.Logger,
	config *MinioConfig,
) (*minioWatcher, error) {
	opts, err := redis.ParseURL(config.RedisAddress)
	if err != nil {
		return nil, err
	}

	client := redis.NewClient(opts)
	if err := redisotel.InstrumentTracing(
		client,
		redisotel.WithTracerProvider(tp),
		redisotel.WithDBStatement(false),
	); err != nil {
		return nil, fmt.Errorf("instrument redis client tracing: %v", err)
	}

	bucketConfig := &bucket.Config{
		Endpoint:  config.Endpoint,
		Bucket:    config.Bucket,
		AccessKey: config.Key,
		SecretKey: config.Secret,
		Token:     config.Token,
		Profile:   config.Profile,
		Region:    config.Region,
		PathStyle: config.PathStyle,
		URL:       config.URL,
	}

	if config.RedisFailedList == "" {
		config.RedisFailedList = config.RedisList + "-failed"
	}

	pollInterval := config.PollInterval
	if pollInterval == 0 {
		pollInterval = time.Minute // Sane default
	} else if pollInterval < time.Second {
		pollInterval = time.Second // Must be at least 1s.
	}

	return &minioWatcher{
		client:       client,
		listName:     config.RedisList,
		failedList:   config.RedisFailedList,
		pollInterval: pollInterval,
		logger:       logger,
		bucketConfig: bucketConfig,
		commonWatcherImpl: &commonWatcherImpl{
			name:             config.Name,
			retentionPeriod:  config.RetentionPeriod,
			stripTopLevelDir: config.StripTopLevelDir,
		},
	}, nil
}

func (w *minioWatcher) Watch(ctx context.Context) (*BlobEvent, Cleanup, error) {
	event, val, err := w.pop(ctx)
	if err != nil {
		if errors.Is(err, redis.Nil) {
			err = ErrWatchTimeout
		}
		return nil, noopCleanup, err
	}
	if event.Bucket != w.bucketConfig.Bucket {
		return nil, noopCleanup, ErrBucketMismatch
	}
	cleanup := w.rem(val)

	return event, cleanup, nil
}

func (w *minioWatcher) Path() string {
	return ""
}

// rem return a function that allows items from the failed list to be safely removed
// in the event of a workflow or some other type of redis error.
func (w *minioWatcher) rem(val string) func(ctx context.Context) error {
	return func(ctx context.Context) error {
		logger := w.logger.WithValues("list", w.failedList)
		if _, err := w.client.LRem(ctx, w.failedList, 1, val).Result(); err != nil {
			logger.Error(err, "Error removing message from failed list.")
			return err
		}
		logger.V(2).Info("Successfully removed message(s).")

		return nil
	}
}

func (w *minioWatcher) pop(ctx context.Context) (*BlobEvent, string, error) {
	val, err := w.client.BLMove(ctx, w.listName, w.failedList, "RIGHT", "LEFT", w.pollInterval).Result()
	if err != nil {
		return nil, "", fmt.Errorf("error retrieving from Redis list: %w", err)
	}

	event, err := w.event(val)
	if err != nil {
		return nil, "", fmt.Errorf("error processing item received: %w", err)
	}

	return event, val, nil
}

// event processes Minio-specific events delivered via Redis. We expect a
// single item array containing a map of {Event: ..., EvenTime: ...}
func (w *minioWatcher) event(blob string) (*BlobEvent, error) {
	container := []json.RawMessage{}
	if err := json.Unmarshal([]byte(blob), &container); err != nil {
		return nil, err
	}

	var set MinioEventSet
	if err := json.Unmarshal(container[0], &set); err != nil {
		return nil, fmt.Errorf("error procesing item received from Redis list: %w", err)
	}
	if len(set.Event) == 0 {
		return nil, fmt.Errorf("error processing item received from Redis list: empty event list")
	}

	key, err := url.QueryUnescape(set.Event[0].S3.Object.Key)
	if err != nil {
		return nil, fmt.Errorf("error processing item received from Redis list: %w", err)
	}

	return NewBlobEventWithBucket(w, set.Event[0].S3.Bucket.Name, key), nil
}

func (w *minioWatcher) OpenBucket(ctx context.Context) (*blob.Bucket, error) {
	return bucket.NewWithConfig(ctx, w.bucketConfig)
}

// Download copies the contents of the blob identified by key to dest.
func (w *minioWatcher) Download(ctx context.Context, dest, key string) error {
	bucket, err := w.OpenBucket(ctx)
	if err != nil {
		return fmt.Errorf("error opening bucket: %w", err)
	}
	defer bucket.Close()

	reader, err := bucket.NewReader(ctx, key, nil)
	if err != nil {
		return fmt.Errorf("error creating reader: %w", err)
	}
	defer reader.Close()

	writer, err := os.Create(dest) // #nosec G304 -- trusted file path.
	if err != nil {
		return fmt.Errorf("error creating writer: %w", err)
	}
	defer writer.Close()

	if _, err := io.Copy(writer, reader); err != nil {
		return fmt.Errorf("error copying blob: %w", err)
	}

	// Try to set the file mode but ignore any errors.
	_ = os.Chmod(dest, 0o600)

	return nil
}
