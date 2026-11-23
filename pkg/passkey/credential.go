package passkey

import (
	"bytes"
	"crypto"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"maps"
	"math"
	"reflect"
	"slices"
	"strings"
	"time"
	"unicode"
)

// schemaVersion is the version of the JSON representation of a credential.
const schemaVersion = 1

// Credential is a public key credential source: the private key of a passkey
// together with everything an authenticator needs to use it.
//
// See https://www.w3.org/TR/webauthn-3/#public-key-credential-source
type Credential struct {
	// ID is the credential ID. It is unique per registration.
	ID []byte
	// RPID is the identifier of the Relying Party the credential is scoped
	// to. The credential must not be used for any other Relying Party.
	RPID string
	// UserHandle identifies the user account at the Relying Party. It is
	// opaque, not meant to be displayed and returned with every assertion.
	UserHandle []byte
	// UserName is the name of the user account as known by the Relying Party.
	// It is only used to tell accounts apart.
	UserName string
	// UserDisplayName is a friendly name of the user account. It is only used
	// for display.
	UserDisplayName string
	// Key is the credential private key. It is an *ecdsa.PrivateKey on the
	// P-256 curve, an ed25519.PrivateKey or an *rsa.PrivateKey.
	Key crypto.Signer
	// SignCount is the signature counter. Credentials that are synced
	// between devices can not maintain a counter, so it is zero and stays
	// zero. A value greater than zero opts the credential in to counting.
	SignCount uint32
	// BackupEligible reports that the credential is a multi-device
	// credential. It must not change after the registration.
	BackupEligible bool
	// BackupState reports that the credential is currently backed up.
	BackupState bool
	// Created is when the credential was created. It may be zero.
	Created time.Time
	// Extensions is the Fido2Extensions dictionary of the FIDO Credential
	// Exchange Format, as JSON. It carries the data of WebAuthn extensions an
	// imported credential was used with. gopass does not implement any of
	// these extensions. The data is only kept so that it survives being
	// imported and exported again.
	Extensions json.RawMessage

	// unknown holds the members of the stored form that this version does
	// not know, e.g. ones a later version of gopass added. They are written
	// back as they are, so that they survive a change made by this version.
	unknown map[string]json.RawMessage
}

// NewOptions are the properties of a credential to generate.
type NewOptions struct {
	// RPID is the identifier of the Relying Party. It is required.
	RPID string
	// UserHandle is the user handle. It is required.
	UserHandle []byte
	// UserName is the name of the user account.
	UserName string
	// UserDisplayName is the friendly name of the user account.
	UserDisplayName string
	// Algorithm selects the type of the key pair. ES256 is used if unset.
	Algorithm Algorithm
	// SignCounter opts in to a signature counter that is incremented with
	// each assertion. Do not use this if the store is used on more than one
	// device at a time: a counter that lags behind makes a Relying Party
	// suspect a cloned authenticator.
	SignCounter bool
}

// New generates a credential with a fresh key pair and a random ID.
func New(opts NewOptions) (*Credential, error) {
	id, err := newCredentialID()
	if err != nil {
		return nil, err
	}

	return newCredential(opts, id)
}

func newCredentialID() ([]byte, error) {
	id := make([]byte, credentialIDLen)
	if _, err := rand.Read(id); err != nil {
		return nil, fmt.Errorf("failed to generate credential ID: %w", err)
	}

	return id, nil
}

func newCredential(opts NewOptions, id []byte) (*Credential, error) {
	alg := opts.Algorithm
	if alg == 0 {
		alg = ES256
	}

	key, err := generateKey(alg)
	if err != nil {
		return nil, err
	}

	cred := &Credential{
		ID:              id,
		RPID:            opts.RPID,
		UserHandle:      bytes.Clone(opts.UserHandle),
		UserName:        opts.UserName,
		UserDisplayName: opts.UserDisplayName,
		Key:             key,
		// gopass is a synced-passkey authenticator: the credential is as
		// available as the store it lives in.
		BackupEligible: true,
		BackupState:    true,
		Created:        time.Now().UTC().Truncate(time.Second),
	}

	if opts.SignCounter {
		// Any non-zero value marks the counter as being maintained. A
		// Relying Party only expects the value to grow from the one it is
		// given at the registration.
		cred.SignCount = 1
	}

	if err := cred.validate(); err != nil {
		return nil, err
	}

	return cred, nil
}

// Algorithm returns the signature algorithm of the credential.
func (c *Credential) Algorithm() Algorithm {
	alg, err := algorithmOf(c.Key)
	if err != nil {
		return 0
	}

	return alg
}

