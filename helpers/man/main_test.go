//go:build linux
// +build linux

package main

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli/v2"
)

func TestMain(t *testing.T) {
	// Setup temporary directory for testing
	tempDir := t.TempDir()
	filename = filepath.Join(tempDir, "VERSION")
	err := os.WriteFile(filename, []byte("1.0.0"), 0o644)
	assert.NoError(t, err)

	// Test man-page generation
	buf := &bytes.Buffer{}
	stdout = buf
	defer func() { stdout = os.Stderr }()

	main()

	assert.Contains(t, buf.String(), "1.0.0")
	assert.Contains(t, buf.String(), "gopass")
	// TODO: Validate man format.
}

func TestGetFlags(t *testing.T) {
	flags := []cli.Flag{
		&cli.BoolFlag{Name: "boolFlag", Usage: "A boolean flag"},
		&cli.IntFlag{Name: "intFlag", Usage: "An integer flag"},
		&cli.StringFlag{Name: "stringFlag", Usage: "A string flag"},
	}

	expected := []flag{
		{Name: "boolFlag", Aliases: []string{"boolFlag"}, Description: "A boolean flag"},
		{Name: "intFlag", Aliases: []string{"intFlag"}, Description: "An integer flag"},
		{Name: "stringFlag", Aliases: []string{"stringFlag"}, Description: "A string flag"},
	}

	result := getFlags(flags)
	assert.Equal(t, expected, result)
}

func TestLookPath(t *testing.T) {
	// Test finding an executable in PATH
	path, err := lookPath("ls")
	assert.NoError(t, err)
	assert.NotEmpty(t, path)

	// Test not finding an executable
	_, err = lookPath("nonexistent")
	assert.Error(t, err)
}
