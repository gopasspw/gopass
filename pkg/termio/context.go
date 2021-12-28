package termio

import "context"

type contextKey int

const (
	ctxKeyPassPromptFunc contextKey = iota
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

// GetPassPromptFunc will return the password prompt func or a default one
// Note: will never return nil.
func GetPassPromptFunc(ctx context.Context) PassPromptFunc {
	ppf, ok := ctx.Value(ctxKeyPassPromptFunc).(PassPromptFunc)
	if !ok || ppf == nil {
		return promptPass
	}
	return ppf
}
