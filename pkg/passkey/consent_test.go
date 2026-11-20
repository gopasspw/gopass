package passkey

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/gopasspw/gopass/internal/config"
	"github.com/gopasspw/gopass/pkg/gopass"
	"github.com/gopasspw/gopass/pkg/gopass/apimock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConsent(t *testing.T) {
	t.Parallel()

	ctx := testContext(t)
	a, rec := newAuthenticator(ctx)

	// declined
	rec.consent = Consent{}
	_, _, err := a.Register(ctx, exampleOrigin, creationOptions("alice"))
	require.ErrorIs(t, err, ErrDeclined)
	assert.Equal(t, NotAllowedError, NameOf(err))
	assert.False(t, IsFallback(err))

	// A verification without a test of user presence is no consent.
	rec.consent = Consent{Verified: true}
	_, _, err = a.Register(ctx, exampleOrigin, creationOptions("alice"))
	require.ErrorIs(t, err, ErrDeclined)

	// the user can not be asked
	rec.consent = Consent{Present: true, Verified: true}
	rec.err = errors.New("no pinentry")
	_, _, err = a.Register(ctx, exampleOrigin, creationOptions("alice"))
	require.ErrorIs(t, err, rec.err)
	assert.Equal(t, NotAllowedError, NameOf(err))
	rec.err = nil

	entries, err := a.store.List(ctx, "")
	require.NoError(t, err)
	assert.Empty(t, entries)

	// The site requires a verification that was not performed.
	rec.consent = Consent{Present: true}
	opts := creationOptions("alice")
	opts.AuthenticatorSelection = &AuthenticatorSelection{UserVerification: "required"}
	_, _, err = a.Register(ctx, exampleOrigin, opts)
	require.ErrorIs(t, err, ErrUnsuitable)
	assert.True(t, rec.requests[len(rec.requests)-1].VerificationRequired)

	// Otherwise the flags say exactly what happened.
	opts.AuthenticatorSelection.UserVerification = "preferred"
	reg, _, err := a.Register(ctx, exampleOrigin, opts)
	require.NoError(t, err)
	assert.Equal(t,
		FlagUserPresent|FlagBackupEligible|FlagBackupState|FlagAttestedCredentialData,
		parseAuthData(t, reg.Response.AuthenticatorData).flags)

	req := requestOptions()
	auth, _, err := a.Authenticate(ctx, exampleOrigin, req, nil)
	require.NoError(t, err)
	assert.Equal(t, FlagUserPresent|FlagBackupEligible|FlagBackupState, parseAuthData(t, auth.Response.AuthenticatorData).flags)

	req.UserVerification = "required"
	_, _, err = a.Authenticate(ctx, exampleOrigin, req, nil)
	require.ErrorIs(t, err, ErrUnsuitable)
	assert.True(t, IsFallback(err))

	rec.consent = Consent{}
	_, _, err = a.Authenticate(ctx, exampleOrigin, requestOptions(), nil)
	require.ErrorIs(t, err, ErrDeclined)

	// Without a way to ask there is no consent.
	_, _, err = NewAuthenticator(a.store, nil).Authenticate(ctx, exampleOrigin, requestOptions(), nil)
	require.ErrorIs(t, err, ErrDeclined)
}

func TestVerificationRequiredWithoutAsking(t *testing.T) {
	t.Parallel()

	ctx := testContext(t)

	asked := 0
	a := NewAuthenticator(NewStore(ctx, apimock.New()), ConsentByConfirmation(func(context.Context, ConsentRequest) (bool, error) {
		asked++

		return true, nil
	}))

	reg, _, err := a.Register(ctx, exampleOrigin, creationOptions("alice"))
	require.NoError(t, err)
	assert.False(t, parseAuthData(t, reg.Response.AuthenticatorData).flags.Has(FlagUserVerified))
	assert.Equal(t, 1, asked)

	// The site insists on a verification that is not going to happen. The
	// request is left to another authenticator and the user is not asked.
	opts := requestOptions()
	opts.UserVerification = "required"

	_, _, err = a.Authenticate(ctx, exampleOrigin, opts, nil)
	require.ErrorIs(t, err, ErrUnsuitable)
	assert.Equal(t, NotAllowedError, NameOf(err))
	assert.True(t, IsFallback(err))
	assert.Contains(t, err.Error(), "passkey.user-verification")
	assert.NotContains(t, err.Error(), "failed to ask")
	assert.Equal(t, 1, asked)

	create := creationOptions("bob")
	create.AuthenticatorSelection = &AuthenticatorSelection{UserVerification: "required"}

	_, _, err = a.Register(ctx, exampleOrigin, create)
	require.ErrorIs(t, err, ErrUnsuitable)
	assert.Equal(t, 1, asked)

	// Telling a site that a passkey exists never needs a verification.
	create = creationOptions("alice")
	create.AuthenticatorSelection = &AuthenticatorSelection{UserVerification: "required"}
	create.ExcludeCredentials = []CredentialDescriptor{{Type: "public-key", ID: reg.RawID}}

	_, _, err = a.Register(ctx, exampleOrigin, create)
	assert.Equal(t, InvalidStateError, NameOf(err))
	assert.Equal(t, 2, asked)
}

