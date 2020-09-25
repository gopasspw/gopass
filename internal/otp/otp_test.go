package otp

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/gokyle/twofactor"
	"github.com/gopasspw/gopass/pkg/gopass/secret/secparse"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const pw string = "password"
const totpSecret string = "GJWTGMTNN5YWW2TNPJXWG2DHMIFA"
const totpURL string = "otpauth://totp/example-otp.com?secret=2m32moqkjmzochgb&issuer=authenticator&digits=6"

func TestCalculate(t *testing.T) {
	testCases := [][]byte{
		[]byte(totpSecret),
		[]byte(fmt.Sprintf("%s\ntotp: %s", pw, totpSecret)),
		[]byte(fmt.Sprintf("%s\n---\ntotp: %s", pw, totpSecret)),
		[]byte(fmt.Sprintf("%s\n%s", pw, totpURL)),
		[]byte(fmt.Sprintf("%s\n---\n%s", pw, totpURL)),
	}

	for _, tc := range testCases {
		s, err := secparse.Parse(tc)
		require.NoError(t, err)
		otp, _, err := Calculate("test", s)
		assert.NoError(t, err, string(tc))
		assert.NotNil(t, otp, string(tc))
	}
}

func TestWrite(t *testing.T) {
	td, err := ioutil.TempDir("", "gopass-")
	assert.NoError(t, err)
	defer func() {
		os.RemoveAll(td)
	}()
	tf := filepath.Join(td, "qr.png")

	otp, label, err := twofactor.FromURL(totpURL)
	assert.NoError(t, err)
	assert.NoError(t, WriteQRFile(otp, label, tf))
}
