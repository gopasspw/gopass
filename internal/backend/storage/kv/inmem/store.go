// Package inmem implements an in memory storage backend for tests.
// TODO(2.x) DEPRECATED and slated for removal in the 2.0.0 release.
package inmem

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"sync"

	"github.com/blang/semver"
)

// InMem is a in-memory store
type InMem struct {
	sync.Mutex
	data map[string][]byte
}

// New creates a new mock
func New() *InMem {
	return &InMem{
		data: make(map[string][]byte, 10),
	}
}

// Get retrieves a value
func (m *InMem) Get(ctx context.Context, name string) ([]byte, error) {
	m.Lock()
	defer m.Unlock()

	sec, found := m.data[name]
	if !found {
		return nil, fmt.Errorf("entry not found")
	}
	return sec, nil
}

// Set writes a value
func (m *InMem) Set(ctx context.Context, name string, value []byte) error {
	m.Lock()
	defer m.Unlock()

	m.data[name] = value
	return nil
}

// Delete removes a value
func (m *InMem) Delete(ctx context.Context, name string) error {
	m.Lock()
	defer m.Unlock()

	delete(m.data, name)
	return nil
}

// Exists checks is a value exists
func (m *InMem) Exists(ctx context.Context, name string) bool {
	m.Lock()
	defer m.Unlock()

	_, found := m.data[name]
	return found
}

// List shows all values
func (m *InMem) List(ctx context.Context, prefix string) ([]string, error) {
	m.Lock()
	defer m.Unlock()

	keys := make([]string, 0, len(m.data))
	for k := range m.data {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys, nil
}

// IsDir returns true if the entry is a directory
func (m *InMem) IsDir(ctx context.Context, name string) bool {
	m.Lock()
	defer m.Unlock()

	for k := range m.data {
		if strings.HasPrefix(k, name+"/") {
			return true
		}
	}
	return false
}

// Prune removes a directory
func (m *InMem) Prune(ctx context.Context, prefix string) error {
	m.Lock()
	defer m.Unlock()

	deleted := 0
	for k := range m.data {
		if strings.HasPrefix(k, prefix+"/") {
			delete(m.data, k)
			deleted++
		}
	}
	if deleted < 1 {
		return fmt.Errorf("not found")
	}
	return nil
}

// Name returns the name of this backend
func (m *InMem) Name() string {
	return "inmem"
}

// Version returns the version of this backend
func (m *InMem) Version(context.Context) semver.Version {
	return semver.Version{Major: 1}
}

// String implement fmt.Stringer
func (m *InMem) String() string {
	return "inmem()"
}

// Available will check if this backend is useable
func (m *InMem) Available(ctx context.Context) error {
	if m.data == nil {
		return fmt.Errorf("not initialized")
	}
	return nil
}

// Fsck always returns nil
func (m *InMem) Fsck(ctx context.Context) error {
	return nil
}
