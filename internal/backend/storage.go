package backend

import (
	"context"
	"fmt"
	"sort"

	"github.com/blang/semver"
	"github.com/gopasspw/gopass/internal/debug"
	"github.com/pkg/errors"
)

// StorageBackend is a type of storage backend
type StorageBackend int

const (
	// FS is a filesystem-backend storage
	FS StorageBackend = iota
	// InMem is an in-memory mock store for tests
	InMem
	// OnDisk is an on-disk store
	OnDisk
)

func (s StorageBackend) String() string {
	return StorageNameFromBackend(s)
}

// Storage is an storage backend
type Storage interface {
	fmt.Stringer
	Get(ctx context.Context, name string) ([]byte, error)
	Set(ctx context.Context, name string, value []byte) error
	Delete(ctx context.Context, name string) error
	Exists(ctx context.Context, name string) bool
	List(ctx context.Context, prefix string) ([]string, error)
	IsDir(ctx context.Context, name string) bool
	Prune(ctx context.Context, prefix string) error
	Available(ctx context.Context) error

	Name() string
	Path() string
	Version(context.Context) semver.Version
	Fsck(context.Context) error
}

// RegisterStorage registers a new storage backend with the registry.
func RegisterStorage(id StorageBackend, name string, loader StorageLoader) {
	storageRegistry[id] = loader
	storageNameToBackendMap[name] = id
	storageBackendToNameMap[id] = name
}

// DetectStorage tries to detect the storage backend being used
func DetectStorage(ctx context.Context, path string) (Storage, error) {
	if HasStorageBackend(ctx) {
		if be, found := storageRegistry[GetStorageBackend(ctx)]; found {
			st, err := be.New(ctx, path)
			if err == nil {
				return st, nil
			}
			return storageRegistry[FS].Init(ctx, path)
		}
	}
	bes := make([]StorageBackend, 0, len(storageRegistry))
	for id := range storageRegistry {
		bes = append(bes, id)
	}
	sort.Slice(bes, func(i, j int) bool {
		return storageRegistry[bes[i]].Priority() < storageRegistry[bes[j]].Priority()
	})
	for _, id := range bes {
		be := storageRegistry[id]
		debug.Log("Trying %s for %s", be, path)
		if err := be.Handles(path); err != nil {
			debug.Log("failed to use %s for %s: %s", id, path, err)
			continue
		}
		debug.Log("Using %s for %s", be, path)
		return be.New(ctx, path)
	}
	return storageRegistry[FS].Init(ctx, path)
}

// NewStorage initializes an existing storage backend.
func NewStorage(ctx context.Context, id StorageBackend, path string) (Storage, error) {
	if be, found := storageRegistry[id]; found {
		return be.New(ctx, path)
	}
	return nil, errors.Wrapf(ErrNotFound, "unknown backend: %s", path)
}

// InitStorage initilizes a new storage location.
func InitStorage(ctx context.Context, id StorageBackend, path string) (Storage, error) {
	if be, found := storageRegistry[id]; found {
		return be.Init(ctx, path)
	}
	return nil, errors.Wrapf(ErrNotFound, "unknown backend: %s", path)
}
