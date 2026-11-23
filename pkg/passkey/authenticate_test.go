package passkey

import (
	"context"
	"testing"

	"github.com/gopasspw/gopass/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAuthenticateSelection(t *testing.T) {
	t.Parallel()

	ctx := testContext(t)
	a, rec := newAuthenticator(ctx)

	// no passkey at all
	_, err := a.Candidates(ctx, exampleOrigin, requestOptions())
	require.ErrorIs(t, err, ErrNoCredentials)
	assert.Equal(t, NotAllowedError, NameOf(err))
	assert.True(t, IsFallback(err))

	_, _, err = a.Authenticate(ctx, exampleOrigin, requestOptions(), nil)
	require.ErrorIs(t, err, ErrNoCredentials)

	alice, _, err := a.Register(ctx, exampleOrigin, creationOptions("alice"))
	require.NoError(t, err)
	bob, _, err := a.Register(ctx, exampleOrigin, creationOptions("bob"))
	require.NoError(t, err)

	other := creationOptions("carol")
	other.RP.ID = ptr("login.example.com")
	_, _, err = a.Register(ctx, exampleOrigin, other)
	require.NoError(t, err)

	// Only the passkeys of the Relying Party are candidates.
	entries, err := a.Candidates(ctx, exampleOrigin, requestOptions())
	require.NoError(t, err)
	assert.Equal(t, [][]byte{alice.RawID, bob.RawID}, ids(entries))

	// With more than one the user has to choose. Nobody is asked before.
	rec.requests = nil
	_, _, err = a.Authenticate(ctx, exampleOrigin, requestOptions(), nil)
	require.ErrorIs(t, err, ErrSelectionRequired)
	assert.Equal(t, NotAllowedError, NameOf(err))
	assert.False(t, IsFallback(err))
	assert.Empty(t, rec.requests)

	auth, entry, err := a.Authenticate(ctx, exampleOrigin, requestOptions(), bob.RawID)
	require.NoError(t, err)
	assert.Equal(t, bob.ID, auth.ID)
	assert.Equal(t, "passkeys/example.com/bob", entry.Name)
	assert.Equal(t, Base64URL("handle-bob"), auth.Response.UserHandle)
	assert.True(t, verifyAssertion(t, bob.Response.PublicKey, auth))
	assert.False(t, verifyAssertion(t, alice.Response.PublicKey, auth))
	assert.Equal(t, "bob", rec.requests[0].UserName)

	// A passkey that is not a candidate can not be selected.
	_, _, err = a.Authenticate(ctx, exampleOrigin, requestOptions(), []byte("unknown credential"))
	require.ErrorIs(t, err, ErrNoCredentials)

	// The Relying Party can restrict the passkeys it accepts.
	opts := requestOptions()
	opts.AllowCredentials = []CredentialDescriptor{
		{Type: "public-key", ID: []byte("unknown credential")},
		{Type: "public-key", ID: alice.RawID, Transports: []string{"internal", "hybrid"}},
	}

	entries, err = a.Candidates(ctx, exampleOrigin, opts)
	require.NoError(t, err)
	assert.Equal(t, [][]byte{alice.RawID}, ids(entries))

	auth, _, err = a.Authenticate(ctx, exampleOrigin, opts, nil)
	require.NoError(t, err)
	assert.Equal(t, alice.ID, auth.ID)

	_, _, err = a.Authenticate(ctx, exampleOrigin, opts, bob.RawID)
	require.ErrorIs(t, err, ErrNoCredentials)

	opts.AllowCredentials = []CredentialDescriptor{{Type: "unknown", ID: alice.RawID}}
	_, _, err = a.Authenticate(ctx, exampleOrigin, opts, nil)
	require.ErrorIs(t, err, ErrNoCredentials)
}

