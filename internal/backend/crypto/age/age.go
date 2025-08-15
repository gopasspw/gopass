package age

import (
	"context"
	"fmt"
	"path/filepath"
	"time"

	"github.com/blang/semver/v4"
	"github.com/gopasspw/gopass/internal/cache"
	"github.com/gopasspw/gopass/internal/cache/ghssh"
	"github.com/gopasspw/gopass/pkg/appdir"
	"github.com/gopasspw/gopass/pkg/debug"
)

const (
	// Ext is the file extension for age encrypted secrets.
	Ext = "age"
	// IDFile is the name for age recipients.
	IDFile = ".age-recipients"
)

type githubSSHCacher interface {
	ListKeys(ctx context.Context, user string) ([]string, error)
	String() string
}

// Age is an age backend.
type Age struct {
	identity   string
	ghCache    githubSSHCacher
	askPass    *askPass
	recpCache  *cache.OnDisk
	sshKeyPath string // custom SSH key or directory path
}

// New creates a new Age backend.
func New(ctx context.Context, sshKeyPath string) (*Age, error) {
	ghc, err := ghssh.New()
	if err != nil {
		return nil, err
	}

	rc, err := cache.NewOnDisk("age-identity-recipients", 30*time.Hour)
	if err != nil {
		return nil, err
	}

	a := &Age{
		ghCache:    ghc,
		recpCache:  rc,
		identity:   filepath.Join(appdir.UserConfig(), "age", "identities"),
		askPass:    newAskPass(ctx),
		sshKeyPath: sshKeyPath,
	}

	debug.Log("age initialized (ghc: %s, recipients: %s, identity: %s, sshKeyPath: %s)", a.ghCache.String(), a.recpCache.String(), a.identity, a.sshKeyPath)

	return a, nil
}

// Initialized returns nil.
func (a *Age) Initialized(ctx context.Context) error {
	if a == nil {
		return fmt.Errorf("Age not initialized")
	}

	return nil
}

// Name returns age.
func (a *Age) Name() string {
	return "age"
}

// Version returns the version of the age dependency being used.
func (a *Age) Version(ctx context.Context) semver.Version {
	return debug.ModuleVersion("filippo.io/age")
}

// Ext returns the extension.
func (a *Age) Ext() string {
	return Ext
}

// IDFile return the recipients file.
func (a *Age) IDFile() string {
	return IDFile
}

// Concurrency returns 1 for `age` since otherwise it prompts for the identity password for each worker.
func (a *Age) Concurrency() int {
	return 1
}

// Add a method to get the SSH key path.
func (a *Age) SSHKeyPath() string {
	return a.sshKeyPath
}
