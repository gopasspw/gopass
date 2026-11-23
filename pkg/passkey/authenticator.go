package passkey

import (
	"bytes"
	"context"
	"errors"
	"slices"
	"time"

	"github.com/gopasspw/gopass/pkg/debug"
)

// Authenticator performs WebAuthn ceremonies with the credentials of a
// [Store]. It takes on the role of the authenticator and those parts of the
// client that do not depend on a browser: validating the options, binding
// the RP ID to the origin and assembling the client data.
//
// Everything that needs a browser is left to the caller, which must
//
//   - determine the [Origin] of a request itself and never take it from the
//     web page,
//   - only forward requests the browser would allow: from a secure context
//     and, for embedded documents, permitted by the Permissions Policy
//     ("publickey-credentials-create" and "publickey-credentials-get") and
//     with the user activation a cross-origin registration requires,
//   - not reveal more than [NameOf] an error to the Relying Party, and not
//     even that sooner for a missing passkey than for a declined ceremony:
//     both the details and the timing can tell a site which accounts exist.
type Authenticator struct {
	store   *Store
	consent ConsentFunc
}

// NewAuthenticator returns an Authenticator using the credentials in store.
// Every ceremony is subject to the approval of consent.
func NewAuthenticator(store *Store, consent ConsentFunc) *Authenticator {
	return &Authenticator{
		store:   store,
		consent: consent,
	}
}

const (
	// The timeout of a ceremony is kept within the range WebAuthn recommends.
	minCeremonyTimeout = 5 * time.Minute
	maxCeremonyTimeout = 10 * time.Minute
)

// withTimeout bounds the lifetime of a ceremony.
//
// See https://www.w3.org/TR/webauthn-3/#sctn-timeout-recommended-range
func withTimeout(ctx context.Context, millis uint64) (context.Context, context.CancelFunc) {
	timeout := minCeremonyTimeout
	if requested := time.Duration(min(millis, uint64(maxCeremonyTimeout/time.Millisecond))) * time.Millisecond; requested > timeout {
		timeout = requested
	}

	return context.WithTimeout(ctx, timeout)
}

// checkContext fails if the ceremony timed out or was cancelled.
func checkContext(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return errorf(NotAllowedError, "ceremony was cancelled: %w", err)
	}

	return nil
}

// clientData validates the origin of a request and returns the client data
// for it together with the RP ID the request is scoped to.
func clientData(typ string, origin Origin, rpID *string, challenge []byte) (ClientData, string, error) {
	if len(challenge) == 0 {
		return ClientData{}, "", errorf(TypeError, "challenge is missing")
	}

	resolved, err := ResolveRPID(origin.Origin, rpID)
	if err != nil {
		return ClientData{}, "", err
	}

	// ResolveRPID has validated the origin already
	serialized, _, _ := parseOrigin(origin.Origin)

	cd := ClientData{
		Type: typ,
		// The challenge is what gets signed. It must not change once the
		// user has seen what the request is about.
		Challenge: bytes.Clone(challenge),
		Origin:    serialized,
	}

	if origin.CrossOrigin {
		top, _, err := parseOrigin(origin.TopOrigin)
		if err != nil {
			return ClientData{}, "", errorf(NotAllowedError, "invalid top-level origin: %w", err)
		}

		cd.CrossOrigin = true
		cd.TopOrigin = top
	}

	return cd, resolved, nil
}

// chooseAlgorithm returns the first algorithm of the list a Relying Party
// accepts that is supported.
func chooseAlgorithm(params []CredentialParameter) (Algorithm, error) {
	if len(params) == 0 {
		// the defaults mandated for a client
		params = []CredentialParameter{
			{Type: credentialType, Alg: ES256},
			{Type: credentialType, Alg: RS256},
		}
	}

	for _, p := range params {
		if p.Type == credentialType && p.Alg.Supported() {
			return p.Alg, nil
		}
	}

	// Another authenticator may well support what this one does not.
	return 0, errorf(NotSupportedError, "%w: none of the requested algorithms is supported", ErrUnsuitable)
}

