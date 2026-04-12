// Package cli implements a GPG CLI crypto backend.
package cli

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/gopasspw/gopass/internal/backend/crypto/gpg"
	"github.com/gopasspw/gopass/internal/backend/crypto/gpg/gpgconf"
	"github.com/gopasspw/gopass/pkg/debug"
	lru "github.com/hashicorp/golang-lru/v2"
)

var (
	// defaultArgs contains the default GPG args for non-interactive use. Note: Do not use '--batch'
	// as this will disable (necessary) passphrase questions!
	defaultArgs = []string{"--quiet", "--yes", "--compress-algo=none", "--no-encrypt-to", "--no-auto-check-trustdb"}
	// Ext is the file extension used by this backend.
	Ext = "gpg"
	// IDFile is the name of the recipients file used by this backend.
	IDFile = ".gpg-id"
	// Name is the name of this backend.
	Name = "gpg"
	// Timeout is the time allow for gpg invocations to complete.
	Timeout = time.Minute
)

// GPG is a gpg wrapper.
type GPG struct {
	binary    string
	args      []string
	pubKeys   gpg.KeyList
	privKeys  gpg.KeyList
	listCache *lru.TwoQueueCache[string, gpg.KeyList]
	throwKids bool
}

// Config is the gpg wrapper config.
type Config struct {
	Binary string
	Args   []string
	Umask  int
}

// New creates a new GPG wrapper.
func New(ctx context.Context, cfg Config) (*GPG, error) {
	// ensure created files don't have group or world perms set
	// this setting should be inherited by sub-processes
	gpgconf.Umask(cfg.Umask)

	// make sure GPG_TTY is set (if possible)
	if gt := os.Getenv("GPG_TTY"); gt == "" {
		if t := gpgconf.TTY(); t != "" {
			_ = os.Setenv("GPG_TTY", t)
		}
	}

	gcfg, err := gpgconf.Config()
	if err != nil {
		debug.Log("failed to read GPG config: %s", err)
	}
	_, hasThrowKids := gcfg["throw-keyids"]

	g := &GPG{
		binary:    "gpg",
		args:      append(defaultArgs, cfg.Args...),
		throwKids: hasThrowKids,
	}

	cache, err := lru.New2Q[string, gpg.KeyList](1024)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize the LRU cache: %w", err)
	}
	g.listCache = cache

	bin, err := gpgconf.Binary(ctx, cfg.Binary)
	if err != nil {
		return nil, fmt.Errorf("failed to detect binary: %w", err)
	}

	g.binary = bin
	debug.Log("binary detected as %s", bin)

	return g, nil
}

// Initialized always returns nil.
func (g *GPG) Initialized(ctx context.Context) error {
	return nil
}

// Name returns gpg.
func (g *GPG) Name() string {
	return Name
}

// Ext returns gpg.
func (g *GPG) Ext() string {
	return Ext
}

// IDFile returns .gpg-id.
func (g *GPG) IDFile() string {
	return IDFile
}

// Concurrency returns 1 to avoid concurrency issues
// with many GPG setups.
func (g *GPG) Concurrency() int {
	return 1
}

// Binary returns the GPG binary location.
func (g *GPG) Binary() string {
	if g == nil {
		return ""
	}

	return g.binary
}

// String implements fmt.Stringer.
func (g *GPG) String() string {
	var sb strings.Builder
	sb.WriteString("gpgcli(")
	if g == nil {
		sb.WriteString("<nil>)")

		return sb.String()
	}
	sb.WriteString("binary:")
	sb.WriteString(g.binary)
	sb.WriteString(",args: [")
	sb.WriteString(strings.Join(g.args, " "))
	sb.WriteString("])")

	return sb.String()
}
