// Package passkey implements a WebAuthn software authenticator backed by a
// gopass password store.
//
// gopass acts as an authenticator in the synced-passkey model: the credential
// private key lives in an encrypted secret and is shared, through the store's
// version control, with every device that has access to the store.
//
// The package is layered. Each layer is usable on its own:
//
//   - [Credential] is a public key credential source. It can be generated,
//     serialized (see [Credential.Encode]) and it implements the two
//     authenticator operations [Credential.Attest] and [Credential.Assert].
//   - [Store] persists credentials in a password store, one "passkey" line
//     per credential, below "passkeys/<rp id>/<account>".
//   - [Authenticator] combines both with the processing a WebAuthn client
//     performs. It consumes the JSON option dictionaries a Relying Party hands
//     to navigator.credentials.create() and navigator.credentials.get() and
//     produces the JSON a client returns from PublicKeyCredential.toJSON().
//
// The implementation follows the W3C Recommendation "Web Authentication: An
// API for accessing Public Key Credentials - Level 3"
// (https://www.w3.org/TR/webauthn-3/), the FIDO Client to Authenticator
// Protocol 2.2 for the CBOR encoding rules and the FIDO Credential Exchange
// Format 1.0 for import and export.
//
// # Stability
//
// This package has no stability guarantee yet. The exported symbols may still
// change while browser integrations are being built on top of it.
package passkey

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
)

// AAGUID is the Authenticator Attestation GUID identifying gopass as the
// authenticator that created a credential. It is the same for every gopass
// installation.
//
// See https://www.w3.org/TR/webauthn-3/#aaguid
var AAGUID = [16]byte{
	0x87, 0x76, 0x3d, 0x3e, 0xd8, 0x5f, 0x41, 0xd2,
	0x8c, 0x6f, 0xbd, 0x10, 0x70, 0x1e, 0xcc, 0x95,
}

const (
	// credentialType is the only PublicKeyCredentialType defined by WebAuthn.
	credentialType = "public-key"
	// credentialIDLen is the length of the credential IDs gopass generates.
	credentialIDLen = 32
	// maxCredentialIDLen is the maximum length of a credential ID.
	maxCredentialIDLen = 1023
	// maxUserHandleLen is the maximum length of a user handle.
	maxUserHandleLen = 64
)

// Base64URL is a byte string that is encoded as unpadded base64url in JSON.
// WebAuthn uses this encoding for every binary value in its JSON types.
//
// See https://www.w3.org/TR/webauthn-3/#base64url-encoding
type Base64URL []byte

// String returns the unpadded base64url encoding.
func (b Base64URL) String() string {
	return base64.RawURLEncoding.EncodeToString(b)
}

// MarshalJSON implements json.Marshaler.
func (b Base64URL) MarshalJSON() ([]byte, error) {
	return json.Marshal(b.String())
}

// UnmarshalJSON implements json.Unmarshaler. A JSON null is decoded to a nil
// slice. Padding is tolerated since not all producers omit it.
func (b *Base64URL) UnmarshalJSON(data []byte) error {
	var s *string
	if err := json.Unmarshal(data, &s); err != nil {
		return fmt.Errorf("base64url value must be a string: %w", err)
	}

	if s == nil {
		*b = nil

		return nil
	}

	buf, err := decodeBase64URL(*s)
	if err != nil {
		return err
	}
	*b = buf

	return nil
}

// ParseBase64URL decodes a base64url string. Padding is tolerated.
func ParseBase64URL(s string) (Base64URL, error) {
	return decodeBase64URL(s)
}

// decodeBase64URL decodes base64url with or without padding. Anything else
// a lenient decoder would skip, like line breaks, is an error.
func decodeBase64URL(s string) ([]byte, error) {
	if strings.ContainsAny(s, "\r\n") {
		return nil, fmt.Errorf("invalid base64url value: contains line breaks")
	}

	// Padding completes a group of four with one or two characters.
	if len(s)%4 == 0 {
		s = strings.TrimSuffix(strings.TrimSuffix(s, "="), "=")
	}

	buf, err := base64.RawURLEncoding.Strict().DecodeString(s)
	if err != nil {
		return nil, fmt.Errorf("invalid base64url value: %w", err)
	}

	return buf, nil
}
