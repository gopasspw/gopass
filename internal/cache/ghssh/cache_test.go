package ghssh

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	// Mock GOPASS_HOMEDIR to point to a temp directory
	tempDir := t.TempDir()
	t.Setenv("GOPASS_HOMEDIR", tempDir)

	c, err := New()
	assert.NoError(t, err)
	assert.NotNil(t, c)
	assert.Equal(t, 30*time.Second, c.Timeout)
	assert.NotNil(t, c.disk)
}

func TestCache_String(t *testing.T) {
	// Mock GOPASS_HOMEDIR to point to a temp directory
	tempDir := t.TempDir()
	t.Setenv("GOPASS_HOMEDIR", tempDir)

	c, err := New()
	assert.NoError(t, err)
	assert.NotNil(t, c)

	assert.Contains(t, c.String(), "Github SSH key cache (OnDisk:")
	assert.Contains(t, c.String(), tempDir)
}
