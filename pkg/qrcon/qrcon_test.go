package qrcon

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestQRCode(t *testing.T) {
	_, err := QRCode("https://www.gopass.pw/")
	assert.NoError(t, err)
}
