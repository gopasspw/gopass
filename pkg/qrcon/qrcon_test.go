package qrcon

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func ExampleQRCode() { //nolint:testableexamples
	code, err := QRCode("foo")
	if err != nil {
		panic(err)
	}

	fmt.Println(code)
}

func TestQRCode(t *testing.T) {
	t.Parallel()

	_, err := QRCode("https://www.gopass.pw/")
	assert.NoError(t, err)
}
