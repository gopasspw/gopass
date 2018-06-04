package clipboard

import (
	"context"
	"fmt"

	"github.com/gopasspw/gopass/pkg/ctxutil"
	"github.com/gopasspw/gopass/pkg/out"

	"github.com/atotto/clipboard"
	"github.com/fatih/color"
	"github.com/pkg/errors"
)

var (
	// ErrNotSupported is returned when the clipboard is not accessible
	ErrNotSupported = fmt.Errorf("WARNING: No clipboard available. Install xsel or xclip or use -p to print to console")
)

// CopyTo copies the given data to the clipboard and enqueues automatic
// clearing of the clipboard
func CopyTo(ctx context.Context, name string, content []byte) error {
	if clipboard.Unsupported {
		out.Yellow(ctx, "%s", ErrNotSupported)
		return nil
	}

	if err := clipboard.WriteAll(string(content)); err != nil {
		return errors.Wrapf(err, "failed to write to clipboard")
	}

	if err := clear(ctx, content, ctxutil.GetClipTimeout(ctx)); err != nil {
		return errors.Wrapf(err, "failed to clear clipboard")
	}

	out.Print(ctx, "✔ Copied %s to clipboard. Will clear in %d seconds.", color.YellowString(name), ctxutil.GetClipTimeout(ctx))
	return nil
}
