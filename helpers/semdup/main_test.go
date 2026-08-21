package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/gopasspw/gopass/helpers/semdup/analyzer"
	"github.com/stretchr/testify/require"
)

func TestRunJSON(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module example.com/sample\n\ngo 1.25.0\n"), 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "sample.go"), []byte(`package sample
func first(a int) int { return a + 1 }
func second(b int) int { return b + 1 }
`), 0o600))

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := run([]string{"-dir", dir, "-json", "-min-instructions", "1", "./..."}, &stdout, &stderr)
	require.Zero(t, code, stderr.String())

	var report analyzer.Report
	require.NoError(t, json.Unmarshal(stdout.Bytes(), &report))
	require.Equal(t, 2, report.FunctionsAnalyzed)
	require.Len(t, report.Groups, 1)
}

func TestSplitNonEmpty(t *testing.T) {
	require.Equal(t, []string{"one", "two"}, splitNonEmpty(" one, ,two "))
	require.Nil(t, splitNonEmpty(""))
}
