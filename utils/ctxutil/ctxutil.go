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
	ctxKeyUseSymbols
	ctxKeyNoColor
	ctxKeyFuzzySearch
	ctxKeyVerbose
)

// WithDebug returns a context with an explizit value for debug
func WithDebug(ctx context.Context, dbg bool) context.Context {
	return context.WithValue(ctx, ctxKeyDebug, dbg)
}

// HasDebug returns true if a value for debug has been set in this context
func HasDebug(ctx context.Context) bool {
	_, ok := ctx.Value(ctxKeyDebug).(bool)
	return ok
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

// HasColor returns true if a value for Color has been set in this context
func HasColor(ctx context.Context) bool {
	_, ok := ctx.Value(ctxKeyColor).(bool)
	return ok
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

// HasTerminal returns true if a value for Terminal has been set in this context
func HasTerminal(ctx context.Context) bool {
	_, ok := ctx.Value(ctxKeyTerminal).(bool)
	return ok
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

// HasInteractive returns true if a value for Interactive has been set in this context
func HasInteractive(ctx context.Context) bool {
	_, ok := ctx.Value(ctxKeyInteractive).(bool)
	return ok
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

// HasStdin returns true if a value for Stdin has been set in this context
func HasStdin(ctx context.Context) bool {
	_, ok := ctx.Value(ctxKeyStdin).(bool)
	return ok
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

// HasAskForMore returns true if a value for AskForMore has been set in this context
func HasAskForMore(ctx context.Context) bool {
	_, ok := ctx.Value(ctxKeyAskForMore).(bool)
	return ok
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

// HasClipTimeout returns true if a value for ClipTimeout has been set in this context
func HasClipTimeout(ctx context.Context) bool {
	_, ok := ctx.Value(ctxKeyClipTimeout).(int)
	return ok
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

// HasNoConfirm returns true if a value for NoConfirm has been set in this context
func HasNoConfirm(ctx context.Context) bool {
	_, ok := ctx.Value(ctxKeyNoConfirm).(bool)
	return ok
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

// HasNoPager returns true if a value for NoPager has been set in this context
func HasNoPager(ctx context.Context) bool {
	_, ok := ctx.Value(ctxKeyNoPager).(bool)
	return ok
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

// HasShowSafeContent returns true if a value for ShowSafeContent has been set in this context
func HasShowSafeContent(ctx context.Context) bool {
	_, ok := ctx.Value(ctxKeyShowSafeContent).(bool)
	return ok
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

// HasGitCommit returns true if a value for GitCommit has been set in this context
func HasGitCommit(ctx context.Context) bool {
	_, ok := ctx.Value(ctxKeyShowSafeContent).(bool)
	return ok
}

// IsGitCommit returns the value of git commit or the default (true)
func IsGitCommit(ctx context.Context) bool {
	bv, ok := ctx.Value(ctxKeyGitCommit).(bool)
	if !ok {
		return true
	}
	return bv
}

// WithUseSymbols returns a context with the value for ask for more set
func WithUseSymbols(ctx context.Context, bv bool) context.Context {
	return context.WithValue(ctx, ctxKeyUseSymbols, bv)
}

// HasUseSymbols returns true if a value for UseSymbols has been set in this context
func HasUseSymbols(ctx context.Context) bool {
	_, ok := ctx.Value(ctxKeyUseSymbols).(bool)
	return ok
}

// IsUseSymbols returns the value of ask for more or the default (false)
func IsUseSymbols(ctx context.Context) bool {
	bv, ok := ctx.Value(ctxKeyUseSymbols).(bool)
	if !ok {
		return false
	}
	return bv
}

// WithNoColor returns a context with the value for ask for more set
func WithNoColor(ctx context.Context, bv bool) context.Context {
	return context.WithValue(ctx, ctxKeyNoColor, bv)
}

// HasNoColor returns true if a value for NoColor has been set in this context
func HasNoColor(ctx context.Context) bool {
	_, ok := ctx.Value(ctxKeyNoColor).(bool)
	return ok
}

// IsNoColor returns the value of ask for more or the default (false)
func IsNoColor(ctx context.Context) bool {
	bv, ok := ctx.Value(ctxKeyNoColor).(bool)
	if !ok {
		return false
	}
	return bv
}

// WithAlwaysYes returns a context with the value of always yes set
func WithAlwaysYes(ctx context.Context, bv bool) context.Context {
	return context.WithValue(ctx, ctxKeyAlwaysYes, bv)
}

// HasAlwaysYes returns true if a value for AlwaysYes has been set in this context
func HasAlwaysYes(ctx context.Context) bool {
	_, ok := ctx.Value(ctxKeyAlwaysYes).(bool)
	return ok
}

// IsAlwaysYes returns the value of always yes or the default (false)
func IsAlwaysYes(ctx context.Context) bool {
	bv, ok := ctx.Value(ctxKeyAlwaysYes).(bool)
	if !ok {
		return false
	}
	return bv
}

// WithFuzzySearch returns a context with the value for fuzzy search set
func WithFuzzySearch(ctx context.Context, fuzzy bool) context.Context {
	return context.WithValue(ctx, ctxKeyFuzzySearch, fuzzy)
}

// HasFuzzySearch returns true if a value for FuzzySearch has been set in this context
func HasFuzzySearch(ctx context.Context) bool {
	_, ok := ctx.Value(ctxKeyFuzzySearch).(bool)
	return ok
}

// IsFuzzySearch return the value of fuzzy search or the default (true)
func IsFuzzySearch(ctx context.Context) bool {
	bv, ok := ctx.Value(ctxKeyFuzzySearch).(bool)
	if !ok {
		return true
	}
	return bv
}

// WithVerbose returns a context with the value for verbose set
func WithVerbose(ctx context.Context, verbose bool) context.Context {
	return context.WithValue(ctx, ctxKeyVerbose, verbose)
}

// HasVerbose returns true if a value for Verbose has been set in this context
func HasVerbose(ctx context.Context) bool {
	_, ok := ctx.Value(ctxKeyVerbose).(bool)
	return ok
}

// IsVerbose returns the value of verbose or the default (false)
func IsVerbose(ctx context.Context) bool {
	bv, ok := ctx.Value(ctxKeyVerbose).(bool)
	if !ok {
		return false
	}
	return bv
}
