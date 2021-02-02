package age

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"filippo.io/age"
	"filippo.io/age/agessh"
	"github.com/blang/semver/v4"
	"github.com/google/go-github/github"
	"github.com/gopasspw/gopass/internal/cache"
	"github.com/gopasspw/gopass/pkg/appdir"
	"github.com/gopasspw/gopass/pkg/debug"
	"github.com/gopasspw/gopass/pkg/termio"
)

const (
	// Ext is the file extension for age encrypted secrets
	Ext = "age"
	// IDFile is the name for age recipients
	IDFile = ".age-ids"
)

// Age is an age backend
type Age struct {
	binary  string
	keyring string
	ghc     *github.Client
	ghCache *cache.OnDisk
	askPass *askPass
	krCache map[string]age.Identity
}

// New creates a new Age backend
func New() (*Age, error) {
	cDir, err := cache.NewOnDisk("age-github", 6*time.Hour)
	if err != nil {
		return nil, err
	}
	return &Age{
		binary:  "age",
		ghc:     github.NewClient(nil),
		ghCache: cDir,
		keyring: filepath.Join(appdir.UserConfig(), "age-keyring.age"),
		askPass: DefaultAskPass,
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

// Version return 1.0.0
func (a *Age) Version(ctx context.Context) semver.Version {
	return semver.Version{
		Patch: 1,
	}
}

// Ext returns the extension
func (a *Age) Ext() string {
	return Ext
}

// IDFile return the recipients file
func (a *Age) IDFile() string {
	return IDFile
}

func (a *Age) parseRecipients(ctx context.Context, recipients []string) ([]age.Recipient, error) {
	out := make([]age.Recipient, 0, len(recipients))
	for _, r := range recipients {
		if strings.HasPrefix(r, "age1") {
			id, err := age.ParseX25519Recipient(r)
			if err != nil {
				debug.Log("Failed to parse recipient '%s' as X25519: %s", r, err)
				continue
			}
			out = append(out, id)
			continue
		}
		if strings.HasPrefix(r, "ssh-") {
			id, err := agessh.ParseRecipient(r)
			if err != nil {
				debug.Log("Failed to parse recipient '%s' as SSH: %s", r, err)
				continue
			}
			out = append(out, id)
			continue
		}
		if strings.HasPrefix(r, "github:") {
			pks, err := a.getPublicKeysGithub(ctx, strings.TrimPrefix(r, "github:"))
			if err != nil {
				return out, err
			}
			for _, pk := range pks {
				id, err := agessh.ParseRecipient(r)
				if err != nil {
					debug.Log("Failed to parse GitHub recipient '%s': '%s': %s", r, pk, err)
					continue
				}
				out = append(out, id)
			}
		}
	}
	return out, nil
}

// ListIdentities lists all identities
func (a *Age) ListIdentities(ctx context.Context) ([]string, error) {
	ids, err := a.getAllIdentities(ctx)
	if err != nil {
		return nil, err
	}

	idStr := make([]string, 0, len(ids))
	for k := range ids {
		idStr = append(idStr, k)
	}
	sort.Strings(idStr)
	return idStr, nil
}

func (a *Age) getAllIds(ctx context.Context) ([]age.Identity, error) {
	ids, err := a.getAllIdentities(ctx)
	if err != nil {
		return nil, err
	}
	idl := make([]age.Identity, 0, len(ids))
	for _, id := range ids {
		idl = append(idl, id)
	}
	return idl, nil
}

func (a *Age) getAllIdentities(ctx context.Context) (map[string]age.Identity, error) {
	native, err := a.getNativeIdentities(ctx)
	if err != nil {
		return nil, err
	}
	ssh, err := a.getSSHIdentities(ctx)
	if err != nil {
		return nil, err
	}
	for k, v := range ssh {
		native[k] = v
	}

	return native, nil
}

func (a *Age) getNativeIdentities(ctx context.Context) (map[string]age.Identity, error) {
	if len(a.krCache) > 0 {
		return a.krCache, nil
	}
	kr, err := a.loadKeyring(ctx)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		debug.Log("failed to load native identities: %+v", err)
		return nil, err
	}
	debug.Log("keyring: %+v", kr)
	if len(kr) < 1 {
		// TODO we shouldn't print in here, use a callback
		ok, err := termio.AskForBool(ctx, "🔑 No existing age identities found. Do you want to generate a new one?", true)
		if err != nil {
			return nil, err
		}
		if !ok {
			return nil, fmt.Errorf("user aborted")
		}
		debug.Log("generating new age keypair")
		id, err := a.genKey(ctx)
		if err != nil {
			return nil, err
		}
		return map[string]age.Identity{
			id.Recipient().String(): id,
		}, nil
	}
	ids := make(map[string]age.Identity, len(kr))
	for _, k := range kr {
		id, err := age.ParseX25519Identity(k.Identity)
		if err != nil {
			debug.Log("Failed to parse identity '%s': %s", k, err)
			continue
		}
		ids[id.Recipient().String()] = id
	}
	a.krCache = ids
	return ids, nil
}
