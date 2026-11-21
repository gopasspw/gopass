package passkey

import (
	"encoding/json"
	"testing"

	"github.com/gopasspw/gopass/pkg/gopass/apimock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRegisterDefaults(t *testing.T) {
	t.Parallel()

	ctx := testContext(t)
	a, _ := newAuthenticator(ctx)

	// Without an RP ID the effective domain of the origin is used. Without
	// parameters the defaults of a client apply.
	opts := creationOptions("alice")
	opts.RP.ID = nil
	opts.PubKeyCredParams = nil

	reg, entry, err := a.Register(ctx, exampleOrigin, opts)
	require.NoError(t, err)
	assert.Equal(t, "passkeys/login.example.com/alice", entry.Name)
	assert.Equal(t, "login.example.com", entry.Credential.RPID)
	assert.Equal(t, ES256, reg.Response.PublicKeyAlgorithm)
	assert.Empty(t, reg.ClientExtensionResults)

	// The passkey is not usable for the parent domain.
	_, _, err = a.Authenticate(ctx, exampleOrigin, requestOptions(), nil)
	require.ErrorIs(t, err, ErrNoCredentials)

	opts2 := requestOptions()
	opts2.RPID = nil
	_, _, err = a.Authenticate(ctx, exampleOrigin, opts2, nil)
	require.NoError(t, err)
}

func TestRegisterRejected(t *testing.T) {
	t.Parallel()

	for name, tc := range map[string]struct {
		origin Origin
		mutate func(*CreationOptions)
		want   ErrorName
	}{
		"no user handle": {exampleOrigin, func(o *CreationOptions) { o.User.ID = nil }, TypeError},
		"long user handle": {
			exampleOrigin, func(o *CreationOptions) { o.User.ID = make([]byte, 65) }, TypeError,
		},
		"no challenge": {exampleOrigin, func(o *CreationOptions) { o.Challenge = nil }, TypeError},
		"other site":   {Origin{Origin: "https://evil.example"}, func(*CreationOptions) {}, SecurityError},
		"insecure":     {Origin{Origin: "http://login.example.com"}, func(*CreationOptions) {}, SecurityError},
		"opaque":       {Origin{Origin: "null"}, func(*CreationOptions) {}, NotAllowedError},
		"public suffix": {
			exampleOrigin, func(o *CreationOptions) { o.RP.ID = ptr("com") }, SecurityError,
		},
		"no top origin": {
			Origin{Origin: "https://login.example.com", CrossOrigin: true}, func(*CreationOptions) {}, NotAllowedError,
		},
		"unsupported algorithm": {
			exampleOrigin,
			func(o *CreationOptions) { o.PubKeyCredParams = []CredentialParameter{{Type: "public-key", Alg: -36}} },
			NotSupportedError,
		},
		"unsupported type": {
			exampleOrigin,
			func(o *CreationOptions) { o.PubKeyCredParams = []CredentialParameter{{Type: "password", Alg: ES256}} },
			NotSupportedError,
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			ctx := testContext(t)
			a, rec := newAuthenticator(ctx)

			opts := creationOptions("alice")
			tc.mutate(opts)

			_, _, err := a.Register(ctx, tc.origin, opts)
			require.Error(t, err)
			assert.Equal(t, tc.want, NameOf(err), err)
			// What this authenticator can not do, another one may.
			assert.Equal(t, tc.want == NotSupportedError, IsFallback(err))

			// The user is not bothered and nothing is stored.
			assert.Empty(t, rec.requests)

			entries, err := a.store.List(ctx, "")
			require.NoError(t, err)
			assert.Empty(t, entries)
		})
	}

	ctx := testContext(t)
	a, _ := newAuthenticator(ctx)

	_, _, err := a.Register(ctx, exampleOrigin, nil)
	assert.Equal(t, TypeError, NameOf(err))
}

func TestRegisterRoamingAuthenticator(t *testing.T) {
	t.Parallel()

	ctx := testContext(t)
	a, rec := newAuthenticator(ctx)

	// A site that asks for a security key does not get a synced passkey.
	opts := creationOptions("alice")
	opts.AuthenticatorSelection = &AuthenticatorSelection{AuthenticatorAttachment: "cross-platform"}

	_, _, err := a.Register(ctx, exampleOrigin, opts)
	require.ErrorIs(t, err, ErrUnsuitable)
	assert.Equal(t, NotAllowedError, NameOf(err))
	assert.True(t, IsFallback(err))
	assert.Empty(t, rec.requests)

	opts.AuthenticatorSelection.AuthenticatorAttachment = "platform"
	opts.AuthenticatorSelection.ResidentKey = "discouraged"
	_, _, err = a.Register(ctx, exampleOrigin, opts)
	require.NoError(t, err)
}

