package out

import (
	"bytes"
	"sync"
	"testing"

	"github.com/gopasspw/gopass/internal/config"
	"github.com/stretchr/testify/assert"
)

// TestWriterContextIsolation verifies that output written through a context
// carrying WithWriter lands in the provided writers and never in the
// package-level Stdout. The leaf store tests rely on this to capture output
// from parallel tests without mutating shared globals (which the race
// detector flagged, see issue #3544).
func TestWriterContextIsolation(t *testing.T) {
	t.Parallel()

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	ctx := WithWriter(config.NewContextInMemory(), stdout, stderr)

	var wg sync.WaitGroup

	// One goroutine writes through the package default ...
	wg.Add(1)

	go func() {
		defer wg.Done()

		Printf(config.NewContextInMemory(), "default %s", "write")
	}()

	// ... while another writes through the context-provided writers.
	wg.Add(1)

	go func() {
		defer wg.Done()

		Printf(ctx, "captured %s", "write")
		Warningf(ctx, "captured %s", "warning")
	}()

	wg.Wait()

	assert.Equal(t, "captured write\n", stdout.String())
	assert.Contains(t, stderr.String(), "captured warning")
}

// TestWriterFallbackToPackageGlobals verifies that contexts without a
// writer keep using the package-level Stdout/Stderr, preserving the
// behavior every existing caller depends on.
func TestWriterFallbackToPackageGlobals(t *testing.T) {
	t.Parallel()

	ctx := config.NewContextInMemory()

	so, se := Writer(ctx)
	assert.Equal(t, Stdout, so)
	assert.Equal(t, Stderr, se)

	captured := &bytes.Buffer{}

	ctx = WithWriter(ctx, captured, captured)
	so, se = Writer(ctx)
	assert.Equal(t, captured, so)
	assert.Equal(t, captured, se)
}
