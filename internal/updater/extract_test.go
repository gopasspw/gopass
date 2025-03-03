package updater

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
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
	assert.NoError(t, err)
	_, err = tw.Write([]byte("test content"))
	assert.NoError(t, err)
	assert.NoError(t, tw.Close())
	assert.NoError(t, gz.Close())

	err = extractFile(buf.Bytes(), "gopass.gz", dest)
	assert.NoError(t, err)

	content, err := os.ReadFile(dest)
	assert.NoError(t, err)
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
	assert.NoError(t, err)
	_, err = tw.Write([]byte("test content"))
	assert.NoError(t, err)
	assert.NoError(t, tw.Close())
	assert.NoError(t, gz.Close())

	tempFile, err := extractToTempFile(buf.Bytes(), "gopass.gz", dest)
	assert.NoError(t, err)

	content, err := os.ReadFile(tempFile)
	assert.NoError(t, err)
	assert.Equal(t, "test content", string(content))
}

func TestExtractZip(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()
	dest := filepath.Join(tempDir, "gopass")

	// Create a sample zip file
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	w, err := zw.Create("gopass.exe")
	assert.NoError(t, err)
	_, err = w.Write([]byte("test content"))
	assert.NoError(t, err)
	assert.NoError(t, zw.Close())

	dfh, err := os.CreateTemp(tempDir, "gopass")
	assert.NoError(t, err)
	defer dfh.Close()

	tempFile, err := extractZip(buf.Bytes(), dfh, dest)
	assert.NoError(t, err)

	content, err := os.ReadFile(tempFile)
	assert.NoError(t, err)
	assert.Equal(t, "test content", string(content))
}

func TestExtractTar(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()
	dest := filepath.Join(tempDir, "gopass")

	// Create a sample tar file
	var buf bytes.Buffer
	tw := tar.NewWriter(&buf)
	err := tw.WriteHeader(&tar.Header{
		Name: "gopass",
		Mode: 0o600,
		Size: int64(len("test content")),
	})
	assert.NoError(t, err)
	_, err = tw.Write([]byte("test content"))
	assert.NoError(t, err)
	assert.NoError(t, tw.Close())

	dfh, err := os.CreateTemp(tempDir, "gopass")
	assert.NoError(t, err)
	defer dfh.Close()

	tempFile, err := extractTar(&buf, dfh, dest)
	assert.NoError(t, err)

	content, err := os.ReadFile(tempFile)
	assert.NoError(t, err)
	assert.Equal(t, "test content", string(content))
}
