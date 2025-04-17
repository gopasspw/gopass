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
