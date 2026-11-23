package passkey

import (
	"bytes"
	"encoding/binary"
	"slices"
)

// This file contains the small subset of a CBOR (RFC 8949) encoder that is
// needed to produce attestation objects and COSE keys. Everything it emits
// is in the CTAP2 canonical CBOR encoding form:
//
//   - integers and lengths use the shortest possible encoding,
//   - only definite length items are used,
//   - map keys are sorted by major type, then by length, then bytewise.
//
// See https://fidoalliance.org/specs/fido-v2.2-ps-20250714/fido-client-to-authenticator-protocol-v2.2-ps-20250714.html#ctap2-canonical-cbor-encoding-form

const (
	cborMajorUint  = 0
	cborMajorNeg   = 1
	cborMajorBytes = 2
	cborMajorText  = 3
	cborMajorMap   = 5
)

// cborHead appends the initial byte(s) of a data item: the major type and its
// argument, encoded as short as possible.
func cborHead(buf []byte, major byte, arg uint64) []byte {
	major <<= 5

	switch {
	case arg < 24:
		return append(buf, major|byte(arg))
	case arg <= 0xff:
		return append(buf, major|24, byte(arg))
	case arg <= 0xffff:
		return binary.BigEndian.AppendUint16(append(buf, major|25), uint16(arg))
	case arg <= 0xffffffff:
		return binary.BigEndian.AppendUint32(append(buf, major|26), uint32(arg))
	default:
		return binary.BigEndian.AppendUint64(append(buf, major|27), arg)
	}
}

// cborInt encodes a signed integer.
func cborInt(v int64) []byte {
	if v < 0 {
		// a negative integer is encoded as -1 - n
		return cborHead(nil, cborMajorNeg, uint64(-1-v))
	}

	return cborHead(nil, cborMajorUint, uint64(v))
}

// cborBytes encodes a byte string.
func cborBytes(v []byte) []byte {
	return append(cborHead(nil, cborMajorBytes, uint64(len(v))), v...)
}

// cborText encodes a UTF-8 text string.
func cborText(v string) []byte {
	return append(cborHead(nil, cborMajorText, uint64(len(v))), v...)
}

// cborPair is an already encoded key and value of a map.
type cborPair struct {
	key   []byte
	value []byte
}

// cborMap encodes a map. The entries are sorted into canonical order, so
// they can be passed in any order.
func cborMap(pairs ...cborPair) []byte {
	slices.SortFunc(pairs, func(a, b cborPair) int {
		// The major type lives in the three most significant bits of the
		// first byte. Within a major type a shorter encoding always denotes
		// a smaller integer or a shorter string, so comparing the length of
		// the encoded keys implements the "shorter sorts earlier" rule.
		if ma, mb := a.key[0]>>5, b.key[0]>>5; ma != mb {
			return int(ma) - int(mb)
		}
		if len(a.key) != len(b.key) {
			return len(a.key) - len(b.key)
		}

		return bytes.Compare(a.key, b.key)
	})

	buf := cborHead(nil, cborMajorMap, uint64(len(pairs)))
	for _, p := range pairs {
		buf = append(buf, p.key...)
		buf = append(buf, p.value...)
	}

	return buf
}
