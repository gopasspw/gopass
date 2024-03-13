package fs

import (
	"context"
	"time"

	"github.com/gopasspw/gopass/internal/backend"
	"github.com/gopasspw/gopass/internal/store"
)

// Add does nothing.
func (s *Store) Add(ctx context.Context, args ...string) error {
	return store.ErrGitNotInit
}

// TryAdd does nothing.
func (s *Store) TryAdd(ctx context.Context, args ...string) error {
	return nil
}

// Commit does nothing.
func (s *Store) Commit(ctx context.Context, msg string) error {
	return store.ErrGitNotInit
}

// TryCommit does nothing.
func (s *Store) TryCommit(ctx context.Context, msg string) error {
	return nil
}

// Push does nothing.
func (s *Store) Push(ctx context.Context, origin, branch string) error {
	return store.ErrGitNotInit
}

// TryPush does nothing.
func (s *Store) TryPush(ctx context.Context, origin, branch string) error {
	return nil
}

// Pull does nothing.
func (s *Store) Pull(ctx context.Context, origin, branch string) error {
	return store.ErrGitNotInit
}

// Cmd does nothing.
func (s *Store) Cmd(ctx context.Context, name string, args ...string) error {
	return nil
}

// Init does nothing.
func (s *Store) Init(context.Context, string, string) error {
	return backend.ErrNotSupported
}

// InitConfig does nothing.
func (s *Store) InitConfig(context.Context, string, string) error {
	return nil
}

// AddRemote does nothing.
func (s *Store) AddRemote(ctx context.Context, remote, url string) error {
	return backend.ErrNotSupported
}

// RemoveRemote does nothing.
func (s *Store) RemoveRemote(ctx context.Context, remote string) error {
	return backend.ErrNotSupported
}

// Revisions is not implemented.
func (s *Store) Revisions(context.Context, string) ([]backend.Revision, error) {
	return []backend.Revision{
		{
			Hash: "latest",
			Date: time.Now(),
		},
	}, backend.ErrNotSupported
}

// GetRevision only supports getting the latest revision.
func (s *Store) GetRevision(ctx context.Context, name string, revision string) ([]byte, error) {
	if revision == "HEAD" || revision == "latest" {
		return s.Get(ctx, name)
	}

	return []byte(""), backend.ErrNotSupported
}

// Status is not implemented.
func (s *Store) Status(context.Context) ([]byte, error) {
	return []byte(""), backend.ErrNotSupported
}

// Compact is not implemented.
func (s *Store) Compact(context.Context) error {
	return nil
}
