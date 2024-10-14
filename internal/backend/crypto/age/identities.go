package age

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"filippo.io/age"
	"filippo.io/age/agessh"
	"filippo.io/age/plugin"
	"github.com/gopasspw/gopass/pkg/appdir"
	"github.com/gopasspw/gopass/pkg/ctxutil"
	"github.com/gopasspw/gopass/pkg/debug"
)

var idRecpCacheKey = "identity"

// wrappedIdentity is a struct that allows us to wrap an `age.Identity` (typically
// a `plugin.Identity` in order to keep track of its corresponding `age.Recipient`
// and its bech32 encoding, since the `age`  plugin system doesn't provide a way
// to easily derive a plugin `Recipient` from a given `Identity`.
// It is very important to instantiate the recipient when instantiating a
// wrappedIdentity.
type wrappedIdentity struct {
	id       age.Identity
	rec      age.Recipient
	encoding string
}

func (w *wrappedIdentity) Recipient() age.Recipient { return w.rec }
func (w *wrappedIdentity) String() string           { return w.encoding }

// SafeStr is implemented in order to avoid logging potentially sensitive data,
// since an `age.Identity` typically contains secret key material.
func (w *wrappedIdentity) SafeStr() string {
	if len(w.encoding) < 12 {
		return "(elided)"
	} else {
		// we return the first 12 char which are typically "AGE-PLUGIN-x" where
		// x is the first letter of the plugin name
		return w.encoding[:12]
	}
}

// Unwrap simply delegates the unwrapping process to its wrapped identity.
func (w *wrappedIdentity) Unwrap(stanzas []*age.Stanza) ([]byte, error) {
	return w.id.Unwrap(stanzas)
}

// wrappedRecipient is meant to wrap an `age.Recipient`, typically a plugin one,
// in order to keep track of its corresponding bech32 encoding since plugins don't
// support deriving a recipient and its encoding from a given identity.
type wrappedRecipient struct {
	rec      age.Recipient
	encoding string
}

func (w *wrappedRecipient) String() string { return w.encoding }

// Wrap simply delegates the wrapping process to its wrapped recipient.
func (w *wrappedRecipient) Wrap(fileKey []byte) ([]*age.Stanza, error) {
	return w.rec.Wrap(fileKey)
}

// Identities returns all identities, used for decryption.
func (a *Age) Identities(ctx context.Context) ([]age.Identity, error) {
	if !ctxutil.HasPasswordCallback(ctx) {
		debug.Log("no password callback found, redirecting to askPass")
		ctx = ctxutil.WithPasswordCallback(ctx, func(prompt string, confirm bool) ([]byte, error) {
			pw, err := a.askPass.Passphrase(prompt, fmt.Sprintf("to read the age keyring from %s", a.identity), confirm)

			return []byte(pw), err
		})
		ctx = ctxutil.WithPasswordPurgeCallback(ctx, a.askPass.Remove)
	}

	debug.Log("reading native identities from %s", a.identity)
	buf, err := a.decryptFile(ctx, a.identity)
	if err != nil {
		debug.Log("failed to decrypt existing identities from %s: %s", a.identity, err)
		if !errors.Is(err, os.ErrNotExist) {
			return nil, fmt.Errorf("failed to decrypt %s: %w", a.identity, err)
		}

		return nil, nil
	}

	ids, err := parseIdentities(bytes.NewReader(buf))
	if err != nil {
		return nil, err
	}

	debug.Log("read %d native identities from %s", len(ids), a.identity)

	return ids, nil
}

// parseIdentity is mostly like `age` parseIdentity, except that it implements
// our custom format we use with wrapped identities to store the encoding of
// both the plugin identity and its corresponding recipient.
// Custom format: `<age identity>"|"<age recipient>`
// This custom format allows us to keep track of a given identity's recipient
// and prevents us from storing secret identity data in our recipient cache.
func parseIdentity(s string) (age.Identity, error) {
	switch {
	case strings.HasPrefix(s, "AGE-PLUGIN-"):
		// sp will have a length at least 1 and will contain either the full string
		// or the first part before | and the second part will be in sp[1].
		sp := strings.Split(s, "|")
		id, err := plugin.NewIdentity(sp[0], pluginTerminalUI)
		if err != nil {
			return nil, fmt.Errorf("unable to parse plugin identity: %w", err)
		}
		var rec age.Recipient
		if len(sp) == 2 {
			rec = &wrappedRecipient{
				rec:      id.Recipient(),
				encoding: sp[1],
			}
		} else {
			rec = id.Recipient()
		}

		return &wrappedIdentity{
			id:       id,
			encoding: s,
			rec:      rec,
		}, nil
	case strings.HasPrefix(s, "AGE-SECRET-KEY-1"):
		sp := strings.Split(s, "|")

		return age.ParseX25519Identity(sp[0])
	default:
		return nil, fmt.Errorf("unknown identity type")
	}
}

