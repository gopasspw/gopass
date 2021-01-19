package updater

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// gpg -u 0x21FF4D8F35D1AEE5 --armor --output testdata.sig --detach-sign testdata
var testSignature = []byte(`
-----BEGIN PGP SIGNATURE-----

iLgEABMKAB0WIQTiWnWvE2Kw5vo0dzgh/02PNdGu5QUCYAcgywAKCRAh/02PNdGu
5TVbAgkBPO1k24iHSVrtfz0Mdy8RQbXxr50Y+I2PpL1Ai2e9f+yrUnmENnmzrlom
/vhR47sCH0KEjJ0xVKsfCgkX9Tv/jZICBjv0s0GoAFkeM5mCSHTYfbZZSjFPnxR9
A3ELxJxspsRt1awBSF79yPrJnuZIWMz81G4JhBBqbCpNM85982uSqXqo
=vhCd
-----END PGP SIGNATURE-----
`)

var testData = []byte(`gopass-sign-test
`)

func TestGPGVerify(t *testing.T) {
	ok, err := gpgVerify(testData, testSignature)
	assert.NoError(t, err)
	assert.True(t, ok)
}
