package main

import (
	"context"
	"os"
	"runtime"

	"golang.org/x/crypto/ssh/terminal"

	"github.com/gopasspw/gopass/pkg/backend/crypto/gpg"
	"github.com/gopasspw/gopass/pkg/config"
	"github.com/gopasspw/gopass/pkg/ctxutil"
	"github.com/gopasspw/gopass/pkg/store/sub"

	"github.com/fatih/color"
)

func initContext(ctx context.Context, cfg *config.Config) context.Context {
	// always trust
	ctx = gpg.WithAlwaysTrust(ctx, true)

	// check recipients conflicts with always trust, make sure it's not enabled
	// when always trust is
	if gpg.IsAlwaysTrust(ctx) {
		ctx = sub.WithCheckRecipients(ctx, false)
	}

	// debug flag
	if gdb := os.Getenv("GOPASS_DEBUG"); gdb != "" {
		ctx = ctxutil.WithDebug(ctx, true)
	}

	// need this override for our integration tests
	if nc := os.Getenv("GOPASS_NOCOLOR"); nc == "true" || ctxutil.IsNoColor(ctx) {
		color.NoColor = true
		ctx = ctxutil.WithColor(ctx, false)
	}

	// support for no-color.org
	if nc := os.Getenv("NO_COLOR"); nc != "" {
		color.NoColor = true
		ctx = ctxutil.WithColor(ctx, false)
	}

	// only emit color codes when stdout is a terminal
	if !terminal.IsTerminal(int(os.Stdout.Fd())) {
		color.NoColor = true
		ctx = ctxutil.WithColor(ctx, false)
		ctx = ctxutil.WithTerminal(ctx, false)
		ctx = ctxutil.WithInteractive(ctx, false)
	}

	// reading from stdin?
	if info, err := os.Stdin.Stat(); err == nil && info.Mode()&os.ModeCharDevice == 0 {
		ctx = ctxutil.WithInteractive(ctx, false)
		ctx = ctxutil.WithStdin(ctx, true)
	}

	// disable colored output on windows since cmd.exe doesn't support ANSI color
	// codes. Other terminal may do, but until we can figure that out better
	// disable this for all terms on this platform
	if runtime.GOOS == "windows" {
		color.NoColor = true
		ctx = ctxutil.WithColor(ctx, false)
	}

	return ctx
}