// askConsent collects the authorization gesture and enforces the requirements
// of the Relying Party on it.
func (a *Authenticator) askConsent(ctx context.Context, req ConsentRequest) (Consent, error) {
	if a.consent == nil {
		return Consent{}, errorf(NotAllowedError, "%w: no way to ask the user", ErrDeclined)
	}

	// Nobody is asked about a ceremony that is over already.
	if err := checkContext(ctx); err != nil {
		return Consent{}, err
	}

	consent, err := a.consent(ctx, req)
	if err != nil {
		if errors.Is(err, ErrUnsuitable) || errors.Is(err, ErrDeclined) {
			return Consent{}, newError(NotAllowedError, err)
		}

		return Consent{}, errorf(NotAllowedError, "failed to ask for confirmation: %w", err)
	}

	// There is no ceremony without a test of user presence.
	if !consent.Present {
		return Consent{}, newError(NotAllowedError, ErrDeclined)
	}

	// An answer that came too late is no consent.
	if err := checkContext(ctx); err != nil {
		return Consent{}, err
	}

	if req.VerificationRequired && !consent.Verified {
		return Consent{}, errorf(NotAllowedError, "%w: the site requires user verification", ErrUnsuitable)
	}

	return consent, nil
}

// Register performs a registration ceremony: it creates a new credential for
// the user account described by the options, stores it and returns what a
// client hands to the Relying Party, together with the stored credential.
//
// A credential that is already stored for the account is kept. The ceremony
// can still fail at the Relying Party, and a passkey that works must not be
// lost to one that does not.
//
// See https://www.w3.org/TR/webauthn-3/#sctn-createCredential
func (a *Authenticator) Register(ctx context.Context, origin Origin, opts *CreationOptions) (*RegistrationResponse, *Entry, error) {
	if opts == nil {
		return nil, nil, errorf(TypeError, "options are missing")
	}

	if len(opts.User.ID) == 0 || len(opts.User.ID) > maxUserHandleLen {
		return nil, nil, errorf(TypeError, "user handle must have 1 to %d bytes", maxUserHandleLen)
	}

	cd, rpID, err := clientData(clientDataCreate, origin, opts.RP.ID, opts.Challenge)
	if err != nil {
		return nil, nil, err
	}

	alg, err := chooseAlgorithm(opts.PubKeyCredParams)
	if err != nil {
		return nil, nil, err
	}

	sel := AuthenticatorSelection{}
	if opts.AuthenticatorSelection != nil {
		sel = *opts.AuthenticatorSelection
	}

	if sel.AuthenticatorAttachment == attachmentCrossPlatform {
		return nil, nil, errorf(NotAllowedError, "%w: the site asked for a roaming authenticator", ErrUnsuitable)
	}

	// Everything the user agrees to is fixed before the question is asked.
	user := User{
		ID:          bytes.Clone(opts.User.ID),
		Name:        opts.User.Name,
		DisplayName: opts.User.DisplayName,
	}
	credProps := string(opts.Extensions[extCredProps]) == "true"

	// The timeout bounds the ceremony. It does not bound what the store
	// does with a credential that was written: a store may defer work, like
	// a commit, that must not be cancelled when the ceremony is over.
	cctx, cancel := withTimeout(ctx, opts.Timeout)
	defer cancel()

	req := ConsentRequest{
		RPID:                 rpID,
		RPName:               opts.RP.Name,
		Origin:               cd.Origin,
		TopOrigin:            cd.TopOrigin,
		UserName:             user.Name,
		UserDisplayName:      user.DisplayName,
		VerificationRequired: sel.UserVerification == requirementRequired,
	}

	// The passkeys of the Relying Party are read once for everything the
	// ceremony needs to know about them. Excluded credentials can be stored
	// anywhere, the account of the user only below the prefix.
	names, err := a.store.Names(cctx, rpID)
	if err != nil {
		return nil, nil, errorf(UnknownError, "failed to look up existing passkeys: %w", err)
	}

	if len(opts.ExcludeCredentials) == 0 {
		names = slices.DeleteFunc(names, func(n string) bool { return !a.store.isAccountPath(cctx, n) })
	}

	entries, err := a.store.load(cctx, names)
	if err != nil {
		// What can not be read is treated like an unknown credential.
		debug.Log("failed to read some passkeys of %s: %s", rpID, err)
	}

	if err := a.checkExcluded(cctx, req, entries, opts.ExcludeCredentials); err != nil {
		return nil, nil, err
	}

	// The user should know if the account has a passkey already. The new
	// one is added to the same secret.
	existing, _ := a.store.accountIn(cctx, entries, rpID, user.ID)
	req.Entry = existing

	req.Operation = OperationRegister

	consent, err := a.askConsent(cctx, req)
	if err != nil {
		return nil, nil, err
	}

	cred, err := New(NewOptions{
		RPID:            rpID,
		UserHandle:      user.ID,
		UserName:        user.Name,
		UserDisplayName: user.DisplayName,
		Algorithm:       alg,
		SignCounter:     SignCounterEnabled(ctx),
	})
	if err != nil {
		return nil, nil, newError(UnknownError, err)
	}

	att, err := cred.Attest(consent)
	if err != nil {
		return nil, nil, newError(UnknownError, err)
	}

	pub, err := cred.PublicKey()
	if err != nil {
		return nil, nil, newError(UnknownError, err)
	}

	// The credential must be stored before the Relying Party learns about
	// it. A passkey that is registered but lost locks the user out.
	if err := checkContext(cctx); err != nil {
		return nil, nil, err
	}

	name, warning := a.store.putTo(ctx, cred, existing)
	if warning != nil && !errors.Is(warning, ErrNotSynced) {
		return nil, nil, errorf(UnknownError, "failed to store the passkey: %w", warning)
	}

	// A ceremony that ended while the credential was written is not
	// answered. The passkey is stored nevertheless, and can be removed.
	if err := checkContext(cctx); err != nil {
		return nil, nil, errorf(NotAllowedError, "the passkey was stored in %s, but %w", name, err)
	}

	results := map[string]any{}
	if credProps {
		// Every credential is stored client-side and hence discoverable.
		results[extCredProps] = map[string]any{"rk": true}
	}

	return &RegistrationResponse{
		ID:    Base64URL(cred.ID).String(),
		RawID: cred.ID,
		Response: AttestationResponse{
			ClientDataJSON:     cd.JSON(),
			AuthenticatorData:  att.AuthenticatorData,
			Transports:         []string{transportInternal},
			PublicKey:          pub,
			PublicKeyAlgorithm: cred.Algorithm(),
			AttestationObject:  att.Object,
		},
		AuthenticatorAttachment: attachmentPlatform,
		ClientExtensionResults:  results,
		Type:                    credentialType,
	}, &Entry{Name: name, Credential: cred, Warning: warning}, nil
}

