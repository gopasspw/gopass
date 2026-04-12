// Package ctxutil provides a set of functions to manage context values
// in a gopass application. It allows to set and get values in the context.
package ctxutil

import (
	"context"
	"time"

	"github.com/urfave/cli/v2"
)

type contextKey int

const (
	// ctxKeyExecConfig holds the consolidated ExecConfig struct.
	ctxKeyExecConfig contextKey = iota
)

// ExecConfig holds all boolean/string/struct configuration flags that are
// threaded through the call stack. Using a single typed struct instead of
// individual context keys makes dependencies explicit at the type level and
// eliminates the boilerplate of per-field With*/Is*/Has*/Get* helpers.
type ExecConfig struct {
	// Terminal indicates whether output is going to a terminal.
	Terminal *bool
	// Interactive indicates whether user interaction is possible.
	Interactive *bool
	// Stdin indicates whether data is being piped on stdin.
	Stdin *bool
	// GitCommit indicates whether changes should be committed to git.
	GitCommit *bool
	// AlwaysYes indicates that all prompts should be answered with yes.
	AlwaysYes *bool
	// GitInit indicates whether a new git repository should be initialized.
	GitInit *bool
	// Force indicates that operations that would otherwise fail should be forced.
	Force *bool
	// NoNetwork indicates that no network operations should be performed.
	NoNetwork *bool
	// ShowParsing indicates whether parsing errors should be shown.
	ShowParsing *bool
	// Hidden indicates whether output should be hidden.
	Hidden *bool
	// FollowRef indicates whether symlink references should be followed.
	FollowRef *bool
	// Alias is the mount alias for the current store.
	Alias string
	// Username is the configured user name.
	Username string
	// Email is the configured e-mail address.
	Email string
	// CommitMessage is the structured git commit message.
	CommitMessage *HeadedText
	// CommitTimestamp is an optional fixed timestamp for reproducible commits.
	CommitTimestamp *time.Time
	// AgePassphrase is an optional passphrase used for the age identity file.
	// It is set from the GOPASS_AGE_PASSWORD environment variable and is used
	// to provide a password for the age encryption backend without interactive
	// prompts (e.g. in CI/CD pipelines or tests).
	AgePassphrase string
	// SetupRemote is set to the remote URL when a git remote is specified during
	// setup, signalling that the automatic initial commit should be suppressed.
	SetupRemote string
}

// WithExecConfig stores e in ctx and returns the updated context.
func WithExecConfig(ctx context.Context, e ExecConfig) context.Context {
	return context.WithValue(ctx, ctxKeyExecConfig, e)
}

// GetExecConfig retrieves the ExecConfig from ctx, returning a zero value if
// none has been set.
func GetExecConfig(ctx context.Context) ExecConfig {
	if e, ok := ctx.Value(ctxKeyExecConfig).(ExecConfig); ok {
		return e
	}

	return ExecConfig{}
}

// withBool is a helper that clones the ExecConfig, sets the field pointed to
// by set, and stores the result back into ctx.
func withBool(ctx context.Context, set func(*ExecConfig, *bool), v bool) context.Context {
	e := GetExecConfig(ctx)
	set(&e, &v)

	return WithExecConfig(ctx, e)
}

// WithGlobalFlags parses any global flags from the cli context and returns
// a regular context. It handles the --yes flag and sets the appropriate
// context value.
func WithGlobalFlags(c *cli.Context) context.Context {
	if c.Bool("yes") {
		return WithAlwaysYes(c.Context, true)
	}

	return c.Context
}

// ProgressCallback is a callback for updating progress.
type ProgressCallback func()

// WithTerminal returns a context with an explicit value for whether or not we are
// in a terminal.
func WithTerminal(ctx context.Context, isTerm bool) context.Context {
	return withBool(ctx, func(e *ExecConfig, v *bool) { e.Terminal = v }, isTerm)
}

// HasTerminal returns true if a value for Terminal has been set in this context.
func HasTerminal(ctx context.Context) bool {
	return GetExecConfig(ctx).Terminal != nil
}

