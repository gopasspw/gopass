package qrcon

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestQRCode(t *testing.T) {
	_, err := QRCode("http://www.justwatch.com/")
	assert.NoError(t, err)
}
