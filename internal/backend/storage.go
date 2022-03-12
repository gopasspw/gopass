package backend

import (
	"context"
	"fmt"

	"github.com/blang/semver/v4"
	"github.com/gopasspw/gopass/pkg/debug"
)

// ErrNotSupported is returned by backends for unsupported calls.
var ErrNotSupported = fmt.Errorf("not supported")

// StorageBackend is a type of storage backend.
type StorageBackend int

const (
	// FS is a filesystem-backed storage.
	FS StorageBackend = iota
	// GitFS is a filesystem-backed storage with Git.
	GitFS
	// FossilFS is a filesystem-backed storage with Fossil.
	FossilFS
)

func (s StorageBackend) String() string {
	if be, err := StorageRegistry.BackendName(s); err == nil {
		return be
	}

	return ""
}

// Storage is an storage backend.
type Storage interface {
	fmt.Stringer
	rcs
	Get(ctx context.Context, name string) ([]byte, error)
	Set(ctx context.Context, name string, value []byte) error
	Delete(ctx context.Context, name string) error
	Exists(ctx context.Context, name string) bool
	List(ctx context.Context, prefix string) ([]string, error)
	IsDir(ctx context.Context, name string) bool
	Prune(ctx context.Context, prefix string) error
	Link(ctx context.Context, from, to string) error

	Name() string
	Path() string
	Version(context.Context) semver.Version
	Fsck(context.Context) error
}

// DetectStorage tries to detect the storage backend being used.
func DetectStorage(ctx context.Context, path string) (Storage, error) {
	// The call to HasStorageBackend is important since GetStorageBackend will always return FS
	// if nothing is found in the context.
	if be, err := StorageRegistry.Get(GetStorageBackend(ctx)); HasStorageBackend(ctx) && err == nil {
		debug.Log("Trying requested %s for %s", be, path)
		st, err := be.New(ctx, path)
		if err == nil {
			debug.Log("Using requested %s for %s", be, path)

			return st, nil
		}
		debug.Log("Failed to use requested %s for %s: %s", be, path, err)

		// fallback to FS
		be, err := StorageRegistry.Get(FS)
		if err != nil {
			return nil, err
		}
		debug.Log("Using fallback %q for %q", be, path)

		return be.Init(ctx, path)
	}

	// Nothing requested in the context. Try to detect the backend.
	for _, be := range StorageRegistry.Prioritized() {
		debug.Log("Trying %s for %s", be, path)
		if err := be.Handles(ctx, path); err != nil {
			debug.Log("failed to use %s for %s: %s", be, path, err)

			continue
		}
		debug.Log("Using detected %s for %s", be, path)

		return be.New(ctx, path)
	}

	// fallback to FS
	be, err := StorageRegistry.Get(FS)
	if err != nil {
		return nil, err
	}
	debug.Log("Using default fallback %q for %q", be, path)

	return be.Init(ctx, path)
}

// NewStorage initializes an existing storage backend.
func NewStorage(ctx context.Context, id StorageBackend, path string) (Storage, error) {
	if be, err := StorageRegistry.Get(id); err == nil {
		debug.Log("Using %s for %s", be, path)

		return be.New(ctx, path)
	}

	return nil, fmt.Errorf("unknown backend %q: %w", path, ErrNotFound)
}

// InitStorage initilizes a new storage location.
func InitStorage(ctx context.Context, id StorageBackend, path string) (Storage, error) {
	if be, err := StorageRegistry.Get(id); err == nil {
		debug.Log("Using %s for %s", be, path)

		return be.Init(ctx, path)
	}

	return nil, fmt.Errorf("unknown backend %q: %w", path, ErrNotFound)
}
