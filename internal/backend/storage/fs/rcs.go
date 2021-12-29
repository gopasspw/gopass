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

// Commit does nothing.
func (s *Store) Commit(ctx context.Context, msg string) error {
	return store.ErrGitNotInit
}

// Push does nothing.
func (s *Store) Push(ctx context.Context, origin, branch string) error {
	return store.ErrGitNotInit
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
	return nil
}

// RemoveRemote does nothing.
func (s *Store) RemoveRemote(ctx context.Context, remote string) error {
	return nil
}

// Revisions is not implemented.
func (s *Store) Revisions(context.Context, string) ([]backend.Revision, error) {
	return []backend.Revision{
		{
			Hash: "latest",
			Date: time.Now(),
		}}, nil
}

// GetRevision is not implemented.
func (s *Store) GetRevision(context.Context, string, string) ([]byte, error) {
	return []byte("foo\nbar"), nil
}

// Status is not implemented.
func (s *Store) Status(context.Context) ([]byte, error) {
	return []byte(""), nil
}

// Compact is not implemented.
func (s *Store) Compact(context.Context) error {
	return nil
}
