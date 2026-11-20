package passkey

import (
	"bytes"
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"path"
	"slices"
	"strings"
	"unicode"

	"github.com/gopasspw/gopass/internal/config"
	"github.com/gopasspw/gopass/pkg/ctxutil"
	"github.com/gopasspw/gopass/pkg/debug"
	"github.com/gopasspw/gopass/pkg/gopass"
	"github.com/gopasspw/gopass/pkg/gopass/secrets"
)

const (
	// SecretKey is the key of the lines of a secret that hold a credential.
	// A secret can have any number of such lines, each line is one
	// self-contained credential encoded as JSON.
	SecretKey = "passkey"
	// Folder is the name of the folder credentials are looked up in. The
	// secrets directly below "passkeys/<rp id>/" hold the credentials of the
	// Relying Party with that ID.
	Folder = "passkeys"
)

// SecretStore is the part of a password store that is needed to persist
// credentials. It is satisfied by gopass.Store.
type SecretStore interface {
	// List returns the names of all secrets.
	List(ctx context.Context) ([]string, error)
	// Get returns the decrypted secret with the given name.
	Get(ctx context.Context, name, revision string) (gopass.Secret, error)
	// Set creates or updates the secret with the given name.
	Set(ctx context.Context, name string, sec gopass.Byter) error
}

// Entry is a credential and the name of the secret that holds it.
type Entry struct {
	// Name is the name of the secret.
	Name string
	// Credential is the credential.
	Credential *Credential
	// Warning is set by a ceremony that wrote the credential if the store
	// reported a failure after the credential was written, see
	// [ErrNotSynced]. The ceremony succeeded nevertheless.
	Warning error
}

// Store persists credentials in a password store.
//
// A credential is a line of the form "passkey: <json>" in a secret named
// "passkeys/<rp id>/<account>". The name of a secret is not encrypted, so the
// secrets that may hold a credential for a Relying Party can be found without
// decrypting anything. The credentials themselves, including the RP ID they
// are bound to, are only ever taken from the encrypted content.
//
// The folder may also live deeper in the tree, e.g. in a mounted store shared
// with a team: "team/passkeys/<rp id>/<account>" is found as well. New
// credentials are only ever written below the configured prefix. They are
// encrypted for whoever has access to that part of the password store.
//
// The password store offers no transactions. A Store reads a secret right
// before it changes it, but two processes changing the same secret at the
// very same time can still overwrite each other.
type Store struct {
	store  SecretStore
	prefix string
}

// NewStore returns a Store that keeps credentials in s. New credentials are
// written below the folder configured as "passkey.prefix", or below
// "passkeys" if that is not set.
//
// The config keys of this package are only read from the configs the user
// controls, never from the config of a store: that is versioned with the
// store, and whoever can push to the store could otherwise redirect new
// credentials to another folder or claim user verification.
func NewStore(ctx context.Context, s SecretStore) *Store {
	return &Store{
		store:  s,
		prefix: prefixFromConfig(ctx),
	}
}

func prefixFromConfig(ctx context.Context) string {
	prefix := path.Clean("/" + config.UserString(ctx, "passkey.prefix"))[1:]
	if prefix == "" {
		return Folder
	}

	return prefix
}

// Prefix returns the folder new credentials are written to.
func (s *Store) Prefix() string {
	return s.prefix
}

// SignCounterEnabled reports whether new credentials should maintain a
// signature counter according to the config key "passkey.signcount". See
// [NewStore] for where the key is read from.
func SignCounterEnabled(ctx context.Context) bool {
	return config.UserBool(ctx, "passkey.signcount")
}

// get reads the secret with the given name as it is stored.
func (s *Store) get(ctx context.Context, name string) (gopass.Secret, error) {
	// A password that is a reference to another secret must not be resolved:
	// a secret that is written back would lose the reference.
	return s.store.Get(ctxutil.WithFollowRef(ctx, false), name, "latest") //nolint:wrapcheck
}

// UserVerificationEnabled reports whether a ceremony the user confirmed is
// reported as user verified, according to the config key
// "passkey.user-verification". See [ConsentByConfirmation], and [NewStore]
// for where the key is read from.
func UserVerificationEnabled(ctx context.Context) bool {
	return config.UserBool(ctx, "passkey.user-verification")
}

