package out

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/fatih/color"
	"github.com/gopasspw/gopass/pkg/ctxutil"
)

var (
	// Stdout is exported for tests
	Stdout io.Writer = os.Stdout
	// Stderr is exported for tests
	Stderr io.Writer = os.Stderr
)

func newline(ctx context.Context) string {
	if HasNewline(ctx) {
		return "\n"
	}
	return ""
}

// Print formats and prints the given string
func Print(ctx context.Context, format string, args ...interface{}) {
	if ctxutil.IsHidden(ctx) {
		return
	}
	fmt.Fprintf(Stdout, Prefix(ctx)+format+newline(ctx), args...)
}

// Notice prints the string with an exclamation mark in front
func Notice(ctx context.Context, format string, args ...interface{}) {
	if ctxutil.IsHidden(ctx) {
		return
	}
	fmt.Fprintf(Stdout, Prefix(ctx)+"⚠ "+format+newline(ctx), args...)
}

// Error prints the string in red to stderr
func Error(ctx context.Context, format string, args ...interface{}) {
	if ctxutil.IsHidden(ctx) {
		return
	}
	fmt.Fprint(Stderr, color.RedString(Prefix(ctx)+"❌ "+format+newline(ctx), args...))
}

// OK prints the string in with an OK checkmark in front
func OK(ctx context.Context, format string, args ...interface{}) {
	if ctxutil.IsHidden(ctx) {
		return
	}
	fmt.Fprintf(Stdout, Prefix(ctx)+"✅ "+format+newline(ctx), args...)
}

// Warning prints the string in yellow to stderr and prepends "Warning: "
func Warning(ctx context.Context, format string, args ...interface{}) {
	if ctxutil.IsHidden(ctx) {
		return
	}
	fmt.Fprint(Stderr, color.YellowString(Prefix(ctx)+"⚠ "+format+newline(ctx), args...))
}
