package backend

import (
	"cmp"
	"context"
	"fmt"
	"maps"
	"slices"
	"sync"

	"github.com/gopasspw/gopass/internal/set"
)

var (
	// CryptoRegistry is the global registry of available crypto backends.
	CryptoRegistry = NewRegistry[CryptoBackend, CryptoLoader]()
	// StorageRegistry is the global registry of available storage backends.
	StorageRegistry = NewRegistry[StorageBackend, StorageLoader]()

	// ErrNotFound is returned if the requested backend was not found.
	ErrNotFound = fmt.Errorf("backend not found")
)

// Prioritized is the interface for prioritized items.
type Prioritized interface {
	Priority() int
}

// CryptoLoader is the interface for creating a new crypto backend.
type CryptoLoader interface {
	fmt.Stringer
	Prioritized
	New(context.Context) (Crypto, error)
	Handles(context.Context, Storage) error
}

// StorageLoader is the interface for creating a new storage backend.
type StorageLoader interface {
	fmt.Stringer
	Prioritized
	New(context.Context, string) (Storage, error)
	Init(context.Context, string) (Storage, error)
	Clone(context.Context, string, string) (Storage, error)
	Handles(context.Context, string) error
}

// NewRegistry returns a new registry.
func NewRegistry[K comparable, V Prioritized]() *Registry[K, V] {
	return &Registry[K, V]{
		backends:      map[K]V{},
		nameToBackend: map[string]K{},
		backendToName: map[K]string{},
	}
}

// Registry is a registry of backends.
type Registry[K comparable, V Prioritized] struct {
	sync.RWMutex

	backends      map[K]V
	nameToBackend map[string]K
	backendToName map[K]string
}

func (r *Registry[K, V]) Register(backend K, name string, loader V) {
	r.Lock()
	defer r.Unlock()

	r.backends[backend] = loader
	r.nameToBackend[name] = backend
	r.backendToName[backend] = name
}

func (r *Registry[K, V]) BackendNames() []string {
	r.RLock()
	defer r.RUnlock()

	return set.SortedKeys(r.nameToBackend)
}

func (r *Registry[K, V]) Backends() []V {
	r.RLock()
	defer r.RUnlock()

	bes := make([]V, 0, len(r.backends))
	for _, be := range r.backends {
		bes = append(bes, be)
	}

	return bes
}

func (r *Registry[K, V]) Prioritized() []V {
	r.RLock()
	defer r.RUnlock()

	bes := maps.Values(r.backends)

	return slices.SortedFunc(bes, func(a, b V) int {
		return cmp.Compare(a.Priority(), b.Priority())
	})
}

func (r *Registry[K, V]) Get(key K) (V, error) {
	r.RLock()
	defer r.RUnlock()

	if be, found := r.backends[key]; found {
		return be, nil
	}
	var zero V

	return zero, ErrNotFound
}

func (r *Registry[K, V]) Backend(name string) (K, error) {
	r.RLock()
	defer r.RUnlock()

	if name == "gpg" {
		name = "gpgcli"
	}
	backend, ok := r.nameToBackend[name]
	if !ok {
		var zero K

		return zero, ErrNotFound
	}

	return backend, nil
}

func (r *Registry[K, V]) BackendName(backend K) (string, error) {
	r.RLock()
	defer r.RUnlock()

	name, ok := r.backendToName[backend]
	if !ok {
		return "", ErrNotFound
	}

	return name, nil
}
