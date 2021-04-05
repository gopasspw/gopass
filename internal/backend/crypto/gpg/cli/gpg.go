// Package cli implements a GPG CLI crypto backend.
package cli

import (
	"context"
	"os"

	"github.com/gopasspw/gopass/internal/backend/crypto/gpg"
	"github.com/gopasspw/gopass/pkg/debug"
	lru "github.com/hashicorp/golang-lru"
)

var (
	// defaultArgs contains the default GPG args for non-interactive use. Note: Do not use '--batch'
	// as this will disable (necessary) passphrase questions!
	defaultArgs = []string{"--quiet", "--yes", "--compress-algo=none", "--no-encrypt-to", "--no-auto-check-trustdb"}
	// Ext is the file extension used by this backend
	Ext = "gpg"
	// IDFile is the name of the recipients file used by this backend
	IDFile = ".gpg-id"
)

// GPG is a gpg wrapper
type GPG struct {
	binary    string
	args      []string
	pubKeys   gpg.KeyList
	privKeys  gpg.KeyList
	listCache *lru.TwoQueueCache
	throwKids bool
}

// Config is the gpg wrapper config
type Config struct {
	Binary string
	Args   []string
	Umask  int
}

// New creates a new GPG wrapper
func New(ctx context.Context, cfg Config) (*GPG, error) {
	// ensure created files don't have group or world perms set
	// this setting should be inherited by sub-processes
	umask(cfg.Umask)

	// make sure GPG_TTY is set (if possible)
	if gt := os.Getenv("GPG_TTY"); gt == "" {
		if t := tty(); t != "" {
			_ = os.Setenv("GPG_TTY", t)
		}
	}

	gcfg, err := gpgConfig()
	if err != nil {
		debug.Log("failed to read GPG config: %s", err)
	}
	_, throwKids := gcfg["throw-keyids"]

	g := &GPG{
		binary:    "gpg",
		args:      append(defaultArgs, cfg.Args...),
		throwKids: throwKids,
	}

	debug.Log("initializing LRU cache")
	cache, err := lru.New2Q(1024)
	if err != nil {
		return nil, err
	}
	g.listCache = cache
	debug.Log("LRU cache initialized")

	debug.Log("detecting binary")
	bin, err := Binary(ctx, cfg.Binary)
	if err != nil {
		return nil, err
	}
	g.binary = bin
	debug.Log("binary detected")

	return g, nil
}

// Initialized always returns nil
func (g *GPG) Initialized(ctx context.Context) error {
	return nil
}

// Name returns gpg
func (g *GPG) Name() string {
	return "gpg"
}

// Ext returns gpg
func (g *GPG) Ext() string {
	return Ext
}

// IDFile returns .gpg-id
func (g *GPG) IDFile() string {
	return IDFile
}
