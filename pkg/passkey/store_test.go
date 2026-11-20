package passkey

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/gopasspw/gopass/internal/config"
	"github.com/gopasspw/gopass/pkg/ctxutil"
	"github.com/gopasspw/gopass/pkg/gopass"
	"github.com/gopasspw/gopass/pkg/gopass/apimock"
	"github.com/gopasspw/gopass/pkg/gopass/secrets"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newCred returns a credential for a user account.
func newCred(tb testing.TB, rpID, user, handle string) *Credential {
	tb.Helper()

	cred, err := New(NewOptions{RPID: rpID, UserHandle: []byte(handle), UserName: user})
	require.NoError(tb, err)

	return cred
}

// ids returns the IDs of the credentials of the entries.
func ids(entries []Entry) [][]byte {
	out := make([][]byte, 0, len(entries))
	for _, e := range entries {
		out = append(out, e.Credential.ID)
	}

	return out
}

// readCreds returns the credentials stored in a secret.
func readCreds(ctx context.Context, tb testing.TB, s SecretStore, name string) []*Credential {
	tb.Helper()

	sec, err := s.Get(ctx, name, "latest")
	require.NoError(tb, err)

	creds, err := FromSecret(sec)
	require.NoError(tb, err)

	return creds
}

// writeCreds stores credentials in a secret, bypassing the Store.
func writeCreds(ctx context.Context, tb testing.TB, s SecretStore, name string, creds ...*Credential) {
	tb.Helper()

	sec := secrets.New()
	require.NoError(tb, ToSecret(sec, creds))
	require.NoError(tb, s.Set(ctx, name, sec))
}

func TestStoreNames(t *testing.T) {
	t.Parallel()

	ctx := testContext(t)
	api := apimock.New()

	for _, name := range []string{
		"passkeys/example.com/alice",
		"passkeys/example.com/bob",
		"passkeys/example.org/alice",
		"team/passkeys/example.com/shared",
		"a/b/passkeys/example.com/deep",
		// not below a folder for a Relying Party
		"passkeys/example.com",
		"passkeys/orphan",
		"passkeys/example.com/sub/alice",
		"websites/example.com/alice",
		"notpasskeys/example.com/alice",
		"passkeys",
	} {
		require.NoError(t, api.Set(ctx, name, secrets.New()))
	}

	s := NewStore(ctx, api)
	assert.Equal(t, "passkeys", s.Prefix())

	// The secrets below the prefix come first.
	names, err := s.Names(ctx, "example.com")
	require.NoError(t, err)
	assert.Equal(t, []string{
		"passkeys/example.com/alice",
		"passkeys/example.com/bob",
		"a/b/passkeys/example.com/deep",
		"team/passkeys/example.com/shared",
	}, names)

	names, err = s.Names(ctx, "example.org")
	require.NoError(t, err)
	assert.Equal(t, []string{"passkeys/example.org/alice"}, names)

	names, err = s.Names(ctx, "")
	require.NoError(t, err)
	assert.Len(t, names, 5)

	names, err = s.Names(ctx, "example.net")
	require.NoError(t, err)
	assert.Empty(t, names)
}

func TestStorePut(t *testing.T) {
	t.Parallel()

	ctx := testContext(t)
	api := apimock.New()
	s := NewStore(ctx, api)

	alice := newCred(t, "example.com", "alice", "handle-alice")

	name, err := s.Put(ctx, alice)
	require.NoError(t, err)
	assert.Equal(t, "passkeys/example.com/alice", name)

	// A credential is a single line. The password of the secret is unused.
	sec, err := api.Get(ctx, name, "latest")
	require.NoError(t, err)
	assert.Empty(t, sec.Password())
	assert.Equal(t, []string{SecretKey}, sec.Keys())

	line, found := sec.Get(SecretKey)
	require.True(t, found)
	assert.True(t, strings.HasPrefix(line, `{"v":1,"rpId":"example.com","userName":"alice",`), line)

	bob := newCred(t, "example.com", "bob", "handle-bob")
	name, err = s.Put(ctx, bob)
	require.NoError(t, err)
	assert.Equal(t, "passkeys/example.com/bob", name)

	other := newCred(t, "example.org", "alice", "handle-alice")
	name, err = s.Put(ctx, other)
	require.NoError(t, err)
	assert.Equal(t, "passkeys/example.org/alice", name)

	entries, err := s.List(ctx, "example.com")
	require.NoError(t, err)
	assert.Equal(t, [][]byte{alice.ID, bob.ID}, ids(entries))
	assert.Equal(t, "passkeys/example.com/alice", entries[0].Name)

	entries, err = s.List(ctx, "")
	require.NoError(t, err)
	assert.Len(t, entries, 3)
}

