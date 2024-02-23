// Package filenotify provides a mechanism for watching file(s) for changes.
// Generally leans on fsnotify, but provides a poll-based notifier which fsnotify does not support.
// These are wrapped up in a common interface so that either can be used interchangeably in your code.
package filenotify

import (
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/radovskyb/watcher"
)

type Config struct {
	PollInterval time.Duration
}

// FileWatcher is an interface for implementing file notification watchers.
type FileWatcher interface {
	Events() <-chan fsnotify.Event
	Errors() <-chan error
	Add(name string) error
	Remove(name string) error
	Close() error
}

// New tries to use an fs-event watcher, and falls back to the poller if there is an error
func New(cfg Config) (FileWatcher, error) {
	if watcher, err := NewEventWatcher(); err == nil {
		return watcher, nil
	}
	return NewPollingWatcher(cfg)
}

// NewEventWatcher returns an fs-event based file watcher
func NewEventWatcher() (FileWatcher, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}
	return &fsNotifyWatcher{
		Watcher: watcher,
	}, nil
}

// NewPollingWatcher returns a poll-based file watcher
func NewPollingWatcher(cfg Config) (FileWatcher, error) {
	poller := &filePoller{
		wr:     watcher.New(),
		events: make(chan fsnotify.Event),
		errors: make(chan error),
	}
	poller.wr.FilterOps(watcher.Create, watcher.Rename, watcher.Move)
	go poller.loop()

	done := make(chan error)
	{
		go func() {
			err := poller.wr.Start(cfg.PollInterval)
			if err != nil {
				done <- err
			}
		}()
		go func() {
			poller.wr.Wait()
			done <- nil
		}()
	}
	if err := <-done; err != nil {
		return nil, err
	}

	return poller, nil
}
