package leaf

import (
	"context"
	"errors"
	"fmt"

	"github.com/gopasspw/gopass/internal/queue"
	"github.com/gopasspw/gopass/internal/store"
	"github.com/gopasspw/gopass/pkg/debug"
)

// Link creates a symlink.
func (s *Store) Link(ctx context.Context, from, to string) error {
	if !s.Exists(ctx, from) {
		return fmt.Errorf("source %q does not exists", from)
	}
	if s.Exists(ctx, to) {
		return fmt.Errorf("destination %q already exists", to)
	}

	if err := s.storage.Link(ctx, s.passfile(from), s.passfile(to)); err != nil {
		return fmt.Errorf("failed to create symlink from %q to %q: %w", from, to, err)
	}
	debug.Log("created symlink from %q to %q", from, to)

	if err := s.storage.Add(ctx, s.passfile(to)); err != nil {
		if errors.Is(err, store.ErrGitNotInit) {
			return nil
		}
		return fmt.Errorf("failed to add %q to git: %w", to, err)
	}

	// try to enqueue this task, if the queue is not available
	// it will return the task and we will execute it inline
	t := queue.GetQueue(ctx).Add(func(ctx context.Context) error {
		return s.gitCommitAndPush(ctx, to)
	})
	return t(ctx)
}
