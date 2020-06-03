package ctxutil

import "context"

// hasString is a helper function for checking if a string has been set in
// the provided context.
func hasString(ctx context.Context, key contextKey) bool {
	_, ok := ctx.Value(key).(string)
	return ok
}

// hasInt is a helper function for checking if a int has been set in
// the provided context.
func hasInt(ctx context.Context, key contextKey) bool {
	_, ok := ctx.Value(key).(int)
	return ok
}

// hasBool is a helper function for checking if a bool has been set in
// the provided context.
func hasBool(ctx context.Context, key contextKey) bool {
	_, ok := ctx.Value(key).(bool)
	return ok
}

// is is a helper function for returning the value of a bool from the context
// or the provided default.
func is(ctx context.Context, key contextKey, def bool) bool {
	bv, ok := ctx.Value(key).(bool)
	if !ok {
		return def
	}
	return bv
}