// IsTerminal returns the value of terminal or the default (true).
func IsTerminal(ctx context.Context) bool {
	if v := GetExecConfig(ctx).Terminal; v != nil {
		return *v
	}

	return true
}

// WithInteractive returns a context with an explicit value for whether or not we are
// in an interactive session.
func WithInteractive(ctx context.Context, isInteractive bool) context.Context {
	return withBool(ctx, func(e *ExecConfig, v *bool) { e.Interactive = v }, isInteractive)
}

// HasInteractive returns true if a value for Interactive has been set in this context.
func HasInteractive(ctx context.Context) bool {
	return GetExecConfig(ctx).Interactive != nil
}

// IsInteractive returns the value of interactive or the default (true).
func IsInteractive(ctx context.Context) bool {
	if v := GetExecConfig(ctx).Interactive; v != nil {
		return *v
	}

	return true
}

// WithStdin returns a context with the value for Stdin set. If true some input
// is available on Stdin (e.g. something is being piped into it).
func WithStdin(ctx context.Context, isStdin bool) context.Context {
	return withBool(ctx, func(e *ExecConfig, v *bool) { e.Stdin = v }, isStdin)
}

// HasStdin returns true if a value for Stdin has been set in this context.
func HasStdin(ctx context.Context) bool {
	return GetExecConfig(ctx).Stdin != nil
}

// IsStdin returns the value of stdin, i.e. if it's true some data is being
// piped to stdin. If not set it returns the default value (false).
func IsStdin(ctx context.Context) bool {
	if v := GetExecConfig(ctx).Stdin; v != nil {
		return *v
	}

	return false
}

// WithShowParsing returns a context with the value for ShowParsing set.
// This is used to control whether to show parsing errors.
func WithShowParsing(ctx context.Context, bv bool) context.Context {
	return withBool(ctx, func(e *ExecConfig, v *bool) { e.ShowParsing = v }, bv)
}

// HasShowParsing returns true if a value for ShowParsing has been set in this context.
func HasShowParsing(ctx context.Context) bool {
	return GetExecConfig(ctx).ShowParsing != nil
}

// IsShowParsing returns the value of ShowParsing or the default (true).
func IsShowParsing(ctx context.Context) bool {
	if v := GetExecConfig(ctx).ShowParsing; v != nil {
		return *v
	}

	return true
}

// WithGitCommit returns a context with the value of git commit set.
// If true, changes will be committed to git.
func WithGitCommit(ctx context.Context, bv bool) context.Context {
	return withBool(ctx, func(e *ExecConfig, v *bool) { e.GitCommit = v }, bv)
}

// HasGitCommit returns true if a value for GitCommit has been set in this context.
func HasGitCommit(ctx context.Context) bool {
	return GetExecConfig(ctx).GitCommit != nil
}

// IsGitCommit returns the value of git commit or the default (true).
func IsGitCommit(ctx context.Context) bool {
	if v := GetExecConfig(ctx).GitCommit; v != nil {
		return *v
	}

	return true
}

// IsFollowRef returns the value of follow-ref or the default (false).
// If true, symlinks will be followed.
func IsFollowRef(ctx context.Context) bool {
	if v := GetExecConfig(ctx).FollowRef; v != nil {
		return *v
	}

	return false
}

// HasFollowRef returns true if a value for follow-ref has been set in this context.
func HasFollowRef(ctx context.Context) bool {
	return GetExecConfig(ctx).FollowRef != nil
}

// WithFollowRef returns a context with the value of follow-ref set.
func WithFollowRef(ctx context.Context, bv bool) context.Context {
	return withBool(ctx, func(e *ExecConfig, v *bool) { e.FollowRef = v }, bv)
}

// WithAlwaysYes returns a context with the value of always yes set.
// If true, any prompts will be answered with yes.
func WithAlwaysYes(ctx context.Context, bv bool) context.Context {
	return withBool(ctx, func(e *ExecConfig, v *bool) { e.AlwaysYes = v }, bv)
}

// HasAlwaysYes returns true if a value for AlwaysYes has been set in this context.
func HasAlwaysYes(ctx context.Context) bool {
	return GetExecConfig(ctx).AlwaysYes != nil
}

