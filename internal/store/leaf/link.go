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

	if err := s.storage.Link(ctx, s.Passfile(from), s.Passfile(to)); err != nil {
		return fmt.Errorf("failed to create symlink from %q to %q: %w", from, to, err)
	}

	debug.Log("created symlink from %q to %q", from, to)

	if err := s.storage.Add(ctx, s.Passfile(to)); err != nil {
		if errors.Is(err, store.ErrGitNotInit) {
			return nil
		}

		return fmt.Errorf("failed to add %q to git: %w", to, err)
	}

	// try to enqueue this task, if the queue is not available
	// it will return the task and we will execute it inline
	t := queue.GetQueue(ctx).Add(func(ctx context.Context) (context.Context, error) {
		return nil, s.gitCommitAndPush(ctx, to)
	})

	_, err := t(ctx)

	return err
}

// IsSymlink returns true if the secret is a symlink to another secret.
func (s *Store) IsSymlink(ctx context.Context, name string) bool {
	// can't be a link if one of the files does not exist
	if !s.Exists(ctx, name) {
		return false
	}

	return s.storage.IsSymlink(ctx, s.Passfile(name))
}