func TestStorePutKeepsWhatIsThere(t *testing.T) {
	t.Parallel()

	ctx := testContext(t)
	api := apimock.New()
	s := NewStore(ctx, api)

	first := newCred(t, "example.com", "alice", "handle")
	name, err := s.Put(ctx, first)
	require.NoError(t, err)

	// The secret may hold more than passkeys. The rest of it is kept.
	sec, err := api.Get(ctx, name, "latest")
	require.NoError(t, err)
	sec.SetPassword("hunter2")
	require.NoError(t, sec.Set("comment", "my passkey"))
	require.NoError(t, api.Set(ctx, name, sec))

	// The user renamed the account at the site, or the secret. A passkey
	// for the same user handle still ends up in the secret of the account.
	require.NoError(t, api.Rename(ctx, name, "passkeys/example.com/work"))

	found, ok, err := s.AccountSecret(ctx, "example.com", []byte("handle"))
	require.NoError(t, err)
	require.True(t, ok)
	assert.Equal(t, "passkeys/example.com/work", found)

	second := newCred(t, "example.com", "alice.renamed", "handle")
	name, err = s.Put(ctx, second)
	require.NoError(t, err)
	assert.Equal(t, "passkeys/example.com/work", name)

	// A passkey that is already stored is never removed: the new one may
	// not work, e.g. because its registration failed at the site.
	creds := readCreds(ctx, t, api, name)
	require.Len(t, creds, 2)
	assert.Equal(t, first.ID, creds[0].ID)
	assert.Equal(t, second.ID, creds[1].ID)

	sec, err = api.Get(ctx, name, "latest")
	require.NoError(t, err)
	assert.Equal(t, "hunter2", sec.Password())

	comment, _ := sec.Get("comment")
	assert.Equal(t, "my passkey", comment)

	// Another account with a name that is taken shares the secret.
	third := newCred(t, "example.com", "work", "another-handle")
	name, err = s.Put(ctx, third)
	require.NoError(t, err)
	assert.Equal(t, "passkeys/example.com/work", name)

	_, ok, err = s.AccountSecret(ctx, "example.com", []byte("unknown"))
	require.NoError(t, err)
	assert.False(t, ok)

	entries, err := s.List(ctx, "example.com")
	require.NoError(t, err)
	assert.Equal(t, [][]byte{first.ID, second.ID, third.ID}, ids(entries))
}

func TestStorePutIsIdempotent(t *testing.T) {
	t.Parallel()

	ctx := testContext(t)
	api := apimock.New()
	s := NewStore(ctx, api)

	first := newCred(t, "example.com", "alice", "handle")
	second := newCred(t, "example.com", "alice", "handle")

	for _, cred := range []*Credential{first, second, first, second} {
		name, err := s.Put(ctx, cred)
		require.NoError(t, err)
		assert.Equal(t, "passkeys/example.com/alice", name)
	}

	entries, err := s.List(ctx, "example.com")
	require.NoError(t, err)
	assert.Equal(t, [][]byte{first.ID, second.ID}, ids(entries))

	// Stores refuse to write a secret that did not change. Storing a
	// credential that is already there does not even try.
	ro := NewStore(ctx, readOnlyStore{api})

	name, err := ro.Put(ctx, first)
	require.NoError(t, err)
	assert.Equal(t, "passkeys/example.com/alice", name)

	_, err = ro.Put(ctx, newCred(t, "example.com", "alice", "handle"))
	require.ErrorIs(t, err, errReadOnly)

	// A credential with a known ID replaces the stored one.
	changed, err := Decode(must(first.Encode()))
	require.NoError(t, err)
	changed.UserDisplayName = "Alice"

	_, err = s.Put(ctx, changed)
	require.NoError(t, err)

	creds := readCreds(ctx, t, api, name)
	require.Len(t, creds, 2)
	assert.Equal(t, first.ID, creds[0].ID)
	assert.Equal(t, "Alice", creds[0].UserDisplayName)
}

