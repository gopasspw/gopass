package updater

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// To update see README.md.
var testSignature = []byte(`
-----BEGIN PGP SIGNATURE-----

iQGzBAABCAAdFiEEofrA/VCsqN4ejERxfqcKNWfm6NIFAmena44ACgkQfqcKNWfm
6NLPcgv+PeNs5OLB9y+kJhcWJXyGMyCCq4fj8ACA/mMkRxi+T9iP+51Di+GWyXvd
iMAHCBNbra2qn6nfiy7YJbgFDWZZVVUOXayqbgoGuxojO3n5AF9sK8Ieou7iYXpd
TXx0Zr8XFrhMMvzHVEDNqMtrRpmuwtixHA1PtGx/8Adv35gHRFZzW8xZ1ar5FVXk
Jk/bjo7h1bVf/jaakN9SDx8xc0D72LniPFNrEeOf8QTxSHZFaAOXuU9GsED8Cx1U
wQKBwveBSFKy17dGx03xcknqF/V3djApIgOIZ3MbaD50gpu3x9ltt9yOtkP9op0B
ANkUpIyrgcv39Trf44Z/rgj/bZz0UaagjMwA/RWtjnA6Kuw93BctVcfxuA2jC00g
GSny65MYtI6ynXnJ3xJVrIlNDawK/PjkS/HFWHFLKF7/K4ycL0KBVm/SETdIoGDK
gGTBIqBqDvHISE686mpH6rBRvyu7VOdbh6WTvztynHbdX/1cwyTKghnHNlw6gtIP
rp7LGb+c
=NeAM
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
