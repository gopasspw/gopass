package passkey

import (
	"encoding/json"
)

// The types in this file mirror the JSON types of the Web Authentication API.
// A client gets the options from the Relying Party, e.g. by serializing the
// argument of navigator.credentials.create(), and can construct a
// PublicKeyCredential from the responses.

// RelyingParty is a PublicKeyCredentialRpEntity.
type RelyingParty struct {
	// ID is the RP ID. The effective domain of the origin is used if it is
	// nil. An empty RP ID is not the same as none: it is invalid.
	ID *string `json:"id,omitempty"`
	// Name is a friendly name of the Relying Party, for display only.
	Name string `json:"name"`
}

// User is a PublicKeyCredentialUserEntityJSON.
type User struct {
	// ID is the user handle.
	ID Base64URL `json:"id"`
	// Name is the name of the user account.
	Name string `json:"name"`
	// DisplayName is the friendly name of the user account.
	DisplayName string `json:"displayName"`
}

// CredentialParameter is a PublicKeyCredentialParameters.
type CredentialParameter struct {
	// Type is the type of credential, "public-key".
	Type string `json:"type"`
	// Alg is the signature algorithm.
	Alg Algorithm `json:"alg"`
}

// CredentialDescriptor is a PublicKeyCredentialDescriptorJSON.
type CredentialDescriptor struct {
	// Type is the type of credential, "public-key".
	Type string `json:"type"`
	// ID is the credential ID.
	ID Base64URL `json:"id"`
	// Transports are hints as to how the authenticator can be reached.
	Transports []string `json:"transports,omitempty"`
}

// AuthenticatorSelection is an AuthenticatorSelectionCriteria.
type AuthenticatorSelection struct {
	// AuthenticatorAttachment restricts the kind of authenticator: "platform"
	// or "cross-platform".
	AuthenticatorAttachment string `json:"authenticatorAttachment,omitempty"`
	// ResidentKey is "discouraged", "preferred" or "required".
	ResidentKey string `json:"residentKey,omitempty"`
	// RequireResidentKey is the legacy form of ResidentKey "required".
	RequireResidentKey bool `json:"requireResidentKey,omitempty"`
	// UserVerification is "required", "preferred" or "discouraged".
	UserVerification string `json:"userVerification,omitempty"`
}

// CreationOptions is a PublicKeyCredentialCreationOptionsJSON: the options of
// a registration ceremony.
//
// See https://www.w3.org/TR/webauthn-3/#dictdef-publickeycredentialcreationoptionsjson
type CreationOptions struct {
	RP                     RelyingParty               `json:"rp"`
	User                   User                       `json:"user"`
	Challenge              Base64URL                  `json:"challenge"`
	PubKeyCredParams       []CredentialParameter      `json:"pubKeyCredParams"`
	Timeout                uint64                     `json:"timeout,omitempty"`
	ExcludeCredentials     []CredentialDescriptor     `json:"excludeCredentials,omitempty"`
	AuthenticatorSelection *AuthenticatorSelection    `json:"authenticatorSelection,omitempty"`
	Hints                  []string                   `json:"hints,omitempty"`
	Attestation            string                     `json:"attestation,omitempty"`
	AttestationFormats     []string                   `json:"attestationFormats,omitempty"`
	Extensions             map[string]json.RawMessage `json:"extensions,omitempty"`
}

// RequestOptions is a PublicKeyCredentialRequestOptionsJSON: the options of
// an authentication ceremony.
//
// See https://www.w3.org/TR/webauthn-3/#dictdef-publickeycredentialrequestoptionsjson
type RequestOptions struct {
	Challenge        Base64URL                  `json:"challenge"`
	Timeout          uint64                     `json:"timeout,omitempty"`
	RPID             *string                    `json:"rpId,omitempty"`
	AllowCredentials []CredentialDescriptor     `json:"allowCredentials,omitempty"`
	UserVerification string                     `json:"userVerification,omitempty"`
	Hints            []string                   `json:"hints,omitempty"`
	Extensions       map[string]json.RawMessage `json:"extensions,omitempty"`
}

// AttestationResponse is an AuthenticatorAttestationResponseJSON.
type AttestationResponse struct {
	// ClientDataJSON is the serialized client data.
	ClientDataJSON Base64URL `json:"clientDataJSON"`
	// AuthenticatorData is the authenticator data. It is also part of the
	// attestation object.
	AuthenticatorData Base64URL `json:"authenticatorData"`
	// Transports are the transports the authenticator is reachable over.
	Transports []string `json:"transports"`
	// PublicKey is the DER encoded SubjectPublicKeyInfo of the credential.
	PublicKey Base64URL `json:"publicKey,omitempty"`
	// PublicKeyAlgorithm is the signature algorithm of the credential.
	PublicKeyAlgorithm Algorithm `json:"publicKeyAlgorithm"`
	// AttestationObject is the attestation object.
	AttestationObject Base64URL `json:"attestationObject"`
}

// RegistrationResponse is a RegistrationResponseJSON: the result of a
// registration ceremony.
//
// See https://www.w3.org/TR/webauthn-3/#dictdef-registrationresponsejson
type RegistrationResponse struct {
	// ID is the base64url encoded credential ID.
	ID string `json:"id"`
	// RawID is the credential ID.
	RawID Base64URL `json:"rawId"`
	// Response is the response of the authenticator.
	Response AttestationResponse `json:"response"`
	// AuthenticatorAttachment is the attachment modality, "platform".
	AuthenticatorAttachment string `json:"authenticatorAttachment,omitempty"`
	// ClientExtensionResults are the outputs of the client extensions.
	ClientExtensionResults map[string]any `json:"clientExtensionResults"`
	// Type is the type of credential, "public-key".
	Type string `json:"type"`
}

// AssertionResponse is an AuthenticatorAssertionResponseJSON.
type AssertionResponse struct {
	// ClientDataJSON is the serialized client data.
	ClientDataJSON Base64URL `json:"clientDataJSON"`
	// AuthenticatorData is the authenticator data.
	AuthenticatorData Base64URL `json:"authenticatorData"`
	// Signature is the assertion signature.
	Signature Base64URL `json:"signature"`
	// UserHandle is the user handle of the credential.
	UserHandle Base64URL `json:"userHandle,omitempty"`
}

// AuthenticationResponse is an AuthenticationResponseJSON: the result of an
// authentication ceremony.
//
// See https://www.w3.org/TR/webauthn-3/#dictdef-authenticationresponsejson
type AuthenticationResponse struct {
	// ID is the base64url encoded credential ID.
	ID string `json:"id"`
	// RawID is the credential ID.
	RawID Base64URL `json:"rawId"`
	// Response is the response of the authenticator.
	Response AssertionResponse `json:"response"`
	// AuthenticatorAttachment is the attachment modality, "platform".
	AuthenticatorAttachment string `json:"authenticatorAttachment,omitempty"`
	// ClientExtensionResults are the outputs of the client extensions.
	ClientExtensionResults map[string]any `json:"clientExtensionResults"`
	// Type is the type of credential, "public-key".
	Type string `json:"type"`
}

const (
	// attachmentPlatform is the attachment modality of gopass: it is part of
	// the client device, not a roaming authenticator.
	attachmentPlatform = "platform"
	// attachmentCrossPlatform is the attachment modality of roaming
	// authenticators like security keys.
	attachmentCrossPlatform = "cross-platform"
	// transportInternal is the transport of platform authenticators.
	transportInternal = "internal"

	requirementRequired = "required"

	extCredProps = "credProps"
)
