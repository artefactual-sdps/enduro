package bundler_test

import (
	"bytes"
	"os"
	"path/filepath"
	"syscall"
	"testing"

	"gotest.tools/v3/assert"
	"gotest.tools/v3/fs"

	"github.com/artefactual-sdps/enduro/internal/bundler"
)

func TestNewBundlerWithTempDir(t *testing.T) {
	t.Parallel()

	tmpDir := fs.NewDir(t, "enduro-bundler")
	b, err := bundler.NewBundlerWithTempDir(tmpDir.Path())
	assert.NilError(t, err)

	// Describe the transfer.
	b.Describe("dc.title", "Transfer")

	// Describe and write foobar.txt and notes.txt into the transfer.
	err = b.Write("foobar.txt", bytes.NewReader([]byte("some text")))
	assert.NilError(t, err)
	b.DescribeFile("foobar.txt", "dc.title", "Foobar")
	b.ChecksumMD5("foobar.txt", "552e21cd4cd9918678e3c1a0df491bc3")
	b.ChecksumSHA1("foobar.txt", "37aa63c77398d954473262e1a0057c1e632eda77")
	b.ChecksumSHA256("foobar.txt", "b94f6f125c79e3a5ffaa826f584c10d52ada669e6762051b826b55776d05aed2")
	err = b.Write("notes.txt", bytes.NewReader([]byte("some notes")))
	assert.NilError(t, err)
	b.DescribeFile("notes.txt", "dc.title", "Notes")
	b.DescribeFile("notes.txt", "dc.description", "Some notes")

	err = b.Bundle()
	assert.NilError(t, err)

	// The actual transfer directory being bundled is created two levels deep,
	// using a root container ("c") and a temporary directory, for example:
	// `/tmp/enduro-bundler-2161969315/c/304672139`.
	dir := transferDir(t, tmpDir.Path())
	assert.Equal(t, b.FullBaseFsPath(), dir)

	umask := os.FileMode(syscall.Umask(0))
	dirMode := fs.WithMode(0o755 &^ umask)
	fileMode := fs.WithMode(0o664 &^ umask)

	assert.Assert(t, fs.Equal(
		dir,
		fs.Expected(t,
			// By default, the bundler uses less restrictive permissions to
			// allow other local users (e.g., the default Archivematica user) to
			// access the contents. This might not be the best approach.
			dirMode,
			fs.WithDir("metadata",
				fs.WithFile("checksum.md5", "552e21cd4cd9918678e3c1a0df491bc3 foobar.txt\n", fileMode),
				fs.WithFile("checksum.sha1", "37aa63c77398d954473262e1a0057c1e632eda77 foobar.txt\n", fileMode),
				fs.WithFile("checksum.sha256", "b94f6f125c79e3a5ffaa826f584c10d52ada669e6762051b826b55776d05aed2 foobar.txt\n", fileMode),
				fs.WithFile("metadata.csv", `filename,dc.description,dc.title
foobar.txt,,Foobar
notes.txt,Some notes,Notes
objects/,,Transfer
`,
					fileMode),
			),
			// The `objects`` directory is empty because the bundler opts to
			// place the files in the root directory, which is also supported.
			fs.WithDir("objects"),
			fs.WithFile("foobar.txt", "some text", fileMode),
			fs.WithFile("notes.txt", "some notes", fileMode),
		),
	))

	// We can remove the transfer but it should not remove the root container
	// since it is the common ancestor to other transfers
	err = b.Destroy()
	assert.NilError(t, err)
	assert.Assert(t, fs.Equal(
		tmpDir.Path(),
		fs.Expected(t,
			fs.WithDir("c"),
		),
	))
}

func transferDir(t *testing.T, dir string) string {
	t.Helper()

	// bundler uses an additional "c" parent directory to avoid polluting the
	// Archivematica transfer source location, which is arguably unnecessary
	// or should maybe be optional.
	rootContainer := filepath.Join(dir, "c")

	// Inside it, it creates a temporary directory named randomly.
	entries, err := os.ReadDir(rootContainer)
	assert.NilError(t, err)

	return filepath.Join(rootContainer, entries[0].Name())
}
