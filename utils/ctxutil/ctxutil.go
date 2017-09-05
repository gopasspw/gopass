package ctxutil

import "context"

type contextKey int

const (
	ctxKeyDebug contextKey = iota
	ctxKeyColor
	ctxKeyTerminal
	ctxKeyInteractive
)

// WithDebug returns a context with an explizit value for debug
func WithDebug(ctx context.Context, dbg bool) context.Context {
	return context.WithValue(ctx, ctxKeyDebug, dbg)
}

// IsDebug returns the value of debug or the default (false)
func IsDebug(ctx context.Context) bool {
	bv, ok := ctx.Value(ctxKeyDebug).(bool)
	if !ok {
		return false
	}
	return bv
}

// WithColor returns a context with an explizit value for color
func WithColor(ctx context.Context, color bool) context.Context {
	return context.WithValue(ctx, ctxKeyColor, color)
}

// IsColor returns the value of color or the default (true)
func IsColor(ctx context.Context) bool {
	bv, ok := ctx.Value(ctxKeyColor).(bool)
	if !ok {
		return true
	}
	return bv
}

// WithTerminal returns a context with an explizit value for terminal
func WithTerminal(ctx context.Context, isTerm bool) context.Context {
	return context.WithValue(ctx, ctxKeyTerminal, isTerm)
}

// IsTerminal returns the value of terminal or the default (true)
func IsTerminal(ctx context.Context) bool {
	bv, ok := ctx.Value(ctxKeyTerminal).(bool)
	if !ok {
		return true
	}
	return bv
}

// WithInteractive returns a context with an explizit value for interactive
func WithInteractive(ctx context.Context, isInteractive bool) context.Context {
	return context.WithValue(ctx, ctxKeyInteractive, isInteractive)
}

// IsInteractive returns the value of interactive or the default (true)
func IsInteractive(ctx context.Context) bool {
	bv, ok := ctx.Value(ctxKeyInteractive).(bool)
	if !ok {
		return true
	}
	return bv
}
