package age

import (
	"context"
	"fmt"
	"path/filepath"
	"runtime"
	"time"

	"github.com/blang/semver/v4"
	"github.com/gopasspw/gopass/internal/cache"
	"github.com/gopasspw/gopass/internal/cache/ghssh"
	"github.com/gopasspw/gopass/pkg/appdir"
	"github.com/gopasspw/gopass/pkg/debug"
)

const (
	// Ext is the file extension for age encrypted secrets
	Ext = "age"
	// IDFile is the name for age recipients
	IDFile = ".age-recipients"
)

// Age is an age backend
type Age struct {
	identity  string
	ghCache   *ghssh.Cache
	askPass   *askPass
	recpCache *cache.OnDisk
}

// New creates a new Age backend
func New() (*Age, error) {
	ghc, err := ghssh.New()
	if err != nil {
		return nil, err
	}
	rc, err := cache.NewOnDisk("age-identity-recipients", 30*time.Hour)
	if err != nil {
		return nil, err
	}
	return &Age{
		ghCache:   ghc,
		recpCache: rc,
		identity:  filepath.Join(appdir.UserConfig(), "age", "identities"),
		askPass:   DefaultAskPass,
	}, nil
}

// Initialized returns nil
func (a *Age) Initialized(ctx context.Context) error {
	if a == nil {
		return fmt.Errorf("Age not initialized")
	}

	return nil
}

// Name returns age
func (a *Age) Name() string {
	return "age"
}

// Version returns the version of the age dependency being used
func (a *Age) Version(ctx context.Context) semver.Version {
	return debug.ModuleVersion("filippo.io/age")
}

// Ext returns the extension
func (a *Age) Ext() string {
	return Ext
}

// IDFile return the recipients file
func (a *Age) IDFile() string {
	return IDFile
}

// Concurrency returns the number of CPUs
func (a *Age) Concurrency() int {
	return runtime.NumCPU()
}
