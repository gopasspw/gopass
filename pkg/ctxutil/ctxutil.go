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
	ctxKeyTerminal contextKey = iota
	ctxKeyInteractive
	ctxKeyStdin
	ctxKeyGitCommit
	ctxKeyAlwaysYes
	ctxKeyVerbose
	ctxKeyProgressCallback
	ctxKeyAlias
	ctxKeyGitInit
	ctxKeyForce
	ctxKeyCommitMessage
	ctxKeyCommitMessageBody
	ctxKeyErrorCollector
	ctxKeyIsMultiSecretOperation
	ctxKeyNoNetwork
	ctxKeyUsername
	ctxKeyEmail
	ctxKeyImportFunc
	ctxKeyPasswordCallback
	ctxKeyPasswordPurgeCallback
	ctxKeyCommitTimestamp
	ctxKeyShowParsing
	ctxKeyHidden
)

// ErrNoCallback is returned when no callback is set in the context.
var ErrNoCallback = fmt.Errorf("no callback")

// WithGlobalFlags parses any global flags from the cli context and returns
// a regular context.
func WithGlobalFlags(c *cli.Context) context.Context {
	if c.Bool("yes") {
		return WithAlwaysYes(c.Context, true)
	}

	return c.Context
}

// ProgressCallback is a callback for updateing progress.
type ProgressCallback func()

// WithTerminal returns a context with an explicit value for terminal.
func WithTerminal(ctx context.Context, isTerm bool) context.Context {
	return context.WithValue(ctx, ctxKeyTerminal, isTerm)
}

// HasTerminal returns true if a value for Terminal has been set in this context.
func HasTerminal(ctx context.Context) bool {
	_, ok := ctx.Value(ctxKeyTerminal).(bool)

	return ok
}

// IsTerminal returns the value of terminal or the default (true).
func IsTerminal(ctx context.Context) bool {
	bv, ok := ctx.Value(ctxKeyTerminal).(bool)
	if !ok {
		return true
	}

	return bv
}

// WithInteractive returns a context with an explicit value for interactive.
func WithInteractive(ctx context.Context, isInteractive bool) context.Context {
	return context.WithValue(ctx, ctxKeyInteractive, isInteractive)
}

// HasInteractive returns true if a value for Interactive has been set in this context.
func HasInteractive(ctx context.Context) bool {
	_, ok := ctx.Value(ctxKeyInteractive).(bool)

	return ok
}

// IsInteractive returns the value of interactive or the default (true).
func IsInteractive(ctx context.Context) bool {
	bv, ok := ctx.Value(ctxKeyInteractive).(bool)
	if !ok {
		return true
	}

	return bv
}

// WithStdin returns a context with the value for Stdin set. If true some input
// is available on Stdin (e.g. something is being piped into it).
func WithStdin(ctx context.Context, isStdin bool) context.Context {
	return context.WithValue(ctx, ctxKeyStdin, isStdin)
}

// HasStdin returns true if a value for Stdin has been set in this context.
func HasStdin(ctx context.Context) bool {
	_, ok := ctx.Value(ctxKeyStdin).(bool)

	return ok
}

// IsStdin returns the value of stdin, i.e. if it's true some data is being
// piped to stdin. If not set it returns the default value (false).
func IsStdin(ctx context.Context) bool {
	bv, ok := ctx.Value(ctxKeyStdin).(bool)
	if !ok {
		return false
	}

	return bv
}

// WithShowParsing returns a context with the value for ShowParsing set.
func WithShowParsing(ctx context.Context, bv bool) context.Context {
	return context.WithValue(ctx, ctxKeyShowParsing, bv)
}

// HasShowParsing returns true if a value for ShowParsing has been set in this context.
func HasShowParsing(ctx context.Context) bool {
	_, ok := ctx.Value(ctxKeyShowParsing).(bool)

	return ok
}

// IsShowParsing returns the value of ShowParsing or the default (true).
func IsShowParsing(ctx context.Context) bool {
	bv, ok := ctx.Value(ctxKeyShowParsing).(bool)
	if !ok {
		return true
	}

	return bv
}

// WithGitCommit returns a context with the value of git commit set.
func WithGitCommit(ctx context.Context, bv bool) context.Context {
	return context.WithValue(ctx, ctxKeyGitCommit, bv)
}

// HasGitCommit returns true if a value for GitCommit has been set in this context.
func HasGitCommit(ctx context.Context) bool {
	_, ok := ctx.Value(ctxKeyGitCommit).(bool)

	return ok
}

// IsGitCommit returns the value of git commit or the default (true).
func IsGitCommit(ctx context.Context) bool {
	return is(ctx, ctxKeyGitCommit, true)
}

// WithAlwaysYes returns a context with the value of always yes set.
func WithAlwaysYes(ctx context.Context, bv bool) context.Context {
	return context.WithValue(ctx, ctxKeyAlwaysYes, bv)
}

// HasAlwaysYes returns true if a value for AlwaysYes has been set in this context.
func HasAlwaysYes(ctx context.Context) bool {
	_, ok := ctx.Value(ctxKeyAlwaysYes).(bool)

	return ok
}

