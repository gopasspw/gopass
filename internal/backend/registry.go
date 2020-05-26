package backend

import (
	"context"
	"fmt"
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
	Handles(Storage) error
	Priority() int
}

// RCSLoader is the interface for creating a new RCS backend.
type RCSLoader interface {
	fmt.Stringer
	Open(context.Context, string) (RCS, error)
	Clone(context.Context, string, string) (RCS, error)
	InitRCS(context.Context, string) (RCS, error)
	Handles(string) error
	Priority() int
}

// StorageLoader is the interface for creating a new storage backend.
type StorageLoader interface {
	fmt.Stringer
	New(context.Context, string) (Storage, error)
	Init(context.Context, string) (Storage, error)
	Handles(string) error
	Priority() int
}
