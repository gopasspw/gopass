package updater

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExtractFile(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()
	dest := filepath.Join(tempDir, "gopass")

	// Create a sample gzip file
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gz)
	err := tw.WriteHeader(&tar.Header{
		Name: "gopass",
		Mode: 0o600,
		Size: int64(len("test content")),
	})
	require.NoError(t, err)
	_, err = tw.Write([]byte("test content"))
	require.NoError(t, err)
	require.NoError(t, tw.Close())
	require.NoError(t, gz.Close())

	err = extractFile(buf.Bytes(), "gopass.gz", dest)
	require.NoError(t, err)

	content, err := os.ReadFile(dest)
	require.NoError(t, err)
	assert.Equal(t, "test content", string(content))
}

func TestExtractToTempFile(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()
	dest := filepath.Join(tempDir, "gopass")

	// Create a sample gzip file
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gz)
	err := tw.WriteHeader(&tar.Header{
		Name: "gopass",
		Mode: 0o600,
		Size: int64(len("test content")),
	})
	require.NoError(t, err)
	_, err = tw.Write([]byte("test content"))
	require.NoError(t, err)
	require.NoError(t, tw.Close())
	require.NoError(t, gz.Close())

	tempFile, err := extractToTempFile(buf.Bytes(), "gopass.gz", dest)
	require.NoError(t, err)

	content, err := os.ReadFile(tempFile)
	require.NoError(t, err)
	assert.Equal(t, "test content", string(content))
}
