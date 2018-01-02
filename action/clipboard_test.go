package action

import (
	"bytes"
	"context"
	"testing"
	"time"

	"github.com/atotto/clipboard"
	"github.com/justwatchcom/gopass/utils/out"
	"github.com/stretchr/testify/assert"
)

func TestCopyToClipboard(t *testing.T) {
	ctx := context.Background()
	clipboard.Unsupported = true

	buf := &bytes.Buffer{}
	out.Stdout = buf
	assert.NoError(t, copyToClipboard(ctx, "foo", []byte("bar")))
	assert.Contains(t, buf.String(), "WARNING")
}

func TestClearClipboard(t *testing.T) {
	ctx := context.Background()
	assert.NoError(t, clearClipboard(ctx, []byte("bar"), 0))
	time.Sleep(50 * time.Millisecond)
}
