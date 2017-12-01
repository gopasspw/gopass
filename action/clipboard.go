package action

import (
	"context"
	"fmt"

	"github.com/atotto/clipboard"
	"github.com/fatih/color"
	"github.com/justwatchcom/gopass/utils/ctxutil"
	"github.com/justwatchcom/gopass/utils/out"
	"github.com/pkg/errors"
)

func (s *Action) copyToClipboard(ctx context.Context, name string, content []byte) error {
	if clipboard.Unsupported {
		out.Yellow(ctx, "WARNING: No clipboard available. Install xsel or xclip or use -p to print to console")
		return nil
	}

	if err := clipboard.WriteAll(string(content)); err != nil {
		return errors.Wrapf(err, "failed to write to clipboard")
	}

	if err := clearClipboard(ctx, content, ctxutil.GetClipTimeout(ctx)); err != nil {
		return errors.Wrapf(err, "failed to clear clipboard")
	}

	fmt.Printf("Copied %s to clipboard. Will clear in %d seconds.\n", color.YellowString(name), ctxutil.GetClipTimeout(ctx))
	return nil
}
