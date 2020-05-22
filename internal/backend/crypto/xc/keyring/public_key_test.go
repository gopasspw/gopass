package keyring

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFingerprint(t *testing.T) {
	pk := PublicKey{}

	assert.Equal(t, "7471b2b8801eb22f1657f0003cab1d0adf9dadd8", pk.Fingerprint())
}