func TestStorePutIsConfinedToPrefix(t *testing.T) {
	t.Parallel()

	cfg := config.NewInMemory()
	require.NoError(t, cfg.Set("", "passkey.prefix", "/personal/keys/"))
	ctx := cfg.WithConfig(context.Background())

	api := apimock.New()
	s := NewStore(ctx, api)
	assert.Equal(t, "personal/keys", s.Prefix())

	// A passkey for the same account lives in a store shared with others.
	shared := newCred(t, "example.com", "alice", "handle")
	writeCreds(ctx, t, api, "team/passkeys/example.com/alice", shared)

	_, ok, err := s.AccountSecret(ctx, "example.com", []byte("handle"))
	require.NoError(t, err)
	assert.False(t, ok)

	// A new passkey is never written there.
	mine := newCred(t, "example.com", "alice", "handle")
	name, err := s.Put(ctx, mine)
	require.NoError(t, err)
	assert.Equal(t, "personal/keys/example.com/alice", name)

	// Both are found, the one below the prefix first.
	entries, err := s.List(ctx, "example.com")
	require.NoError(t, err)
	assert.Equal(t, [][]byte{mine.ID, shared.ID}, ids(entries))

	got := readCreds(ctx, t, api, "team/passkeys/example.com/alice")
	require.Len(t, got, 1)
	assert.Equal(t, shared.ID, got[0].ID)

	// A prefix can not point outside of the store.
	for in, want := range map[string]string{
		"":              "passkeys",
		"/":             "passkeys",
		"..":            "passkeys",
		"../../outside": "outside",
		"a//b/./c/":     "a/b/c",
	} {
		cfg := config.NewInMemory()
		require.NoError(t, cfg.Set("", "passkey.prefix", in))
		assert.Equal(t, want, NewStore(cfg.WithConfig(context.Background()), api).Prefix(), in)
	}
}

func TestStoreNamesDifferingInCase(t *testing.T) {
	t.Parallel()

	ctx := testContext(t)
	api := apimock.New()
	s := NewStore(ctx, api)

	// Not all file systems tell names apart that only differ in case. A
	// folder that is spelled differently is still found.
	old := newCred(t, "example.com", "Alice", "handle")
	writeCreds(ctx, t, api, "Passkeys/Example.COM/Alice", old)

	entries, err := s.List(ctx, "example.com")
	require.NoError(t, err)
	require.Equal(t, [][]byte{old.ID}, ids(entries))

	// The secret of the account is used as it is spelled.
	second := newCred(t, "example.com", "alice", "handle")
	name, err := s.Put(ctx, second)
	require.NoError(t, err)
	assert.Equal(t, "Passkeys/Example.COM/Alice", name)

	// So is a secret that only differs in case from the one to write.
	third := newCred(t, "example.com", "ALICE", "another-handle")
	name, err = s.Put(ctx, third)
	require.NoError(t, err)
	assert.Equal(t, "Passkeys/Example.COM/Alice", name)

	// But a secret with exactly the name is preferred if both exist.
	exact := newCred(t, "example.com", "bob", "handle-bob")
	writeCreds(ctx, t, api, "Passkeys/example.com/BOB", newCred(t, "example.com", "BOB", "handle-BOB"))
	writeCreds(ctx, t, api, "passkeys/example.com/bob", exact)

	fourth := newCred(t, "example.com", "bob", "handle-bob-2")
	name, err = s.Put(ctx, fourth)
	require.NoError(t, err)
	assert.Equal(t, "passkeys/example.com/bob", name)

	entries, err = s.List(ctx, "example.com")
	require.NoError(t, err)
	assert.Len(t, entries, 6)
}