// IsAlwaysYes returns the value of always yes or the default (false).
func IsAlwaysYes(ctx context.Context) bool {
	bv, ok := ctx.Value(ctxKeyAlwaysYes).(bool)
	if !ok {
		return false
	}

	return bv
}

// WithVerbose returns a context with the value for verbose set.
func WithVerbose(ctx context.Context, verbose bool) context.Context {
	return context.WithValue(ctx, ctxKeyVerbose, verbose)
}

// HasVerbose returns true if a value for Verbose has been set in this context.
func HasVerbose(ctx context.Context) bool {
	_, ok := ctx.Value(ctxKeyVerbose).(bool)

	return ok
}

// IsVerbose returns the value of verbose or the default (false).
func IsVerbose(ctx context.Context) bool {
	return is(ctx, ctxKeyVerbose, false)
}

// WithProgressCallback returns a context with the value of ProgressCallback set.
func WithProgressCallback(ctx context.Context, cb ProgressCallback) context.Context {
	return context.WithValue(ctx, ctxKeyProgressCallback, cb)
}

// HasProgressCallback returns true if a ProgressCallback has been set.
func HasProgressCallback(ctx context.Context) bool {
	_, ok := ctx.Value(ctxKeyProgressCallback).(ProgressCallback)

	return ok
}

// GetProgressCallback return the set progress callback or a default one.
// It never returns nil.
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

// WithForce returns a context with the force flag set.
func WithForce(ctx context.Context, bv bool) context.Context {
	return context.WithValue(ctx, ctxKeyForce, bv)
}

// HasForce returns true if the context has the force flag set.
func HasForce(ctx context.Context) bool {
	return hasBool(ctx, ctxKeyForce)
}

// IsForce returns the force flag value of the default (false).
func IsForce(ctx context.Context) bool {
	return is(ctx, ctxKeyForce, false)
}

// WithMultiSecretOperation returns a context with a flag set
func WithMultiSecretOperation (ctx context.Context, bv bool) context.Context {
	return context.WithValue(ctx, ctxKeyIsMultiSecretOperation, bv)
}

// IsMultiSecretOperation returns true if the flag is set to true
func IsMultiSecretOperation (ctx context.Context) bool {
	if hasBool(ctx,ctxKeyIsMultiSecretOperation) {
		return is(ctx, ctxKeyForce, true)
	} else {
		return false
	}
}


// AddToCommitMessageBody returns a context with something added to the commit's body
func AddToCommitMessageBody(ctx context.Context, sv string) context.Context {
	if hasString(ctx, ctxKeyCommitMessageBody) {
		current_body, ok := ctx.Value(ctxKeyCommitMessageBody).(string)
		if ok {
			current_body += "\n"
			current_body += sv
			sv = current_body
		}
	}
	return context.WithValue(ctx, ctxKeyCommitMessageBody, sv)
}

// HasCommitMessageBody returns true if the commit message body is nonempty.
func HasCommitMessageBody(ctx context.Context) bool {
	return hasString(ctx, ctxKeyCommitMessageBody)
}

// GetCommitMessageBody returns the set commit message body or an empty string.
func GetCommitMessageBody(ctx context.Context) string {
	sv, ok := ctx.Value(ctxKeyCommitMessageBody).(string)
	if !ok {
		return ""
	}

	return sv
}

// WithCommitMessage returns a context with a commit message (head) set.
// (full commit message is the commit message's body is not defined, commit messahe head otherwise)
func WithCommitMessage(ctx context.Context, sv string) context.Context {
	return context.WithValue(ctx, ctxKeyCommitMessage, sv)
}

// HasCommitMessage returns true if the commit message (head) was set.
func HasCommitMessage(ctx context.Context) bool {
	return hasString(ctx, ctxKeyCommitMessage)
}

// GetCommitMessage returns the set commit message (head) or an empty string.
func GetCommitMessage(ctx context.Context) string {
	sv, ok := ctx.Value(ctxKeyCommitMessage).(string)
	if !ok {
		return ""
	}

	return sv
}

// GetCommitMessageFull returns the set commit message (head+body, of either are defined) or an empty string.
func GetCommitMessageFull(ctx context.Context) string {
	sv_head, ok_head := ctx.Value(ctxKeyCommitMessage).(string)
	sv_body, ok_body := ctx.Value(ctxKeyCommitMessageBody).(string)
	if !(ok_body||ok_head) {
		return ""
	} else if !ok_body {
		return sv_head
	} else if !ok_head {
		return sv_body
	} else {
		return sv_head + "\n\n" + sv_body
	}
}

// ClearErrorCollector returns a context with where the error collector is emptied.
func ClearErrorCollector(ctx context.Context) context.Context {
	return context.WithValue(ctx, ctxKeyErrorCollector, nil)
}
// AddToErrorCollector returns a context with a new error logged as a new line into the error collector
func AddToErrorCollector(ctx context.Context, sv string) context.Context {
	if hasString(ctx, ctxKeyErrorCollector) {
		current_body, ok := ctx.Value(ctxKeyErrorCollector).(string)
		if ok {
			current_body += "\n"
			current_body += sv
			sv = current_body
		}
	}
	return context.WithValue(ctx, ctxKeyErrorCollector, sv)
}
// HasErrorCollector returns true if at least one error was added to the collector.
func HasErrorCollector(ctx context.Context) bool {
	return hasString(ctx, ctxKeyErrorCollector)
}
// GetErrorCollector returns the content, if any, of the error collector.
func GetErrorCollector(ctx context.Context) string {
	sv, ok := ctx.Value(ctxKeyErrorCollector).(string)
	if !ok {
		return ""
	}
	return sv
}

