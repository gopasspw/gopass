package backend

import (
	"context"
	"fmt"
)

var (
	cryptoRegistry  = map[CryptoBackend]CryptoLoader{}
	storageRegistry = map[StorageBackend]StorageLoader{}

	// ErrNotFound is returned if the requested backend was not found.
	ErrNotFound = fmt.Errorf("backend not found")
)

// CryptoLoader is the interface for creating a new crypto backend.
type CryptoLoader interface {
	fmt.Stringer
	New(context.Context) (Crypto, error)
	Handles(Storage) error
	Priority() int
}

// StorageLoader is the interface for creating a new storage backend.
type StorageLoader interface {
	fmt.Stringer
	New(context.Context, string) (Storage, error)
	Init(context.Context, string) (Storage, error)
	Clone(context.Context, string, string) (Storage, error)
	Handles(string) error
	Priority() int
}
