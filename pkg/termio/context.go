package termio

import "context"

type contextKey int

const (
	ctxKeyPassPromptFunc contextKey = iota
	ctxKeyWorkdir
)

// PassPromptFunc is a password prompt function.
type PassPromptFunc func(context.Context, string) (string, error)

// WithPassPromptFunc returns a context with the password prompt function set.
func WithPassPromptFunc(ctx context.Context, ppf PassPromptFunc) context.Context {
	return context.WithValue(ctx, ctxKeyPassPromptFunc, ppf)
}

// HasPassPromptFunc returns true if a value for the pass prompt func has been
// set in this context.
func HasPassPromptFunc(ctx context.Context) bool {
	ppf, ok := ctx.Value(ctxKeyPassPromptFunc).(PassPromptFunc)

	return ok && ppf != nil
}

// GetPassPromptFunc will return the password prompt func or a default one.
// Note: will never return nil.
func GetPassPromptFunc(ctx context.Context) PassPromptFunc {
	ppf, ok := ctx.Value(ctxKeyPassPromptFunc).(PassPromptFunc)
	if !ok || ppf == nil {
		return promptPass
	}

	return ppf
}

// WithWorkdir returns a context with the working directory option set.
// The working directory is used to resolve relative paths.
func WithWorkdir(ctx context.Context, dir string) context.Context {
	return context.WithValue(ctx, ctxKeyWorkdir, dir)
}

// GetWorkdir returns the working directory from the context or an empty
// string if it is not set.
// The working directory is used to resolve relative paths.
func GetWorkdir(ctx context.Context) string {
	sv, ok := ctx.Value(ctxKeyWorkdir).(string)
	if !ok {
		return ""
	}

	return sv
}
