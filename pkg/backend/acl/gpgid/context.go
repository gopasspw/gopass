package gpgid

import "context"

type contextKey int

const (
	ctxKeyCommitMsg contextKey = iota
)

func withCommitMsg(ctx context.Context, msg string) context.Context {
	return context.WithValue(ctx, ctxKeyCommitMsg, msg)
}

func getCommitMsg(ctx context.Context) string {
	sv, ok := ctx.Value(ctxKeyCommitMsg).(string)
	if !ok {
		return ""
	}
	return sv
}
