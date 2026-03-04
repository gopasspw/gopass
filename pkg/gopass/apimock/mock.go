// Package apimock provides a mock implementation of the gopass API.
// This is useful for testing purposes and allows to simulate different
// scenarios without relying on a real backend.
package apimock

import (
	"context"
	"fmt"

	"github.com/gopasspw/gopass/internal/store/mockstore"
	"github.com/gopasspw/gopass/pkg/gopass"
)

// ErrNotImplemented is returned when a method is not implemented.
var ErrNotImplemented = fmt.Errorf("not yet implemented")

// Secret is a mock secret for writing.
type Secret struct {
	Buf []byte
}

// Bytes returns the underlying bytes.
func (m *Secret) Bytes() []byte {
	return m.Buf
}

// MockAPI is a gopass API mock.
type MockAPI struct {
	store *mockstore.MockStore
}

// New creates a new gopass API mock.
// It uses an in-memory store.
func New() *MockAPI {
	return &MockAPI{
		store: mockstore.New(""),
	}
}

// String returns the name of the mock API.
func (a *MockAPI) String() string {
	return "mockapi"
}

// List returns a list of all secrets in the mock store.
func (a *MockAPI) List(ctx context.Context) ([]string, error) {
	return a.store.List(ctx, "") //nolint:wrapcheck
}

// Get returns a secret from the mock store.
func (a *MockAPI) Get(ctx context.Context, name, _ string) (gopass.Secret, error) {
	return a.store.Get(ctx, name) //nolint:wrapcheck
}

// Revisions returns a list of all revisions of a secret in the mock store.
func (a *MockAPI) Revisions(ctx context.Context, name string) ([]string, error) {
	rs, err := a.store.ListRevisions(ctx, name)
	if err != nil {
		return nil, err //nolint:wrapcheck
	}

	revs := make([]string, 0, len(rs))
	for _, r := range rs {
		revs = append(revs, r.Hash)
	}

	return revs, nil
}

// Set sets a secret in the mock store.
func (a *MockAPI) Set(ctx context.Context, name string, sec gopass.Byter) error {
	return a.store.Set(ctx, name, sec) //nolint:wrapcheck
}

// Remove removes a secret from the mock store.
func (a *MockAPI) Remove(ctx context.Context, name string) error {
	return a.store.Delete(ctx, name) //nolint:wrapcheck
}

// RemoveAll removes all secrets with a given prefix from the mock store.
func (a *MockAPI) RemoveAll(ctx context.Context, prefix string) error {
	return a.store.Prune(ctx, prefix) //nolint:wrapcheck
}

// Rename moves a secret in the mock store.
func (a *MockAPI) Rename(ctx context.Context, src, dest string) error {
	return a.store.Move(ctx, src, dest) //nolint:wrapcheck
}

// Sync does nothing.
func (a *MockAPI) Sync(ctx context.Context) error {
	return ErrNotImplemented
}

// Close does nothing.
func (a *MockAPI) Close(ctx context.Context) error {
	return nil
}
