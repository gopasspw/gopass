package passkey

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"testing"

	"github.com/gopasspw/gopass/internal/config"
	"github.com/gopasspw/gopass/pkg/gopass/apimock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// consentRecorder is a ConsentFunc that records what it was asked.
type consentRecorder struct {
	consent  Consent
	err      error
	requests []ConsentRequest
}

func (c *consentRecorder) ask(_ context.Context, req ConsentRequest) (Consent, error) {
	c.requests = append(c.requests, req)

	return c.consent, c.err
}

// operations returns the operations consent was asked for.
func (c *consentRecorder) operations() []Operation {
	ops := make([]Operation, 0, len(c.requests))
	for _, r := range c.requests {
		ops = append(ops, r.Operation)
	}

	return ops
}

// newAuthenticator returns an authenticator with an empty store whose user
// approves and is verified.
func newAuthenticator(ctx context.Context) (*Authenticator, *consentRecorder) {
	rec := &consentRecorder{consent: Consent{Present: true, Verified: true}}

	return NewAuthenticator(NewStore(ctx, apimock.New()), rec.ask), rec
}

var exampleOrigin = Origin{Origin: "https://login.example.com"}

// ptr returns a pointer to s.
func ptr(s string) *string {
	return &s
}

// creationOptions returns the options of a typical registration.
func creationOptions(user string) *CreationOptions {
	return &CreationOptions{
		RP:        RelyingParty{ID: ptr("example.com"), Name: "Example Inc."},
		User:      User{ID: []byte("handle-" + user), Name: user, DisplayName: "The " + user},
		Challenge: []byte("registration challenge"),
		PubKeyCredParams: []CredentialParameter{
			{Type: "public-key", Alg: Algorithm(-36)}, // not supported
			{Type: "public-key", Alg: ES256},
			{Type: "public-key", Alg: RS256},
		},
	}
}

// requestOptions returns the options of a typical authentication.
func requestOptions() *RequestOptions {
	return &RequestOptions{
		Challenge: []byte("authentication challenge"),
		RPID:      ptr("example.com"),
	}
}

func TestRegisterAndAuthenticate(t *testing.T) {
	t.Parallel()

	for _, alg := range SupportedAlgorithms() {
		t.Run(alg.String(), func(t *testing.T) {
			t.Parallel()

			ctx := testContext(t)
			a, rec := newAuthenticator(ctx)

			opts := creationOptions("alice")
			opts.PubKeyCredParams = []CredentialParameter{{Type: "public-key", Alg: alg}}
			opts.Extensions = map[string]json.RawMessage{"credProps": json.RawMessage("true"), "unknown": json.RawMessage("{}")}

			reg, entry, err := a.Register(ctx, exampleOrigin, opts)
			require.NoError(t, err)
			assert.Equal(t, "passkeys/example.com/alice", entry.Name)
			assert.Equal(t, []Operation{OperationRegister}, rec.operations())
			assert.Equal(t, ConsentRequest{
				Operation:       OperationRegister,
				RPID:            "example.com",
				RPName:          "Example Inc.",
				Origin:          "https://login.example.com",
				UserName:        "alice",
				UserDisplayName: "The alice",
				Destination:     "passkeys/example.com/alice",
			}, rec.requests[0])

			// what the Relying Party gets
			assert.Equal(t, "public-key", reg.Type)
			assert.Equal(t, "platform", reg.AuthenticatorAttachment)
			assert.Equal(t, reg.RawID.String(), reg.ID)
			assert.Len(t, reg.RawID, 32)
			assert.Equal(t, []string{"internal"}, reg.Response.Transports)
			assert.Equal(t, alg, reg.Response.PublicKeyAlgorithm)
			assert.Equal(t, map[string]any{"credProps": map[string]any{"rk": true}}, reg.ClientExtensionResults)
			equalBytes(t,
				`{"type":"webauthn.create","challenge":"cmVnaXN0cmF0aW9uIGNoYWxsZW5nZQ","origin":"https://login.example.com","crossOrigin":false}`,
				reg.Response.ClientDataJSON)

			decoded, rest := cborDecode(t, reg.Response.AttestationObject)
			require.Empty(t, rest)

			obj, ok := decoded.(map[any]any)
			require.True(t, ok)
			assert.Equal(t, "none", obj["fmt"])
			assert.Equal(t, []byte(reg.Response.AuthenticatorData), obj["authData"])

			ad := parseAuthData(t, reg.Response.AuthenticatorData)
			rpIDHash := sha256.Sum256([]byte("example.com"))
			assert.Equal(t, rpIDHash[:], ad.rpIDHash)
			assert.Equal(t, FlagUserPresent|FlagUserVerified|FlagBackupEligible|FlagBackupState|FlagAttestedCredentialData, ad.flags)
			assert.Equal(t, []byte(reg.RawID), ad.credentialID)
			assert.Equal(t, AAGUID[:], ad.aaguid)
			assert.Zero(t, ad.signCount)

			// the response serializes to a RegistrationResponseJSON
			buf, err := json.Marshal(reg)
			require.NoError(t, err)

			var asJSON map[string]any
			require.NoError(t, json.Unmarshal(buf, &asJSON))
			assert.Equal(t, reg.ID, asJSON["rawId"])
			assert.Contains(t, asJSON["response"], "attestationObject")
			assert.Contains(t, asJSON["response"], "publicKey")

			// and the passkey can be used to sign in
			auth, used, err := a.Authenticate(ctx, exampleOrigin, requestOptions(), nil)
			require.NoError(t, err)
			assert.Equal(t, entry.Name, used.Name)
			assert.Equal(t, []Operation{OperationRegister, OperationAuthenticate}, rec.operations())
			assert.Equal(t, ConsentRequest{
				Operation:       OperationAuthenticate,
				RPID:            "example.com",
				Origin:          "https://login.example.com",
				UserName:        "alice",
				UserDisplayName: "The alice",
				Entry:           "passkeys/example.com/alice",
			}, rec.requests[1])

			assert.Equal(t, reg.ID, auth.ID)
			assert.Equal(t, reg.RawID, auth.RawID)
			assert.Equal(t, "public-key", auth.Type)
			assert.Equal(t, "platform", auth.AuthenticatorAttachment)
			assert.Equal(t, Base64URL("handle-alice"), auth.Response.UserHandle)
			assert.Empty(t, auth.ClientExtensionResults)
			equalBytes(t,
				`{"type":"webauthn.get","challenge":"YXV0aGVudGljYXRpb24gY2hhbGxlbmdl","origin":"https://login.example.com","crossOrigin":false}`,
				auth.Response.ClientDataJSON)

			ad = parseAuthData(t, auth.Response.AuthenticatorData)
			assert.Equal(t, rpIDHash[:], ad.rpIDHash)
			assert.Equal(t, FlagUserPresent|FlagUserVerified|FlagBackupEligible|FlagBackupState, ad.flags)
			assert.Zero(t, ad.signCount)

			assert.True(t, verifyAssertion(t, reg.Response.PublicKey, auth))

			buf, err = json.Marshal(auth)
			require.NoError(t, err)
			require.NoError(t, json.Unmarshal(buf, &asJSON))
			assert.Contains(t, asJSON, "clientExtensionResults")
			assert.Contains(t, asJSON["response"], "userHandle")
		})
	}
}

