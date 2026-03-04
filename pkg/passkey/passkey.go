// Package passkey implements the support of Webauthn credentials for authentication.
// It is based on the W3C's "Web Authentication: An API for accessing Public Key Credentials"
// https://www.w3.org/TR/webauthn-2/
package passkey

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"fmt"
)

// CredentialFlags for the credential parameters.
// See: https://www.w3.org/TR/webauthn-2/#flags
type CredentialFlags struct {
	UserPresent     bool
	UserVerified    bool
	AttestationData bool
	ExtensionData   bool
}

// Credential is a structure to store information and key of public key credential.
type Credential struct {
	ID        string
	Rp        string
	UserName  string
	Algorithm string
	SecretKey *ecdsa.PrivateKey
	Counter   uint32
	Flags     CredentialFlags
}

// ClientData for signature.
// See: https://www.w3.org/TR/webauthn-1/#sec-client-data
type ClientData struct {
	Challenge string `json:"challenge"`
	Origin    string `json:"origin"`
	CredType  string `json:"type"`
}

// Response from a GetAssertion.
// See: https://www.w3.org/TR/webauthn-1/#authenticatorGetAssertion-return-values
type Response struct {
	AuthenticatorData []byte `json:"authdata"`
	ClientDataJSON    []byte `json:"client_data_json"`
	Signature         []byte `json:"signature"`
	Login             string `json:"login"`
}

func authDataFlags(options CredentialFlags) uint8 {
	flags := uint8(0)

	if options.ExtensionData {
		flags |= 0b10000000
	}

	if options.AttestationData {
		flags |= 0b01000000
	}

	if options.UserVerified {
		flags |= 0b00000100
	}

	if options.UserPresent {
		flags |= 0b00000001
	}

	return flags
}

// CreateCredential is an implementation of the authenticatorMakeCredential Operation.
// See: https://www.w3.org/TR/webauthn-2/#sctn-op-make-cred
func CreateCredential(rp string, user string, flags CredentialFlags) (*Credential, error) {
	rawID := make([]byte, 32)
	_, err := rand.Read(rawID)
	if err != nil {
		return nil, fmt.Errorf("error while generating random ID: %w", err)
	}
	privateKey, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)

	return &Credential{
		ID:        base64.RawURLEncoding.EncodeToString(rawID),
		Rp:        rp,
		UserName:  user,
		Algorithm: "ECDSA",
		SecretKey: privateKey,
		Counter:   0,
		Flags:     flags,
	}, nil
}

// GetAssertion is an implementation of the authenticatorGetAssertion Operation.
// See: https://www.w3.org/TR/webauthn-2/#authenticatorgetassertion
func (cred *Credential) GetAssertion(challenge string, origin string) (*Response, error) {
	credType := "webauthn.get"
	clientData := ClientData{
		CredType:  credType,
		Challenge: challenge,
		Origin:    origin,
	}
	clientDataJSON, err := json.Marshal(clientData)
	if err != nil {
		return nil, fmt.Errorf("error while reading client data: %w", err)
	}

	clientDataHash := sha256.Sum256(clientDataJSON)
	rpIDHash := sha256.Sum256([]byte(cred.Rp))
	flags := []byte{authDataFlags(cred.Flags)}

	// Signature counter is incremented according to https://www.w3.org/TR/webauthn-2/#signature-counter
	cred.Counter += 1
	signCount := make([]byte, 4)
	binary.BigEndian.PutUint32(signCount, cred.Counter)
	authData := append(rpIDHash[:], flags...)
	authData = append(authData[:], signCount[:]...)
	message := sha256.Sum256(append(authData[:], clientDataHash[:]...))
	signature, serr := ecdsa.SignASN1(rand.Reader, cred.SecretKey, message[:])
	if err != nil {
		return nil, serr
	}

	return &Response{
		AuthenticatorData: authData,
		ClientDataJSON:    clientDataJSON,
		Signature:         signature,
		Login:             cred.UserName,
	}, nil
}
