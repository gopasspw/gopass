package ghssh

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	// Mock GOPASS_HOMEDIR to point to a temp directory
	tempDir := t.TempDir()
	os.Setenv("GOPASS_HOMEDIR", tempDir)
	defer os.Unsetenv("GOPASS_HOMEDIR")

	c, err := New()
	assert.NoError(t, err)
	assert.NotNil(t, c)
	assert.Equal(t, 30*time.Second, c.Timeout)
	assert.NotNil(t, c.client)
	assert.NotNil(t, c.disk)
}

func TestCache_String(t *testing.T) {
	// Mock GOPASS_HOMEDIR to point to a temp directory
	tempDir := t.TempDir()
	os.Setenv("GOPASS_HOMEDIR", tempDir)
	defer os.Unsetenv("GOPASS_HOMEDIR")

	c, err := New()
	assert.NoError(t, err)
	assert.NotNil(t, c)

	expected := "Github SSH key cache (OnDisk: " + filepath.Join(tempDir, "github-ssh") + ")"
	assert.Equal(t, expected, c.String())
}