func TestStoreAdvanceAndRemove(t *testing.T) {
	t.Parallel()

	ctx := testContext(t)
	api := apimock.New()
	s := NewStore(ctx, api)

	a := newCred(t, "example.com", "alice", "handle")

	b, err := New(NewOptions{RPID: "example.com", UserHandle: []byte("handle"), UserName: "alice", SignCounter: true})
	require.NoError(t, err)

	name, err := s.Put(ctx, a)
	require.NoError(t, err)
	_, err = s.Put(ctx, b)
	require.NoError(t, err)

	// The counter that is returned is the one that is stored.
	for want := uint32(2); want < 5; want++ {
		got, err := s.Advance(ctx, name, b)
		require.NoError(t, err)
		assert.Equal(t, want, got.SignCount)

		creds := readCreds(ctx, t, api, name)
		require.Len(t, creds, 2)
		assert.Zero(t, creds[0].SignCount)
		assert.Equal(t, want, creds[1].SignCount)
	}

	// It is taken from the store, not from the caller: a counter that was
	// incremented elsewhere in the meantime is not used again.
	assert.Equal(t, uint32(1), b.SignCount)

	got, err := s.Advance(ctx, name, b)
	require.NoError(t, err)
	assert.Equal(t, uint32(5), got.SignCount)

	// A credential without a counter is not written.
	got, err = NewStore(ctx, readOnlyStore{api}).Advance(ctx, name, a)
	require.NoError(t, err)
	assert.Zero(t, got.SignCount)

	unknown := newCred(t, "example.com", "alice", "handle")
	_, err = s.Advance(ctx, name, unknown)
	require.ErrorIs(t, err, ErrNoCredentials)
	require.ErrorIs(t, s.Remove(ctx, name, unknown.ID), ErrNoCredentials)

	// The credential in the store must be the one that is meant, not just
	// one with its ID: another key or another account is refused.
	for _, other := range []*Credential{
		must(New(NewOptions{RPID: "example.com", UserHandle: []byte("handle"), SignCounter: true})),
		must(New(NewOptions{RPID: "example.com", UserHandle: []byte("another"), SignCounter: true})),
		must(New(NewOptions{RPID: "example.org", UserHandle: []byte("handle"), SignCounter: true})),
	} {
		other.ID = b.ID
		if string(other.UserHandle) == "another" {
			other.Key = b.Key
		}

		_, err = s.Advance(ctx, name, other)
		require.ErrorIs(t, err, ErrNoCredentials)
		assert.Contains(t, err.Error(), "replaced")
	}

	assert.Equal(t, uint32(5), readCreds(ctx, t, api, name)[1].SignCount)

	require.NoError(t, s.Remove(ctx, name, a.ID))

	entries, err := s.List(ctx, "example.com")
	require.NoError(t, err)
	assert.Equal(t, [][]byte{b.ID}, ids(entries))
}

func TestStoreIgnoresMisplacedCredentials(t *testing.T) {
	t.Parallel()

	ctx := testContext(t)
	api := apimock.New()
	s := NewStore(ctx, api)

	// The name of a secret is not encrypted and not authenticated. What a
	// credential is bound to is decided by its content alone.
	cred := newCred(t, "bank.example", "alice", "handle")
	writeCreds(ctx, t, api, "passkeys/evil.example/alice", cred)

	entries, err := s.List(ctx, "evil.example")
	require.ErrorIs(t, err, ErrInvalidCredential)
	assert.Empty(t, entries)

	entries, err = s.List(ctx, "bank.example")
	require.NoError(t, err)
	assert.Empty(t, entries)
}

