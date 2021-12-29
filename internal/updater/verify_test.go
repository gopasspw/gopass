package updater

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// To generate use:
// `gpg -u 0x2F752B2CA00248FC --armor --output testdata.sig --detach-sign testdata`.
var testSignature = []byte(`
-----BEGIN PGP SIGNATURE-----

iQGzBAABCgAdFiEEeXE+gccft5Z7UYXQL3UrLKACSPwFAmAH4zYACgkQL3UrLKAC
SPzcbgv9GwH+bxGJ2IwPPm2IsCCpZyef4Sa7PK8O/4lbEEDcpinap6LZvxzga+xb
8pLV/lFj/QXkxQM3qJbdongkiBu5A77FUmgUi2j1pmK0g9ExF9mYKE858zBBcWIj
P+NyVXQ4OIYmpmTvBa9h4Fbtb3R9J0gXdTh9v/8txAkKNEAWH8C+4lp+MHUmHC/N
O/dwoYF+K7eFlFIei/E0+T2tT2TPgUX3DIxbBqucG9ZERBz6SfjJSjzpQ4R9jQvs
VuUjbTpzmg1elY1aUN41VaiMHY4QGgDILwrs8Y63pRY65Yrox9wAYwxqxlpFbcHT
5FZTJi2W1fKDecrewD7SJ4iEvXawa0DuWZp1bS5rQyk9OJTHudzYAbQuDece6S3E
H5n6eSuBp+fooGijs56IbHQxTclHdRpLNCDRCMf/0vIAtnLy2mcMvIiUCzMolQtx
umpKYk4AOrn1Q56ChcLVx7PIKu618UYq37Tnn7hZqjxL1mTDKcmhU1gPflbLRXiQ
WubSpxPm
=xTH+
-----END PGP SIGNATURE-----
`)

var testData = []byte(`gopass-sign-test
`)

func TestGPGVerify(t *testing.T) {
	ok, err := gpgVerify(testData, testSignature)
	assert.NoError(t, err)
	assert.True(t, ok)
}
