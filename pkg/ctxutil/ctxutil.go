package ctxutil

import (
	"context"
	"fmt"
	"time"

	"github.com/gopasspw/gopass/internal/store"
	"github.com/urfave/cli/v2"
)

type contextKey int

const (
	ctxKeyColor contextKey = iota
	ctxKeyTerminal
	ctxKeyInteractive
	ctxKeyStdin
	ctxKeyClipTimeout
	ctxKeyConcurrency
	ctxKeyNoPager
	ctxKeyShowSafeContent
	ctxKeyGitCommit
	ctxKeyAlwaysYes
	ctxKeyNoColor
	ctxKeyFuzzySearch
	ctxKeyVerbose
	ctxKeyAutoClip
	ctxKeyNotifications
	ctxKeyProgressCallback
	ctxKeyAlias
	ctxKeyGitInit
	ctxKeyForce
	ctxKeyCommitMessage
	ctxKeyNoNetwork
	ctxKeyUsername
	ctxKeyEmail
	ctxKeyImportFunc
	ctxKeyExportKeys
	ctxKeyPasswordCallback
	ctxKeyCommitTimestamp
)

// WithGlobalFlags parses any global flags from the cli context and returns
// a regular context
func WithGlobalFlags(c *cli.Context) context.Context {
	if c.Bool("yes") {
		return WithAlwaysYes(c.Context, true)
	}
	return c.Context
}

// ProgressCallback is a callback for updateing progress
type ProgressCallback func()

// WithColor returns a context with an explicit value for color
func WithColor(ctx context.Context, color bool) context.Context {
	return context.WithValue(ctx, ctxKeyColor, color)
}

// HasColor returns true if a value for Color has been set in this context
func HasColor(ctx context.Context) bool {
	return hasBool(ctx, ctxKeyColor)
}

// IsColor returns the value of color or the default (true)
func IsColor(ctx context.Context) bool {
	return is(ctx, ctxKeyColor, true)
}

// WithTerminal returns a context with an explicit value for terminal
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

// WithInteractive returns a context with an explicit value for interactive
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

// WithNoPager returns a context with the value for pager set
func WithNoPager(ctx context.Context, bv bool) context.Context {
	return context.WithValue(ctx, ctxKeyNoPager, bv)
}

// HasNoPager returns true if a value for NoPager has been set in this context
func HasNoPager(ctx context.Context) bool {
	_, ok := ctx.Value(ctxKeyNoPager).(bool)
	return ok
}

// IsNoPager returns the value of pager or the default (false)
func IsNoPager(ctx context.Context) bool {
	bv, ok := ctx.Value(ctxKeyNoPager).(bool)
	if !ok {
		return false
	}
	return bv
}

// WithShowSafeContent returns a context with the value for ShowSafeContent set
func WithShowSafeContent(ctx context.Context, bv bool) context.Context {
	return context.WithValue(ctx, ctxKeyShowSafeContent, bv)
}

// HasShowSafeContent returns true if a value for ShowSafeContent has been set in this context
func HasShowSafeContent(ctx context.Context) bool {
	_, ok := ctx.Value(ctxKeyShowSafeContent).(bool)
	return ok
}

// IsShowSafeContent returns the value of ShowSafeContent or the default (false)
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
	_, ok := ctx.Value(ctxKeyGitCommit).(bool)
	return ok
}

// IsGitCommit returns the value of git commit or the default (true)
func IsGitCommit(ctx context.Context) bool {
	return is(ctx, ctxKeyGitCommit, true)
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
	return is(ctx, ctxKeyVerbose, false)
}

// WithNotifications returns a context with the value for Notifications set
func WithNotifications(ctx context.Context, verbose bool) context.Context {
	return context.WithValue(ctx, ctxKeyNotifications, verbose)
}

// HasNotifications returns true if a value for Notifications has been set in this context
func HasNotifications(ctx context.Context) bool {
	return hasBool(ctx, ctxKeyNotifications)
}

// IsNotifications returns the value of Notifications or the default (true)
func IsNotifications(ctx context.Context) bool {
	return is(ctx, ctxKeyNotifications, true)
}

// WithAutoClip returns a context with the value for AutoClip set
func WithAutoClip(ctx context.Context, bv bool) context.Context {
	return context.WithValue(ctx, ctxKeyAutoClip, bv)
}

