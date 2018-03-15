package cli

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGpgOpts(t *testing.T) {
	for _, vn := range []string{"GOPASS_GPG_OPTS", "PASSWORD_STORE_GPG_OPTS"} {
		for in, out := range map[string][]string{
			"": nil,
			"--decrypt --armor --recipient 0xDEADBEEF": {"--decrypt", "--armor", "--recipient", "0xDEADBEEF"},
		} {
			assert.NoError(t, os.Setenv(vn, in))
			assert.Equal(t, out, GPGOpts())
			assert.NoError(t, os.Unsetenv(vn))
		}
	}
}

func TestSplitPacket(t *testing.T) {
	for in, out := range map[string]map[string]string{
		"": {},
		":pubkey enc packet: version 3, algo 1, keyid 00F0FF00FFC00F0F": {
			"algo":    "1",
			"keyid":   "00F0FF00FFC00F0F",
			"version": "3",
		},
		":encrypted data packet:": {},
	} {
		assert.Equal(t, out, splitPacket(in))
	}
}

func TestTTY(t *testing.T) {
	fd0 = "/tmp/foobar"
	assert.Equal(t, "", tty())
}
