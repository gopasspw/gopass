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
func New() *MockAPI {
	return &MockAPI{
		store: mockstore.New(""),
	}
}

// String returns mockapi.
func (a *MockAPI) String() string {
	return "mockapi"
}

// List does nothing.
func (a *MockAPI) List(ctx context.Context) ([]string, error) {
	return a.store.List(ctx, "") //nolint:wrapcheck
}

// Get does nothing.
func (a *MockAPI) Get(ctx context.Context, name, _ string) (gopass.Secret, error) {
	return a.store.Get(ctx, name) //nolint:wrapcheck
}

// Revisions does nothing.
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

// Set does nothing.
func (a *MockAPI) Set(ctx context.Context, name string, sec gopass.Byter) error {
	return a.store.Set(ctx, name, sec) //nolint:wrapcheck
}

// Remove does nothing.
func (a *MockAPI) Remove(ctx context.Context, name string) error {
	return a.store.Delete(ctx, name) //nolint:wrapcheck
}

// RemoveAll does nothing.
func (a *MockAPI) RemoveAll(ctx context.Context, prefix string) error {
	return a.store.Prune(ctx, prefix) //nolint:wrapcheck
}

// Rename does nothing.
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
