package cache

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gopasspw/gopass/pkg/appdir"
	"github.com/gopasspw/gopass/pkg/debug"
	"github.com/gopasspw/gopass/pkg/fsutil"
)

// OnDisk is a simple on disk cache.
type OnDisk struct {
	ttl  time.Duration
	name string
	dir  string
}

// NewOnDisk creates a new on disk cache.
func NewOnDisk(name string, ttl time.Duration) (*OnDisk, error) {
	d := filepath.Join(appdir.UserCache(), "gopass", name)
	debug.Log("New on disk cache %s created at %s", name, d)

	o := &OnDisk{
		ttl:  ttl,
		name: name,
		dir:  d,
	}
	return o, o.ensureDir()
}

func (o *OnDisk) ensureDir() error {
	if err := os.MkdirAll(o.dir, 0700); err != nil {
		return fmt.Errorf("failed to create ondisk cache dir %s: %w", o.dir, err)
	}
	return nil
}

// Get fetches an entry from the cache.
func (o *OnDisk) Get(key string) ([]string, error) {
	key = fsutil.CleanFilename(key)
	fn := filepath.Join(o.dir, key)
	fi, err := os.Stat(fn)
	if err != nil {
		return nil, fmt.Errorf("failed to stat %s: %w", fn, err)
	}

	if time.Now().After(fi.ModTime().Add(o.ttl)) {
		return nil, fmt.Errorf("expired")
	}

	buf, err := os.ReadFile(fn)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", fn, err)
	}

	return strings.Split(string(buf), "\n"), nil
}

// Set adds an entry to the cache.
func (o *OnDisk) Set(key string, value []string) error {
	if err := o.ensureDir(); err != nil {
		return err
	}
	key = fsutil.CleanFilename(key)
	fn := filepath.Join(o.dir, key)
	if err := os.WriteFile(fn, []byte(strings.Join(value, "\n")), 0644); err != nil {
		return fmt.Errorf("failed to write %s to %s: %w", key, fn, err)
	}
	return nil
}

// ModTime returns the modification time of the cache entry.
func (o *OnDisk) ModTime(key string) time.Time {
	key = fsutil.CleanFilename(key)
	fn := filepath.Join(o.dir, key)
	fi, err := os.Stat(fn)
	if err != nil {
		return time.Time{}
	}
	return fi.ModTime()
}

// Remove removes an entry from the cache.
func (o *OnDisk) Remove(key string) error {
	if err := o.ensureDir(); err != nil {
		return err
	}
	key = fsutil.CleanFilename(key)
	fn := filepath.Join(o.dir, key)
	return os.Remove(fn)
}

// Purge removes all entries from the cache.
func (o *OnDisk) Purge() error {
	return os.RemoveAll(o.dir)
}