// parseIdentities is like age.ParseIdentities, but supports plugin identities,
// it is a copy of https://github.com/FiloSottile/age/blob/2214a556f60400ad19f2ca43d3cbbb4a5a0fe5ab/cmd/age/parse.go#L123-L126
func parseIdentities(f io.Reader) ([]age.Identity, error) {
	const privateKeySizeLimit = 1 << 24 // 16 MiB
	var ids []age.Identity
	scanner := bufio.NewScanner(io.LimitReader(f, privateKeySizeLimit))
	var n int
	for scanner.Scan() {
		n++
		line := scanner.Text()
		if strings.HasPrefix(line, "#") || line == "" {
			continue
		}

		i, err := parseIdentity(line)
		if err != nil {
			return nil, fmt.Errorf("error at line %d: %w", n, err)
		}
		ids = append(ids, i)
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed to read secret keys file: %w", err)
	}
	if len(ids) == 0 {
		return nil, fmt.Errorf("no secret keys found")
	}

	return ids, nil
}

// IdentityRecipients returns a slice of recipients derived from our identities.
// Since the identity file is encrypted we try to use a cached copy of the recipients
// derived from the identities.
func (a *Age) IdentityRecipients(ctx context.Context) ([]age.Recipient, error) {
	if ids := a.cachedIDRecipients(); len(ids) > 0 {
		debug.Log("successfully retrieved identities from cache")

		return ids, nil
	}

	ids, err := a.Identities(ctx)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}

		return nil, err
	}

	var r []age.Recipient
	for _, id := range ids {
		if rec := IdentityToRecipient(id); rec != nil {
			r = append(r, rec)
		}
	}
	debug.Log("got %d recipients from %d age identities", len(r), len(ids))

	if err := a.recpCache.Set(idRecpCacheKey, recipientsToString(r)); err != nil {
		debug.Log("failed to cache identity recipients: %s", err)
	}

	return r, nil
}

func IdentityToRecipient(id age.Identity) age.Recipient {
	switch id := id.(type) {
	case *age.X25519Identity:
		debug.Log("parsed age identity as X25519Identity")

		return id.Recipient()
	case *wrappedIdentity:
		debug.Log("parsed age identity as wrappedIdentity")

		return id.Recipient()
	case *plugin.Identity:
		debug.Log("parsed age identity as plugin.Identity")

		return id.Recipient()
	case *agessh.RSAIdentity:
		debug.Log("parsed age identity as RSAIdentity")

		return id.Recipient()
	case *agessh.Ed25519Identity:
		debug.Log("parsed age identity as Ed25519Identity")

		return id.Recipient()
	case *agessh.EncryptedSSHIdentity:
		debug.Log("parsed age identity as encrypted SSHIdentity")

		return id.Recipient()
	default:
		debug.Log("unexpected age identity type: %T", id)

		return nil
	}
}

// GenerateIdentity creates a new identity.
func (a *Age) GenerateIdentity(ctx context.Context, _ string, _ string, pw string) error {
	// we don't check if the password callback is set, since it could only be
	// set through an env variable, and here pw can only be set through an
	// actual user input.
	if pw != "" {
		debug.Log("age GenerateIdentity using provided pw")
		ctx = ctxutil.WithPasswordCallback(ctx, func(prompt string, confirm bool) ([]byte, error) {
			return []byte(pw), nil
		})
	}

	id, err := age.GenerateX25519Identity()
	if err != nil {
		return err
	}

	return a.addIdentity(ctx, id)
}

// ListIdentities lists all identities.
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

