package clipboard

import (
	"bytes"
	"context"
	"testing"
	"time"

	"github.com/atotto/clipboard"
	"github.com/justwatchcom/gopass/pkg/out"
	"github.com/stretchr/testify/assert"
)

func TestCopyToClipboard(t *testing.T) {
	ctx := context.Background()
	clipboard.Unsupported = true

	buf := &bytes.Buffer{}
	out.Stdout = buf
	assert.NoError(t, CopyTo(ctx, "foo", []byte("bar")))
	assert.Contains(t, buf.String(), "WARNING")
}

func TestClearClipboard(t *testing.T) {
	ctx := context.Background()
	assert.NoError(t, clear(ctx, []byte("bar"), 0))
	time.Sleep(50 * time.Millisecond)
}