// rpIDOf returns the RP ID encoded in the name of a secret and whether the
// name denotes a secret that may hold credentials at all. Names are compared
// without regard to case, because not all file systems tell them apart.
func (s *Store) rpIDOf(name string) (string, bool) {
	// <parent>/<rp id>/<account>
	dir, account, found := cutLast(name)
	if !found || account == "" {
		return "", false
	}

	parent, rpID, found := cutLast(dir)
	if !found || rpID == "" {
		return "", false
	}

	_, folder, _ := cutLast(parent)
	if strings.EqualFold(parent, s.prefix) || strings.EqualFold(folder, Folder) {
		return strings.ToLower(rpID), true
	}

	return "", false
}

// Owns reports whether a secret is located directly below the prefix, i.e.
// where new credentials are written to, and not in a folder that was only
// found because of its name.
func (s *Store) Owns(name string) bool {
	dir, _, _ := cutLast(name)
	parent, _, found := cutLast(dir)

	return found && strings.EqualFold(parent, s.prefix)
}

// cutLast splits a secret name at its last separator.
func cutLast(name string) (string, string, bool) {
	i := strings.LastIndex(name, "/")
	if i < 0 {
		return "", name, false
	}

	return name[:i], name[i+1:], true
}

// Names returns the names of the secrets that may hold credentials for the
// Relying Party. All such secrets are returned if rpID is empty. Nothing is
// decrypted. Secrets below the prefix come first.
func (s *Store) Names(ctx context.Context, rpID string) ([]string, error) {
	all, err := s.store.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list secrets: %w", err)
	}

	names := make([]string, 0, 8)
	for _, name := range all {
		id, ok := s.rpIDOf(name)
		if !ok {
			continue
		}

		if rpID != "" && id != rpID {
			continue
		}

		names = append(names, name)
	}

	slices.SortFunc(names, func(a, b string) int {
		if oa, ob := s.Owns(a), s.Owns(b); oa != ob {
			if oa {
				return -1
			}

			return 1
		}

		return strings.Compare(a, b)
	})

	return names, nil
}

// List returns the credentials of a Relying Party, or all credentials if
// rpID is empty. The secrets holding them are decrypted.
//
// What can not be read or parsed does not make the call fail: it is skipped
// and reported in the returned error, together with the credentials that
// could be loaded.
func (s *Store) List(ctx context.Context, rpID string) ([]Entry, error) {
	names, err := s.Names(ctx, rpID)
	if err != nil {
		return nil, err
	}

	return s.load(ctx, names)
}

// load decrypts the secrets with the given names, which must be names
// returned by [Store.Names], and returns their credentials. See [Store.List].
func (s *Store) load(ctx context.Context, names []string) ([]Entry, error) {
	var errs []error

	entries := make([]Entry, 0, len(names))
	for _, name := range names {
		sec, err := s.get(ctx, name)
		if err != nil {
			errs = append(errs, fmt.Errorf("failed to read %s: %w", name, err))

			continue
		}

		creds, err := FromSecret(sec)
		if err != nil {
			errs = append(errs, fmt.Errorf("failed to parse %s: %w", name, err))
		}

		nameID, _ := s.rpIDOf(name)
		for _, cred := range creds {
			// The name of a secret is only an index. What a credential is
			// bound to is decided by its encrypted content alone.
			if cred.RPID != nameID {
				errs = append(errs, fmt.Errorf("%w: %s holds a passkey for %q", ErrInvalidCredential, name, cred.RPID))

				continue
			}

			entries = append(entries, Entry{Name: name, Credential: cred})
		}
	}

	return entries, errors.Join(errs...)
}

// Load returns the credentials stored in the secret with the given name,
// wherever it is located. Only this secret is decrypted. A secret that holds
// no credentials results in neither entries nor an error.
//
// Unlike [Store.List], Load does not compare the RP ID of a credential with
// the name of the secret. The caller chose the secret, and a credential is
// bound to the RP ID it holds. Lines that can not be parsed are reported in
// the returned error, wrapping [ErrInvalidCredential], together with the
// credentials of the lines that could.
func (s *Store) Load(ctx context.Context, name string) ([]Entry, error) {
	sec, err := s.get(ctx, name)
	if err != nil {
		return nil, fmt.Errorf("failed to read %s: %w", name, err)
	}

	creds, err := FromSecret(sec)

	entries := make([]Entry, 0, len(creds))
	for _, cred := range creds {
		entries = append(entries, Entry{Name: name, Credential: cred})
	}

	if err != nil {
		return entries, fmt.Errorf("failed to parse %s: %w", name, err)
	}

	return entries, nil
}

