package updater

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// To generate use:
// `gpg -u 0x4B75BB5995892BE3 --armor --output testdata.sig --detach-sign testdata`.
var testSignature = []byte(`
-----BEGIN PGP SIGNATURE-----

iQGzBAABCgAdFiEEwhyMrSlNNb9aO7sVs8WxoFYNhSIFAmPW2HQACgkQs8WxoFYN
hSLmNgv8CLlft+O7vTolDPM/kZNOlM3UvAzbeA+JkeMyl7snWHnmWgtggtZhMIbq
DIj1OjfW/JKiEqJZy2LCaXUxSshXJ2WHRfxTBvDprSQK5PHiVZwGmsJPXn1aOXSm
OUCPNhv0/wl729reQ7VrLNVM6zXwY91+77XePMsKXV90Vdc+RXucEtZULNeZlhvE
nnJ03ZHLeUpN61CJh6UhBP3dF1aFWiW4+oyONnNyxlC1QNh4oiwcP3iAe8m+gHWj
cQ1z8sTyJFl2l6Mk9cq6wmMwhzyrPgxdre+YDa1oWb/hmq8U3qFH6kkaoYS6b9x3
rEedoxQYh0N7B74IFSgjnKgtUkfQPRXFEUfbGpz03NVtwMnqg8IiCO5bMcbHiqDG
UezkZM1wpxVWCgoGcZaCv6c1gu4KAx8iVhovxekJcEVf+BUahMWBPwIjsmJg5kCU
L63+5ieE6wuYAcVPKZBUG3v9J6VjKpc3puv0sKUPw6swYOF973u3vfChs0oHjpiS
/+EULiNZ
=dS8F
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
