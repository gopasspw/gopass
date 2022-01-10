package gpgconf

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDetectBinaryCandidates(t *testing.T) {
	bins, err := detectBinaryCandidates("foobar")
	assert.NoError(t, err)
	// the install locations differ depending on :
	// - chocolatey install path prefix
	// - 64bit/32bit windows
	var stripped []string
	for _, bin := range bins {
		stripped = append(stripped, filepath.Base(bin))
	}
	assert.Contains(t, stripped, "gpg.exe")
}
