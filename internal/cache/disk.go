package cache

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/gopasspw/gopass/pkg/fsutil"
)

// OnDisk is a simple on disk cache.
type OnDisk struct {
	name string
	dir  string
}

// NewOnDisk creates a new on disk cache.
func NewOnDisk(name string) (*OnDisk, error) {
	ucd, err := os.UserCacheDir()
	if err != nil {
		return nil, err
	}
	d := filepath.Join(ucd, "gopass", name)
	if err := os.MkdirAll(d, 0755); err != nil {
		return nil, err
	}
	return &OnDisk{
		name: name,
		dir:  d,
	}, nil
}

// Get fetches an entry from the cache.
func (o *OnDisk) Get(key string) ([]string, error) {
	key = fsutil.CleanFilename(key)
	buf, err := ioutil.ReadFile(filepath.Join(o.dir, key))
	if err != nil {
		return nil, err
	}
	return strings.Split(string(buf), "\n"), nil
}

// Set adds an entry to the cache.
func (o *OnDisk) Set(key string, value []string) error {
	key = fsutil.CleanFilename(key)
	return ioutil.WriteFile(filepath.Join(o.dir, key), []byte(strings.Join(value, "\n")), 0644)
}
