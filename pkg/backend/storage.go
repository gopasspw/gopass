package backend

import (
	"context"

	"github.com/blang/semver"
)

// StorageBackend is a type of storage backend
type StorageBackend int

const (
	// FS is a filesystem-backend storage
	FS StorageBackend = iota
	// InMem is an in-memory mock store for tests
	InMem
	// Consul is a consul backend storage
	Consul
)

func (s StorageBackend) String() string {
	return storageNameFromBackend(s)
}

// Storage is an storage backend
type Storage interface {
	Get(ctx context.Context, name string) ([]byte, error)
	Set(ctx context.Context, name string, value []byte) error
	Delete(ctx context.Context, name string) error
	Exists(ctx context.Context, name string) bool
	List(ctx context.Context, prefix string) ([]string, error)
	IsDir(ctx context.Context, name string) bool
	Prune(ctx context.Context, prefix string) error
	Available(ctx context.Context) error

	Name() string
	Version() semver.Version
	Fsck(context.Context) error
}