func TestExcludeCredentials(t *testing.T) {
	t.Parallel()

	ctx := testContext(t)
	a, rec := newAuthenticator(ctx)

	reg, _, err := a.Register(ctx, exampleOrigin, creationOptions("alice"))
	require.NoError(t, err)

	opts := creationOptions("alice")
	opts.ExcludeCredentials = []CredentialDescriptor{
		{Type: "public-key", ID: []byte("unknown credential")},
		{Type: "public-key", ID: reg.RawID},
	}

	// The site learns that the passkey exists only if the user agrees.
	rec.requests = nil
	_, _, err = a.Register(ctx, exampleOrigin, opts)
	require.Error(t, err)
	assert.Equal(t, InvalidStateError, NameOf(err))
	require.Equal(t, []Operation{OperationDisclose}, rec.operations())
	assert.Equal(t, "passkeys/example.com/alice", rec.requests[0].Entry)
	assert.Equal(t, "alice", rec.requests[0].UserName)

	rec.requests = nil
	rec.consent = Consent{}
	_, _, err = a.Register(ctx, exampleOrigin, opts)
	require.ErrorIs(t, err, ErrDeclined)
	assert.Equal(t, NotAllowedError, NameOf(err))
	assert.Equal(t, []Operation{OperationDisclose}, rec.operations())

	// Nothing was replaced.
	entries, err := a.store.List(ctx, "example.com")
	require.NoError(t, err)
	assert.Equal(t, [][]byte{reg.RawID}, ids(entries))

	// Credentials of an unknown type and unknown credentials are ignored.
	rec.requests = nil
	rec.consent = Consent{Present: true}
	opts.ExcludeCredentials = []CredentialDescriptor{
		{Type: "unknown", ID: reg.RawID},
		{Type: "public-key", ID: []byte("unknown credential")},
	}
	opts.User.ID = []byte("another handle")
	opts.User.Name = "bob"
	_, _, err = a.Register(ctx, exampleOrigin, opts)
	require.NoError(t, err)
	assert.Equal(t, []Operation{OperationRegister}, rec.operations())
}

func TestRegisterKeepsCredentialOfAccount(t *testing.T) {
	t.Parallel()

	ctx := testContext(t)
	a, rec := newAuthenticator(ctx)

	first, _, err := a.Register(ctx, exampleOrigin, creationOptions("alice"))
	require.NoError(t, err)
	assert.Empty(t, rec.requests[0].Entry)

	// The user is told that the account has a passkey already.
	second, entry, err := a.Register(ctx, exampleOrigin, creationOptions("alice"))
	require.NoError(t, err)
	assert.Equal(t, "passkeys/example.com/alice", entry.Name)
	assert.NotEqual(t, first.ID, second.ID)
	assert.Equal(t, OperationRegister, rec.requests[1].Operation)
	assert.Equal(t, "passkeys/example.com/alice", rec.requests[1].Entry)

	// The passkey of another account is not mentioned.
	_, _, err = a.Register(ctx, exampleOrigin, creationOptions("bob"))
	require.NoError(t, err)
	assert.Empty(t, rec.requests[2].Entry)

	// The first passkey is kept. Whether the site accepted the second one
	// is not known here, and a passkey that works must not be lost.
	entries, err := a.store.List(ctx, "example.com")
	require.NoError(t, err)
	assert.Equal(t, [][]byte{first.RawID, second.RawID}, ids(entries)[:2])
	assert.Len(t, entries, 3)

	for _, reg := range []*RegistrationResponse{first, second} {
		auth, _, err := a.Authenticate(ctx, exampleOrigin, requestOptions(), reg.RawID)
		require.NoError(t, err)
		assert.True(t, verifyAssertion(t, reg.Response.PublicKey, auth))
	}
}

func TestRegisterFailsIfNotStored(t *testing.T) {
	t.Parallel()

	ctx := testContext(t)
	rec := &consentRecorder{consent: Consent{Present: true}}
	a := NewAuthenticator(NewStore(ctx, readOnlyStore{apimock.New()}), rec.ask)

	// The site must not learn about a passkey that was lost right away.
	reg, entry, err := a.Register(ctx, exampleOrigin, creationOptions("alice"))
	require.ErrorIs(t, err, errReadOnly)
	assert.Equal(t, UnknownError, NameOf(err))
	assert.Nil(t, reg)
	assert.Nil(t, entry)
}

func TestCredProps(t *testing.T) {
	t.Parallel()

	ctx := testContext(t)
	a, _ := newAuthenticator(ctx)

	// The extension is only processed if it is requested with true.
	for in, want := range map[string]bool{"true": true, "false": false, "null": false, `"true"`: false, "1": false} {
		opts := creationOptions("alice")
		opts.Extensions = map[string]json.RawMessage{"credProps": json.RawMessage(in)}

		reg, _, err := a.Register(ctx, exampleOrigin, opts)
		require.NoError(t, err)
		assert.Equal(t, want, len(reg.ClientExtensionResults) == 1, in)
	}
}
