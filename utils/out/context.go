package out

import "context"

type contextKey int

const (
	ctxKeyPrefix contextKey = iota
	ctxKeyHidden
	ctxKeyNewline
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
