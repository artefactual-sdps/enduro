package fsutil_test

import (
	"errors"
	"os"
	"testing"

	"gotest.tools/v3/assert"
	"gotest.tools/v3/fs"

	"github.com/artefactual-sdps/enduro/internal/fsutil"
)

var Renamer = os.Rename

var dirOpts = []fs.PathOp{
	fs.WithDir(
		"child1",
		fs.WithFile(
			"foo.txt",
			"foo",
		),
	),
	fs.WithDir(
		"child2",
		fs.WithFile(
			"bar.txt",
			"bar",
		),
	),
}

func TestBaseNoExt(t *testing.T) {
	t.Parallel()

	assert.Equal(t, fsutil.BaseNoExt("/tmp/dir"), "dir")
	assert.Equal(t, fsutil.BaseNoExt("/tmp/dir/small.txt"), "small")
}

func TestMove(t *testing.T) {
	t.Parallel()

	t.Run("It fails if destination already exists", func(t *testing.T) {
		t.Parallel()

		tmpDir := fs.NewDir(t, "enduro")
		fs.Apply(t, tmpDir, fs.WithFile("foobar.txt", ""))
		fs.Apply(t, tmpDir, fs.WithFile("barfoo.txt", ""))

		src := tmpDir.Join("foobar.txt")
		dst := tmpDir.Join("barfoo.txt")
		err := fsutil.Move(src, dst)

		assert.Error(t, err, "destination already exists")
	})

	t.Run("It moves files", func(t *testing.T) {
		t.Parallel()

		tmpDir := fs.NewDir(t, "enduro")
		fs.Apply(t, tmpDir, fs.WithFile("foobar.txt", ""))

		src := tmpDir.Join("foobar.txt")
		dst := tmpDir.Join("barfoo.txt")
		err := fsutil.Move(src, dst)

		assert.NilError(t, err)
		assert.Equal(t, fsutil.FileExists(src), false)
		assert.Equal(t, fsutil.FileExists(dst), true)
	})

	t.Run("It moves directories", func(t *testing.T) {
		t.Parallel()

		tmpSrc := fs.NewDir(t, "enduro", dirOpts...)
		src := tmpSrc.Path()
		srcManifest := fs.ManifestFromDir(t, src)
		tmpDst := fs.NewDir(t, "enduro")
		dst := tmpDst.Join("nested")

		err := fsutil.Move(src, dst)

		assert.NilError(t, err)
		assert.Equal(t, fsutil.FileExists(src), false)
		assert.Assert(t, fs.Equal(dst, srcManifest))
	})

	t.Run("It copies directories when using different filesystems", func(t *testing.T) {
		fsutil.Renamer = func(src, dst string) error {
			return &os.LinkError{
				Op:  "rename",
				Old: src,
				New: dst,
				Err: errors.New("invalid cross-device link"),
			}
		}
		t.Cleanup(func() {
			fsutil.Renamer = os.Rename
		})

		tmpSrc := fs.NewDir(t, "enduro", dirOpts...)
		src := tmpSrc.Path()
		srcManifest := fs.ManifestFromDir(t, src)
		tmpDst := fs.NewDir(t, "enduro")
		dst := tmpDst.Join("nested")

		err := fsutil.Move(src, dst)

		assert.NilError(t, err)
		assert.Equal(t, fsutil.FileExists(src), false)
		assert.Assert(t, fs.Equal(dst, srcManifest))
	})
}

func TestSetFileModes(t *testing.T) {
	td := fs.NewDir(t, "enduro-test-fsutil",
		fs.WithDir("transfer", fs.WithMode(0o755),
			fs.WithFile("test1", "I'm a test file.", fs.WithMode(0o644)),
			fs.WithDir("subdir", fs.WithMode(0o755),
				fs.WithFile("test2", "Another test file.", fs.WithMode(0o644)),
			),
		),
	)

	err := fsutil.SetFileModes(td.Join("transfer"), 0o700, 0o600)
	assert.NilError(t, err)
	assert.Assert(t, fs.Equal(
		td.Path(),
		fs.Expected(t,
			fs.WithDir("transfer", fs.WithMode(0o700),
				fs.WithFile("test1", "I'm a test file.", fs.WithMode(0o600)),
				fs.WithDir("subdir", fs.WithMode(0o700),
					fs.WithFile("test2", "Another test file.", fs.WithMode(0o600)),
				),
			),
		),
	))
}

func TestFileExists(t *testing.T) {
	t.Parallel()

	td := fs.NewDir(t, "enduro-test", fs.WithFile("small.txt", "I'm a small file."))
	assert.Equal(t, fsutil.FileExists(td.Path()), true)
	assert.Equal(t, fsutil.FileExists(td.Join("small.txt")), true)
	assert.Equal(t, fsutil.FileExists(td.Join("nope")), false)
}
