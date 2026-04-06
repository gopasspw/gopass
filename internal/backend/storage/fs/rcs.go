// This file contains stub implementations for the rcs interface that is
// embedded in backend.Storage. The fs backend intentionally has no VCS
// support — it exists primarily as a lightweight storage layer for tests
// and for users who manage versioning through an external mechanism (e.g.
// a FUSE overlay, a network filesystem with versioning, etc.).
//
// # Leaky abstraction note
//
// The backend.Storage interface embeds the rcs interface, which means every
// Storage implementation must also satisfy all VCS operations (Add, Commit,
// Push, Pull, Revisions, …). This is a deliberate trade-off: in practice the
// overwhelming majority of storage backends (gitfs, fossilfs, jjfs) are also
// RCS backends, so a single unified interface keeps the API surface small and
// avoids an extra type-assertion at every call site in the leaf store.
//
// The downside is that the fs backend ends up with ~15 stub methods here, most
// of which return store.ErrGitNotInit or backend.ErrNotSupported. A previous
// iteration of the codebase kept Storage and RCS as separate interfaces; the
// decision to merge them was made to simplify the leaf store and reduce the
// indirection layer. If the number of pure-storage backends grows, or if the
// stub overhead becomes a maintenance burden, see
// docs/adr/A-3-separate-storage-rcs.md for a ready-made plan to split them
// again.
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
