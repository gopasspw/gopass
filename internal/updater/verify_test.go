package updater

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// To update see README.md.
var testSignature = []byte(`-----BEGIN PGP SIGNATURE-----

iQGzBAABCAAdFiEEofrA/VCsqN4ejERxfqcKNWfm6NIFAmenWNkACgkQfqcKNWfm
6NLpzAwAzjLpYduk+X6JvMvpEh/KmJnLZfV0wA8YitmNxu3Ap0B1pVo/q6lyMHW8
rGCEgY4kpgJ0MEdD4mYYNhDzpPSv00NymrtlTfiel42ksMNBjH1/EVOFy+qFVEsj
OS7pCVHlGhPOYjjs5hMLMGvLkxXiuT0rKi2GluglGfTiYkbmsJxfj/alvb9rVQJ1
eQV6DbGDiIPdDTqeGUZBv3xX6YMAtuzly/WpXohCIpVK6ckKmqpufwavaVmBuk8F
U4+S/2OzuKGMySlYk8YHwaRDHeQAcpgtu6B+6h6B8rOpkI2OH6tihATo2vjQw8vd
093guOpDwqHV8AxBksCyYEyFwVOA71De0Sm75EUQUqRskUtAEQCJcYzNacBmZxvt
qBMi1E2U1mbv5doG+Y7zV36M33pQ/OtsHoIrXuJrqtldgQ5fXdIyD+qrn2/P39Do
mkMsGZfH7H5TjXEuwoDXNGXEo5D7dnCTNLq6gw8fjTTVfTMC3xKwqbyajxo9SnWi
jVGtKFWL
=QgDD
-----END PGP SIGNATURE-----

`)

var testData = []byte(`gopass-sign-test
`)

func TestGPGVerify(t *testing.T) {
	t.Parallel()

	ok, err := gpgVerify(testData, testSignature)
	require.NoError(t, err)
	assert.True(t, ok)
}

// TestGPGVerifyIn6Months tests that the signature is still valid in 6 months.
// This is supposed to act as a canary so we don't forget to roll the key
// before it expires. See README.md for details.
func TestGPGVerifyIn6Months(t *testing.T) {
	// Can not run parallel because we're overwriting the global timeNow variable.
	timeNow = func() time.Time {
		return time.Now().Add(6 * 30 * 24 * time.Hour)
	}
	defer func() {
		timeNow = func() time.Time {
			return time.Now()
		}
	}()

	ok, err := gpgVerify(testData, testSignature)
	require.NoError(t, err, "If TestGPGVerify succeeds but this test fails the self-updater key is about to expire. Please open an issue to update the key. Thank you.")
	assert.True(t, ok)
}