// FromSecret returns the credentials stored in a secret. Lines that can not
// be parsed are reported in the returned error, together with the
// credentials of the lines that could.
func FromSecret(sec gopass.Secret) ([]*Credential, error) {
	values, _ := sec.Values(SecretKey)

	var errs []error

	creds := make([]*Credential, 0, len(values))
	for _, v := range values {
		// A line without a value is not skipped: it is not understood, and
		// what is not understood must not be dropped when the secret is
		// written.
		cred, err := Decode([]byte(strings.TrimSpace(v)))
		if err != nil {
			errs = append(errs, err)

			continue
		}

		creds = append(creds, cred)
	}

	return creds, errors.Join(errs...)
}

// ToSecret replaces the credentials stored in a secret. Everything else in
// the secret is left alone.
func ToSecret(sec gopass.Secret, creds []*Credential) error {
	lines := make([]string, 0, len(creds))
	for _, cred := range creds {
		buf, err := cred.Encode()
		if err != nil {
			return err
		}

		lines = append(lines, string(buf))
	}

	sec.Del(SecretKey)

	if len(lines) == 0 {
		return nil
	}

	// Lines are appended. A secret that was not parsed but assembled may
	// lack the line break that keeps the first one apart from what is there.
	if buf := sec.Bytes(); len(buf) > 0 && buf[len(buf)-1] != '\n' {
		w, ok := sec.(io.Writer)
		if !ok {
			return fmt.Errorf("secret does not end with a line break")
		}

		if _, err := w.Write([]byte("\n")); err != nil {
			return fmt.Errorf("failed to add passkey to secret: %w", err)
		}
	}

	for _, line := range lines {
		if err := sec.Add(SecretKey, line); err != nil {
			return fmt.Errorf("failed to add passkey to secret: %w", err)
		}
	}

	return nil
}

// Put stores a credential and returns the name of the secret it was written
// to.
//
// All credentials of a user account are kept in one secret. A credential is
// added to the secret below the prefix that already holds credentials of its
// account, or to a new one named after the user. Credentials that are already
// there are never removed, with one exception: a credential with the same ID
// is replaced, so that storing a credential twice does not duplicate it.
//
// An error wrapping [ErrNotSynced] is returned together with the name if the
// credential was written but the store reported a failure afterwards.
func (s *Store) Put(ctx context.Context, cred *Credential) (string, error) {
	if err := cred.validate(); err != nil {
		return "", err
	}

	target, _, err := s.AccountSecret(ctx, cred.RPID, cred.UserHandle)
	if err != nil {
		return "", err
	}

	return s.putTo(ctx, cred, target)
}

// putTo stores a credential in the secret with the given name, which must be
// the account secret of the credential, see [Store.AccountSecret]. If target
// is empty the account has none yet and a new secret is created.
func (s *Store) putTo(ctx context.Context, cred *Credential, target string) (string, error) {
	if err := cred.validate(); err != nil {
		return "", err
	}

	if target == "" {
		target = s.prefix + "/" + cred.RPID + "/" + accountName(cred)
	}

	var stored bool

	return s.modify(ctx, target, func(creds []*Credential) ([]*Credential, error) {
		// A credential takes the place of the one it replaces, so that
		// storing it again leaves the secret as it is.
		for i, c := range creds {
			if bytes.Equal(c.ID, cred.ID) {
				creds[i] = cred
				stored = true
			}
		}

		if stored {
			return creds, nil
		}

		return append(creds, cred), nil
	})
}

// AccountSecret returns the name of the secret below the prefix that holds
// the credentials of a user account, if there is one. It is the secret a new
// credential for the account is written to, see [Store.Put].
func (s *Store) AccountSecret(ctx context.Context, rpID string, userHandle []byte) (string, bool, error) {
	names, err := s.Names(ctx, rpID)
	if err != nil {
		return "", false, err
	}

	for _, name := range names {
		// Only secrets below the prefix are candidates for an update.
		// Secrets in other folders may be shared with other people, who must
		// not end up with a credential that was created for this user alone.
		if !s.Owns(name) {
			continue
		}

		sec, err := s.get(ctx, name)
		if err != nil {
			debug.Log("skipping unreadable passkey secret %s: %s", name, err)

			continue
		}

		creds, err := FromSecret(sec)
		if err != nil {
			debug.Log("skipping malformed passkeys in %s: %s", name, err)
		}

		if slices.ContainsFunc(creds, func(c *Credential) bool {
			return c.RPID == rpID && bytes.Equal(c.UserHandle, userHandle)
		}) {
			return name, true, nil
		}
	}

	return "", false, nil
}

