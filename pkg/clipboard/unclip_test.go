package clipboard

import (
	"bytes"
	"context"
	"os"
	"testing"

	"github.com/gopasspw/gopass/internal/out"
	"github.com/gopasspw/gopass/pkg/ctxutil"
	"github.com/stretchr/testify/assert"
)

func TestNotExistingClipboardClearCommand(t *testing.T) { //nolint:paralleltest
	ctx := context.Background()
	ctx = ctxutil.WithAlwaysYes(ctx, true)

	t.Setenv("GOPASS_CLIPBOARD_CLEAR_CMD", "not_existing_command")

	maybeErr := Clear(ctx, "", "", false)
	assert.Error(t, maybeErr)
	assert.Contains(t, maybeErr.Error(), "\"not_existing_command\": executable file not found in")
}

func TestUnclip(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	ctx = ctxutil.WithAlwaysYes(ctx, true)

	buf := &bytes.Buffer{}
	out.Stdout = buf

	defer func() {
		out.Stdout = os.Stdout
	}()

	assert.EqualError(t, Clear(ctx, "", "", false), ErrNotSupported.Error())
}