// PublicKey returns the public key as a DER encoded SubjectPublicKeyInfo.
func (c *Credential) PublicKey() ([]byte, error) {
	if _, err := algorithmOf(c.Key); err != nil {
		return nil, err
	}

	der, err := x509.MarshalPKIXPublicKey(c.Key.Public())
	if err != nil {
		return nil, fmt.Errorf("failed to encode public key: %w", err)
	}

	return der, nil
}

// Label returns a short, human readable description of the user account.
//
// The names of an account are chosen by the Relying Party. The label is safe
// to print: it has no control characters or line breaks and is limited in
// length.
func (c *Credential) Label() string {
	name, display := printable(c.UserName), printable(c.UserDisplayName)

	switch {
	case name != "" && display != "" && name != display:
		return fmt.Sprintf("%s (%s)", name, display)
	case name != "":
		return name
	case display != "":
		return display
	default:
		return Base64URL(c.ID).String()
	}
}

// maxLabelLen is the number of characters of a name that are shown.
const maxLabelLen = 64

// printable reduces a string to its printable characters, with single
// spaces between words, and shortens it.
func printable(s string) string {
	s = strings.Join(strings.FieldsFunc(s, func(r rune) bool {
		return unicode.IsSpace(r) || !unicode.IsPrint(r)
	}), " ")

	if r := []rune(s); len(r) > maxLabelLen {
		s = string(r[:maxLabelLen]) + "…"
	}

	return s
}

// String returns a description of the credential without its private key.
//
// String, GoString and MarshalJSON have value receivers, so that a copy of a
// credential is described the same way and never with the private key.
func (c Credential) String() string {
	return fmt.Sprintf("passkey(%s at %s, %s)", c.Label(), c.RPID, Base64URL(c.ID))
}

// GoString implements fmt.GoStringer. Like String, it leaves out the private
// key, which "%#v" would otherwise print.
func (c Credential) GoString() string {
	return fmt.Sprintf("passkey.Credential{ID: %q, RPID: %q, UserName: %q, Algorithm: %s}",
		Base64URL(c.ID), c.RPID, c.UserName, c.Algorithm())
}

// validate checks the invariants every usable credential must satisfy.
func (c *Credential) validate() error {
	if c == nil {
		return fmt.Errorf("%w: no credential", ErrInvalidCredential)
	}

	if len(c.ID) == 0 || len(c.ID) > maxCredentialIDLen {
		return fmt.Errorf("%w: credential ID must have 1 to %d bytes", ErrInvalidCredential, maxCredentialIDLen)
	}

	// The RP ID is hashed as it is and used as the name of a folder. Only
	// the canonical form of a domain is accepted for both.
	if domain, err := normalizeDomain(c.RPID); err != nil || domain != c.RPID {
		return fmt.Errorf("%w: RP ID %q is not a domain in its canonical form", ErrInvalidCredential, c.RPID)
	}

	// The user handle is returned with every assertion. Relying Parties
	// identify the account by it.
	if len(c.UserHandle) == 0 || len(c.UserHandle) > maxUserHandleLen {
		return fmt.Errorf("%w: user handle must have 1 to %d bytes", ErrInvalidCredential, maxUserHandleLen)
	}

	if c.Key == nil {
		return fmt.Errorf("%w: private key is missing", ErrInvalidCredential)
	}

	if _, err := algorithmOf(c.Key); err != nil {
		return err
	}

	return nil
}

// Attestation is the result of the authenticatorMakeCredential operation.
type Attestation struct {
	// AuthenticatorData is the authenticator data. It contains the attested
	// credential data: the AAGUID, the credential ID and the public key.
	AuthenticatorData []byte
	// Object is the attestation object: the authenticator data wrapped
	// together with an attestation statement.
	Object []byte
}

// Assertion is the result of the authenticatorGetAssertion operation.
type Assertion struct {
	// AuthenticatorData is the authenticator data.
	AuthenticatorData []byte
	// Signature is the assertion signature.
	Signature []byte
}

// Attest performs the signing part of the authenticatorMakeCredential
// operation for a freshly created credential.
//
// gopass has no attestation key, so the attestation object uses the "none"
// attestation statement format. Nothing is signed in that format, which is
// why this method does not take the hash of the client data.
//
// See https://www.w3.org/TR/webauthn-3/#sctn-op-make-cred
func (c *Credential) Attest(consent Consent) (*Attestation, error) {
	if err := c.validate(); err != nil {
		return nil, err
	}

	attested, err := c.attestedCredentialData(AAGUID)
	if err != nil {
		return nil, err
	}

	authData := authenticatorData(c.RPID, c.flags(consent), c.SignCount, attested)

	return &Attestation{
		AuthenticatorData: authData,
		Object:            attestationObject(authData),
	}, nil
}