// HasAutoClip returns true if a value for AutoClip has been set in this context
func HasAutoClip(ctx context.Context) bool {
	return hasBool(ctx, ctxKeyAutoClip)
}

// IsAutoClip returns the value of AutoClip or the default (true)
func IsAutoClip(ctx context.Context) bool {
	return is(ctx, ctxKeyAutoClip, true)
}

// WithConcurrency returns a context with the value for clip timeout set
func WithConcurrency(ctx context.Context, to int) context.Context {
	return context.WithValue(ctx, ctxKeyConcurrency, to)
}

// HasConcurrency returns true if a value for Concurrency has been set in this context and is bigger than 1
// since if it is equal to 1, we are not working concurrently.
func HasConcurrency(ctx context.Context) bool {
	return hasInt(ctx, ctxKeyConcurrency)
}

// GetConcurrency returns the value of concurrent threads or the default (1)
func GetConcurrency(ctx context.Context) int {
	iv, ok := ctx.Value(ctxKeyConcurrency).(int)
	if !ok || iv < 1 {
		return 1
	}
	return iv
}

// WithProgressCallback returns a context with the value of ProgressCallback set
func WithProgressCallback(ctx context.Context, cb ProgressCallback) context.Context {
	return context.WithValue(ctx, ctxKeyProgressCallback, cb)
}

// HasProgressCallback returns true if a ProgressCallback has been set
func HasProgressCallback(ctx context.Context) bool {
	_, ok := ctx.Value(ctxKeyProgressCallback).(ProgressCallback)
	return ok
}

// GetProgressCallback return the set progress callback or a default one.
// It never returns nil
func GetProgressCallback(ctx context.Context) ProgressCallback {
	cb, ok := ctx.Value(ctxKeyProgressCallback).(ProgressCallback)
	if !ok || cb == nil {
		return func() {}
	}
	return cb
}

// WithAlias returns an context with the alias set.
func WithAlias(ctx context.Context, alias string) context.Context {
	return context.WithValue(ctx, ctxKeyAlias, alias)
}

// HasAlias returns true if a value for alias has been set.
func HasAlias(ctx context.Context) bool {
	return hasString(ctx, ctxKeyAlias)
}

// GetAlias returns an alias if it has been set or an empty string otherwise.
func GetAlias(ctx context.Context) string {
	a, ok := ctx.Value(ctxKeyAlias).(string)
	if !ok {
		return ""
	}
	return a
}

// WithGitInit returns a context with the value for the git init flag set.
func WithGitInit(ctx context.Context, bv bool) context.Context {
	return context.WithValue(ctx, ctxKeyGitInit, bv)
}

// HasGitInit returns true if the git init flag was set.
func HasGitInit(ctx context.Context) bool {
	return hasBool(ctx, ctxKeyGitInit)
}

// IsGitInit returns the value of the git init flag or ture if none was set.
func IsGitInit(ctx context.Context) bool {
	return is(ctx, ctxKeyGitInit, true)
}

// WithForce returns a context with the force flag set
func WithForce(ctx context.Context, bv bool) context.Context {
	return context.WithValue(ctx, ctxKeyForce, bv)
}

// HasForce returns true if the context has the force flag set
func HasForce(ctx context.Context) bool {
	return hasBool(ctx, ctxKeyForce)
}

// IsForce returns the force flag value of the default (false)
func IsForce(ctx context.Context) bool {
	return is(ctx, ctxKeyForce, false)
}

// WithCommitMessage returns a context with a commit message set
func WithCommitMessage(ctx context.Context, sv string) context.Context {
	return context.WithValue(ctx, ctxKeyCommitMessage, sv)
}

// HasCommitMessage returns true if the commit message was set
func HasCommitMessage(ctx context.Context) bool {
	return hasString(ctx, ctxKeyCommitMessage)
}

// GetCommitMessage returns the set commit message or an empty string
func GetCommitMessage(ctx context.Context) string {
	sv, ok := ctx.Value(ctxKeyCommitMessage).(string)
	if !ok {
		return ""
	}
	return sv
}

// WithNoNetwork returns a context with the value of no network set
func WithNoNetwork(ctx context.Context, bv bool) context.Context {
	return context.WithValue(ctx, ctxKeyNoNetwork, bv)
}

