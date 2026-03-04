//go:build linux

package main

import (
	"bufio"
	"os"
	"strings"
	"testing"
)

func TestMain(t *testing.T) {
	// Create a temporary file
	tmpfile, err := os.CreateTemp("", "changelog_test_*.md")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	// Write test data to the temporary file
	content := `# Changelog
## [1.0.1] - 2021-01-01
### Added
- New feature

## [1.0.0] - 2020-12-31
### Added
- Initial release
`
	if _, err := tmpfile.Write([]byte(content)); err != nil {
		t.Fatal(err)
	}
	if err := tmpfile.Close(); err != nil {
		t.Fatal(err)
	}

	// Override the global variable filename
	filename = tmpfile.Name()

	// Capture the output
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	main()

	w.Close()
	os.Stdout = old

	var output strings.Builder
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		output.WriteString(scanner.Text() + "\n")
	}

	expected := `## [1.0.1] - 2021-01-01
### Added
- New feature

`
	if output.String() != expected {
		t.Errorf("expected %q, got %q", expected, output.String())
	}
}
