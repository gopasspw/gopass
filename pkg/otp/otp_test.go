package otp

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/gopasspw/gopass/pkg/gopass/secrets/secparse"
	"github.com/pquerna/otp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	pw         string = "password"
	totpSecret string = "GJWTGMTNN5YWW2TNPJXWG2DHMIFA"
	totpURL    string = "otpauth://totp/example-otp.com?secret=2m32moqkjmzochgb&issuer=authenticator&digits=6"
)

func TestCalculate(t *testing.T) {
	t.Parallel()

	testCases := [][]byte{
		[]byte(totpSecret),
		[]byte(fmt.Sprintf("%s\ntotp: %s", pw, totpSecret)),
		[]byte(fmt.Sprintf("%s\n---\ntotp: %s", pw, totpSecret)),
		[]byte(fmt.Sprintf("%s\n%s", pw, totpURL)),
		[]byte(fmt.Sprintf("%s\n---\n%s", pw, totpURL)),
	}

	for _, tc := range testCases { //nolint:paralleltest
		tc := tc

		t.Run(fmt.Sprintf("%s", tc), func(t *testing.T) {
			t.Parallel()

			s, err := secparse.Parse(tc)
			require.NoError(t, err)
			otp, err := Calculate("test", s)
			assert.NoError(t, err, string(tc))
			assert.NotNil(t, otp, string(tc))
		})
	}
}

func TestWrite(t *testing.T) {
	t.Parallel()

	td, err := os.MkdirTemp("", "gopass-")
	assert.NoError(t, err)

	defer func() {
		_ = os.RemoveAll(td)
	}()

	tf := filepath.Join(td, "qr.png")

	key, err := otp.NewKeyFromURL(totpURL)
	assert.NoError(t, err)
	assert.NoError(t, WriteQRFile(key, tf))
}
