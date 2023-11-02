package otp

import (
	"fmt"
	"path/filepath"
	"testing"

	"github.com/gopasspw/gopass/pkg/gopass"
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

	for _, tc := range testCases {
		tc := tc

		t.Run(fmt.Sprintf("%s", tc), func(t *testing.T) {
			t.Parallel()

			s, err := secparse.Parse(tc)
			require.NoError(t, err)
			otp, err := Calculate("test", s)
			require.NoError(t, err, string(tc))
			assert.NotNil(t, otp, string(tc))
		})
	}
}

func TestWrite(t *testing.T) {
	t.Parallel()

	td := t.TempDir()

	tf := filepath.Join(td, "qr.png")

	key, err := otp.NewKeyFromURL(totpURL)
	require.NoError(t, err)
	require.NoError(t, WriteQRFile(key, tf))
}

func TestGetOTPURL(t *testing.T) {
	for _, tc := range []struct {
		name string
		sec  gopass.Secret
		url  string
	}{
		{
			name: "url-only-in-body",
			sec:  secparse.MustParse(fmt.Sprintf("%s\n%s", pw, totpURL)),
			url:  totpURL,
		},
		{
			name: "url-and-other-text-in-body",
			sec:  secparse.MustParse(fmt.Sprintf("%s\n%s\nfoo bar\nbaz\n", pw, totpURL)),
			url:  totpURL,
		},
		{
			name: "url-in-kvp",
			sec:  secparse.MustParse(fmt.Sprintf("%s\notpauth: %s\nfoo bar\nbaz\n", pw, totpURL)),
			url:  totpURL,
		},
	} {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tc.url, getOTPURL(tc.sec))
		})
	}
}
