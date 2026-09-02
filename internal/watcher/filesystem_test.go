package watcher_test

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"syscall"
	"testing"
	"time"

	"gotest.tools/v3/assert"
	"gotest.tools/v3/fs"

	"github.com/artefactual-sdps/enduro/internal/watcher"
)

type file struct {
	name     string
	contents []byte
}

func TestFileSystemWatcher(t *testing.T) {
	t.Parallel()

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
				PollInterval: time.Millisecond * 100,
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

			ctx, cancel := context.WithTimeout(t.Context(), time.Millisecond*300)
			defer cancel()

			w, err := watcher.NewFilesystemWatcher(ctx, tt.config)
			assert.NilError(t, err)

			if err = os.WriteFile(
				filepath.Join(tt.config.Path, tt.file.name),
				tt.file.contents,
				0o600,
			); err != nil {
				t.Fatalf("Couldn't create %q in %q", tt.file.name, tt.config.Path)
			}

			got, _, err := w.Watch(ctx)
			assert.NilError(t, err, "watcher error: %v", err)
			assert.Equal(t, got.Key, tt.want.Key)
			assert.Equal(t, got.IsDir, tt.want.IsDir)
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

	t.Run("Dispose moves SIP to CompletedDir", func(t *testing.T) {
		t.Parallel()

		src := fs.NewDir(t, "enduro-test-fswatcher",
			fs.WithDir("example-sip",
				fs.WithDir("objects", fs.WithFile("test.txt", "A test file.")),
			),
		)
		dest := fs.NewDir(t, "enduro-test-fswatcher")

		ctx := t.Context()

		w, err := watcher.NewFilesystemWatcher(ctx, &watcher.FilesystemConfig{
			Name:            "filesystem",
			Path:            src.Path(),
			CompletedDir:    dest.Path(),
			RetentionPeriod: -1 * time.Second,
		})
		assert.NilError(t, err)

		srcPath := filepath.Join(w.Path(), "example-sip")
		info, err := os.Stat(srcPath)
		assert.NilError(t, err)
		assert.Assert(t, info.IsDir())

		err = w.Dispose("example-sip")
		assert.NilError(t, err)

		_, err = os.Stat(srcPath)
		assert.Assert(t, os.IsNotExist(err))

		assert.Assert(t, fs.Equal(dest.Path(), fs.Expected(t,
			fs.WithDir("example-sip",
				fs.WithDir("objects", fs.WithFile("test.txt", "A test file.")),
			),
		)))
	})

	t.Run("Dispose preserves mtimes across filesystems", func(t *testing.T) {
		watchedDir := t.TempDir()
		completedDir, err := os.MkdirTemp("/dev/shm", "enduro-completed-")
		if err != nil {
			t.Skipf("cross-filesystem test requires writable /dev/shm: %v", err)
		}
		t.Cleanup(func() { _ = os.RemoveAll(completedDir) })

		fileStat := func(path string) *syscall.Stat_t {
			info, err := os.Stat(path)
			assert.NilError(t, err)
			stat, ok := info.Sys().(*syscall.Stat_t)
			assert.Assert(t, ok)
			return stat
		}
		if fileStat(watchedDir).Dev == fileStat(completedDir).Dev {
			t.Skip("temporary and completed directories are on the same filesystem")
		}

		sourceDir := filepath.Join(watchedDir, "example-sip")
		sourceFile := filepath.Join(sourceDir, "objects", "test.txt")
		assert.NilError(t, os.MkdirAll(filepath.Dir(sourceFile), 0o755))
		assert.NilError(t, os.WriteFile(sourceFile, []byte("A test file."), 0o600))
		wantModTime := time.Date(2001, time.February, 3, 4, 5, 6, 0, time.UTC)
		assert.NilError(t, os.Chtimes(sourceFile, wantModTime, wantModTime))

		w, err := watcher.NewFilesystemWatcher(t.Context(), &watcher.FilesystemConfig{
			Name:            "filesystem",
			Path:            watchedDir,
			CompletedDir:    completedDir,
			RetentionPeriod: -1 * time.Second,
		})
		assert.NilError(t, err)
		assert.NilError(t, w.Dispose("example-sip"))

		info, err := os.Stat(filepath.Join(completedDir, "example-sip", "objects", "test.txt"))
		assert.NilError(t, err)
		assert.Equal(t, info.ModTime().Unix(), wantModTime.Unix())
	})

	t.Run("Download copies a file to the destination path", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()

		src := fs.NewDir(t, "enduro-test-fswatcher",
			fs.WithFile("sip.zip", "A test file."),
		)
		wantModTime := time.Date(2001, time.February, 3, 4, 5, 6, 0, time.UTC)
		assert.NilError(t, os.Chtimes(src.Join("sip.zip"), wantModTime, wantModTime))
		dest := fs.NewDir(t, "enduro-test-fswatcher")

		w, err := watcher.NewFilesystemWatcher(ctx, &watcher.FilesystemConfig{
			Name: "filesystem",
			Path: src.Path(),
		})
		assert.NilError(t, err)

		destPath := filepath.Join(dest.Path(), "sip.zip")
		err = w.Download(context.Background(), destPath, "sip.zip")
		assert.NilError(t, err)

		info, err := os.Stat(destPath)
		assert.NilError(t, err)
		assert.Assert(t, !info.IsDir())
		assert.Equal(t, info.ModTime().Unix(), wantModTime.Unix())

		got, err := os.ReadFile(destPath)
		assert.NilError(t, err)
		assert.Equal(t, string(got), "A test file.")
	})

	t.Run("Download copies a directory", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()

		src := fs.NewDir(t, "enduro-test-fswatcher",
			fs.WithDir("example-sip",
				fs.WithFile("test.txt", "A test file."),
				fs.WithFile("test2", "Another test file."),
			),
		)
		wantModTime := time.Date(2001, time.February, 3, 4, 5, 6, 0, time.UTC)
		assert.NilError(t, os.Chtimes(src.Join("example-sip", "test.txt"), wantModTime, wantModTime))
		dest := fs.NewDir(t, "enduro-test-fswatcher")

		w, err := watcher.NewFilesystemWatcher(ctx, &watcher.FilesystemConfig{
			Name:    "filesystem",
			Path:    src.Path(),
			Inotify: true,
		})
		assert.NilError(t, err)

		err = w.Download(context.Background(), filepath.Join(dest.Path(), "example-sip"), "example-sip")
		assert.NilError(t, err)
		assert.Assert(t, fs.Equal(dest.Path(), fs.Expected(t, fs.WithMode(0o700),
			fs.WithDir("example-sip", fs.WithMode(0o755),
				fs.WithFile("test.txt", "A test file.", fs.WithMode(0o644)),
				fs.WithFile("test2", "Another test file.", fs.WithMode(0o644)),
			),
		)))
		info, err := os.Stat(filepath.Join(dest.Path(), "example-sip", "test.txt"))
		assert.NilError(t, err)
		assert.Equal(t, info.ModTime().Unix(), wantModTime.Unix())
	})
}
