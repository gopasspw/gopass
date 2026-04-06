// Package queue implements a background queue for cleanup jobs.
package queue

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/gopasspw/gopass/internal/out"
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

// WithQueue adds the given queue to the context. Add a nil
// queue to disable queuing in this context.
func WithQueue(ctx context.Context, q *Queue) context.Context {
	return context.WithValue(ctx, ctxKeyQueue, q)
}

// GetQueue returns an existing queue from the context or
// returns a noop one.
func GetQueue(ctx context.Context) Queuer {
	if q, ok := ctx.Value(ctxKeyQueue).(*Queue); ok && q != nil {
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
type Task func(ctx context.Context) (context.Context, error)

// Queue is a serialized background processing unit.
type Queue struct {
	work   chan Task
	done   chan struct{}
	wg     sync.WaitGroup
	mu     sync.Mutex
	closed bool
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
		ctx2, err := t(ctx)
		if err != nil {
			out.Errorf(ctx, "Task failed: %s", err)
		}
		if ctx2 != nil {
			// if a task returns a context, it is to transmit information to the next tasks in line
			// so replace the in-queue context with the new one
			// (each task has access to two contexts: one from the queue, and one from the function creating the task)
			ctx = ctx2
		}
		q.wg.Done()
		debug.Log("Task done")
	}
	debug.Log("all tasks done")
	q.done <- struct{}{}
}

// Add enqueues a new task. If the queue is already closed, the task is
// returned to the caller for inline execution (matching noop behaviour).
func (q *Queue) Add(t Task) Task {
	q.mu.Lock()
	if q.closed {
		q.mu.Unlock()
		debug.Log("queue closed, returning task for inline execution")
		return t
	}
	q.wg.Add(1)
	q.work <- t
	q.mu.Unlock()
	debug.Log("enqueued task")

	return func(ctx2 context.Context) (context.Context, error) {
		return ctx2, nil
	}
}

// Idle waits until all currently enqueued tasks have finished executing.
// Returns an error if maxWait elapses before the queue drains.
func (q *Queue) Idle(maxWait time.Duration) error {
	done := make(chan struct{})
	go func() {
		q.wg.Wait()
		close(done)
	}()
	select {
	case <-done:
		return nil
	case <-time.After(maxWait):
		return fmt.Errorf("timed out waiting for queue to drain")
	}
}

// Close waits for all tasks to be processed. Must only be called once on
// shutdown.
func (q *Queue) Close(ctx context.Context) error {
	q.mu.Lock()
	q.closed = true
	close(q.work)
	q.mu.Unlock()
	select {
	case <-q.done:
		return nil
	case <-ctx.Done():
		debug.Log("context canceled")

		return ctx.Err()
	}
}
