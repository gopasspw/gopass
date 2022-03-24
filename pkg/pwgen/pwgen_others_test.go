//go:build !windows
// +build !windows

package pwgen

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPwgenExternal(t *testing.T) { //nolint:paralleltest
	t.Setenv("GOPASS_EXTERNAL_PWGEN", "echo foobar")

	pw, err := GenerateExternal(4)

	assert.NoError(t, err)
	assert.Equal(t, "foobar 4", pw)
}