// checkExcluded fails if one of the credentials a Relying Party does not
// want to be registered again is among the stored entries. Whether one is
// stored is not disclosed without the consent of the user.
//
// See https://www.w3.org/TR/webauthn-3/#sctn-op-make-cred step 3.
func (a *Authenticator) checkExcluded(ctx context.Context, req ConsentRequest, entries []Entry, excluded []CredentialDescriptor) error {
	if len(excluded) == 0 {
		return nil
	}

	matches := filterEntries(entries, excluded)
	if len(matches) == 0 {
		return nil
	}

	req.Operation = OperationDisclose
	req.Entry = matches[0].Name
	req.UserName = matches[0].Credential.UserName
	req.UserDisplayName = matches[0].Credential.UserDisplayName
	// only a test of user presence is needed to disclose a credential
	req.VerificationRequired = false

	if _, err := a.askConsent(ctx, req); err != nil {
		return err
	}

	return errorf(InvalidStateError, "a passkey for this account is already stored in %s", req.Entry)
}

// filterEntries returns the entries whose credential is described by one of
// the descriptors.
func filterEntries(entries []Entry, descriptors []CredentialDescriptor) []Entry {
	return slices.DeleteFunc(slices.Clone(entries), func(e Entry) bool {
		return !slices.ContainsFunc(descriptors, func(d CredentialDescriptor) bool {
			return d.Type == credentialType && bytes.Equal(d.ID, e.Credential.ID)
		})
	})
}

