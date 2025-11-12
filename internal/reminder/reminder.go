// Package reminder provides a reminder store for gopass.
// It stores timestamps on disk to remind the user of certain events.
package reminder

import (
	"fmt"
	"time"

	"github.com/gopasspw/gopass/internal/cache"
	"github.com/gopasspw/gopass/pkg/debug"
)

// Store stores timestamps on disk.
type Store struct {
	cache *cache.OnDisk
}

// New creates a new persistent timestamp store.
func New() (*Store, error) {
	od, err := cache.NewOnDisk("reminder", 90*24*time.Hour)
	if err != nil {
		return nil, fmt.Errorf("failed to init reminder cache: %w", err)
	}

	return &Store{
		cache: od,
	}, nil
}

// LastSeen returns the time when the key was last reset.
func (s *Store) LastSeen(key string) time.Time {
	t := time.Time{}
	if s == nil {
		return t
	}

	res, err := s.cache.Get(key)
	if err != nil {
		debug.V(2).Log("failed to read %q from cache: %s", key, err)

		return t
	}

	if len(res) < 1 {
		debug.V(1).Log("cache result is empty")

		return t
	}

	ts, err := time.Parse(time.RFC3339, res[0])
	if err != nil {
		debug.V(1).Log("failed to parse stored time %q: %s", err)

		return t
	}

	return ts
}

// Reset marks a key as just see.
func (s *Store) Reset(key string) error {
	if s == nil {
		return nil
	}

	return s.cache.Set(key, []string{time.Now().Format(time.RFC3339)})
}

// Overdue returns true iff (a) overdue did not return true within 24h AND (b)
// the key wasn't updated within the last 90 day.
func (s *Store) Overdue(key string) bool {
	if s == nil {
		return false
	}

	if time.Since(s.LastSeen("overdue")) < 24*time.Hour {
		return false
	}

	_ = s.Reset("overdue")

	return time.Since(s.LastSeen(key)) > 90*24*time.Hour
}
