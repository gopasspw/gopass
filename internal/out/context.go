package out

import (
	"context"
	"io"
)

type contextKey int

const (
	ctxKeyPrefix contextKey = iota
	ctxKeyNewline
	ctxKeyWriter
)

// WithPrefix returns a context with the given prefix set.
func WithPrefix(ctx context.Context, prefix string) context.Context {
	return context.WithValue(ctx, ctxKeyPrefix, prefix)
}

// AddPrefix returns a context with the given prefix added to end of the
// existing prefix.
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

// Prefix returns the prefix or an empty string.
func Prefix(ctx context.Context) string {
	sv, ok := ctx.Value(ctxKeyPrefix).(string)
	if !ok {
		return ""
	}

	return sv
}

// WithNewline returns a context with the flag value for newline set.
func WithNewline(ctx context.Context, nl bool) context.Context {
	return context.WithValue(ctx, ctxKeyNewline, nl)
}

// HasNewline returns the value of newline or the default (true).
func HasNewline(ctx context.Context) bool {
	bv, ok := ctx.Value(ctxKeyNewline).(bool)
	if !ok {
		return true
	}

	return bv
}

// WithWriter returns a context whose output is written to the given writers
// instead of the package-level Stdout and Stderr. It allows callers, most
// notably tests running in parallel, to capture or discard output without
// mutating shared package state.
func WithWriter(ctx context.Context, stdout, stderr io.Writer) context.Context {
	return context.WithValue(ctx, ctxKeyWriter, writerPair{stdout, stderr})
}

// Writer returns the per-context stdout and stderr writers, falling back to
// the package-level Stdout and Stderr when none were set.
func Writer(ctx context.Context) (io.Writer, io.Writer) {
	if wp, ok := ctx.Value(ctxKeyWriter).(writerPair); ok {
		return wp.stdout, wp.stderr
	}

	return Stdout, Stderr
}

type writerPair struct {
	stdout io.Writer
	stderr io.Writer
}
