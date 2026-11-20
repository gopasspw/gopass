package passkey

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"
)

// Flags are the flags of the authenticator data.
//
// See https://www.w3.org/TR/webauthn-3/#authdata-flags
type Flags byte

const (
	// FlagUserPresent is set if a test of user presence was performed.
	FlagUserPresent Flags = 1 << 0
	// FlagUserVerified is set if the user was verified.
	FlagUserVerified Flags = 1 << 2
	// FlagBackupEligible is set if the credential is a multi-device
	// credential. It never changes after the registration.
	FlagBackupEligible Flags = 1 << 3
	// FlagBackupState is set if the credential is currently backed up.
	FlagBackupState Flags = 1 << 4
	// FlagAttestedCredentialData is set if the authenticator data contains
	// attested credential data. That is the case for registrations only.
	FlagAttestedCredentialData Flags = 1 << 6
	// FlagExtensionData is set if the authenticator data has extensions.
	FlagExtensionData Flags = 1 << 7
)

// Has reports whether all bits of f are set.
func (fl Flags) Has(f Flags) bool {
	return fl&f == f
}

// flags returns the authenticator data flags for a ceremony with the
// credential. The flags attested about the user are taken from the consent,
// the backup flags from the credential.
func (c *Credential) flags(consent Consent) Flags {
	var fl Flags

	if consent.Present {
		fl |= FlagUserPresent
	}

	if consent.Verified {
		fl |= FlagUserVerified
	}

	if c.BackupEligible {
		fl |= FlagBackupEligible
		// A credential that is not backup eligible can not be backed up.
		if c.BackupState {
			fl |= FlagBackupState
		}
	}

	return fl
}

// authenticatorData encodes the authenticator data structure. The attested
// credential data is appended if, and only if, attested is not nil.
//
// See https://www.w3.org/TR/webauthn-3/#sctn-authenticator-data
func authenticatorData(rpID string, flags Flags, signCount uint32, attested []byte) []byte {
	rpIDHash := sha256.Sum256([]byte(rpID))

	flags &^= FlagAttestedCredentialData | FlagExtensionData
	if attested != nil {
		flags |= FlagAttestedCredentialData
	}

	buf := make([]byte, 0, len(rpIDHash)+1+4+len(attested))
	buf = append(buf, rpIDHash[:]...)
	buf = append(buf, byte(flags))
	buf = binary.BigEndian.AppendUint32(buf, signCount)
	buf = append(buf, attested...)

	return buf
}

// attestedCredentialData encodes the attested credential data of the
// credential.
//
// See https://www.w3.org/TR/webauthn-3/#sctn-attested-credential-data
func (c *Credential) attestedCredentialData(aaguid [16]byte) ([]byte, error) {
	if len(c.ID) == 0 || len(c.ID) > maxCredentialIDLen {
		return nil, fmt.Errorf("%w: credential ID must have 1 to %d bytes", ErrInvalidCredential, maxCredentialIDLen)
	}

	pub, err := coseKey(c.Key)
	if err != nil {
		return nil, err
	}

	buf := make([]byte, 0, len(aaguid)+2+len(c.ID)+len(pub))
	buf = append(buf, aaguid[:]...)
	buf = binary.BigEndian.AppendUint16(buf, uint16(len(c.ID)))
	buf = append(buf, c.ID...)
	buf = append(buf, pub...)

	return buf, nil
}

// attestationObject encodes an attestation object without an attestation
// statement, i.e. in the "none" attestation statement format.
//
// See https://www.w3.org/TR/webauthn-3/#sctn-none-attestation
func attestationObject(authData []byte) []byte {
	return cborMap(
		cborPair{cborText("fmt"), cborText("none")},
		cborPair{cborText("attStmt"), cborMap()},
		cborPair{cborText("authData"), cborBytes(authData)},
	)
}
