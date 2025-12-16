package filenotify_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/fsnotify/fsnotify"
	"gotest.tools/v3/poll"

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
	assertEvent(t, w, fsnotify.Create)
}

func assertEvent(t *testing.T, w filenotify.FileWatcher, eType fsnotify.Op) {
	t.Helper()

	poll.WaitOn(t, func(t poll.LogT) poll.Result {
		select {
		case e := <-w.Events():
			if e.Op == eType {
				return poll.Success()
			}
			return poll.Continue("got wrong event type, expected %q: %v", eType, e.Op)
		case err := <-w.Errors():
			return poll.Error(err)
		default:
			return poll.Continue("no event yet")
		}
	},
		poll.WithTimeout(pollInterval*20),
		poll.WithDelay(pollInterval),
	)
}