// IsAlwaysYes returns the value of always yes or the default (false).
func IsAlwaysYes(ctx context.Context) bool {
	if v := GetExecConfig(ctx).AlwaysYes; v != nil {
		return *v
	}

	return false
}

// WithAlias returns a context with the alias set.
func WithAlias(ctx context.Context, alias string) context.Context {
	e := GetExecConfig(ctx)
	e.Alias = alias

	return WithExecConfig(ctx, e)
}

// HasAlias returns true if a value for alias has been set.
func HasAlias(ctx context.Context) bool {
	return GetExecConfig(ctx).Alias != ""
}

// GetAlias returns an alias if it has been set or an empty string otherwise.
func GetAlias(ctx context.Context) string {
	return GetExecConfig(ctx).Alias
}

// WithGitInit returns a context with the value for the git init flag set.
// If true, a git repository will be initialized.
func WithGitInit(ctx context.Context, bv bool) context.Context {
	return withBool(ctx, func(e *ExecConfig, v *bool) { e.GitInit = v }, bv)
}

// HasGitInit returns true if the git init flag was set.
func HasGitInit(ctx context.Context) bool {
	return GetExecConfig(ctx).GitInit != nil
}

// IsGitInit returns the value of the git init flag or true if none was set.
func IsGitInit(ctx context.Context) bool {
	if v := GetExecConfig(ctx).GitInit; v != nil {
		return *v
	}

	return true
}

// WithForce returns a context with the force flag set.
// If true, operations that would otherwise fail will be forced to succeed.
func WithForce(ctx context.Context, bv bool) context.Context {
	return withBool(ctx, func(e *ExecConfig, v *bool) { e.Force = v }, bv)
}

// HasForce returns true if the context has the force flag set.
func HasForce(ctx context.Context) bool {
	return GetExecConfig(ctx).Force != nil
}

// IsForce returns the force flag value of the default (false).
func IsForce(ctx context.Context) bool {
	if v := GetExecConfig(ctx).Force; v != nil {
		return *v
	}

	return false
}

// AddToCommitMessageBody returns a context with something added to the commit's body.
func AddToCommitMessageBody(ctx context.Context, sv string) context.Context {
	e := GetExecConfig(ctx)
	if e.CommitMessage == nil {
		var ht HeadedText
		e.CommitMessage = &ht
	}
	e.CommitMessage.AddToBody(sv)

	return WithExecConfig(ctx, e)
}

// HasCommitMessageBody returns true if the commit message body is nonempty.
func HasCommitMessageBody(ctx context.Context) bool {
	ht := GetExecConfig(ctx).CommitMessage
	if ht == nil {
		return false
	}

	return ht.HasBody()
}

// GetCommitMessageBody returns the set commit message body or an empty string.
func GetCommitMessageBody(ctx context.Context) string {
	ht := GetExecConfig(ctx).CommitMessage
	if ht == nil {
		return ""
	}

	return ht.GetBody()
}

// WithCommitMessage returns a context with a commit message (head) set.
// (full commit message is the commit message's body is not defined, commit message head otherwise).
func WithCommitMessage(ctx context.Context, sv string) context.Context {
	e := GetExecConfig(ctx)
	if e.CommitMessage == nil {
		var ht HeadedText
		e.CommitMessage = &ht
	}
	e.CommitMessage.SetHead(sv)

	return WithExecConfig(ctx, e)
}

// HasCommitMessage returns true if the commit message (head) was set.
func HasCommitMessage(ctx context.Context) bool {
	ht := GetExecConfig(ctx).CommitMessage

	return ht != nil && ht.head != "" // not the most intuitive answer, but a backwards-compatible one. for now.
}

// GetCommitMessage returns the set commit message (head) or an empty string.
func GetCommitMessage(ctx context.Context) string {
	ht := GetExecConfig(ctx).CommitMessage
	if ht == nil {
		return ""
	}

	return ht.head
}