func TestCancellation(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(testContext(t))
	a, rec := newAuthenticator(ctx)

	_, _, err := a.Register(ctx, exampleOrigin, creationOptions("alice"))
	require.NoError(t, err)

	// The ceremony is cancelled while the user is asked, and the user
	// confirms anyway. Nothing is signed and nothing is stored.
	confirmLate := NewAuthenticator(a.store, func(context.Context, ConsentRequest) (Consent, error) {
		cancel()

		return Consent{Present: true, Verified: true}, nil
	})

	rsp, _, err := confirmLate.Authenticate(ctx, exampleOrigin, requestOptions(), nil)
	require.ErrorIs(t, err, context.Canceled)
	assert.Equal(t, NotAllowedError, NameOf(err))
	assert.Nil(t, rsp)

	reg, _, err := confirmLate.Register(ctx, exampleOrigin, creationOptions("bob"))
	require.ErrorIs(t, err, context.Canceled)
	assert.Equal(t, NotAllowedError, NameOf(err))
	assert.Nil(t, reg)

	// A ceremony that is cancelled from the start: nobody is asked.
	rec.requests = nil
	_, _, err = a.Authenticate(ctx, exampleOrigin, requestOptions(), nil)
	require.ErrorIs(t, err, context.Canceled)

	_, _, err = a.Register(ctx, exampleOrigin, creationOptions("carol"))
	require.ErrorIs(t, err, context.Canceled)
	assert.Empty(t, rec.requests)

	names, err := a.store.Names(context.WithoutCancel(ctx), "")
	require.NoError(t, err)
	assert.Equal(t, []string{"passkeys/example.com/alice"}, names)
}

func TestRequestIsFixedBeforeConsent(t *testing.T) {
	t.Parallel()

	ctx := testContext(t)

	// The request is changed while the user looks at it. What the user
	// agreed to is what is done.
	creation := creationOptions("alice")
	request := requestOptions()

	a := NewAuthenticator(NewStore(ctx, apimock.New()), func(_ context.Context, req ConsentRequest) (Consent, error) {
		assert.Equal(t, "alice", req.UserName)

		if req.Operation == OperationRegister {
			creation.User.Name = "mallory"
			creation.User.DisplayName = "Mallory"
			creation.User.ID[0] ^= 0xff
			creation.Challenge[0] ^= 0xff
			creation.Extensions = map[string]json.RawMessage{"credProps": json.RawMessage("true")}
		} else {
			request.Challenge[0] ^= 0xff
		}

		return Consent{Present: true}, nil
	})

	reg, entry, err := a.Register(ctx, exampleOrigin, creation)
	require.NoError(t, err)
	assert.Equal(t, "passkeys/example.com/alice", entry.Name)
	assert.Equal(t, "alice", entry.Credential.UserName)
	assert.Equal(t, "The alice", entry.Credential.UserDisplayName)
	assert.Equal(t, []byte("handle-alice"), entry.Credential.UserHandle)
	assert.Contains(t, string(reg.Response.ClientDataJSON), `"challenge":"cmVnaXN0cmF0aW9uIGNoYWxsZW5nZQ"`)
	assert.Empty(t, reg.ClientExtensionResults)

	auth, _, err := a.Authenticate(ctx, exampleOrigin, request, nil)
	require.NoError(t, err)
	assert.Contains(t, string(auth.Response.ClientDataJSON), `"challenge":"YXV0aGVudGljYXRpb24gY2hhbGxlbmdl"`)
	assert.Equal(t, Base64URL("handle-alice"), auth.Response.UserHandle)
	assert.True(t, verifyAssertion(t, reg.Response.PublicKey, auth))
}

func TestConsentByConfirmation(t *testing.T) {
	t.Parallel()

	var (
		asked  int
		answer bool
		fail   error
	)

	consent := ConsentByConfirmation(func(context.Context, ConsentRequest) (bool, error) {
		asked++

		return answer, fail
	})

	// A confirmation is a test of user presence, nothing more.
	ctx := testContext(t)
	assert.False(t, SignCounterEnabled(ctx))

	got, err := consent(ctx, ConsentRequest{})
	require.NoError(t, err)
	assert.Equal(t, Consent{}, got)

	answer = true
	got, err = consent(ctx, ConsentRequest{})
	require.NoError(t, err)
	assert.Equal(t, Consent{Present: true}, got)
	assert.Equal(t, 2, asked)

	// A request that needs more is refused without bothering the user.
	got, err = consent(ctx, ConsentRequest{VerificationRequired: true})
	require.ErrorIs(t, err, ErrUnsuitable)
	assert.Contains(t, err.Error(), "passkey.user-verification")
	assert.Equal(t, Consent{}, got)
	assert.Equal(t, 2, asked)

	// Unless the user says that using the store implies a verification.
	cfg := config.NewInMemory()
	require.NoError(t, cfg.Set("", "passkey.signcount", "true"))
	require.NoError(t, cfg.Set("", "passkey.user-verification", "true"))
	ctx = cfg.WithConfig(context.Background())

	assert.True(t, SignCounterEnabled(ctx))

	for _, required := range []bool{false, true} {
		got, err = consent(ctx, ConsentRequest{VerificationRequired: required})
		require.NoError(t, err)
		assert.Equal(t, Consent{Present: true, Verified: true}, got)
	}
	assert.Equal(t, 4, asked)

	answer = false
	got, err = consent(ctx, ConsentRequest{VerificationRequired: true})
	require.NoError(t, err)
	assert.Equal(t, Consent{}, got)

	// A question that could not be asked is no consent, whatever it says.
	answer, fail = true, errors.New("no terminal")
	got, err = consent(ctx, ConsentRequest{})
	require.ErrorIs(t, err, fail)
	assert.Equal(t, Consent{}, got)
}