// Candidates returns the credentials that can satisfy an authentication
// request: those of the Relying Party, restricted to the allowed credentials
// if the Relying Party named any. It is meant to let the user pick an account
// and to pass the choice to [Authenticator.AuthenticateWith], which does not
// read the store again.
//
// The result is for the eyes of the user only. Handing any of it to the
// Relying Party before the user consented reveals which accounts exist.
func (a *Authenticator) Candidates(ctx context.Context, origin Origin, opts *RequestOptions) ([]Entry, error) {
	if opts == nil {
		return nil, errorf(TypeError, "options are missing")
	}

	_, rpID, err := clientData(clientDataGet, origin, opts.RPID, opts.Challenge)
	if err != nil {
		return nil, err
	}

	ctx, cancel := withTimeout(ctx, opts.Timeout)
	defer cancel()

	return a.candidates(ctx, rpID, opts.AllowCredentials)
}

func (a *Authenticator) candidates(ctx context.Context, rpID string, allowed []CredentialDescriptor) ([]Entry, error) {
	entries, lerr := a.store.List(ctx, rpID)
	if lerr != nil {
		debug.Log("failed to read some passkeys of %s: %s", rpID, lerr)
	}

	if len(allowed) > 0 {
		entries = filterEntries(entries, allowed)
	}

	// The same credential can be stored more than once, e.g. as a copy in a
	// shared store. It is offered once. The entries below the prefix come
	// first, so what others can write never takes precedence over what the
	// user stored.
	seen := make([][]byte, 0, len(entries))
	entries = slices.DeleteFunc(entries, func(e Entry) bool {
		if slices.ContainsFunc(seen, func(id []byte) bool { return bytes.Equal(id, e.Credential.ID) }) {
			return true
		}
		seen = append(seen, e.Credential.ID)

		return false
	})

	if len(entries) == 0 {
		if lerr != nil {
			return nil, errorf(NotAllowedError, "%w: %w", ErrNoCredentials, lerr)
		}

		return nil, errorf(NotAllowedError, "%w for %s", ErrNoCredentials, rpID)
	}

	return entries, nil
}

// Authenticate performs an authentication ceremony: it signs the challenge
// of the Relying Party with a stored credential and returns what a client
// hands to the Relying Party, together with the credential that was used.
//
// credentialID selects the credential to use among the [Authenticator.Candidates].
// It may be nil if there is only one. Otherwise the ceremony fails with
// [ErrSelectionRequired].
//
// The candidates are read from the store again. A caller that has them
// already, e.g. to let the user pick one, can use
// [Authenticator.AuthenticateWith] instead.
//
// See https://www.w3.org/TR/webauthn-3/#sctn-getAssertion
func (a *Authenticator) Authenticate(ctx context.Context, origin Origin, opts *RequestOptions, credentialID []byte) (*AuthenticationResponse, *Entry, error) {
	if opts == nil {
		return nil, nil, errorf(TypeError, "options are missing")
	}

	cd, rpID, err := clientData(clientDataGet, origin, opts.RPID, opts.Challenge)
	if err != nil {
		return nil, nil, err
	}

	// See Register for why the store is not given this context to write.
	cctx, cancel := withTimeout(ctx, opts.Timeout)
	defer cancel()

	entries, err := a.candidates(cctx, rpID, opts.AllowCredentials)
	if err != nil {
		return nil, nil, err
	}

	entry, err := selectEntry(entries, credentialID)
	if err != nil {
		return nil, nil, err
	}

	return a.authenticate(ctx, cctx, cd, rpID, opts, entry)
}

// AuthenticateWith performs an authentication ceremony like
// [Authenticator.Authenticate], with a credential the caller has read
// already: one of the [Authenticator.Candidates], or one of a secret the user
// named, see [Store.Load]. Nothing is read from the store, unless the
// credential maintains a signature counter.
//
// The credential must be able to satisfy the request: it must be bound to
// the RP ID the request is scoped to and, if the Relying Party lists the
// credentials it allows, be one of them.
func (a *Authenticator) AuthenticateWith(ctx context.Context, origin Origin, opts *RequestOptions, entry Entry) (*AuthenticationResponse, *Entry, error) {
	if opts == nil {
		return nil, nil, errorf(TypeError, "options are missing")
	}

	cd, rpID, err := clientData(clientDataGet, origin, opts.RPID, opts.Challenge)
	if err != nil {
		return nil, nil, err
	}

	if entry.Credential == nil {
		return nil, nil, errorf(TypeError, "credential is missing")
	}

	// A credential is never used for another Relying Party, whatever the
	// name of the secret holding it says.
	if entry.Credential.RPID != rpID {
		return nil, nil, errorf(NotAllowedError, "%w for %s in %s", ErrNoCredentials, rpID, entry.Name)
	}

	if len(opts.AllowCredentials) > 0 && len(filterEntries([]Entry{entry}, opts.AllowCredentials)) == 0 {
		return nil, nil, errorf(NotAllowedError, "%w: the site does not allow the passkey in %s", ErrNoCredentials, entry.Name)
	}

	cctx, cancel := withTimeout(ctx, opts.Timeout)
	defer cancel()

	return a.authenticate(ctx, cctx, cd, rpID, opts, entry)
}