// WithNoNetwork returns a context with the value of no network set.
func WithNoNetwork(ctx context.Context, bv bool) context.Context {
	return context.WithValue(ctx, ctxKeyNoNetwork, bv)
}

// HasNoNetwork returns true if no network was set.
func HasNoNetwork(ctx context.Context) bool {
	return hasBool(ctx, ctxKeyNoNetwork)
}

// IsNoNetwork returns the value of no network or false.
func IsNoNetwork(ctx context.Context) bool {
	return is(ctx, ctxKeyNoNetwork, false)
}

// WithUsername returns a context with the username set in the context.
func WithUsername(ctx context.Context, sv string) context.Context {
	return context.WithValue(ctx, ctxKeyUsername, sv)
}

// GetUsername returns the username from the context.
func GetUsername(ctx context.Context) string {
	sv, ok := ctx.Value(ctxKeyUsername).(string)
	if !ok {
		return ""
	}

	return sv
}

// WithEmail returns a context with the email set in the context.
func WithEmail(ctx context.Context, sv string) context.Context {
	return context.WithValue(ctx, ctxKeyEmail, sv)
}

// GetEmail returns the email from the context.
func GetEmail(ctx context.Context) string {
	sv, ok := ctx.Value(ctxKeyEmail).(string)
	if !ok {
		return ""
	}

	return sv
}

// WithImportFunc will return a context with the import callback set.
func WithImportFunc(ctx context.Context, imf store.ImportCallback) context.Context {
	return context.WithValue(ctx, ctxKeyImportFunc, imf)
}

// HasImportFunc returns true if a value for import func has been set in this
// context.
func HasImportFunc(ctx context.Context) bool {
	imf, ok := ctx.Value(ctxKeyImportFunc).(store.ImportCallback)

	return ok && imf != nil
}

// GetImportFunc will return the import callback or a default one returning true
// Note: will never return nil.
func GetImportFunc(ctx context.Context) store.ImportCallback {
	imf, ok := ctx.Value(ctxKeyImportFunc).(store.ImportCallback)
	if !ok || imf == nil {
		return func(context.Context, string, []string) bool {
			return true
		}
	}

	return imf
}

// PasswordCallback is a password prompt callback.
type PasswordCallback func(string, bool) ([]byte, error)

// WithPasswordCallback returns a context with the password callback set.
func WithPasswordCallback(ctx context.Context, cb PasswordCallback) context.Context {
	return context.WithValue(ctx, ctxKeyPasswordCallback, cb)
}

// HasPasswordCallback returns true if a password callback was set in the context.
func HasPasswordCallback(ctx context.Context) bool {
	_, ok := ctx.Value(ctxKeyPasswordCallback).(PasswordCallback)

	return ok
}

// GetPasswordCallback returns the password callback or a default (which always fails).
func GetPasswordCallback(ctx context.Context) PasswordCallback {
	pwcb, ok := ctx.Value(ctxKeyPasswordCallback).(PasswordCallback)
	if !ok || pwcb == nil {
		return func(string, bool) ([]byte, error) {
			return nil, ErrNoCallback
		}
	}

	return pwcb
}

// PasswordPurgeCallback is a callback to purge a password cached by PasswordCallback.
type PasswordPurgeCallback func(string)

// WithPasswordPurgeCallback returns a context with the password purge callback set.
func WithPasswordPurgeCallback(ctx context.Context, cb PasswordPurgeCallback) context.Context {
	return context.WithValue(ctx, ctxKeyPasswordPurgeCallback, cb)
}

// HasPasswordPurgeCallback returns true if a password purge callback was set in the context.
func HasPasswordPurgeCallback(ctx context.Context) bool {
	_, ok := ctx.Value(ctxKeyPasswordPurgeCallback).(PasswordPurgeCallback)

	return ok
}

// GetPasswordPurgeCallback returns the password purge callback or a default (which is a no-op).
func GetPasswordPurgeCallback(ctx context.Context) PasswordPurgeCallback {
	ppcb, ok := ctx.Value(ctxKeyPasswordPurgeCallback).(PasswordPurgeCallback)
	if !ok || ppcb == nil {
		return func(string) {}
	}

	return ppcb
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

// WithHidden returns a context with the flag value for hidden set.
func WithHidden(ctx context.Context, hidden bool) context.Context {
	return context.WithValue(ctx, ctxKeyHidden, hidden)
}

// IsHidden returns true if any output should be hidden in this context.
func IsHidden(ctx context.Context) bool {
	bv, ok := ctx.Value(ctxKeyHidden).(bool)
	if !ok {
		return false
	}

	return bv
}
