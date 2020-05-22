package keyring

import (
	"encoding/binary"
	"fmt"
	"time"

	"golang.org/x/crypto/sha3"

	"github.com/gopasspw/gopass/internal/backend/crypto/xc/xcpb"
)

// PublicKeyAlgorithm is a type of public key algorithm
type PublicKeyAlgorithm uint8

const (
	// PubKeyNaCl is a NaCl (Salt) based public key
	PubKeyNaCl PublicKeyAlgorithm = iota
)

// PublicKey is the public part of a keypair
type PublicKey struct {
	CreationTime time.Time
	PubKeyAlgo   PublicKeyAlgorithm
	PublicKey    [32]byte
	Identity     *xcpb.Identity
}

// Fingerprint calculates the unique ID of a public key
func (p PublicKey) Fingerprint() string {
	h := make([]byte, 20)
	d := sha3.NewShake256()
	_, _ = d.Write([]byte{0x42})
	_ = binary.Write(d, binary.LittleEndian, p.PubKeyAlgo)
	_, _ = d.Write(p.PublicKey[:])
	_, _ = d.Read(h)
	return fmt.Sprintf("%x", h)
}
