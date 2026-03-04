package fossilfs

import "context"

type contextKey int

const (
	ctxKeyPathOverride contextKey = iota
)

func withPathOverride(ctx context.Context, path string) context.Context {
	return context.WithValue(ctx, ctxKeyPathOverride, path)
}

func getPathOverride(ctx context.Context, def string) string {
	if sv, ok := ctx.Value(ctxKeyPathOverride).(string); ok && sv != "" {
		return sv
	}

	return def
}
