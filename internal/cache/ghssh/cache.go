package ghssh

import (
	"fmt"
	"time"

	"github.com/google/go-github/github"
	"github.com/gopasspw/gopass/internal/cache"
)

// Cache is a disk-backed GitHub SSH public key cache.
type Cache struct {
	disk    *cache.OnDisk
	client  *github.Client
	Timeout time.Duration
}

// New creates a new github cache.
func New() (*Cache, error) {
	cDir, err := cache.NewOnDisk("github-ssh", 6*time.Hour)
	if err != nil {
		return nil, err
	}

	return &Cache{
		disk:    cDir,
		client:  github.NewClient(nil),
		Timeout: 30 * time.Second,
	}, nil
}

func (c *Cache) String() string {
	return fmt.Sprintf("Github SSH key cache (OnDisk: %s)", c.disk.String())
}
