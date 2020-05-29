package out

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/fatih/color"
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
	if IsHidden(ctx) {
		return
	}
	fmt.Fprintf(Stdout, Prefix(ctx)+format+newline(ctx), args...)
}

// Black prints the string in black
func Black(ctx context.Context, format string, args ...interface{}) {
	if IsHidden(ctx) {
		return
	}
	fmt.Fprint(Stdout, color.BlackString(Prefix(ctx)+format+newline(ctx), args...))
}

// Blue prints the string in blue
func Blue(ctx context.Context, format string, args ...interface{}) {
	if IsHidden(ctx) {
		return
	}
	fmt.Fprint(Stdout, color.BlueString(Prefix(ctx)+format+newline(ctx), args...))
}

// Cyan prints the string in cyan
func Cyan(ctx context.Context, format string, args ...interface{}) {
	if IsHidden(ctx) {
		return
	}
	fmt.Fprint(Stdout, color.CyanString(Prefix(ctx)+format+newline(ctx), args...))
}

// Green prints the string in green
func Green(ctx context.Context, format string, args ...interface{}) {
	if IsHidden(ctx) {
		return
	}
	fmt.Fprint(Stdout, color.GreenString(Prefix(ctx)+format+newline(ctx), args...))
}

// Magenta prints the string in magenta
func Magenta(ctx context.Context, format string, args ...interface{}) {
	if IsHidden(ctx) {
		return
	}
	fmt.Fprint(Stdout, color.MagentaString(Prefix(ctx)+format+newline(ctx), args...))
}

// Red prints the string in red
func Red(ctx context.Context, format string, args ...interface{}) {
	if IsHidden(ctx) {
		return
	}
	fmt.Fprint(Stdout, color.RedString(Prefix(ctx)+format+newline(ctx), args...))
}

// Error prints the string in red to stderr
func Error(ctx context.Context, format string, args ...interface{}) {
	if IsHidden(ctx) {
		return
	}
	fmt.Fprint(Stderr, color.RedString(Prefix(ctx)+format+newline(ctx), args...))
}

// White prints the string in white
func White(ctx context.Context, format string, args ...interface{}) {
	if IsHidden(ctx) {
		return
	}
	fmt.Fprint(Stdout, color.WhiteString(Prefix(ctx)+format+newline(ctx), args...))
}

// Yellow prints the string in yellow
func Yellow(ctx context.Context, format string, args ...interface{}) {
	if IsHidden(ctx) {
		return
	}
	fmt.Fprint(Stdout, color.YellowString(Prefix(ctx)+format+newline(ctx), args...))
}
