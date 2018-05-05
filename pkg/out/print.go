package out

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/justwatchcom/gopass/pkg/ctxutil"

	"github.com/fatih/color"
)

// Stdout is exported for tests
var Stdout io.Writer = os.Stdout

func newline(ctx context.Context) string {
	if HasNewline(ctx) {
		return "\n"
	}
	return ""
}

// Print formats and prints the given string
func Print(ctx context.Context, format string, args ...interface{}) {
	if IsHidden(ctx) && !ctxutil.IsDebug(ctx) {
		return
	}
	fmt.Fprintf(Stdout, Prefix(ctx)+format+newline(ctx), args...)
}

// Debug prints the given string if the debug flag is set
func Debug(ctx context.Context, format string, args ...interface{}) {
	if !ctxutil.IsDebug(ctx) {
		return
	}
	fmt.Fprintf(Stdout, Prefix(ctx)+"[DEBUG] "+format+newline(ctx), args...)
}

// Black prints the string in black
func Black(ctx context.Context, format string, args ...interface{}) {
	if IsHidden(ctx) && !ctxutil.IsDebug(ctx) {
		return
	}
	fmt.Fprint(Stdout, color.BlackString(Prefix(ctx)+format+newline(ctx), args...))
}

// Blue prints the string in blue
func Blue(ctx context.Context, format string, args ...interface{}) {
	if IsHidden(ctx) && !ctxutil.IsDebug(ctx) {
		return
	}
	fmt.Fprint(Stdout, color.BlueString(Prefix(ctx)+format+newline(ctx), args...))
}

// Cyan prints the string in cyan
func Cyan(ctx context.Context, format string, args ...interface{}) {
	if IsHidden(ctx) && !ctxutil.IsDebug(ctx) {
		return
	}
	fmt.Fprint(Stdout, color.CyanString(Prefix(ctx)+format+newline(ctx), args...))
}

// Green prints the string in green
func Green(ctx context.Context, format string, args ...interface{}) {
	if IsHidden(ctx) && !ctxutil.IsDebug(ctx) {
		return
	}
	fmt.Fprint(Stdout, color.GreenString(Prefix(ctx)+format+newline(ctx), args...))
}

// Magenta prints the string in magenta
func Magenta(ctx context.Context, format string, args ...interface{}) {
	if IsHidden(ctx) && !ctxutil.IsDebug(ctx) {
		return
	}
	fmt.Fprint(Stdout, color.MagentaString(Prefix(ctx)+format+newline(ctx), args...))
}

// Red prints the string in red
func Red(ctx context.Context, format string, args ...interface{}) {
	if IsHidden(ctx) && !ctxutil.IsDebug(ctx) {
		return
	}
	fmt.Fprint(Stdout, color.RedString(Prefix(ctx)+format+newline(ctx), args...))
}

// White prints the string in white
func White(ctx context.Context, format string, args ...interface{}) {
	if IsHidden(ctx) && !ctxutil.IsDebug(ctx) {
		return
	}
	fmt.Fprint(Stdout, color.WhiteString(Prefix(ctx)+format+newline(ctx), args...))
}

// Yellow prints the string in yellow
func Yellow(ctx context.Context, format string, args ...interface{}) {
	if IsHidden(ctx) && !ctxutil.IsDebug(ctx) {
		return
	}
	fmt.Fprint(Stdout, color.YellowString(Prefix(ctx)+format+newline(ctx), args...))
}