// accountIn returns the account secret of a user account among entries
// loaded by [Store.List]. It is what [Store.AccountSecret] returns for the
// same state of the store.
func (s *Store) accountIn(entries []Entry, rpID string, userHandle []byte) (string, bool) {
	for _, e := range entries {
		if s.Owns(e.Name) && e.Credential.RPID == rpID && bytes.Equal(e.Credential.UserHandle, userHandle) {
			return e.Name, true
		}
	}

	return "", false
}

// FindUser returns a credential of the user account with the given name at
// a Relying Party, if one is stored in the secret below the prefix that is
// named after the user. That is the secret the credentials of a new account
// are written to, see [Store.Put]. Only this secret is decrypted.
func (s *Store) FindUser(ctx context.Context, rpID, userName string) (*Credential, bool) {
	account := sanitizeName(userName)
	if account == "" {
		return nil, false
	}

	names, err := s.Names(ctx, rpID)
	if err != nil {
		debug.Log("failed to list the passkeys of %s: %s", rpID, err)

		return nil, false
	}

	want := s.prefix + "/" + rpID + "/" + account

	i := slices.IndexFunc(names, func(n string) bool { return s.Owns(n) && strings.EqualFold(n, want) })
	if i < 0 {
		return nil, false
	}

	entries, err := s.load(ctx, names[i:i+1])
	if err != nil {
		debug.Log("failed to read some passkeys in %s: %s", names[i], err)
	}

	for _, e := range entries {
		if e.Credential.UserName == userName {
			return e.Credential, true
		}
	}

	return nil, false
}

// Advance increments the signature counter of a credential in the secret
// with the given name, and returns the credential as it is stored now. The
// counter is read from the store right before it is written back, so that an
// assertion is never signed with a counter that was not persisted.
//
// The credential in the store must still be the given one: the same key for
// the same account. Otherwise nothing is written.
//
// An error wrapping [ErrNotSynced] is returned together with the credential
// if the counter was written but the store reported a failure afterwards.
func (s *Store) Advance(ctx context.Context, name string, cred *Credential) (*Credential, error) {
	var stored *Credential

	_, err := s.modify(ctx, name, func(creds []*Credential) ([]*Credential, error) {
		i := slices.IndexFunc(creds, func(c *Credential) bool { return bytes.Equal(c.ID, cred.ID) })
		if i < 0 {
			return nil, fmt.Errorf("%w in %s", ErrNoCredentials, name)
		}

		if !sameCredential(creds[i], cred) {
			return nil, fmt.Errorf("%w in %s: the passkey was replaced", ErrNoCredentials, name)
		}

		if err := creds[i].Advance(); err != nil {
			return nil, err
		}
		stored = creds[i]

		return creds, nil
	})
	if err != nil && !errors.Is(err, ErrNotSynced) {
		return nil, err
	}

	return stored, err
}

// sameCredential reports whether two credentials are the same key of the
// same user account.
func sameCredential(a, b *Credential) bool {
	if a.RPID != b.RPID || !bytes.Equal(a.ID, b.ID) || !bytes.Equal(a.UserHandle, b.UserHandle) {
		return false
	}

	ka, errA := a.PublicKey()
	kb, errB := b.PublicKey()

	return errA == nil && errB == nil && bytes.Equal(ka, kb)
}

// Remove deletes the credential with the given ID from the secret with the
// given name and returns the number of credentials left in the secret. The
// secret itself is kept, even if no credential is left.
//
// An error wrapping [ErrNotSynced] is returned together with the number if
// the credential was removed but the store reported a failure afterwards.
func (s *Store) Remove(ctx context.Context, name string, id []byte) (int, error) {
	left := 0

	_, err := s.modify(ctx, name, func(creds []*Credential) ([]*Credential, error) {
		n := len(creds)
		creds = slices.DeleteFunc(creds, func(c *Credential) bool { return bytes.Equal(c.ID, id) })

		if len(creds) == n {
			return nil, fmt.Errorf("%w in %s", ErrNoCredentials, name)
		}
		left = len(creds)

		return creds, nil
	})
	if err != nil && !errors.Is(err, ErrNotSynced) {
		return 0, err
	}

	return left, err
}

