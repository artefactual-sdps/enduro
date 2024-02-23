package filenotify_test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/fsnotify/fsnotify"

	"github.com/artefactual-sdps/enduro/internal/filenotify"
)

const pollInterval = time.Millisecond * 5

func TestPollerEvent(t *testing.T) {
	w, err := filenotify.NewPollingWatcher(
		filenotify.Config{PollInterval: pollInterval},
	)
	if err != nil {
		t.Fatal("error creating poller")
	}
	defer w.Close()

	tmpdir, err := os.MkdirTemp("", "test-poller")
	if err != nil {
		t.Fatal("error creating temp file")
	}
	defer os.RemoveAll(tmpdir)

	if err := w.Add(tmpdir); err != nil {
		t.Fatal(err)
	}

	select {
	case <-w.Events():
		t.Fatal("got event before anything happened")
	case <-w.Errors():
		t.Fatal("got error before anything happened")
	default:
	}

	path := filepath.Join(tmpdir, "hello.txt")
	if err := os.WriteFile(path, []byte("hello"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := assertEvent(w, fsnotify.Create); err != nil {
		t.Fatal(err)
	}
}

func assertEvent(w filenotify.FileWatcher, eType fsnotify.Op) error {
	var err error
	select {
	case e := <-w.Events():
		if e.Op != eType {
			err = fmt.Errorf("got wrong event type, expected %q: %v", eType, e.Op)
		}
	case e := <-w.Errors():
		err = fmt.Errorf("got unexpected error waiting for events %v: %v", eType, e)
	case <-time.After(pollInterval * 2):
		err = fmt.Errorf("timeout waiting for event %v", eType)
	}
	return err
}
