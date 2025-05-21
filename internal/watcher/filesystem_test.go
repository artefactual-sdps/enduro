package watcher_test

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"gotest.tools/v3/assert"
	"gotest.tools/v3/fs"
	"gotest.tools/v3/poll"

	"github.com/artefactual-sdps/enduro/internal/watcher"
)

type file struct {
	name     string
	contents []byte
}

func TestFileSystemWatcher(t *testing.T) {
	t.Parallel()

	td := fs.NewDir(t, "enduro-test-fs-watcher")
	type test struct {
		name   string
		config *watcher.FilesystemConfig
		file   file
		want   *watcher.BlobEvent
	}
	for _, tt := range []test{
		{
			name: "Polling watcher returns a blob event",
			config: &watcher.FilesystemConfig{
				Name:         "filesystem",
				Path:         t.TempDir(),
				PollInterval: time.Millisecond * 5,
			},
			file: file{name: "test.txt"},
			want: &watcher.BlobEvent{Key: "test.txt"},
		},
		{
			name: "Inotify watcher returns a blob event",
			config: &watcher.FilesystemConfig{
				Name:    "filesystem",
				Path:    t.TempDir(),
				Inotify: true,
			},
			file: file{name: "test.txt"},
			want: &watcher.BlobEvent{Key: "test.txt"},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctx := t.Context()

			w, err := watcher.NewFilesystemWatcher(ctx, tt.config)
			assert.NilError(t, err)

			check := func(t poll.LogT) poll.Result {
				got, _, err := w.Watch(ctx)
				if err != nil {
					return poll.Error(fmt.Errorf("watcher error: %w", err))
				}
				if got.Key != tt.want.Key || got.IsDir != tt.want.IsDir {
					return poll.Error(fmt.Errorf(
						"expected: *watcher.BlobEvent(Key: %q, IsDir: %t); got: *watcher.BlobEvent(Key: %q, IsDir: %t)",
						tt.want.Key, tt.want.IsDir, got.Key, got.IsDir,
					))
				}

				return poll.Success()
			}

			if err = os.WriteFile(
				filepath.Join(tt.config.Path, tt.file.name),
				tt.file.contents,
				0o600,
			); err != nil {
				t.Fatalf("Couldn't create text.txt in %q", td.Path())
			}

			poll.WaitOn(t, check, poll.WithTimeout(time.Millisecond*15))
		})
	}

	t.Run("Path returns the watcher path", func(t *testing.T) {
		t.Parallel()

		td := t.TempDir()
		ctx := t.Context()

		w, err := watcher.NewFilesystemWatcher(ctx, &watcher.FilesystemConfig{
			Name: "filesystem",
			Path: td,
		})
		assert.NilError(t, err)
		assert.Equal(t, w.Path(), td)
	})

	t.Run("OpenBucket returns a bucket", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()

		w, err := watcher.NewFilesystemWatcher(ctx, &watcher.FilesystemConfig{
			Name: "filesystem",
			Path: t.TempDir(),
		})
		assert.NilError(t, err)

		b, err := w.OpenBucket(ctx)
		assert.NilError(t, err)
		assert.Equal(t, fmt.Sprintf("%T", b), "*blob.Bucket")
		b.Close()
	})

	t.Run("RemoveAll deletes a directory", func(t *testing.T) {
		t.Parallel()

		td := fs.NewDir(t, "enduro-test-fswatcher",
			fs.WithDir("transfer", fs.WithFile("test.txt", "A test file.")),
		)

		ctx := t.Context()

		w, err := watcher.NewFilesystemWatcher(ctx, &watcher.FilesystemConfig{
			Name: "filesystem",
			Path: td.Path(),
		})
		assert.NilError(t, err)

		err = w.RemoveAll("transfer")
		assert.NilError(t, err)
		assert.Assert(t, fs.Equal(w.Path(), fs.Expected(t)))
	})

	t.Run("Dispose moves transfer to CompletedDir", func(t *testing.T) {
		t.Parallel()

		src := fs.NewDir(t, "enduro-test-fswatcher",
			fs.WithDir("transfer", fs.WithFile("test.txt", "A test file.")),
		)
		dest := fs.NewDir(t, "enduro-test-fswatcher")

		ctx := t.Context()

		w, err := watcher.NewFilesystemWatcher(ctx, &watcher.FilesystemConfig{
			Name:         "filesystem",
			Path:         src.Path(),
			CompletedDir: dest.Path(),
		})
		assert.NilError(t, err)

		err = w.Dispose("transfer")
		assert.NilError(t, err)
		assert.Assert(t, fs.Equal(dest.Path(), fs.Expected(t,
			fs.WithDir("transfer", fs.WithFile("test.txt", "A test file.")),
		)))
	})

	t.Run("Download copies a directory", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()

		src := fs.NewDir(t, "enduro-test-fswatcher",
			fs.WithDir("transfer",
				fs.WithFile("test.txt", "A test file."),
				fs.WithFile("test2", "Another test file."),
			),
		)
		dest := fs.NewDir(t, "enduro-test-fswatcher")

		w, err := watcher.NewFilesystemWatcher(ctx, &watcher.FilesystemConfig{
			Name:    "filesystem",
			Path:    src.Path(),
			Inotify: true,
		})
		assert.NilError(t, err)

		err = w.Download(context.Background(), dest.Path(), "transfer")
		assert.NilError(t, err)
		assert.Assert(t, fs.Equal(dest.Path(), fs.Expected(t, fs.WithMode(0o700),
			fs.WithDir("transfer", fs.WithMode(0o755),
				fs.WithFile("test.txt", "A test file.", fs.WithMode(0o644)),
				fs.WithFile("test2", "Another test file.", fs.WithMode(0o644)),
			),
		)))
	})
}
