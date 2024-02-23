package watcher

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"runtime"

	"github.com/fsnotify/fsnotify"
	cp "github.com/otiai10/copy"
	"gocloud.dev/blob"
	_ "gocloud.dev/blob/fileblob"

	"github.com/artefactual-sdps/enduro/internal/bucket"
	"github.com/artefactual-sdps/enduro/internal/filenotify"
	"github.com/artefactual-sdps/enduro/internal/fsutil"
)

// filesystemWatcher implements a Watcher for watching paths in a local filesystem.
type filesystemWatcher struct {
	ctx   context.Context
	cfg   *FilesystemConfig
	fw    filenotify.FileWatcher
	ch    chan *fsnotify.Event
	path  string
	regex *regexp.Regexp
	*commonWatcherImpl
}

var _ Watcher = (*filesystemWatcher)(nil)

func NewFilesystemWatcher(ctx context.Context, config *FilesystemConfig) (*filesystemWatcher, error) {
	config.setDefaults()

	stat, err := os.Stat(config.Path)
	if err != nil {
		return nil, fmt.Errorf("error looking up stat info: %w", err)
	}
	if !stat.IsDir() {
		return nil, errors.New("given path is not a directory")
	}
	abspath, err := filepath.Abs(config.Path)
	if err != nil {
		return nil, fmt.Errorf("error generating absolute path of %s: %v", config.Path, err)
	}

	var regex *regexp.Regexp
	if config.Ignore != "" {
		if regex, err = regexp.Compile(config.Ignore); err != nil {
			return nil, fmt.Errorf("error compiling regular expression (ignore): %v", err)
		}
	}

	if config.CompletedDir != "" && config.RetentionPeriod != nil {
		return nil, errors.New("cannot use completedDir and retentionPeriod simultaneously")
	}

	fw, err := fileWatcher(config)
	if err != nil {
		return nil, err
	}

	w := &filesystemWatcher{
		ctx:   ctx,
		cfg:   config,
		fw:    fw,
		ch:    make(chan *fsnotify.Event, 100),
		path:  abspath,
		regex: regex,
		commonWatcherImpl: &commonWatcherImpl{
			name:             config.Name,
			retentionPeriod:  config.RetentionPeriod,
			completedDir:     config.CompletedDir,
			stripTopLevelDir: config.StripTopLevelDir,
		},
	}

	go w.loop()

	if err := fw.Add(abspath); err != nil {
		return nil, fmt.Errorf("error configuring filesystem watcher: %w", err)
	}

	return w, nil
}

func fileWatcher(cfg *FilesystemConfig) (filenotify.FileWatcher, error) {
	var (
		fsw filenotify.FileWatcher
		err error
	)

	// The inotify API isn't always available, fall back to polling.
	if cfg.Inotify && runtime.GOOS != "windows" {
		fsw, err = filenotify.New(filenotify.Config{PollInterval: cfg.PollInterval})
	} else {
		fsw, err = filenotify.NewPollingWatcher(
			filenotify.Config{PollInterval: cfg.PollInterval},
		)
	}
	if err != nil {
		return nil, fmt.Errorf("error creating filesystem watcher: %w", err)
	}

	return fsw, nil
}

func (w *filesystemWatcher) loop() {
	for {
		select {
		case event, ok := <-w.fw.Events():
			if !ok {
				continue
			}
			if event.Op != fsnotify.Create && event.Op != fsnotify.Rename {
				continue
			}
			if path, err := filepath.Abs(event.Name); err != nil || path == w.path {
				continue
			}
			if w.regex != nil && w.regex.MatchString(filepath.Base(event.Name)) {
				continue
			}
			w.ch <- &event
		case _, ok := <-w.fw.Errors():
			if !ok {
				continue
			}
		case <-w.ctx.Done():
			_ = w.fw.Close()
			close(w.ch)
			return
		}
	}
}

func (w *filesystemWatcher) Watch(ctx context.Context) (*BlobEvent, Cleanup, error) {
	fsevent, ok := <-w.ch
	if !ok {
		return nil, noopCleanup, ErrWatchTimeout
	}
	info, err := os.Stat(fsevent.Name)
	if err != nil {
		return nil, noopCleanup, fmt.Errorf("error in file stat check: %s", err)
	}
	rel, err := filepath.Rel(w.path, fsevent.Name)
	if err != nil {
		return nil, noopCleanup, fmt.Errorf("error generating relative path of fsvent.Name %s - %w", fsevent.Name, err)
	}
	return NewBlobEvent(w, rel, info.IsDir()), noopCleanup, nil
}

func (w *filesystemWatcher) Path() string {
	return w.path
}

func (w *filesystemWatcher) OpenBucket(ctx context.Context) (*blob.Bucket, error) {
	return bucket.Open(ctx, &bucket.Config{
		URL: fmt.Sprintf("file://%s", w.path),
	})
}

func (w *filesystemWatcher) RemoveAll(key string) error {
	return os.RemoveAll(filepath.Join(w.path, key))
}

func (w *filesystemWatcher) Dispose(key string) error {
	if w.completedDir == "" {
		return nil
	}

	src := filepath.Join(w.path, key)
	dst := filepath.Join(w.completedDir, key)

	return fsutil.Move(src, dst)
}

// Download recursively copies the contents of key to dest. Key may be the name
// of a directory or file.
func (w *filesystemWatcher) Download(ctx context.Context, dest, key string) error {
	src := filepath.Clean(filepath.Join(w.path, key))
	dest = filepath.Clean(filepath.Join(dest, key))
	if err := cp.Copy(src, dest); err != nil {
		return fmt.Errorf("filesystem watcher: download: %v", err)
	}

	return nil
}