// GetCommitMessageFull returns the set commit message (head+body, if either are defined) or an empty string.
func GetCommitMessageFull(ctx context.Context) string {
	ht := GetExecConfig(ctx).CommitMessage
	if ht == nil {
		return ""
	}

	return ht.GetText()
}

// WithNoNetwork returns a context with the value of no network set.
// If true, no network operations will be performed.
func WithNoNetwork(ctx context.Context, bv bool) context.Context {
	return withBool(ctx, func(e *ExecConfig, v *bool) { e.NoNetwork = v }, bv)
}

// HasNoNetwork returns true if no network was set.
func HasNoNetwork(ctx context.Context) bool {
	return GetExecConfig(ctx).NoNetwork != nil
}

// IsNoNetwork returns the value of no network or false.
func IsNoNetwork(ctx context.Context) bool {
	if v := GetExecConfig(ctx).NoNetwork; v != nil {
		return *v
	}

	return false
}

// WithUsername returns a context with the username set in the context.
func WithUsername(ctx context.Context, sv string) context.Context {
	e := GetExecConfig(ctx)
	e.Username = sv

	return WithExecConfig(ctx, e)
}

// GetUsername returns the username from the context.
func GetUsername(ctx context.Context) string {
	return GetExecConfig(ctx).Username
}

// WithEmail returns a context with the email set in the context.
func WithEmail(ctx context.Context, sv string) context.Context {
	e := GetExecConfig(ctx)
	e.Email = sv

	return WithExecConfig(ctx, e)
}

// GetEmail returns the email from the context.
func GetEmail(ctx context.Context) string {
	return GetExecConfig(ctx).Email
}

// WithAgePassphrase returns a context with the age passphrase set.
// This is used by the age crypto backend to encrypt/decrypt the identity file
// without interactive prompts (e.g. set from the GOPASS_AGE_PASSWORD env variable).
func WithAgePassphrase(ctx context.Context, pw string) context.Context {
	e := GetExecConfig(ctx)
	e.AgePassphrase = pw

	return WithExecConfig(ctx, e)
}

// GetAgePassphrase returns the age passphrase from the context, or an empty
// string if none has been set.
func GetAgePassphrase(ctx context.Context) string {
	return GetExecConfig(ctx).AgePassphrase
}

// WithCommitTimestamp returns a context with the value for the commit
// timestamp set.
// This is used to allow for reproducible builds.
func WithCommitTimestamp(ctx context.Context, ts time.Time) context.Context {
	e := GetExecConfig(ctx)
	e.CommitTimestamp = &ts

	return WithExecConfig(ctx, e)
}

// HasCommitTimestamp returns true if the value for the commit timestamp
// was set in the context.
func HasCommitTimestamp(ctx context.Context) bool {
	return GetExecConfig(ctx).CommitTimestamp != nil
}

// GetCommitTimestamp returns the commit timestamp from the context if
// set or the default (now) otherwise.
func GetCommitTimestamp(ctx context.Context) time.Time {
	if ts := GetExecConfig(ctx).CommitTimestamp; ts != nil {
		return *ts
	}

	return time.Now()
}

// WithHidden returns a context with the flag value for hidden set.
// This is used to hide secrets from the output.
func WithHidden(ctx context.Context, hidden bool) context.Context {
	return withBool(ctx, func(e *ExecConfig, v *bool) { e.Hidden = v }, hidden)
}

// IsHidden returns true if any output should be hidden in this context.
func IsHidden(ctx context.Context) bool {
	if v := GetExecConfig(ctx).Hidden; v != nil {
		return *v
	}

	return false
}

// WithSetupRemote returns a context with the remote URL stored for use during
// "gopass setup". When non-empty it signals that the automatic initial commit
// in gitfs.Init() should be suppressed.
func WithSetupRemote(ctx context.Context, remote string) context.Context {
	e := GetExecConfig(ctx)
	e.SetupRemote = remote

	return WithExecConfig(ctx, e)
}

// HasSetupRemote returns true if a non-empty setup remote was stored in ctx.
func HasSetupRemote(ctx context.Context) bool {
	return GetExecConfig(ctx).SetupRemote != ""
}