// Assert performs the signing part of the authenticatorGetAssertion operation.
// The assertion signature covers the authenticator data and the hash of the
// serialized client data.
//
// The signature counter is used as it is. For a credential that maintains a
// counter, see [Credential.Counting], call [Credential.Advance] and persist
// the credential before every assertion.
//
// See https://www.w3.org/TR/webauthn-3/#sctn-op-get-assertion
func (c *Credential) Assert(clientDataHash []byte, consent Consent) (*Assertion, error) {
	if err := c.validate(); err != nil {
		return nil, err
	}

	if len(clientDataHash) != sha256.Size {
		return nil, fmt.Errorf("client data hash must have %d bytes, got %d", sha256.Size, len(clientDataHash))
	}

	authData := authenticatorData(c.RPID, c.flags(consent), c.SignCount, nil)

	// The signature covers the plain concatenation of both values.
	signature, err := sign(c.Key, append(bytes.Clone(authData), clientDataHash...))
	if err != nil {
		return nil, err
	}

	return &Assertion{
		AuthenticatorData: authData,
		Signature:         signature,
	}, nil
}

// Counting reports whether the credential maintains a signature counter.
func (c *Credential) Counting() bool {
	return c.SignCount > 0
}

// Advance increments the signature counter of a credential that maintains
// one. It does nothing for other credentials: their counter stays at zero.
//
// See https://www.w3.org/TR/webauthn-3/#sctn-sign-counter
func (c *Credential) Advance() error {
	if !c.Counting() {
		return nil
	}

	if c.SignCount == math.MaxUint32 {
		return fmt.Errorf("%w: signature counter is exhausted", ErrInvalidCredential)
	}
	c.SignCount++

	return nil
}

// credentialJSON is the serialized form of a credential. The names of the
// members match the passkey dictionary of the FIDO Credential Exchange Format
// where both have the same member.
//
// Members are only ever added to this version of the schema. A member that
// changes the meaning of others requires a new version, which older versions
// of gopass refuse to read and hence to change.
type credentialJSON struct {
	Version         int             `json:"v"`
	RPID            string          `json:"rpId"`
	UserName        string          `json:"userName"`
	UserDisplayName string          `json:"userDisplayName"`
	UserHandle      Base64URL       `json:"userHandle"`
	CredentialID    Base64URL       `json:"credentialId"`
	Key             Base64URL       `json:"key"`
	SignCount       uint32          `json:"signCount"`
	COSEAlg         int             `json:"coseAlg,omitempty"`
	BackupEligible  *bool           `json:"backupEligible,omitempty"`
	BackupState     *bool           `json:"backupState,omitempty"`
	CreationDate    string          `json:"creationDate,omitempty"`
	Extensions      json.RawMessage `json:"fido2Extensions,omitempty"`
}

// knownMembers are the names of the members of credentialJSON.
var knownMembers = func() []string {
	t := reflect.TypeFor[credentialJSON]()

	names := make([]string, 0, t.NumField())
	for i := range t.NumField() {
		name, _, _ := strings.Cut(t.Field(i).Tag.Get("json"), ",")
		names = append(names, name)
	}

	return names
}()

// isKnownMember reports whether a member is one of credentialJSON. Like the
// JSON decoder, it does not tell apart names that only differ in case.
func isKnownMember(name string) bool {
	return slices.ContainsFunc(knownMembers, func(k string) bool { return strings.EqualFold(k, name) })
}

// Encode returns the form in which a credential is stored: a single line of
// JSON that includes the private key as a PKCS#8 structure. Binary members
// are base64url encoded. Members a decoded credential had that this version
// does not know follow the known ones.
//
// HTML characters are not escaped. The result is stored, not rendered, and
// names like "Alice & Bob" should stay readable.
func (c *Credential) Encode() ([]byte, error) {
	if err := c.validate(); err != nil {
		return nil, err
	}

	key, err := marshalPrivateKey(c.Key)
	if err != nil {
		return nil, err
	}

	cj := credentialJSON{
		Version:         schemaVersion,
		RPID:            c.RPID,
		UserName:        c.UserName,
		UserDisplayName: c.UserDisplayName,
		UserHandle:      c.UserHandle,
		CredentialID:    c.ID,
		Key:             key,
		SignCount:       c.SignCount,
		COSEAlg:         int(c.Algorithm()),
		BackupEligible:  &c.BackupEligible,
		BackupState:     &c.BackupState,
		Extensions:      c.Extensions,
	}

	if !c.Created.IsZero() {
		cj.CreationDate = c.Created.UTC().Format(time.RFC3339)
	}

	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(cj); err != nil {
		return nil, fmt.Errorf("failed to encode credential: %w", err)
	}

	line := bytes.TrimRight(buf.Bytes(), "\n")
	if len(c.unknown) == 0 {
		return line, nil
	}

	return appendMembers(line, c.unknown)
}