func TestStoreMalformedLines(t *testing.T) {
	t.Parallel()

	ctx := testContext(t)
	api := apimock.New()
	s := NewStore(ctx, api)

	good := newCred(t, "example.com", "alice", "handle")
	name, err := s.Put(ctx, good)
	require.NoError(t, err)

	// A line that can not be parsed sits next to a passkey that can.
	sec, err := api.Get(ctx, name, "latest")
	require.NoError(t, err)
	require.NoError(t, sec.Add(SecretKey, `{"v":1,"rpId":"example.com"}`))
	require.NoError(t, api.Set(ctx, name, sec))

	before, err := api.Get(ctx, name, "latest")
	require.NoError(t, err)

	// The passkey is still found and usable.
	creds, err := FromSecret(before)
	require.ErrorIs(t, err, ErrInvalidCredential)
	require.Len(t, creds, 1)

	entries, err := s.List(ctx, "example.com")
	require.ErrorIs(t, err, ErrInvalidCredential)
	assert.Equal(t, [][]byte{good.ID}, ids(entries))

	found, ok, err := s.AccountSecret(ctx, "example.com", []byte("handle"))
	require.NoError(t, err)
	require.True(t, ok)
	assert.Equal(t, name, found)

	// But the secret is not changed: writing it would drop the line.
	_, err = s.Put(ctx, newCred(t, "example.com", "alice", "handle"))
	require.ErrorIs(t, err, ErrInvalidCredential)
	require.ErrorIs(t, s.Remove(ctx, name, good.ID), ErrInvalidCredential)
	_, err = s.Advance(ctx, name, good)
	require.ErrorIs(t, err, ErrInvalidCredential)

	after, err := api.Get(ctx, name, "latest")
	require.NoError(t, err)
	assert.Equal(t, string(before.Bytes()), string(after.Bytes()))
}

func TestStoreBlankLine(t *testing.T) {
	t.Parallel()

	ctx := testContext(t)
	api := apimock.New()
	s := NewStore(ctx, api)

	good := newCred(t, "example.com", "alice", "handle")
	name, err := s.Put(ctx, good)
	require.NoError(t, err)

	// A line without a value is not a passkey, but it is not nothing
	// either: it is kept, which means that the secret is not rewritten.
	sec, err := api.Get(ctx, name, "latest")
	require.NoError(t, err)
	require.NoError(t, sec.Add(SecretKey, ""))
	require.NoError(t, api.Set(ctx, name, sec))

	before, err := api.Get(ctx, name, "latest")
	require.NoError(t, err)
	require.Contains(t, string(before.Bytes()), "\npasskey: \n")

	creds, err := FromSecret(before)
	require.ErrorIs(t, err, ErrInvalidCredential)
	require.Len(t, creds, 1)

	_, err = s.Put(ctx, newCred(t, "example.com", "alice", "handle"))
	require.ErrorIs(t, err, ErrInvalidCredential)

	after, err := api.Get(ctx, name, "latest")
	require.NoError(t, err)
	assert.Equal(t, string(before.Bytes()), string(after.Bytes()))
}

// refStore is a store that resolves references in passwords when it is asked
// to, like the root store does.
type refStore struct {
	SecretStore
}

func (r refStore) Get(ctx context.Context, name, revision string) (gopass.Secret, error) {
	sec, err := r.SecretStore.Get(ctx, name, revision)
	if err != nil {
		return nil, err
	}

	if _, ok := sec.Ref(); ok && ctxutil.IsFollowRef(ctx) {
		sec.SetPassword("the password of the referenced secret")
	}

	return sec, nil
}

func TestStoreKeepsPasswordReference(t *testing.T) {
	t.Parallel()

	// References are followed for everything this context is used for.
	ctx := ctxutil.WithFollowRef(testContext(t), true)
	api := apimock.New()

	sec := secrets.New()
	sec.SetPassword("gopass://websites/example.com/alice")
	require.NoError(t, api.Set(ctx, "passkeys/example.com/alice", sec))

	// A secret that is written back keeps the reference. It does not get
	// the password of the secret that is referenced.
	s := NewStore(ctx, refStore{api})

	name, err := s.Put(ctx, newCred(t, "example.com", "alice", "handle"))
	require.NoError(t, err)
	assert.Equal(t, "passkeys/example.com/alice", name)

	stored, err := api.Get(ctx, name, "latest")
	require.NoError(t, err)
	assert.Equal(t, "gopass://websites/example.com/alice", stored.Password())
	assert.Len(t, readCreds(ctx, t, api, name), 1)
}

