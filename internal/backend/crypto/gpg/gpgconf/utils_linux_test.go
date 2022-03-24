//go:build linux
// +build linux

package gpgconf

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTTY(t *testing.T) {
	t.Parallel()

	fd0 = "/tmp/foobar"
	assert.Equal(t, "", TTY())
}