// FindIdentities returns all usable identities (native only).
func (a *Age) FindIdentities(ctx context.Context, keys ...string) ([]string, error) {
	ids, err := a.IdentityRecipients(ctx)
	if err != nil {
		return nil, err
	}
	matches := make([]string, 0, len(ids))
OUTER:
	for _, k := range keys {
		for _, r := range recipientsToString(ids) {
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

func (a *Age) cachedIDRecipients() []age.Recipient {
	if a.recpCache.ModTime(idRecpCacheKey).Before(modTime(a.identity)) {
		debug.Log("identity cache expired")
		if err := a.recpCache.Remove(idRecpCacheKey); err != nil {
			debug.Log("error invalidating age id recipient cache: %s", err)
		}

		return nil
	}

	recps, err := a.recpCache.Get(idRecpCacheKey)
	if err != nil {
		debug.Log("failed to get recipients from cache: %s", err)

		return nil
	}

	rs, err := a.parseRecipients(context.Background(), recps)
	if err != nil {
		debug.Log("cachedIDRecipients failed to parse some age recipients: %s", err)
	}

	return rs
}

func (a *Age) addIdentity(ctx context.Context, id age.Identity) error {
	// we invalidate our recipient id cache when we add a new identity
	if err := a.recpCache.Remove(idRecpCacheKey); err != nil {
		debug.Log("error invalidating age id recipient cache: %s", err)
	}

	ids, _ := a.Identities(ctx)

	ids = append(ids, id)

	return a.saveIdentities(ctx, identitiesToString(ids), true)
}

func (a *Age) saveIdentities(ctx context.Context, ids []string, newFile bool) error {
	// only force a password prompt if running interactively
	// TODO: this doesn't really cut it. the purpose is to avoid a password prompt
	// from popping up during tests. but no combination of existing flags really
	// does convey that correctly. I think we need to cleanup and document the
	// different flags conveyed by ctxutil.
	//
	// Note: if running in a test, we don't want to prompt for a password and just fail.
	// Not perfect but we don't support password-less age, yet.
	// TODO(#2108): remove this hack
	if !ctxutil.HasPasswordCallback(ctx) && !ctxutil.IsAlwaysYes(ctx) {
		debug.Log("no password callback found, redirecting to askPass")
		ctx = ctxutil.WithPasswordCallback(ctx, func(prompt string, confirm bool) ([]byte, error) {
			pw, err := a.askPass.Passphrase(prompt, fmt.Sprintf("to save the age keyring to %s", a.identity), confirm)

			return []byte(pw), err
		})
		ctx = ctxutil.WithPasswordPurgeCallback(ctx, a.askPass.Remove)
	}

	// ensure directory exists.
	if err := os.MkdirAll(filepath.Dir(a.identity), 0o700); err != nil {
		debug.Log("failed to create directory for the keyring at %s: %s", a.identity, err)

		return fmt.Errorf("failed to create directory for %s: %w", a.identity, err)
	}

	if err := a.encryptFile(ctx, a.identity, []byte(strings.Join(ids, "\n")), newFile); err != nil {
		return fmt.Errorf("failed to write encrypted identity to %s: %w", a.identity, err)
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
		if errors.Is(err, ErrNoSSHDir) {
			return native, nil
		}

		return nil, err
	}

	debug.Log("got %d ssh identities", len(ssh))

	// merge both.
	for k, v := range ssh {
		native[k] = v
	}
	debug.Log("got %d merged identities", len(native))

	ps, err := a.getPassageIdentities(ctx)
	if err != nil {
		debug.Log("unable to load passage identities: %s", err)
	}

	// merge
	for k, v := range ps {
		native[k] = v
	}

	return native, nil
}

func (a *Age) getPassageIdentities(ctx context.Context) (map[string]age.Identity, error) {
	fn := PassageIdFile()
	fh, err := os.Open(fn)
	if err != nil {
		return nil, fmt.Errorf("failed to open %s: %w", fn, err)
	}
	defer func() { _ = fh.Close() }()

	ids, err := age.ParseIdentities(fh)
	if err != nil {
		return nil, err
	}

	// TODO(gh/2059) support encrypted passage identities

	return idMap(ids), nil
}

// PassageIdFile returns the location of the passage identities file.
func PassageIdFile() string {
	return filepath.Join(appdir.UserHome(), ".passage", "identities")
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
		switch i := id.(type) {
		case *age.X25519Identity:
			m[i.Recipient().String()] = id

			continue
		case *wrappedIdentity:
			m[i.String()] = id

		default:
			debug.Log("unknown Identity type: %T", id)
		}
	}

	return m
}

func recipientsToString(recps []age.Recipient) []string {
	r := make([]string, 0, len(recps))
	for _, recp := range recps {
		r = append(r, fmt.Sprintf("%s", recp))
	}

	return r
}

func identitiesToString(ids []age.Identity) []string {
	r := make([]string, 0, len(ids))
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
