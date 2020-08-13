// +build !windows

package pwgen

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPwgenExternal(t *testing.T) {
	_ = os.Setenv("GOPASS_EXTERNAL_PWGEN", "echo foobar")
	defer os.Unsetenv("GOPASS_EXTERNAL_PWGEN")
	pw, err := GenerateExternal(4)
	assert.NoError(t, err)
	assert.Equal(t, "foobar 4", pw)
}