// deferringStore is a store that keeps working with the context of a write
// after the write returned, like a store that queues its commits.
type deferringStore struct {
	SecretStore

	writes []context.Context //nolint:containedctx
}

func (d *deferringStore) Set(ctx context.Context, name string, sec gopass.Byter) error {
	d.writes = append(d.writes, ctx)

	return d.SecretStore.Set(ctx, name, sec)
}

func TestStoreOutlivesCeremony(t *testing.T) {
	t.Parallel()

	cfg := config.NewInMemory()
	require.NoError(t, cfg.Set("", "passkey.signcount", "true"))
	ctx := cfg.WithConfig(context.Background())

	// The ceremony is over when the response is returned. What the store
	// still has to do with what was written, like committing it, is not:
	// the context of the write must not end with the ceremony.
	store := &deferringStore{SecretStore: apimock.New()}
	a := NewAuthenticator(NewStore(ctx, store), func(context.Context, ConsentRequest) (Consent, error) {
		return Consent{Present: true}, nil
	})

	_, _, err := a.Register(ctx, exampleOrigin, creationOptions("alice"))
	require.NoError(t, err)

	_, _, err = a.Authenticate(ctx, exampleOrigin, requestOptions(), nil)
	require.NoError(t, err)

	require.Len(t, store.writes, 2, "the passkey and its counter")

	for _, written := range store.writes {
		require.NoError(t, written.Err())

		_, bounded := written.Deadline()
		assert.False(t, bounded, "the timeout of the ceremony does not apply to the store")
	}
}

// cancellingStore is a store that takes so long to write that the ceremony
// ends in the meantime.
type cancellingStore struct {
	SecretStore

	cancel context.CancelFunc
}

func (c cancellingStore) Set(ctx context.Context, name string, sec gopass.Byter) error {
	defer c.cancel()

	return c.SecretStore.Set(ctx, name, sec)
}

func TestCeremonyEndsWhileStoring(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(testContext(t))
	api := apimock.New()

	a := NewAuthenticator(NewStore(ctx, cancellingStore{api, cancel}), func(context.Context, ConsentRequest) (Consent, error) {
		return Consent{Present: true}, nil
	})

	// The site is not answered anymore. The error says where the passkey
	// is, so that it can be removed.
	reg, entry, err := a.Register(ctx, exampleOrigin, creationOptions("alice"))
	require.ErrorIs(t, err, context.Canceled)
	assert.Equal(t, NotAllowedError, NameOf(err))
	assert.Contains(t, err.Error(), "passkeys/example.com/alice")
	assert.Nil(t, reg)
	assert.Nil(t, entry)
}

func TestCredentialChangesWhileAsking(t *testing.T) {
	t.Parallel()

	cfg := config.NewInMemory()
	require.NoError(t, cfg.Set("", "passkey.signcount", "true"))
	ctx := cfg.WithConfig(context.Background())

	api := apimock.New()
	store := NewStore(ctx, api)

	var swap func()

	a := NewAuthenticator(store, func(context.Context, ConsentRequest) (Consent, error) {
		if swap != nil {
			swap()
		}

		return Consent{Present: true}, nil
	})

	_, entry, err := a.Register(ctx, exampleOrigin, creationOptions("alice"))
	require.NoError(t, err)

	// While the user looks at the question the passkey is replaced by
	// another key with the same ID. That is not what the user agreed to
	// use: nothing is signed, and the counter of the other key is left
	// alone.
	other, err := New(NewOptions{RPID: "example.com", UserHandle: []byte("handle-alice"), UserName: "alice", SignCounter: true})
	require.NoError(t, err)
	other.ID = entry.Credential.ID

	swap = func() { writeCreds(ctx, t, api, entry.Name, other) }

	auth, _, err := a.Authenticate(ctx, exampleOrigin, requestOptions(), nil)
	require.ErrorIs(t, err, ErrNoCredentials)
	assert.Nil(t, auth)
	assert.Equal(t, uint32(1), readCreds(ctx, t, api, entry.Name)[0].SignCount)

	// The same goes for a passkey that is gone.
	swap = func() {
		_, err := store.Remove(ctx, entry.Name, entry.Credential.ID)
		require.NoError(t, err)
	}

	auth, _, err = a.Authenticate(ctx, exampleOrigin, requestOptions(), nil)
	require.ErrorIs(t, err, ErrNoCredentials)
	assert.Nil(t, auth)
}