func TestAuthenticateRejected(t *testing.T) {
	t.Parallel()

	ctx := testContext(t)
	a, rec := newAuthenticator(ctx)

	_, _, err := a.Register(ctx, exampleOrigin, creationOptions("alice"))
	require.NoError(t, err)

	rec.requests = nil

	for name, tc := range map[string]struct {
		origin Origin
		mutate func(*RequestOptions)
		want   ErrorName
	}{
		"no challenge":  {exampleOrigin, func(o *RequestOptions) { o.Challenge = nil }, TypeError},
		"other site":    {Origin{Origin: "https://evil.example"}, func(*RequestOptions) {}, SecurityError},
		"lookalike":     {Origin{Origin: "https://example.com.evil.example"}, func(*RequestOptions) {}, SecurityError},
		"insecure":      {Origin{Origin: "http://example.com"}, func(*RequestOptions) {}, SecurityError},
		"public suffix": {exampleOrigin, func(o *RequestOptions) { o.RPID = ptr("com") }, SecurityError},
		"empty rp id":   {exampleOrigin, func(o *RequestOptions) { o.RPID = ptr("") }, SecurityError},
		"opaque":        {Origin{Origin: "null"}, func(*RequestOptions) {}, NotAllowedError},
		"bad top origin": {
			Origin{Origin: "https://example.com", CrossOrigin: true, TopOrigin: "null"}, func(*RequestOptions) {}, NotAllowedError,
		},
	} {
		opts := requestOptions()
		tc.mutate(opts)

		_, _, err := a.Authenticate(ctx, tc.origin, opts, nil)
		require.Error(t, err, name)
		assert.Equal(t, tc.want, NameOf(err), "%s: %s", name, err)

		_, err = a.Candidates(ctx, tc.origin, opts)
		require.Error(t, err, name)
		assert.Equal(t, tc.want, NameOf(err), "%s: %s", name, err)
	}

	// The passkey of example.com is not offered to another site that merely
	// did not say which RP ID it wants.
	opts := requestOptions()
	opts.RPID = nil
	_, _, err = a.Authenticate(ctx, Origin{Origin: "https://evil.example"}, opts, nil)
	require.ErrorIs(t, err, ErrNoCredentials)

	assert.Empty(t, rec.requests)

	_, _, err = a.Authenticate(ctx, exampleOrigin, nil, nil)
	assert.Equal(t, TypeError, NameOf(err))
	_, err = a.Candidates(ctx, exampleOrigin, nil)
	assert.Equal(t, TypeError, NameOf(err))
}

func TestCrossOrigin(t *testing.T) {
	t.Parallel()

	ctx := testContext(t)
	a, rec := newAuthenticator(ctx)

	reg, _, err := a.Register(ctx, exampleOrigin, creationOptions("alice"))
	require.NoError(t, err)

	// A request from an embedded document names the top-level origin.
	origin := Origin{Origin: "https://login.example.com", CrossOrigin: true, TopOrigin: "https://shop.example/"}

	auth, _, err := a.Authenticate(ctx, origin, requestOptions(), nil)
	require.NoError(t, err)
	equalBytes(t,
		`{"type":"webauthn.get","challenge":"YXV0aGVudGljYXRpb24gY2hhbGxlbmdl","origin":"https://login.example.com","crossOrigin":true,"topOrigin":"https://shop.example"}`,
		auth.Response.ClientDataJSON)
	assert.True(t, verifyAssertion(t, reg.Response.PublicKey, auth))

	// The user is told which site embeds the one that asks.
	rec.requests = nil
	_, _, err = a.Authenticate(ctx, origin, requestOptions(), nil)
	require.NoError(t, err)
	assert.Equal(t, "https://login.example.com", rec.requests[0].Origin)
	assert.Equal(t, "https://shop.example", rec.requests[0].TopOrigin)

	// The top-level origin is ignored for same-origin requests.
	origin.CrossOrigin = false
	rec.requests = nil
	auth, _, err = a.Authenticate(ctx, origin, requestOptions(), nil)
	require.NoError(t, err)
	assert.NotContains(t, string(auth.Response.ClientDataJSON), "topOrigin")
	assert.Empty(t, rec.requests[0].TopOrigin)
}

