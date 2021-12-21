package age

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"filippo.io/age"
	"github.com/gopasspw/gopass/pkg/ctxutil"
	"github.com/gopasspw/gopass/pkg/debug"
)

var (
	idRecpCacheKey = "identity"
)

// Identities returns all identities, used for decryption
func (a *Age) Identities(ctx context.Context) ([]age.Identity, error) {
	if !ctxutil.HasPasswordCallback(ctx) {
		debug.Log("no password callback found, redirecting to askPass")
		ctx = ctxutil.WithPasswordCallback(ctx, func(prompt string, confirm bool) ([]byte, error) {
			pw, err := a.askPass.Passphrase(prompt, fmt.Sprintf("to read the age keyring from %s", a.identity), confirm)
			return []byte(pw), err
		})
	}

	debug.Log("reading native identities from %s", a.identity)
	buf, err := a.decryptFile(ctx, a.identity)
	if err != nil {
		debug.Log("failed to decrypt existing identities from %s: %s", a.identity, err)
		if !os.IsNotExist(err) {
			return nil, fmt.Errorf("failed to decrypt %s: %s", a.identity, err)
		}
	}

	ids, err := age.ParseIdentities(bytes.NewReader(buf))
	if err != nil {
		return nil, err
	}
	debug.Log("read %d native identities from %s", len(ids), a.identity)
	return ids, nil
}

// IdentityRecipients returns a slice of recipients dervied from our identities.
// Since the identity file is encrypted we try to use a cached copy of the recipients
// dervied from the identities.
func (a *Age) IdentityRecipients(ctx context.Context) ([]age.Recipient, error) {
	if ids := a.cachedIDRecpipients(); len(ids) > 0 {
		return ids, nil
	}

	ids, err := a.Identities(ctx)
	if err != nil {
		return nil, err
	}

	var r []age.Recipient
	for _, id := range ids {
		if x, ok := id.(*age.X25519Identity); ok {
			r = append(r, x.Recipient())
		}
	}
	if err := a.recpCache.Set(idRecpCacheKey, recipientsToBech32(r)); err != nil {
		debug.Log("failed to cache identity recipients: %s", err)
	}
	return r, nil
}

// GenerateIdentity creates a new identity
func (a *Age) GenerateIdentity(ctx context.Context, _ string, _ string, pw string) error {
	if pw != "" {
		ctx = ctxutil.WithPasswordCallback(ctx, func(prompt string, confirm bool) ([]byte, error) {
			return []byte(pw), nil
		})
	}
	_, err := a.addIdentity(ctx)
	return err
}

// ListIdentities lists all identities
func (a *Age) ListIdentities(ctx context.Context) ([]string, error) {
	debug.Log("checking existing identities")
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

// FindIdentities returns all usable identities (native only)
func (a *Age) FindIdentities(ctx context.Context, keys ...string) ([]string, error) {
	ids, err := a.IdentityRecipients(ctx)
	if err != nil {
		return nil, err
	}
	matches := make([]string, 0, len(ids))
OUTER:
	for _, k := range keys {
		for _, r := range recipientsToBech32(ids) {
			if r == k {
				matches = append(matches, k)
				debug.Log("found matching recipient %s", k)
				continue OUTER
			}
		}
		debug.Log("%s not found in %q", k, ids)
	}
	sort.Strings(matches)
	return matches, nil
}

func (a *Age) cachedIDRecpipients() []age.Recipient {
	if a.recpCache.ModTime(idRecpCacheKey).Before(modTime(a.identity)) {
		debug.Log("identity cache expired")
		a.recpCache.Remove(idRecpCacheKey)
		return nil
	}
	recps, err := a.recpCache.Get(idRecpCacheKey)
	if err != nil {
		debug.Log("failed to get recipients from cache: %s", err)
		return nil
	}
	var rs []age.Recipient
	for _, recp := range recps {
		r, err := age.ParseX25519Recipient(recp)
		if err != nil {
			debug.Log("failed to parse recipient %s: %s", recp, err)
			continue
		}
		rs = append(rs, r)
	}
	return rs
}

func (a *Age) addIdentity(ctx context.Context) ([]age.Identity, error) {
	ids, _ := a.Identities(ctx)
	id, err := age.GenerateX25519Identity()
	if err != nil {
		return nil, err
	}

	ids = append(ids, id)
	if err := a.saveIdentities(ctx, identitiesToString(ids), true); err != nil {
		return nil, err
	}

	return ids, nil
}

func (a *Age) saveIdentities(ctx context.Context, ids []string, newFile bool) error {
	if !ctxutil.HasPasswordCallback(ctx) {
		debug.Log("no password callback found, redirecting to askPass")
		ctx = ctxutil.WithPasswordCallback(ctx, func(prompt string, confirm bool) ([]byte, error) {
			pw, err := a.askPass.Passphrase(prompt, fmt.Sprintf("to save the age keyring to %s", a.identity), confirm)
			return []byte(pw), err
		})
	}

	// ensure directory exists
	if err := os.MkdirAll(filepath.Dir(a.identity), 0700); err != nil {
		debug.Log("failed to create directory for the keyring at %s: %s", a.identity, err)
		return err
	}

	if err := a.encryptFile(ctx, a.identity, []byte(strings.Join(ids, "\n")), newFile); err != nil {
		return err
	}

	debug.Log("saved %d identities to %s", len(ids), a.identity)
	return nil
}

func (a *Age) getAllIdentities(ctx context.Context) (map[string]age.Identity, error) {
	debug.Log("checking native identities")
	native, err := a.getNativeIdentities(ctx)
	if err != nil {
		return nil, err
	}
	debug.Log("got %d native identities", len(native))

	if IsOnlyNative(ctx) {
		debug.Log("returning only native identities")
		return native, nil
	}

	debug.Log("checking ssh identities")
	ssh, err := a.getSSHIdentities(ctx)
	if err != nil {
		return nil, err
	}
	debug.Log("got %d ssh identities", len(ssh))

	// merge both
	for k, v := range ssh {
		native[k] = v
	}
	debug.Log("got %d merged identities", len(native))

	// TODO add passage identities, too
	// $HOME/.passage/identities

	return native, nil
}

func (a *Age) getNativeIdentities(ctx context.Context) (map[string]age.Identity, error) {
	ids, err := a.Identities(ctx)
	if err != nil {
		return nil, err
	}

	return idMap(ids), nil
}

func idMap(ids []age.Identity) map[string]age.Identity {
	m := make(map[string]age.Identity)
	for _, id := range ids {
		if x, ok := id.(*age.X25519Identity); ok {
			m[x.Recipient().String()] = id
			continue
		}
		debug.Log("unknown Identity type: %T", id)
	}
	return m
}

func recipientsToBech32(recps []age.Recipient) []string {
	var r []string
	for _, recp := range recps {
		r = append(r, fmt.Sprintf("%s", recp))
	}
	return r
}

func identitiesToString(ids []age.Identity) []string {
	var r []string
	for _, id := range ids {
		r = append(r, fmt.Sprintf("%s", id))
	}
	return r
}

func modTime(path string) time.Time {
	fi, err := os.Stat(path)
	if err != nil {
		debug.Log("failed to stat %s: %s", path, err)
		return time.Time{}
	}
	return fi.ModTime()
}
