package ctxutil

import "context"

type contextKey int

const (
	ctxKeyDebug contextKey = iota
	ctxKeyColor
	ctxKeyTerminal
	ctxKeyInteractive
	ctxKeyStdin
	ctxKeyAskForMore
	ctxKeyClipTimeout
	ctxKeyNoConfirm
	ctxKeyNoPager
	ctxKeyShowSafeContent
	ctxKeyGitCommit
	ctxKeyAlwaysYes
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

// WithStdin returns a context with the value for Stdin set. If true some input
// is available on Stdin (e.g. something is being piped into it)
func WithStdin(ctx context.Context, isStdin bool) context.Context {
	return context.WithValue(ctx, ctxKeyStdin, isStdin)
}

// IsStdin returns the value of stdin, i.e. if it's true some data is being
// piped to stdin. If not set it returns the default value (false)
func IsStdin(ctx context.Context) bool {
	bv, ok := ctx.Value(ctxKeyStdin).(bool)
	if !ok {
		return false
	}
	return bv
}

// WithAskForMore returns a context with the value for ask for more set
func WithAskForMore(ctx context.Context, afm bool) context.Context {
	return context.WithValue(ctx, ctxKeyAskForMore, afm)
}

// IsAskForMore returns the value of ask for more or the default (false)
func IsAskForMore(ctx context.Context) bool {
	bv, ok := ctx.Value(ctxKeyAskForMore).(bool)
	if !ok {
		return false
	}
	return bv
}

// WithClipTimeout returns a context with the value for clip timeout set
func WithClipTimeout(ctx context.Context, to int) context.Context {
	return context.WithValue(ctx, ctxKeyClipTimeout, to)
}

// GetClipTimeout returns the value of clip timeout or the default (45)
func GetClipTimeout(ctx context.Context) int {
	iv, ok := ctx.Value(ctxKeyClipTimeout).(int)
	if !ok || iv < 1 {
		return 45
	}
	return iv
}

// WithNoConfirm returns a context with the value for ask for more set
func WithNoConfirm(ctx context.Context, bv bool) context.Context {
	return context.WithValue(ctx, ctxKeyNoConfirm, bv)
}

// IsNoConfirm returns the value of ask for more or the default (false)
func IsNoConfirm(ctx context.Context) bool {
	bv, ok := ctx.Value(ctxKeyNoConfirm).(bool)
	if !ok {
		return false
	}
	return bv
}

// WithNoPager returns a context with the value for ask for more set
func WithNoPager(ctx context.Context, bv bool) context.Context {
	return context.WithValue(ctx, ctxKeyNoPager, bv)
}

// IsNoPager returns the value of ask for more or the default (false)
func IsNoPager(ctx context.Context) bool {
	bv, ok := ctx.Value(ctxKeyNoPager).(bool)
	if !ok {
		return false
	}
	return bv
}

// WithShowSafeContent returns a context with the value for ask for more set
func WithShowSafeContent(ctx context.Context, bv bool) context.Context {
	return context.WithValue(ctx, ctxKeyShowSafeContent, bv)
}

// IsShowSafeContent returns the value of ask for more or the default (false)
func IsShowSafeContent(ctx context.Context) bool {
	bv, ok := ctx.Value(ctxKeyShowSafeContent).(bool)
	if !ok {
		return false
	}
	return bv
}

// WithGitCommit returns a context with the value of git commit set
func WithGitCommit(ctx context.Context, bv bool) context.Context {
	return context.WithValue(ctx, ctxKeyGitCommit, bv)
}

// IsGitCommit returns the value of git commit or the default (true)
func IsGitCommit(ctx context.Context) bool {
	bv, ok := ctx.Value(ctxKeyGitCommit).(bool)
	if !ok {
		return true
	}
	return bv
}

// WithAlwaysYes returns a context with the value of always yes set
func WithAlwaysYes(ctx context.Context, bv bool) context.Context {
	return context.WithValue(ctx, ctxKeyAlwaysYes, bv)
}

// IsAlwaysYes returns the value of always yes or the default (false)
func IsAlwaysYes(ctx context.Context) bool {
	bv, ok := ctx.Value(ctxKeyAlwaysYes).(bool)
	if !ok {
		return false
	}
	return bv
}