// authenticate asks for consent and signs the assertion with the credential
// of entry. cctx bounds the ceremony, ctx what is written to the store.
func (a *Authenticator) authenticate(ctx, cctx context.Context, cd ClientData, rpID string, opts *RequestOptions, entry Entry) (*AuthenticationResponse, *Entry, error) {
	consent, err := a.askConsent(cctx, ConsentRequest{
		Operation:            OperationAuthenticate,
		RPID:                 rpID,
		Origin:               cd.Origin,
		TopOrigin:            cd.TopOrigin,
		UserName:             entry.Credential.UserName,
		UserDisplayName:      entry.Credential.UserDisplayName,
		Entry:                entry.Name,
		VerificationRequired: opts.UserVerification == requirementRequired,
	})
	if err != nil {
		return nil, nil, err
	}

	var warning error

	cred := entry.Credential
	if cred.Counting() {
		// The counter is incremented in the store, on what is stored there
		// by now, and before anything is signed: an assertion must not be
		// released with a counter that was used before or was not persisted.
		// Either would make the Relying Party suspect a cloned authenticator.
		// The credential in the store has to be the one the user was asked
		// about.
		cred, warning = a.store.Advance(ctx, entry.Name, cred)
		if warning != nil && !errors.Is(warning, ErrNotSynced) {
			return nil, nil, errorf(UnknownError, "failed to persist the signature counter: %w", warning)
		}
	}

	// Defense in depth: the store only returns credentials of the RP ID,
	// and the one that is used is the one the user was asked about.
	if cred.RPID != rpID || !bytes.Equal(cred.ID, entry.Credential.ID) {
		return nil, nil, errorf(NotAllowedError, "%w for %s", ErrNoCredentials, rpID)
	}

	assertion, err := cred.Assert(cd.Hash(), consent)
	if err != nil {
		return nil, nil, newError(UnknownError, err)
	}

	if err := checkContext(cctx); err != nil {
		return nil, nil, err
	}

	return &AuthenticationResponse{
		ID:    Base64URL(cred.ID).String(),
		RawID: cred.ID,
		Response: AssertionResponse{
			ClientDataJSON:    cd.JSON(),
			AuthenticatorData: assertion.AuthenticatorData,
			Signature:         assertion.Signature,
			UserHandle:        cred.UserHandle,
		},
		AuthenticatorAttachment: attachmentPlatform,
		ClientExtensionResults:  map[string]any{},
		Type:                    credentialType,
	}, &Entry{Name: entry.Name, Credential: cred, Warning: warning}, nil
}

// selectEntry picks the credential to use for an assertion.
func selectEntry(entries []Entry, credentialID []byte) (Entry, error) {
	if len(credentialID) == 0 {
		if len(entries) == 1 {
			return entries[0], nil
		}

		return Entry{}, newError(NotAllowedError, ErrSelectionRequired)
	}

	i := slices.IndexFunc(entries, func(e Entry) bool { return bytes.Equal(e.Credential.ID, credentialID) })
	if i < 0 {
		return Entry{}, newError(NotAllowedError, ErrNoCredentials)
	}

	return entries[i], nil
}

// IsFallback reports whether a failed ceremony should be retried with
// another authenticator, e.g. the one built into the browser: gopass has no
// passkey for the request or is not the kind of authenticator asked for.
func IsFallback(err error) bool {
	return errors.Is(err, ErrNoCredentials) || errors.Is(err, ErrUnsuitable)
}
