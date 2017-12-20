package out

import (
	"context"
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/justwatchcom/gopass/utils/ctxutil"
)

type contextKey int

const (
	ctxKeyPrefix contextKey = iota
	ctxKeyHidden
	ctxKeyNewline
)

var (
	stdout = os.Stdout
)

// WithPrefix returns a context with the given prefix set
func WithPrefix(ctx context.Context, prefix string) context.Context {
	return context.WithValue(ctx, ctxKeyPrefix, prefix)
}

// AddPrefix returns a context with the given prefix added to end of the
// existing prefix
func AddPrefix(ctx context.Context, prefix string) context.Context {
	if prefix == "" {
		return ctx
	}
	pfx := Prefix(ctx)
	if pfx == "" {
		return WithPrefix(ctx, prefix)
	}
	return WithPrefix(ctx, pfx+prefix)
}

// Prefix returns the prefix or an empty string
func Prefix(ctx context.Context) string {
	sv, ok := ctx.Value(ctxKeyPrefix).(string)
	if !ok {
		return ""
	}
	return sv
}

// WithHidden returns a context with the flag value for hidden set
func WithHidden(ctx context.Context, hidden bool) context.Context {
	return context.WithValue(ctx, ctxKeyHidden, hidden)
}

// IsHidden returns true if any output should be hidden in this context
func IsHidden(ctx context.Context) bool {
	bv, ok := ctx.Value(ctxKeyHidden).(bool)
	if !ok {
		return false
	}
	return bv
}

// WithNewline returns a context with the flag value for newline set
func WithNewline(ctx context.Context, nl bool) context.Context {
	return context.WithValue(ctx, ctxKeyNewline, nl)
}

// HasNewline returns the value of newline or the default (true)
func HasNewline(ctx context.Context) bool {
	bv, ok := ctx.Value(ctxKeyNewline).(bool)
	if !ok {
		return true
	}
	return bv
}

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
	fmt.Fprintf(stdout, Prefix(ctx)+format+newline(ctx), args...)
}

// Debug prints the given string if the debug flag is set
func Debug(ctx context.Context, format string, args ...interface{}) {
	if !ctxutil.IsDebug(ctx) {
		return
	}
	fmt.Fprintf(stdout, Prefix(ctx)+"[DEBUG] "+format+newline(ctx), args...)
}

// Black prints the string in black
func Black(ctx context.Context, format string, args ...interface{}) {
	if IsHidden(ctx) && !ctxutil.IsDebug(ctx) {
		return
	}
	fmt.Fprint(stdout, color.BlackString(Prefix(ctx)+format+newline(ctx), args...))
}

// Blue prints the string in blue
func Blue(ctx context.Context, format string, args ...interface{}) {
	if IsHidden(ctx) && !ctxutil.IsDebug(ctx) {
		return
	}
	fmt.Fprint(stdout, color.BlueString(Prefix(ctx)+format+newline(ctx), args...))
}

// Cyan prints the string in cyan
func Cyan(ctx context.Context, format string, args ...interface{}) {
	if IsHidden(ctx) && !ctxutil.IsDebug(ctx) {
		return
	}
	fmt.Fprint(stdout, color.CyanString(Prefix(ctx)+format+newline(ctx), args...))
}

// Green prints the string in green
func Green(ctx context.Context, format string, args ...interface{}) {
	if IsHidden(ctx) && !ctxutil.IsDebug(ctx) {
		return
	}
	fmt.Fprint(stdout, color.GreenString(Prefix(ctx)+format+newline(ctx), args...))
}

// Magenta prints the string in magenta
func Magenta(ctx context.Context, format string, args ...interface{}) {
	if IsHidden(ctx) && !ctxutil.IsDebug(ctx) {
		return
	}
	fmt.Fprint(stdout, color.MagentaString(Prefix(ctx)+format+newline(ctx), args...))
}

// Red prints the string in red
func Red(ctx context.Context, format string, args ...interface{}) {
	if IsHidden(ctx) && !ctxutil.IsDebug(ctx) {
		return
	}
	fmt.Fprint(stdout, color.RedString(Prefix(ctx)+format+newline(ctx), args...))
}

// White prints the string in white
func White(ctx context.Context, format string, args ...interface{}) {
	if IsHidden(ctx) && !ctxutil.IsDebug(ctx) {
		return
	}
	fmt.Fprint(stdout, color.WhiteString(Prefix(ctx)+format+newline(ctx), args...))
}

// Yellow prints the string in yellow
func Yellow(ctx context.Context, format string, args ...interface{}) {
	if IsHidden(ctx) && !ctxutil.IsDebug(ctx) {
		return
	}
	fmt.Fprint(stdout, color.YellowString(Prefix(ctx)+format+newline(ctx), args...))
}
