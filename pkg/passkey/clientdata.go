package passkey

import (
	"crypto/sha256"
	"fmt"
	"strings"
)

const (
	// clientDataCreate is the type of the client data of a registration.
	clientDataCreate = "webauthn.create"
	// clientDataGet is the type of the client data of an authentication.
	clientDataGet = "webauthn.get"
)

// ClientData represents the contextual bindings of the Relying Party and the
// client. The authenticator signs its hash.
//
// See https://www.w3.org/TR/webauthn-3/#dictionary-client-data
type ClientData struct {
	// Type is "webauthn.create" or "webauthn.get".
	Type string
	// Challenge is the challenge provided by the Relying Party.
	Challenge []byte
	// Origin is the serialized origin of the requester.
	Origin string
	// CrossOrigin is true if the request was not made from a context that is
	// same-origin with its ancestors.
	CrossOrigin bool
	// TopOrigin is the serialized top-level origin. It is only set if
	// CrossOrigin is true.
	TopOrigin string
}

// JSON returns the JSON-compatible serialization of client data.
//
// The serialization is not the output of a generic JSON encoder. The order of
// the members and the way strings are escaped are fixed, so that a Relying
// Party can verify the result without a JSON parser.
//
// See https://www.w3.org/TR/webauthn-3/#clientdatajson-serialization
func (c ClientData) JSON() []byte {
	var sb strings.Builder

	sb.WriteString(`{"type":`)
	ccdToString(&sb, c.Type)
	sb.WriteString(`,"challenge":`)
	ccdToString(&sb, Base64URL(c.Challenge).String())
	sb.WriteString(`,"origin":`)
	ccdToString(&sb, c.Origin)
	sb.WriteString(`,"crossOrigin":`)

	if c.CrossOrigin {
		sb.WriteString("true")
	} else {
		sb.WriteString("false")
	}

	if c.CrossOrigin && c.TopOrigin != "" {
		sb.WriteString(`,"topOrigin":`)
		ccdToString(&sb, c.TopOrigin)
	}

	sb.WriteString("}")

	return []byte(sb.String())
}

// Hash returns the hash of the serialized client data.
func (c ClientData) Hash() []byte {
	h := sha256.Sum256(c.JSON())

	return h[:]
}

// ccdToString implements the CCDToString function of the serialization.
func ccdToString(sb *strings.Builder, s string) {
	sb.WriteByte('"')

	for _, r := range s {
		switch {
		case r == '"':
			sb.WriteString(`\"`)
		case r == '\\':
			sb.WriteString(`\\`)
		case r < 0x20:
			fmt.Fprintf(sb, `\u%04x`, r)
		default:
			sb.WriteRune(r)
		}
	}

	sb.WriteByte('"')
}
