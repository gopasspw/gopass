// Package queue implements an experimental background queue for cleanup jobs.
// Beware: It's likely broken.
// We can easily close a channel which might later be written to.
// The current locking is but a poor workaround.
// A better implementation would create a queue object in main, pass
// it through and wait for the channel to be empty before leaving main.
// Will do that later.
package queue

import (
	"context"
	"sync"

	"github.com/gopasspw/gopass/internal/debug"
)

var (
	tasks  = make(chan chan error, 1024)
	done   = make(chan struct{}, 1)
	closed = false
	mux    = sync.Mutex{}
)

// Task is a background task
type Task func() error

func init() {
	go func() {
		for t := range tasks {
			if err := <-t; err != nil {
				debug.Log("Task failed: %s", err)
			}
			debug.Log("Task done")
		}
		debug.Log("all tasks done")
		done <- struct{}{}
	}()
}

// Add enqueues a new task
func Add(t Task) {
	mux.Lock()
	defer mux.Unlock()
	if closed {
		debug.Log("ERROR: Attempting to enqueue in closed queue")
		return
	}
	go func() {
		ec := make(chan error, 1)
		tasks <- ec
		ec <- t()
	}()
	debug.Log("enqueued task")
}

// Close closes the queue for new entries and processes the remaining ones
func Close(ctx context.Context) error {
	mux.Lock()
	closed = true
	mux.Unlock()
	close(tasks)
	select {
	case <-done:
		return nil
	case <-ctx.Done():
		debug.Log("context canceled")
		return ctx.Err()
	}
}
