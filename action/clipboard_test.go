package action

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/atotto/clipboard"
	"github.com/justwatchcom/gopass/utils/out"
)

func TestCopyToClipboard(t *testing.T) {
	ctx := context.Background()
	clipboard.Unsupported = true

	buf := &bytes.Buffer{}
	out.Stdout = buf
	if err := copyToClipboard(ctx, "foo", []byte("bar")); err != nil {
		t.Fatalf("Error: %s", err)
	}

	if !strings.Contains(buf.String(), "WARNING") {
		t.Errorf("Should warn about missing tools")
	}
}
