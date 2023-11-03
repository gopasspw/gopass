package gpgconf

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDetectBinaryCandidates(t *testing.T) {
	bins, err := detectBinaryCandidates("foobar")
	require.NoError(t, err)
	// the install locations differ depending on :
	// - chocolatey install path prefix
	// - 64bit/32bit windows
	var stripped []string
	for _, bin := range bins {
		stripped = append(stripped, filepath.Base(bin))
	}
	assert.Contains(t, stripped, "gpg.exe")
}