// HasNoNetwork returns true if no network was set
func HasNoNetwork(ctx context.Context) bool {
	return hasBool(ctx, ctxKeyNoNetwork)
}

// IsNoNetwork returns the value of no network or false
func IsNoNetwork(ctx context.Context) bool {
	return is(ctx, ctxKeyNoNetwork, false)
}

// WithUsername returns a context with the username set in the context
func WithUsername(ctx context.Context, sv string) context.Context {
	return context.WithValue(ctx, ctxKeyUsername, sv)
}

// GetUsername returns the username from the context
func GetUsername(ctx context.Context) string {
	sv, ok := ctx.Value(ctxKeyUsername).(string)
	if !ok {
		return ""
	}
	return sv
}

// WithEmail returns a context with the email set in the context
func WithEmail(ctx context.Context, sv string) context.Context {
	return context.WithValue(ctx, ctxKeyEmail, sv)
}

// GetEmail returns the email from the context
func GetEmail(ctx context.Context) string {
	sv, ok := ctx.Value(ctxKeyEmail).(string)
	if !ok {
		return ""
	}
	return sv
}

// WithImportFunc will return a context with the import callback set
func WithImportFunc(ctx context.Context, imf store.ImportCallback) context.Context {
	return context.WithValue(ctx, ctxKeyImportFunc, imf)
}

// HasImportFunc returns true if a value for import func has been set in this
// context
func HasImportFunc(ctx context.Context) bool {
	imf, ok := ctx.Value(ctxKeyImportFunc).(store.ImportCallback)
	return ok && imf != nil
}

// GetImportFunc will return the import callback or a default one returning true
// Note: will never return nil
func GetImportFunc(ctx context.Context) store.ImportCallback {
	imf, ok := ctx.Value(ctxKeyImportFunc).(store.ImportCallback)
	if !ok || imf == nil {
		return func(context.Context, string, []string) bool {
			return true
		}
	}
	return imf
}

// WithExportKeys returns a context with the value for export keys set.
func WithExportKeys(ctx context.Context, d bool) context.Context {
	return context.WithValue(ctx, ctxKeyExportKeys, d)
}

// HasExportKeys returns true if Export Keys was set in the context.
func HasExportKeys(ctx context.Context) bool {
	return hasBool(ctx, ctxKeyExportKeys)
}

// IsExportKeys returns the value of export keys or the default (true).
func IsExportKeys(ctx context.Context) bool {
	return is(ctx, ctxKeyExportKeys, true)
}

// PasswordCallback is a password prompt callback
type PasswordCallback func(string) ([]byte, error)

// WithPasswordCallback returns a context with the password callback set
func WithPasswordCallback(ctx context.Context, cb PasswordCallback) context.Context {
	return context.WithValue(ctx, ctxKeyPasswordCallback, cb)
}

// HasPasswordCallback returns true if a password callback was set in the context
func HasPasswordCallback(ctx context.Context) bool {
	_, ok := ctx.Value(ctxKeyPasswordCallback).(PasswordCallback)
	return ok
}

// GetPasswordCallback returns the password callback or a default (which always fails)
func GetPasswordCallback(ctx context.Context) PasswordCallback {
	pwcb, ok := ctx.Value(ctxKeyPasswordCallback).(PasswordCallback)
	if !ok || pwcb == nil {
		return func(string) ([]byte, error) {
			return nil, fmt.Errorf("no callback")
		}
	}
	return pwcb
}

// WithCommitTimestamp returns a context with the value for the commit
// timestamp set.
func WithCommitTimestamp(ctx context.Context, ts time.Time) context.Context {
	return context.WithValue(ctx, ctxKeyCommitTimestamp, ts)
}

// HasCommitTimestamp returns true if the value for the commit timestamp
// was set in the context.
func HasCommitTimestamp(ctx context.Context) bool {
	_, ok := ctx.Value(ctxKeyCommitTimestamp).(time.Time)
	return ok
}

// GetCommitTimestamp returns the commit timestamp from the context if
// set or the default (now) otherwise.
func GetCommitTimestamp(ctx context.Context) time.Time {
	if ts, ok := ctx.Value(ctxKeyCommitTimestamp).(time.Time); ok {
		return ts
	}
	return time.Now()
}
