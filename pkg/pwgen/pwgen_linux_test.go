package pwgen

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPwgenExternal(t *testing.T) {
	_ = os.Setenv("GOPASS_EXTERNAL_PWGEN", "echo foobar")
	assert.Equal(t, "foobar", GeneratePassword(4, true))
}
