package age

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/gopasspw/gopass/internal/backend"
	"github.com/gopasspw/gopass/internal/out"
	"github.com/gopasspw/gopass/pkg/appdir"
	"github.com/gopasspw/gopass/pkg/ctxutil"
	"github.com/gopasspw/gopass/pkg/debug"
	"github.com/gopasspw/gopass/pkg/fsutil"
)

var (
	// OldIDFile is the old file name for the recipients.
	OldIDFile = ".age-ids"
	// OldKeyring is the old file name for the keyring.
	OldKeyring = filepath.Join(appdir.UserConfig(), "age-keyring.age")
)

func migrate(ctx context.Context, s backend.Storage) error {
	out.Noticef(ctx, "Attempting to migrate age backend. You will need to unlock your identities keyring.")

	oldIDPath := filepath.Join(s.Path(), OldIDFile)
	newIDPath := filepath.Join(s.Path(), IDFile)
	if fsutil.IsFile(oldIDPath) && fsutil.IsFile(newIDPath) {
		out.Warningf(ctx, "Both %s and %s exist. Removing the old one (%s).", oldIDPath, newIDPath, oldIDPath)
		if err := os.Remove(oldIDPath); err != nil {
			out.Errorf(ctx, "Failed to remove %s: %s", oldIDPath, err)
		}
	}
	if fsutil.IsFile(oldIDPath) {
		out.Noticef(ctx, "Found %s. Migrating to %s.", oldIDPath, newIDPath)
		if err := os.Rename(oldIDPath, newIDPath); err != nil {
			out.Errorf(ctx, "Failed to rename %s to %s: %s", oldIDPath, newIDPath, err)
		}
	}

	// create a new instance so we can use decryptFile.
	a, err := New()
	if err != nil {
		return err
	}

	if !ctxutil.HasPasswordCallback(ctx) {
		debug.Log("no password callback found, redirecting to askPass")
		ctx = ctxutil.WithPasswordCallback(ctx, func(prompt string, _ bool) ([]byte, error) {
			pw, err := a.askPass.Passphrase(prompt, fmt.Sprintf("to load the age keyring at %s", OldKeyring), false)
			return []byte(pw), err
		})
	}

	if fsutil.IsFile(OldKeyring) && fsutil.IsFile(a.identity) {
		out.Warningf(ctx, "Both %s and %s exist. Keeping both. Recover any identities from %s as needed.", OldKeyring, a.identity, OldKeyring)
		return nil
	}
	if !fsutil.IsFile(OldKeyring) {
		// nothing to do.
		return nil
	}

	debug.Log("loading old identities from %s", OldKeyring)
	ids, err := a.loadIdentitiesFromKeyring(ctx)
	if err != nil {
		return err
	}

	debug.Log("writing new identities to %s", a.identity)
	if err := a.saveIdentities(ctx, ids, false); err != nil {
		return err
	}
	return os.Remove(OldKeyring)
}

// Keyring is an age keyring.
// Deprecated: Only used for backwards compatibility. Will be removed soon.
type Keyring []Keypair

// Keypair is a public / private keypair.
// Deprecated: Only used for backwards compatibility. Will be removed soon.
type Keypair struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Identity string `json:"identity"`
}

func (a *Age) loadIdentitiesFromKeyring(ctx context.Context) ([]string, error) {
	buf, err := a.decryptFile(ctx, OldKeyring)
	if err != nil {
		debug.Log("can't decrypt keyring at %s: %s", OldKeyring, err)
		return nil, err
	}

	var kr Keyring
	if err := json.Unmarshal(buf, &kr); err != nil {
		debug.Log("can't parse keyring at %s: %s", OldKeyring, err)
		return nil, err
	}

	// remove invalid IDs.
	valid := make([]string, 0, len(kr))
	for _, k := range kr {
		if k.Identity == "" {
			continue
		}
		valid = append(valid, k.Identity)
	}
	debug.Log("loaded keyring with %d valid entries from %s", len(kr), OldKeyring)
	return valid, nil
}