// failingStore is a store that lists secrets it can not read.
type failingStore struct {
	*apimock.MockAPI
}

var errDecrypt = errors.New("decryption failed")

func (f failingStore) Get(context.Context, string, string) (gopass.Secret, error) {
	return nil, errDecrypt
}

func TestStoreNeverOverwritesUnreadableSecrets(t *testing.T) {
	t.Parallel()

	ctx := testContext(t)
	api := apimock.New()

	cred := newCred(t, "example.com", "alice", "handle")
	name, err := NewStore(ctx, api).Put(ctx, cred)
	require.NoError(t, err)

	before, err := api.Get(ctx, name, "latest")
	require.NoError(t, err)

	s := NewStore(ctx, failingStore{api})

	entries, err := s.List(ctx, "example.com")
	require.ErrorIs(t, err, errDecrypt)
	assert.Empty(t, entries)

	// The secret to write exists, but can not be read.
	_, err = s.Put(ctx, newCred(t, "example.com", "alice", "other"))
	require.ErrorIs(t, err, errDecrypt)

	// The names differ in case only. On a case insensitive file system both
	// denote the same file.
	_, err = s.Put(ctx, newCred(t, "example.com", "Alice", "other"))
	require.ErrorIs(t, err, errDecrypt)

	_, err = s.Advance(ctx, name, cred)
	require.ErrorIs(t, err, errDecrypt)
	require.ErrorIs(t, s.Remove(ctx, name, cred.ID), errDecrypt)

	after, err := api.Get(ctx, name, "latest")
	require.NoError(t, err)
	assert.Equal(t, string(before.Bytes()), string(after.Bytes()))
}

func TestStoreNotSynced(t *testing.T) {
	t.Parallel()

	ctx := testContext(t)
	api := apimock.New()
	s := NewStore(ctx, unsyncedStore{api})

	// The store wrote the secret, but could not push it to its remote. The
	// passkey is stored, and that is what the caller is told.
	cred, err := New(NewOptions{RPID: "example.com", UserHandle: []byte("handle"), UserName: "alice", SignCounter: true})
	require.NoError(t, err)

	name, err := s.Put(ctx, cred)
	require.ErrorIs(t, err, ErrNotSynced)
	require.ErrorIs(t, err, errPush)
	assert.Equal(t, "passkeys/example.com/alice", name)

	creds := readCreds(ctx, t, api, name)
	require.Len(t, creds, 1)
	assert.Equal(t, cred.ID, creds[0].ID)

	// The same goes for a counter that was written.
	got, err := s.Advance(ctx, name, cred)
	require.ErrorIs(t, err, ErrNotSynced)
	require.NotNil(t, got)
	assert.Equal(t, uint32(2), got.SignCount)
	assert.Equal(t, uint32(2), readCreds(ctx, t, api, name)[0].SignCount)

	require.ErrorIs(t, s.Remove(ctx, name, cred.ID), ErrNotSynced)
	assert.Empty(t, readCreds(ctx, t, api, name))

	// A secret that was not written is a failure, whatever the reason.
	_, err = NewStore(ctx, readOnlyStore{api}).Put(ctx, cred)
	require.ErrorIs(t, err, errReadOnly)
	require.NotErrorIs(t, err, ErrNotSynced)
}

func TestStoreDoesNotWriteAfterCancellation(t *testing.T) {
	t.Parallel()

	api := apimock.New()

	ctx, cancel := context.WithCancel(testContext(t))
	s := NewStore(ctx, api)
	cancel()

	_, err := s.Put(ctx, newCred(t, "example.com", "alice", "handle"))
	require.ErrorIs(t, err, context.Canceled)

	names, err := api.List(ctx)
	require.NoError(t, err)
	assert.Empty(t, names)
}