// modify applies fn to the credentials of the secret with the given name and
// writes the result back. Anything else the secret contains is kept. The
// secret is created if it does not exist. The name the secret was written to
// is returned: it can differ in case from the requested name. It is also
// returned with an error wrapping ErrNotSynced.
func (s *Store) modify(ctx context.Context, name string, fn func([]*Credential) ([]*Credential, error)) (string, error) {
	all, err := s.store.List(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to list secrets: %w", err)
	}

	// Whether the secret exists is decided by the list of secrets, not by
	// trying to read it: a secret that exists but can not be decrypted must
	// never be mistaken for a new one and overwritten. A secret whose name
	// only differs in case is the same file on some file systems.
	sec := secrets.New()
	creds := []*Credential{}

	i := slices.Index(all, name)
	if i < 0 {
		i = slices.IndexFunc(all, func(n string) bool { return strings.EqualFold(n, name) })
	}

	exists := i >= 0
	if exists {
		name = all[i]

		sec, err = s.get(ctx, name)
		if err != nil {
			return "", fmt.Errorf("failed to read %s: %w", name, err)
		}

		// Writing drops lines that could not be parsed. Refuse to change a
		// secret that is not fully understood.
		creds, err = FromSecret(sec)
		if err != nil {
			return "", fmt.Errorf("failed to parse %s: %w", name, err)
		}
	}

	creds, err = fn(creds)
	if err != nil {
		return "", err
	}

	before := bytes.Clone(sec.Bytes())

	if err := ToSecret(sec, creds); err != nil {
		return "", err
	}

	// Stores refuse to write a secret that did not change. Not having to
	// write is a success, e.g. when a credential is imported twice.
	if exists && bytes.Equal(before, sec.Bytes()) {
		return name, nil
	}

	if err := ctx.Err(); err != nil {
		return "", fmt.Errorf("not writing %s: %w", name, err)
	}

	if err := s.store.Set(ctx, name, sec); err != nil {
		// A store also reports what goes wrong after a secret was written,
		// like a push to a remote that is not reachable. That must not be
		// mistaken for a secret that was not stored: a passkey would be
		// thrown away, or created over and over again.
		if !s.holds(ctx, name, sec) {
			return "", fmt.Errorf("failed to write %s: %w", name, err)
		}

		return name, fmt.Errorf("%w: %s: %w", ErrNotSynced, name, err)
	}

	return name, nil
}

// holds reports whether the secret with the given name has the content of
// sec in the store.
func (s *Store) holds(ctx context.Context, name string, sec gopass.Byter) bool {
	// The context may be the reason the store gave up.
	stored, err := s.get(context.WithoutCancel(ctx), name)
	if err != nil {
		return false
	}

	return bytes.Equal(stored.Bytes(), sec.Bytes())
}

// accountName derives the last element of the name of a secret from the user
// account of a credential.
func accountName(cred *Credential) string {
	for _, candidate := range []string{cred.UserName, cred.UserDisplayName} {
		if name := sanitizeName(candidate); name != "" {
			return name
		}
	}

	// Credential IDs can be too long for a file name.
	if len(cred.ID) > credentialIDLen {
		sum := sha256.Sum256(cred.ID)

		return Base64URL(sum[:]).String()
	}

	return Base64URL(cred.ID).String()
}

// maxNameLen is the maximum length in bytes of a name derived from a user.
const maxNameLen = 100

// sanitizeName turns an arbitrary string into something that is safe to use
// as a single element of the name of a secret on every platform. An empty
// string is returned if nothing usable is left.
func sanitizeName(s string) string {
	s = strings.Map(func(r rune) rune {
		switch {
		case r == '/' || r == '\\':
			return '_'
		case strings.ContainsRune(`<>:"|?*`, r):
			// reserved in file names on Windows
			return '_'
		case unicode.IsControl(r) || !unicode.IsPrint(r):
			return -1
		case unicode.IsSpace(r):
			return ' '
		default:
			return r
		}
	}, s)

	if len(s) > maxNameLen {
		s = strings.ToValidUTF8(s[:maxNameLen], "")
	}

	// Leading dots would hide the secret or, as "..", leave the folder.
	// Windows does not allow trailing dots and spaces.
	s = strings.TrimLeft(s, ". ")
	s = strings.TrimRight(s, ". ")

	// Windows reserves the names of devices, with any extension.
	device, _, _ := strings.Cut(strings.ToUpper(s), ".")
	switch strings.TrimRight(device, " ") {
	case "CON", "PRN", "AUX", "NUL",
		"COM1", "COM2", "COM3", "COM4", "COM5", "COM6", "COM7", "COM8", "COM9",
		"LPT1", "LPT2", "LPT3", "LPT4", "LPT5", "LPT6", "LPT7", "LPT8", "LPT9":
		s = "_" + s
	}

	return s
}
