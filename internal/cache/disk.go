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
	if err := os.MkdirAll(d, 0755); err != nil {
		return nil, fmt.Errorf("failed to create ondisk cache dir %s: %w", d, err)
	}

	debug.Log("New on disk cache %s created at %s", name, d)

	return &OnDisk{
		ttl:  ttl,
		name: name,
		dir:  d,
	}, nil
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
	key = fsutil.CleanFilename(key)
	fn := filepath.Join(o.dir, key)
	if err := os.WriteFile(fn, []byte(strings.Join(value, "\n")), 0644); err != nil {
		return fmt.Errorf("failed to write %s to %s: %w", key, fn, err)
	}
	return nil
}
