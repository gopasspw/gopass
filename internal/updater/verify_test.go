package updater

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// To update see README.md.
var testSignature = []byte(`
-----BEGIN PGP SIGNATURE-----

iQGzBAABCgAdFiEE7Tt9z6mw3TnwoHhUKIPSk2aqDz4FAmsD/c8ACgkQKIPSk2aq
Dz73VAv+MOj96+q4a8y6e9AGagZO955mrBoaXwi1Qff4akc52BxF8iD/KWkYoFGt
EXxuA5pDwCJ6seQOBe+7405b2H0ChCyl1GURdT2KuGf/B0YvuskslM3Y7gQ943bV
fBcLeLPH+uLYQTPuY9tZwnRspYEjb/HuYvcaGyHWr2FLCJppy6Z6xg6XA3MAK4HX
3mP6WRY2/3uX/etT+zH/yfxYwkKwTC5nKgZCGlyXOXzYBiNUdDU9w4PWDLPRth20
d2XTSJnlCfyO8Lx/EEFS4DPFtqSvAqjBcbITTMzqSoqkFnmdYVKbjYdViQoBwfBp
knOHxAyJaPk/+Q3Do+HhRdHtXfEgPCtRSdHGn7wcCXhuMUwwBzRf+B+ADsh8Rt6A
AbREsfrLiU8FU6eCNVJ3XxXvbSTWEGrYvXb6tLG1KgNHA/slN75JQSMVURAOUtg6
c5Hq2ZaJT9Otq5LTTxu/qYbpk6+1RSvCtGIhVznYu2uJGsaqdxKty/wne0EZYpW6
HQD1IuC5
=TkFY
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
