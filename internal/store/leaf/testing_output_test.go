package leaf

import (
	"context"
	"io"
	"os"
	"testing"

	"github.com/gopasspw/gopass/internal/out"
)

// captureOutput returns a context whose out-package output is written to the
// returned files instead of the process stdout/stderr. It allows tests to
// capture output without mutating the package-level out.Stdout/out.Stderr
// globals, which is not safe in tests running in parallel.
func captureOutput(t *testing.T, ctx context.Context) (context.Context, *os.File, *os.File) {
	t.Helper()

	stdout, err := os.CreateTemp(t.TempDir(), "stdout")
	if err != nil {
		t.Fatalf("failed to create stdout capture file: %s", err)
	}

	stderr, err := os.CreateTemp(t.TempDir(), "stderr")
	if err != nil {
		t.Fatalf("failed to create stderr capture file: %s", err)
	}

	t.Cleanup(func() {
		stdout.Close()
		stderr.Close()
	})

	return out.WithWriter(ctx, stdout, stderr), stdout, stderr
}

// seekStart rewinds the capture file so subsequent writes overwrite previous
// output, matching the previous obuf.Reset() semantics.
func seekStart(t *testing.T, f *os.File) {
	t.Helper()

	if _, err := f.Seek(0, io.SeekStart); err != nil {
		t.Fatalf("failed to rewind capture file: %s", err)
	}
}

// stderrContents reads back everything written to the stderr capture file.
func stderrContents(t *testing.T, f *os.File) string {
	t.Helper()

	if _, err := f.Seek(0, io.SeekStart); err != nil {
		t.Fatalf("failed to rewind capture file: %s", err)
	}

	//nolint:gosec // capture files are test-created and bounded
	got, err := io.ReadAll(f)
	if err != nil {
		t.Fatalf("failed to read capture file: %s", err)
	}

	return string(got)
}
