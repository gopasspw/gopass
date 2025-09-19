package agecrypto

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/gopasspw/gopass/internal/backend"
	"github.com/gopasspw/gopass/internal/config"
	"github.com/gopasspw/gopass/internal/out"
	"github.com/gopasspw/gopass/pkg/appdir"
	"github.com/gopasspw/gopass/pkg/ctxutil"
	"github.com/gopasspw/gopass/pkg/debug"
	"github.com/gopasspw/gopass/pkg/fsutil"
)

// OldIDFile is the old file name for the recipients.
var OldIDFile = ".age-ids"

// OldKeyringPath is the old file name for the keyring.
// Must be a func to allow us to honor GOPASS_HOMEDIR in tests.
// Otherwise it would be read at init time and setting GOPASS_HOMEDIR
// later would have no effect.
func OldKeyringPath() string {
	return filepath.Join(appdir.UserConfig(), "age-keyring.age")
}

func migrate(ctx context.Context, s backend.Storage) error {
	out.Noticef(ctx, "Attempting to migrate age backend. You will need to unlock your identities keyring.")

	if s.Exists(ctx, OldIDFile) && s.Exists(ctx, IDFile) {
		out.Warningf(ctx, "Both %s and %s exist. Removing the old one (%s).", OldIDFile, IDFile, OldIDFile)
		if err := s.Delete(ctx, OldIDFile); err != nil {
			out.Errorf(ctx, "Failed to remove %s: %s", OldIDFile, err)
		} else {
			out.OKf(ctx, "Removed the old IDFile at %s", OldIDFile)
		}
	}

	if s.Exists(ctx, OldIDFile) {
		out.Noticef(ctx, "Found %s. Migrating to %s.", OldIDFile, IDFile)
		buf, err := s.Get(ctx, OldIDFile)
		if err != nil {
			return err
		}
		if err := s.Set(ctx, IDFile, buf); err != nil {
			out.Errorf(ctx, "Failed to rename %s to %s: %s", OldIDFile, IDFile, err)
		}

		debug.Log("Renamed the old IDFile at %s to %s", OldIDFile, IDFile)
	} else {
		debug.Log("Old IDFile %s does not exist, nothing to do", OldIDFile)
	}

	// create a new instance so we can use decryptFile.
	a, err := New(ctx, config.String(ctx, "age.ssh-key-path"))
	if err != nil {
		return err
	}

	oldKeyring := OldKeyringPath()
	if !ctxutil.HasPasswordCallback(ctx) {
		debug.Log("no password callback found, redirecting to askPass")
		ctx = ctxutil.WithPasswordCallback(ctx, func(prompt string, _ bool) ([]byte, error) {
			pw, err := a.askPass.Passphrase(prompt, fmt.Sprintf("to load the age keyring at %s", oldKeyring), false)

			return []byte(pw), err
		})
		ctx = ctxutil.WithPasswordPurgeCallback(ctx, a.askPass.Remove)
	}

	if fsutil.IsFile(oldKeyring) && fsutil.IsFile(a.identity) {
		out.Warningf(ctx, "Both %s and %s exist. Keeping both. Recover any identities from %s as needed.", oldKeyring, a.identity, oldKeyring)

		return nil
	}
	if !fsutil.IsFile(oldKeyring) {
		debug.Log("old keyring %s does not exist, nothing to do", oldKeyring)

		// nothing to do.
		return nil
	}

	debug.Log("loading old identities from %s", oldKeyring)
	ids, err := a.loadIdentitiesFromKeyring(ctx)
	if err != nil {
		return err
	}

	debug.Log("writing new identities to %s", a.identity)
	if err := a.saveIdentities(ctx, ids, false); err != nil {
		return err
	}

	return os.Remove(oldKeyring)
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
	oldKeyring := OldKeyringPath()
	buf, err := a.decryptFile(ctx, oldKeyring)
	if err != nil {
		debug.Log("can't decrypt keyring at %s: %s", oldKeyring, err)

		return nil, fmt.Errorf("can not decrypt old keyring at %s: %w", oldKeyring, err)
	}

	var kr Keyring
	if err := json.Unmarshal(buf, &kr); err != nil {
		debug.Log("can't parse keyring at %s: %s", oldKeyring, err)

		return nil, fmt.Errorf("can not parse old keyring at %s: %w", oldKeyring, err)
	}

	// remove invalid IDs.
	valid := make([]string, 0, len(kr))
	for _, k := range kr {
		if k.Identity == "" {
			continue
		}
		valid = append(valid, k.Identity)
	}
	debug.Log("loaded keyring with %d valid entries from %s", len(kr), oldKeyring)

	return valid, nil
}
