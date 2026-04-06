package queue

import (
	"context"
	"testing"
	"time"
)

func TestQueue_Add(t *testing.T) {
	ctx := t.Context()
	q := New(ctx)

	task := func(ctx context.Context) (context.Context, error) {
		return ctx, nil
	}

	q.Add(task)

	if len(q.work) != 1 {
		t.Errorf("expected 1 task in queue, got %d", len(q.work))
	}
}

func TestQueue_Close(t *testing.T) {
	ctx := t.Context()
	q := New(ctx)

	task := func(ctx context.Context) (context.Context, error) {
		time.Sleep(100 * time.Millisecond)

		return ctx, nil
	}

	q.Add(task)
	q.Add(task)

	err := q.Close(ctx)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

func TestQueue_Idle(t *testing.T) {
	ctx := t.Context()
	q := New(ctx)

	task := func(ctx context.Context) (context.Context, error) {
		time.Sleep(100 * time.Millisecond)

		return ctx, nil
	}

	q.Add(task)

	err := q.Idle(200 * time.Millisecond)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	q.Add(task)

	err = q.Idle(50 * time.Millisecond)
	if err == nil {
		t.Errorf("expected timeout error, got nil")
	}
}

func TestWithQueue(t *testing.T) {
	ctx := t.Context()
	q := New(ctx)

	ctxWithQueue := WithQueue(ctx, q)
	if GetQueue(ctxWithQueue) != q {
		t.Errorf("expected queue to be set in context")
	}
}

func TestGetQueue(t *testing.T) {
	ctx := t.Context()

	q := GetQueue(ctx)
	if _, ok := q.(*noop); !ok {
		t.Errorf("expected noop queue, got %T", q)
	}
}

// TestQueue_Idle_WaitsForExecution verifies that Idle() waits until the task
// has fully finished executing, not merely been dequeued from the channel.
func TestQueue_Idle_WaitsForExecution(t *testing.T) {
	ctx := t.Context()
	q := New(ctx)

	started := make(chan struct{})
	finished := make(chan struct{})

	task := func(ctx context.Context) (context.Context, error) {
		close(started)
		time.Sleep(150 * time.Millisecond)
		close(finished)

		return ctx, nil
	}

	q.Add(task)

	// Wait until the task has started so the channel is empty but execution is ongoing.
	<-started

	err := q.Idle(500 * time.Millisecond)
	if err != nil {
		t.Fatalf("Idle returned unexpected error: %v", err)
	}

	select {
	case <-finished:
		// expected: task completed before Idle returned
	default:
		t.Error("Idle returned before the task finished executing")
	}

	_ = q.Close(ctx)
}

// TestQueue_Add_AfterClose verifies that calling Add after Close does not panic
// and instead returns the task for inline execution.
func TestQueue_Add_AfterClose(t *testing.T) {
	ctx := t.Context()
	q := New(ctx)

	if err := q.Close(ctx); err != nil {
		t.Fatalf("Close failed: %v", err)
	}

	executed := false
	task := func(ctx context.Context) (context.Context, error) {
		executed = true

		return ctx, nil
	}

	// Must not panic; returned task should be callable inline.
	returned := q.Add(task)
	_, err := returned(ctx)
	if err != nil {
		t.Fatalf("inline task returned error: %v", err)
	}

	if !executed {
		t.Error("expected task to be executed inline after queue was closed")
	}
}
