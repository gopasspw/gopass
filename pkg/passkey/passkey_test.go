package passkey_test

import (
	"crypto/ecdsa"
	"crypto/sha256"
	"encoding/base64"
	"testing"

	"github.com/gopasspw/gopass/pkg/passkey"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var flags passkey.CredentialFlags = passkey.CredentialFlags{
	UserPresent:     true,
	UserVerified:    true,
	AttestationData: false,
	ExtensionData:   false,
}

func TestCreate(t *testing.T) {
	cred, err := passkey.CreateCredential("test.com", "user", flags)
	require.NoError(t, err)
	assert.Equal(t, "test.com", cred.Rp)
	assert.Equal(t, uint32(0), cred.Counter)
}

func TestGetAssertion(t *testing.T) {
	cred, err := passkey.CreateCredential("test.com", "user", flags)
	require.NoError(t, err)

	rsp, err := cred.GetAssertion(base64.RawURLEncoding.EncodeToString([]byte("test_challenge")), "test")
	require.NoError(t, err)

	// Verify signature
	clientDataHash := sha256.Sum256(rsp.ClientDataJSON)

	authData := rsp.AuthenticatorData
	require.NoError(t, err)

	message := sha256.Sum256(append(authData[:], clientDataHash[:]...))
	assert.True(t, ecdsa.VerifyASN1(&cred.SecretKey.PublicKey, message[:], rsp.Signature))
}