// appendMembers adds members to an encoded JSON object, sorted by name. The
// values are compacted, so that the result stays on a single line.
func appendMembers(object []byte, members map[string]json.RawMessage) ([]byte, error) {
	var buf bytes.Buffer

	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)

	buf.Write(bytes.TrimSuffix(object, []byte("}")))

	for _, name := range slices.Sorted(maps.Keys(members)) {
		buf.WriteByte(',')

		if err := enc.Encode(name); err != nil {
			return nil, fmt.Errorf("failed to encode credential: %w", err)
		}
		buf.Truncate(buf.Len() - 1) // the encoder ends every value with a line break
		buf.WriteByte(':')

		if err := json.Compact(&buf, members[name]); err != nil {
			return nil, fmt.Errorf("failed to encode credential: member %q: %w", name, err)
		}
	}

	buf.WriteByte('}')

	return buf.Bytes(), nil
}

// Decode parses a credential encoded by [Credential.Encode].
func Decode(data []byte) (*Credential, error) {
	var cj credentialJSON
	if err := json.Unmarshal(data, &cj); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrInvalidCredential, err)
	}

	if cj.Version != schemaVersion {
		return nil, fmt.Errorf("%w: unsupported schema version %d", ErrInvalidCredential, cj.Version)
	}

	// The data is a JSON object, or the first Unmarshal had failed.
	var members map[string]json.RawMessage
	if err := json.Unmarshal(data, &members); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrInvalidCredential, err)
	}

	var unknown map[string]json.RawMessage

	for name, value := range members {
		if isKnownMember(name) {
			continue
		}

		if unknown == nil {
			unknown = make(map[string]json.RawMessage)
		}
		unknown[name] = value
	}

	key, err := parsePrivateKey(cj.Key)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrInvalidCredential, err)
	}

	cred := &Credential{
		ID:              cj.CredentialID,
		RPID:            cj.RPID,
		UserHandle:      cj.UserHandle,
		UserName:        cj.UserName,
		UserDisplayName: cj.UserDisplayName,
		Key:             key,
		SignCount:       cj.SignCount,
		// Credentials stored by gopass are synced with the store.
		BackupEligible: true,
		BackupState:    true,
		Extensions:     cj.Extensions,
		unknown:        unknown,
	}

	if cj.BackupEligible != nil {
		cred.BackupEligible = *cj.BackupEligible
	}

	if cj.BackupState != nil {
		cred.BackupState = *cj.BackupState
	}

	if cj.COSEAlg != 0 && Algorithm(cj.COSEAlg) != cred.Algorithm() {
		return nil, fmt.Errorf("%w: algorithm %s does not match the %s key", ErrInvalidCredential, Algorithm(cj.COSEAlg), cred.Algorithm())
	}

	if cj.CreationDate != "" {
		created, err := time.Parse(time.RFC3339, cj.CreationDate)
		if err != nil {
			return nil, fmt.Errorf("%w: invalid creation date: %w", ErrInvalidCredential, err)
		}
		cred.Created = created
	}

	if err := cred.validate(); err != nil {
		return nil, err
	}

	return cred, nil
}

// MarshalJSON returns a description of the credential that is safe to hand
// to a user interface. It does not contain the private key. Use
// [Credential.Encode] to serialize a credential for storage.
func (c Credential) MarshalJSON() ([]byte, error) {
	desc := struct {
		ID              Base64URL `json:"id"`
		RPID            string    `json:"rpId"`
		UserName        string    `json:"userName"`
		UserDisplayName string    `json:"userDisplayName"`
		Algorithm       Algorithm `json:"algorithm"`
		Created         string    `json:"created,omitempty"`
	}{
		ID:              c.ID,
		RPID:            c.RPID,
		UserName:        c.UserName,
		UserDisplayName: c.UserDisplayName,
		Algorithm:       c.Algorithm(),
	}

	if !c.Created.IsZero() {
		desc.Created = c.Created.UTC().Format(time.RFC3339)
	}

	return json.Marshal(desc)
}
