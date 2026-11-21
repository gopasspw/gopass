package passkey

import (
	"context"
	"strings"
	"sync"
	"testing"

	"github.com/gopasspw/gopass/internal/config"
	"github.com/gopasspw/gopass/pkg/gopass"
	"github.com/gopasspw/gopass/pkg/gopass/apimock"
	"github.com/gopasspw/gopass/pkg/gopass/secrets"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// countingStore counts how often a secret is read, i.e. decrypted. With a
// key on a token, or without an agent, every read can be a prompt.
type countingStore struct {
	SecretStore

	mu    sync.Mutex
	reads []string
}

func (c *countingStore) Get(ctx context.Context, name, revision string) (gopass.Secret, error) {
	c.mu.Lock()
	c.reads = append(c.reads, name)
	c.mu.Unlock()

	return c.SecretStore.Get(ctx, name, revision) //nolint:wrapcheck
}

// take returns the names of the secrets read since the last call.
func (c *countingStore) take() []string {
	c.mu.Lock()
	defer c.mu.Unlock()

	reads := c.reads
	c.reads = nil

	return reads
}

func TestCeremoniesReadEachSecretOnce(t *testing.T) {
	t.Parallel()

	ctx := testContext(t)
	api := apimock.New()
	cs := &countingStore{SecretStore: api}
	rec := &consentRecorder{consent: Consent{Present: true, Verified: true}}
	a := NewAuthenticator(NewStore(ctx, cs), rec.ask)

	// A new account: there is nothing to read.
	_, _, err := a.Register(ctx, exampleOrigin, creationOptions("alice"))
	require.NoError(t, err)
	assert.Empty(t, cs.take())

	// Another passkey for the account: its secret is read to find the
	// account, and once more right before it is changed.
	_, _, err = a.Register(ctx, exampleOrigin, creationOptions("alice"))
	require.NoError(t, err)
	assert.Equal(t, []string{"passkeys/example.com/alice", "passkeys/example.com/alice"}, cs.take())

	for _, user := range []string{"bob", "carol"} {
		_, _, err = a.Register(ctx, exampleOrigin, creationOptions(user))
		require.NoError(t, err)
	}

	cs.take()

	// The candidates are read once each.
	entries, err := a.Candidates(ctx, exampleOrigin, requestOptions())
	require.NoError(t, err)
	require.Len(t, entries, 4)
	assert.Len(t, cs.take(), 3)

	// Signing with one of them reads nothing more.
	rsp, used, err := a.AuthenticateWith(ctx, exampleOrigin, requestOptions(), entries[2])
	require.NoError(t, err)
	assert.Empty(t, cs.take())
	assert.Equal(t, entries[2].Name, used.Name)
	assert.Equal(t, entries[2].Credential.ID, []byte(rsp.RawID))

	spki, err := entries[2].Credential.PublicKey()
	require.NoError(t, err)
	assert.True(t, verifyAssertion(t, spki, rsp))

	// Authenticate reads the candidates itself.
	_, _, err = a.Authenticate(ctx, exampleOrigin, requestOptions(), entries[2].Credential.ID)
	require.NoError(t, err)
	assert.Len(t, cs.take(), 3)

	// Passkeys in a shared folder are only read for a registration if the
	// site names credentials it already knows: those can be anywhere.
	shared := newCred(t, "example.com", "dave", "handle-dave")
	writeCreds(ctx, t, api, "team/passkeys/example.com/dave", shared)

	_, _, err = a.Register(ctx, exampleOrigin, creationOptions("erin"))
	require.NoError(t, err)
	assert.NotContains(t, cs.take(), "team/passkeys/example.com/dave")

	opts := creationOptions("dave")
	opts.ExcludeCredentials = []CredentialDescriptor{{Type: "public-key", ID: shared.ID}}

	_, _, err = a.Register(ctx, exampleOrigin, opts)
	require.Error(t, err)
	assert.Equal(t, InvalidStateError, NameOf(err))

	reads := cs.take()
	assert.Contains(t, reads, "team/passkeys/example.com/dave")
	assert.Len(t, reads, 5)
}

func TestAuthenticateWith(t *testing.T) {
	t.Parallel()

	ctx := testContext(t)
	api := apimock.New()
	rec := &consentRecorder{consent: Consent{Present: true}}
	st := NewStore(ctx, api)
	a := NewAuthenticator(st, rec.ask)

	// A passkey the user keeps next to the password, outside of any
	// passkeys folder. It is not found by the RP ID, but can be named.
	cred := newCred(t, "example.com", "alice", "handle")
	writeCreds(ctx, t, api, "websites/example.com", cred)

	listed, err := st.List(ctx, "example.com")
	require.NoError(t, err)
	assert.Empty(t, listed)

	entries, err := st.Load(ctx, "websites/example.com")
	require.NoError(t, err)
	require.Len(t, entries, 1)

	rsp, used, err := a.AuthenticateWith(ctx, exampleOrigin, requestOptions(), entries[0])
	require.NoError(t, err)
	assert.Equal(t, "websites/example.com", used.Name)
	require.Len(t, rec.requests, 1)
	assert.Equal(t, OperationAuthenticate, rec.requests[0].Operation)
	assert.Equal(t, "websites/example.com", rec.requests[0].Entry)

	spki, err := cred.PublicKey()
	require.NoError(t, err)
	assert.True(t, verifyAssertion(t, spki, rsp))

	// The request decides what can be used, not the caller. Nobody is
	// asked about a passkey that can not be used.
	rec.requests = nil

	other := requestOptions()
	other.RPID = ptr("example.org")

	_, _, err = a.AuthenticateWith(ctx, Origin{Origin: "https://example.org"}, other, entries[0])
	require.ErrorIs(t, err, ErrNoCredentials)
	assert.Equal(t, NotAllowedError, NameOf(err))
	assert.True(t, IsFallback(err))

	allowed := requestOptions()
	allowed.AllowCredentials = []CredentialDescriptor{{Type: "public-key", ID: []byte("another passkey")}}

	_, _, err = a.AuthenticateWith(ctx, exampleOrigin, allowed, entries[0])
	require.ErrorIs(t, err, ErrNoCredentials)

	_, _, err = a.AuthenticateWith(ctx, Origin{Origin: "https://evil.example"}, requestOptions(), entries[0])
	assert.Equal(t, SecurityError, NameOf(err))

	_, _, err = a.AuthenticateWith(ctx, exampleOrigin, requestOptions(), Entry{Name: "nothing"})
	assert.Equal(t, TypeError, NameOf(err))

	_, _, err = a.AuthenticateWith(ctx, exampleOrigin, nil, entries[0])
	assert.Equal(t, TypeError, NameOf(err))

	assert.Empty(t, rec.requests)

	allowed.AllowCredentials = append(allowed.AllowCredentials, CredentialDescriptor{Type: "public-key", ID: cred.ID})

	_, _, err = a.AuthenticateWith(ctx, exampleOrigin, allowed, entries[0])
	require.NoError(t, err)

	// A counter is advanced in the secret the passkey was read from.
	counting, err := New(NewOptions{RPID: "example.com", UserHandle: []byte("handle"), UserName: "bob", SignCounter: true})
	require.NoError(t, err)
	writeCreds(ctx, t, api, "websites/bob@example.com", counting)

	entries, err = st.Load(ctx, "websites/bob@example.com")
	require.NoError(t, err)
	require.Len(t, entries, 1)

	rsp, used, err = a.AuthenticateWith(ctx, exampleOrigin, requestOptions(), entries[0])
	require.NoError(t, err)
	assert.Equal(t, uint32(2), used.Credential.SignCount)
	assert.Equal(t, uint32(2), parseAuthData(t, rsp.Response.AuthenticatorData).signCount)
	assert.Equal(t, uint32(2), readCreds(ctx, t, api, "websites/bob@example.com")[0].SignCount)
}

func TestStoreLoad(t *testing.T) {
	t.Parallel()

	ctx := testContext(t)
	api := apimock.New()
	cs := &countingStore{SecretStore: api}
	st := NewStore(ctx, cs)

	_, err := st.Load(ctx, "missing")
	require.Error(t, err)
	cs.take()

	// A secret without passkeys is no error. It just holds none.
	plain := secrets.New()
	plain.SetPassword("hunter2")
	require.NoError(t, api.Set(ctx, "example.com", plain))

	entries, err := st.Load(ctx, "example.com")
	require.NoError(t, err)
	assert.Empty(t, entries)
	assert.Equal(t, []string{"example.com"}, cs.take())

	// Lines that can not be parsed are reported with the ones that can.
	good := newCred(t, "example.org", "alice", "handle")
	line, err := good.Encode()
	require.NoError(t, err)

	mixed := secrets.ParseAKV([]byte("\npasskey: {\"v\":1}\npasskey: " + string(line) + "\n"))
	require.NoError(t, api.Set(ctx, "passkeys/example.com/alice", mixed))

	entries, err = st.Load(ctx, "passkeys/example.com/alice")
	require.ErrorIs(t, err, ErrInvalidCredential)
	require.Len(t, entries, 1)
	assert.Equal(t, good.ID, entries[0].Credential.ID)
	// The RP ID is the one the passkey holds, whatever the name says.
	assert.Equal(t, "example.org", entries[0].Credential.RPID)
	assert.Equal(t, []string{"passkeys/example.com/alice"}, cs.take())
}

func TestStoreFindUser(t *testing.T) {
	t.Parallel()

	ctx := testContext(t)
	api := apimock.New()
	cs := &countingStore{SecretStore: api}
	st := NewStore(ctx, cs)

	alice := newCred(t, "example.com", "alice", "handle-alice")
	_, err := st.Put(ctx, alice)
	require.NoError(t, err)

	writeCreds(ctx, t, api, "team/passkeys/example.com/carol", newCred(t, "example.com", "carol", "handle-carol"))
	writeCreds(ctx, t, api, "passkeys/example.com/eve", newCred(t, "example.org", "eve", "handle-eve"))
	cs.take()

	// Only the secret named after the user is read.
	got, found := st.FindUser(ctx, "example.com", "alice")
	require.True(t, found)
	assert.Equal(t, alice.UserHandle, got.UserHandle)
	assert.Equal(t, []string{"passkeys/example.com/alice"}, cs.take())

	_, found = st.FindUser(ctx, "example.com", "bob")
	assert.False(t, found)
	assert.Empty(t, cs.take())

	// User names are compared as they are.
	_, found = st.FindUser(ctx, "example.com", "Alice")
	assert.False(t, found)

	// Accounts in shared folders are not this user's.
	_, found = st.FindUser(ctx, "example.com", "carol")
	assert.False(t, found)

	// A passkey is bound to the RP ID it holds.
	_, found = st.FindUser(ctx, "example.com", "eve")
	assert.False(t, found)

	_, found = st.FindUser(ctx, "example.com", "..")
	assert.False(t, found)
}

func TestStoreKeepsUnknownMembers(t *testing.T) {
	t.Parallel()

	ctx := testContext(t)
	api := apimock.New()
	st := NewStore(ctx, api)

	// A passkey written by a later version of gopass, with a member this
	// version does not know.
	first := newCred(t, "example.com", "alice", "handle")
	line, err := first.Encode()
	require.NoError(t, err)

	extended := strings.Replace(string(line), `{"v":1,`, `{"v":1,"note":"from the future",`, 1)
	require.NoError(t, api.Set(ctx, "passkeys/example.com/alice", secrets.ParseAKV([]byte("\npasskey: "+extended+"\n"))))

	// Another passkey for the account rewrites the secret.
	name, err := st.Put(ctx, newCred(t, "example.com", "alice", "handle"))
	require.NoError(t, err)
	assert.Equal(t, "passkeys/example.com/alice", name)

	sec, err := api.Get(ctx, name, "latest")
	require.NoError(t, err)

	values, _ := sec.Values(SecretKey)
	require.Len(t, values, 2)
	assert.Contains(t, values[0], `"note":"from the future"`)
	assert.NotContains(t, values[1], `"note"`)
}

func TestConfigIsNotReadFromTheStore(t *testing.T) {
	t.Setenv("GOPASS_HOMEDIR", t.TempDir())
	t.Setenv("GOPASS_CONFIG_NOSYSTEM", "true")

	// The first run records the path of the root store, the second loads
	// its config.
	_ = config.New()
	cfg := config.New()
	ctx := cfg.WithConfig(t.Context())

	// The config of a store is versioned with it. Whoever can push to the
	// store could otherwise claim user verification, or have new passkeys
	// written to a folder that is shared with them.
	for _, key := range []string{"passkey.user-verification", "passkey.signcount"} {
		require.NoError(t, cfg.Set("<root>", key, "true"))
	}
	require.NoError(t, cfg.Set("<root>", "passkey.prefix", "shared/passkeys"))
	require.True(t, config.Bool(ctx, "passkey.user-verification"), "the store config is in effect")

	assert.False(t, UserVerificationEnabled(ctx))
	assert.False(t, SignCounterEnabled(ctx))
	assert.Equal(t, Folder, NewStore(ctx, apimock.New()).Prefix())

	consent, err := ConsentByConfirmation(func(context.Context, ConsentRequest) (bool, error) { return true, nil })(ctx, ConsentRequest{})
	require.NoError(t, err)
	assert.Equal(t, Consent{Present: true}, consent)

	// The config of the user is.
	for _, key := range []string{"passkey.user-verification", "passkey.signcount"} {
		require.NoError(t, cfg.Set("", key, "true"))
	}
	require.NoError(t, cfg.Set("", "passkey.prefix", "mine"))

	assert.True(t, UserVerificationEnabled(ctx))
	assert.True(t, SignCounterEnabled(ctx))
	assert.Equal(t, "mine", NewStore(ctx, apimock.New()).Prefix())
}
