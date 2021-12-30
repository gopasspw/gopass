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
	"fmt"
	"time"

	"github.com/gopasspw/gopass/pkg/debug"
)

type contextKey int

const (
	ctxKeyQueue contextKey = iota
)

// Queuer is a queue interface.
type Queuer interface {
	Add(Task) Task
	Close(context.Context) error
	Idle(time.Duration) error
}

// WithQueue adds the given queue to the context.
func WithQueue(ctx context.Context, q *Queue) context.Context {
	return context.WithValue(ctx, ctxKeyQueue, q)
}

// GetQueue returns an existing queue from the context or
// returns a noop one.
func GetQueue(ctx context.Context) Queuer {
	if q, ok := ctx.Value(ctxKeyQueue).(*Queue); ok {
		return q
	}
	return &noop{}
}

type noop struct{}

// Add always returns the task.
func (n *noop) Add(t Task) Task {
	return t
}

// Close always returns nil.
func (n *noop) Close(_ context.Context) error {
	return nil
}

// Idle always returns nil.
func (n *noop) Idle(_ time.Duration) error {
	return nil
}

// Task is a background task.
type Task func(ctx context.Context) error

// Queue is a serialized background processing unit.
type Queue struct {
	work chan Task
	done chan struct{}
}

// New creates a new queue.
func New(ctx context.Context) *Queue {
	q := &Queue{
		work: make(chan Task, 1024),
		done: make(chan struct{}, 1),
	}
	go q.run(ctx)
	return q
}

func (q *Queue) run(ctx context.Context) {
	for t := range q.work {
		if err := t(ctx); err != nil {
			debug.Log("Task failed: %s", err)
		}
		debug.Log("Task done")
	}
	debug.Log("all tasks done")
	q.done <- struct{}{}
}

// Add enqueues a new task.
func (q *Queue) Add(t Task) Task {
	q.work <- t
	debug.Log("enqueued task")
	return func(_ context.Context) error { return nil }
}

// Idle returns nil the next time the queue is empty.
func (q *Queue) Idle(maxWait time.Duration) error {
	done := make(chan struct{})
	go func() {
		for {
			if len(q.work) < 1 {
				done <- struct{}{}
			}
			time.Sleep(20 * time.Millisecond)
		}
	}()
	select {
	case <-done:
		return nil
	case <-time.After(maxWait):
		return fmt.Errorf("timed out waiting for empty queue")
	}
}

// Close waits for all tasks to be processed. Must only be called once on
// shutdown.
func (q *Queue) Close(ctx context.Context) error {
	close(q.work)
	select {
	case <-q.done:
		return nil
	case <-ctx.Done():
		debug.Log("context canceled")
		return ctx.Err()
	}
}