func TestAuthenticateSignCounter(t *testing.T) {
	t.Parallel()

	cfg := config.NewInMemory()
	require.NoError(t, cfg.Set("", "passkey.signcount", "true"))
	ctx := cfg.WithConfig(context.Background())

	a, _ := newAuthenticator(ctx)

	reg, entry, err := a.Register(ctx, exampleOrigin, creationOptions("alice"))
	require.NoError(t, err)

	last := parseAuthData(t, reg.Response.AuthenticatorData).signCount
	assert.Equal(t, uint32(1), last)

	for range 3 {
		auth, _, err := a.Authenticate(ctx, exampleOrigin, requestOptions(), nil)
		require.NoError(t, err)
		assert.True(t, verifyAssertion(t, reg.Response.PublicKey, auth))

		count := parseAuthData(t, auth.Response.AuthenticatorData).signCount
		assert.Greater(t, count, last)
		last = count

		// The counter the Relying Party saw is the one that is stored.
		creds := readCreds(ctx, t, a.store.store, entry.Name)
		require.Len(t, creds, 1)
		assert.Equal(t, count, creds[0].SignCount)
	}

	// The counter is taken from the store when the assertion is made. Two
	// ceremonies that overlap do not hand out the same value.
	var counts []uint32

	nested := NewAuthenticator(a.store, func(ctx context.Context, _ ConsentRequest) (Consent, error) {
		// another ceremony completes while this one waits for the user
		auth, _, err := a.Authenticate(ctx, exampleOrigin, requestOptions(), nil)
		require.NoError(t, err)
		counts = append(counts, parseAuthData(t, auth.Response.AuthenticatorData).signCount)

		return Consent{Present: true}, nil
	})

	auth, _, err := nested.Authenticate(ctx, exampleOrigin, requestOptions(), nil)
	require.NoError(t, err)
	assert.True(t, verifyAssertion(t, reg.Response.PublicKey, auth))

	counts = append(counts, parseAuthData(t, auth.Response.AuthenticatorData).signCount)
	assert.Equal(t, []uint32{last + 1, last + 2}, counts)

	// An assertion is not released if its counter can not be persisted.
	failing := NewAuthenticator(NewStore(ctx, readOnlyStore{a.store.store}), a.consent)
	_, _, err = failing.Authenticate(ctx, exampleOrigin, requestOptions(), nil)
	require.ErrorIs(t, err, errReadOnly)
	assert.Equal(t, UnknownError, NameOf(err))
}

func TestOwnCredentialTakesPrecedence(t *testing.T) {
	t.Parallel()

	ctx := testContext(t)
	a, rec := newAuthenticator(ctx)

	reg, entry, err := a.Register(ctx, exampleOrigin, creationOptions("alice"))
	require.NoError(t, err)

	// Somebody with access to a shared store plants a credential with the
	// ID of the passkey of the user, but another key. Its name sorts first.
	planted, err := New(NewOptions{RPID: "example.com", UserHandle: []byte("handle-alice"), UserName: "alice"})
	require.NoError(t, err)
	planted.ID = entry.Credential.ID
	writeCreds(ctx, t, a.store.store, "a-team/passkeys/example.com/alice", planted)

	// and a copy of a passkey that is fine to be shared
	shared := newCred(t, "example.com", "shared", "handle-shared")
	writeCreds(ctx, t, a.store.store, "a-team/passkeys/example.com/shared", shared)

	entries, err := a.Candidates(ctx, exampleOrigin, requestOptions())
	require.NoError(t, err)
	require.Len(t, entries, 2)
	assert.Equal(t, "passkeys/example.com/alice", entries[0].Name)
	assert.Equal(t, "a-team/passkeys/example.com/shared", entries[1].Name)

	// The passkey of the user keeps working.
	rec.requests = nil

	auth, used, err := a.Authenticate(ctx, exampleOrigin, requestOptions(), reg.RawID)
	require.NoError(t, err)
	assert.Equal(t, "passkeys/example.com/alice", used.Name)
	assert.Equal(t, "passkeys/example.com/alice", rec.requests[0].Entry)
	assert.True(t, verifyAssertion(t, reg.Response.PublicKey, auth))
}
