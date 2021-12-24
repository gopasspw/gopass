package cli

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

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
