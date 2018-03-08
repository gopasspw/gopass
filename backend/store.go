package backend

import (
	"context"

	"github.com/blang/semver"
)

// StoreBackend is a type of storage backend
type StoreBackend int

const (
	// FS is a filesystem-backend storage
	FS StoreBackend = iota
	// KVMock is an in-memory mock store for tests
	KVMock
	// Consul is a consul backend storage
	Consul
)

func (s StoreBackend) String() string {
	return storeNameFromBackend(s)
}

// Store is an storage backend
type Store interface {
	Get(ctx context.Context, name string) ([]byte, error)
	Set(ctx context.Context, name string, value []byte) error
	Delete(ctx context.Context, name string) error
	Exists(ctx context.Context, name string) bool
	List(ctx context.Context, prefix string) ([]string, error)
	IsDir(ctx context.Context, name string) bool
	Prune(ctx context.Context, prefix string) error

	Name() string
	Version() semver.Version
}