func TestCeremoniesWithUnsyncedStore(t *testing.T) {
	t.Parallel()

	cfg := config.NewInMemory()
	require.NoError(t, cfg.Set("", "passkey.signcount", "true"))
	ctx := cfg.WithConfig(context.Background())

	// The store saves what it is given, but can not reach its remote. That
	// is no reason to fail a ceremony: the passkey is stored. A registration
	// that failed for this reason would be repeated, leaving a passkey
	// behind each time.
	api := apimock.New()
	rec := &consentRecorder{consent: Consent{Present: true}}
	a := NewAuthenticator(NewStore(ctx, unsyncedStore{api}), rec.ask)

	reg, entry, err := a.Register(ctx, exampleOrigin, creationOptions("alice"))
	require.NoError(t, err)
	require.ErrorIs(t, entry.Warning, ErrNotSynced)
	require.ErrorIs(t, entry.Warning, errPush)
	assert.Equal(t, "passkeys/example.com/alice", entry.Name)

	creds := readCreds(ctx, t, api, entry.Name)
	require.Len(t, creds, 1)
	assert.Equal(t, []byte(reg.RawID), creds[0].ID)

	// The counter that is signed is the one that is stored.
	auth, used, err := a.Authenticate(ctx, exampleOrigin, requestOptions(), nil)
	require.NoError(t, err)
	require.ErrorIs(t, used.Warning, ErrNotSynced)
	assert.True(t, verifyAssertion(t, reg.Response.PublicKey, auth))

	count := parseAuthData(t, auth.Response.AuthenticatorData).signCount
	assert.Equal(t, uint32(2), count)
	assert.Equal(t, count, readCreds(ctx, t, api, entry.Name)[0].SignCount)

	// With a store that is in sync there is nothing to warn about.
	synced := NewAuthenticator(NewStore(ctx, api), rec.ask)

	_, used, err = synced.Authenticate(ctx, exampleOrigin, requestOptions(), nil)
	require.NoError(t, err)
	require.NoError(t, used.Warning)
}

func TestErrors(t *testing.T) {
	t.Parallel()

	assert.Equal(t, UnknownError, NameOf(errors.New("boom")))
	assert.Equal(t, UnknownError, NameOf(nil))

	err := error(newError(SecurityError, ErrUnsuitable))
	assert.Equal(t, SecurityError, NameOf(err))
	assert.Equal(t, "SecurityError: "+ErrUnsuitable.Error(), err.Error())
	assert.Equal(t, "NotAllowedError", (&Error{Name: NotAllowedError}).Error())
	require.ErrorIs(t, err, ErrUnsuitable)

	assert.Equal(t, "register", OperationRegister.String())
	assert.Equal(t, "authenticate", OperationAuthenticate.String())
	assert.Equal(t, "disclose", OperationDisclose.String())
	assert.Equal(t, "Operation(7)", Operation(7).String())
}

func TestCeremonyTimeout(t *testing.T) {
	t.Parallel()

	for millis, want := range map[uint64]float64{
		0:                    300,
		1000:                 300,
		300000:               300,
		420000:               420,
		600000:               600,
		3600000:              600,
		18446744073709551615: 600,
	} {
		ctx, cancel := withTimeout(context.Background(), millis)
		deadline, ok := ctx.Deadline()
		cancel()

		require.True(t, ok)
		assert.InDelta(t, want, timeUntil(deadline), 5, "%d", millis)
	}
}
