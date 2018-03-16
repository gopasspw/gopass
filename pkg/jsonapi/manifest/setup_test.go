package manifest

import (
	"bytes"
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPrintSummary(t *testing.T) {
	oldOut := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	done := make(chan string)
	go func() {
		buf := &bytes.Buffer{}
		_, _ = io.Copy(buf, r)
		done <- buf.String()
	}()
	assert.NoError(t, PrintSummary("chrome", "/usr/lib/wrapper", "/usr/lib", false))
	assert.NoError(t, w.Close())
	os.Stdout = oldOut
	assert.Contains(t, <-done, "Native Messaging Host Manifest")
}
