package otp

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/gokyle/twofactor"
	"github.com/gopasspw/gopass/internal/store/secret"
	"github.com/stretchr/testify/assert"
)

const pw string = "password"
const totpSecret string = "GJWTGMTNN5YWW2TNPJXWG2DHMIFA===="
const totpURL string = "otpauth://totp/example-otp.com?secret=2m32moqkjmzochgb&issuer=authenticator&digits=6"

func TestCalculate(t *testing.T) {
	testCases := []struct {
		password       string
		secretContents string
	}{
		{totpSecret, ""},
		{pw, fmt.Sprintf("---\ntotp:%s", totpSecret)},
		{pw, fmt.Sprintf("---\ntotp: %s", totpSecret)},
		{pw, totpURL},
		{pw, fmt.Sprintf("---\n%s", totpURL)},
	}

	for _, tc := range testCases {
		s := secret.New(tc.password, tc.secretContents)
		otp, _, err := Calculate(context.Background(), "test", s)
		assert.Nil(t, err)
		assert.NotNil(t, otp)
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
	assert.NoError(t, WriteQRFile(context.Background(), otp, label, tf))
}
