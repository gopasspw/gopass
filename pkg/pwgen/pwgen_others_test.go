//go:build !windows
// +build !windows

package pwgen

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPwgenExternal(t *testing.T) {
	t.Setenv("GOPASS_EXTERNAL_PWGEN", "echo foobar")

	pw, err := GenerateExternal(4)

	require.NoError(t, err)
	assert.Equal(t, "foobar 4", pw)
}
