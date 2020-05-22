package backend

import (
	"context"
	"fmt"

	"github.com/gopasspw/gopass/internal/out"
	"github.com/pkg/errors"
)

var (
	cryptoRegistry  = map[CryptoBackend]CryptoLoader{}
	rcsRegistry     = map[RCSBackend]RCSLoader{}
	storageRegistry = map[StorageBackend]StorageLoader{}

	// ErrNotFound is returned if the requested backend was not found.
	ErrNotFound = fmt.Errorf("backend not found")
)

// CryptoLoader is the interface for creating a new crypto backend.
type CryptoLoader interface {
	fmt.Stringer
	New(context.Context) (Crypto, error)
}

// RCSLoader is the interface for creating a new RCS backend.
type RCSLoader interface {
	fmt.Stringer
	Open(context.Context, string) (RCS, error)
	Clone(context.Context, string, string) (RCS, error)
	Init(context.Context, string, string, string) (RCS, error)
}

// StorageLoader is the interface for creating a new storage backend.
type StorageLoader interface {
	fmt.Stringer
	New(context.Context, *URL) (Storage, error)
}

// RegisterCrypto registers a new crypto backend with the backend registry.
func RegisterCrypto(id CryptoBackend, name string, loader CryptoLoader) {
	cryptoRegistry[id] = loader
	cryptoNameToBackendMap[name] = id
	cryptoBackendToNameMap[id] = name
}

// NewCrypto instantiates a new crypto backend.
func NewCrypto(ctx context.Context, id CryptoBackend) (Crypto, error) {
	if be, found := cryptoRegistry[id]; found {
		return be.New(ctx)
	}
	return nil, errors.Wrapf(ErrNotFound, "unknown backend: %d", id)
}

// RegisterRCS registers a new RCS backend with the backend registry.
func RegisterRCS(id RCSBackend, name string, loader RCSLoader) {
	rcsRegistry[id] = loader
	rcsNameToBackendMap[name] = id
	rcsBackendToNameMap[id] = name
}

// OpenRCS opens an existing repository.
func OpenRCS(ctx context.Context, id RCSBackend, path string) (RCS, error) {
	if be, found := rcsRegistry[id]; found {
		return be.Open(ctx, path)
	}
	return nil, errors.Wrapf(ErrNotFound, "unknown backend: %d", id)
}

// CloneRCS clones an existing repository from a remote.
func CloneRCS(ctx context.Context, id RCSBackend, repo, path string) (RCS, error) {
	if be, found := rcsRegistry[id]; found {
		out.Debug(ctx, "Cloning with %s", be.String())
		return be.Clone(ctx, repo, path)
	}
	return nil, errors.Wrapf(ErrNotFound, "unknown backend: %d", id)
}

// InitRCS initializes a new repository.
func InitRCS(ctx context.Context, id RCSBackend, path, name, email string) (RCS, error) {
	if be, found := rcsRegistry[id]; found {
		return be.Init(ctx, path, name, email)
	}
	return nil, errors.Wrapf(ErrNotFound, "unknown backend: %d", id)
}

// RegisterStorage registers a new storage backend with the registry.
func RegisterStorage(id StorageBackend, name string, loader StorageLoader) {
	storageRegistry[id] = loader
	storageNameToBackendMap[name] = id
	storageBackendToNameMap[id] = name
}

// NewStorage initializes a new storage backend.
func NewStorage(ctx context.Context, id StorageBackend, url *URL) (Storage, error) {
	if be, found := storageRegistry[id]; found {
		return be.New(ctx, url)
	}
	return nil, errors.Wrapf(ErrNotFound, "unknown backend: %s", url.String())
}
