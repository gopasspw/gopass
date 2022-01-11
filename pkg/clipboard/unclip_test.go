package clipboard

import (
	"bytes"
	"context"
	"os"
	"testing"

	"github.com/gopasspw/gopass/internal/out"
	"github.com/gopasspw/gopass/pkg/ctxutil"
	"github.com/gopasspw/gopass/tests/gptest"
	"github.com/stretchr/testify/assert"
)

func TestNotExistingClipboardClearCommand(t *testing.T) {
	r1 := gptest.UnsetVars("GOPASS_CLIPBOARD_CLEAR_CMD")
	defer r1()

	ctx := context.Background()
	ctx = ctxutil.WithAlwaysYes(ctx, true)

	_ = os.Setenv("GOPASS_CLIPBOARD_CLEAR_CMD", "not_existing_command")

	maybeErr := Clear(ctx, "", "", false)
	assert.Error(t, maybeErr)
	assert.Contains(t, maybeErr.Error(), "\"not_existing_command\": executable file not found in")
}

func TestUnclip(t *testing.T) {
	ctx := context.Background()
	ctx = ctxutil.WithAlwaysYes(ctx, true)

	buf := &bytes.Buffer{}
	out.Stdout = buf
	defer func() {
		out.Stdout = os.Stdout
	}()

	assert.EqualError(t, Clear(ctx, "", "", false), ErrNotSupported.Error())
}
