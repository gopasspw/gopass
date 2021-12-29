package leaf

import (
	"context"

	"github.com/gopasspw/gopass/internal/store"
)

type contextKey int

const (
	ctxKeyFsckCheck contextKey = iota
	ctxKeyFsckForce
	ctxKeyFsckFunc
	ctxKeyCheckRecipients
	ctxKeyFsckDecrypt
	ctxKeyNoGitOps
)

// WithFsckCheck returns a context with the flag for fscks check set.
func WithFsckCheck(ctx context.Context, check bool) context.Context {
	return context.WithValue(ctx, ctxKeyFsckCheck, check)
}

// HasFsckCheck returns true if a value for fsck check has been set in this
// context.
func HasFsckCheck(ctx context.Context) bool {
	return hasBool(ctx, ctxKeyFsckCheck)
}

// IsFsckCheck returns the value of fsck check.
func IsFsckCheck(ctx context.Context) bool {
	bv, ok := ctx.Value(ctxKeyFsckCheck).(bool)
	if !ok {
		return false
	}
	return bv
}

// WithFsckForce returns a context with the flag for fsck force set.
func WithFsckForce(ctx context.Context, force bool) context.Context {
	return context.WithValue(ctx, ctxKeyFsckForce, force)
}

// HasFsckForce returns true if a value for fsck force has been set in this
// context.
func HasFsckForce(ctx context.Context) bool {
	return hasBool(ctx, ctxKeyFsckForce)
}

// IsFsckForce returns the value of fsck force.
func IsFsckForce(ctx context.Context) bool {
	bv, ok := ctx.Value(ctxKeyFsckForce).(bool)
	if !ok {
		return false
	}
	return bv
}

// WithFsckFunc will return a context with the fsck confirmation callback set.
func WithFsckFunc(ctx context.Context, imf store.FsckCallback) context.Context {
	return context.WithValue(ctx, ctxKeyFsckFunc, imf)
}

// HasFsckFunc returns true if a fsck func has been set in this context.
func HasFsckFunc(ctx context.Context) bool {
	imf, ok := ctx.Value(ctxKeyFsckFunc).(store.FsckCallback)
	return ok && imf != nil
}

// GetFsckFunc will return the fsck confirmation callback or a default one
// returning true.
// Note: will never return nil.
func GetFsckFunc(ctx context.Context) store.FsckCallback {
	imf, ok := ctx.Value(ctxKeyFsckFunc).(store.FsckCallback)
	if !ok || imf == nil {
		return func(context.Context, string) bool {
			return true
		}
	}
	return imf
}

// WithCheckRecipients will return a context with the flag for check recipients
// set.
func WithCheckRecipients(ctx context.Context, cr bool) context.Context {
	return context.WithValue(ctx, ctxKeyCheckRecipients, cr)
}

// HasCheckRecipients returns true if check recipients has been set in this
// context.
func HasCheckRecipients(ctx context.Context) bool {
	return hasBool(ctx, ctxKeyCheckRecipients)
}

// IsCheckRecipients will return the value of the check recipients flag or the
// default value (false).
func IsCheckRecipients(ctx context.Context) bool {
	bv, ok := ctx.Value(ctxKeyCheckRecipients).(bool)
	if !ok {
		return false
	}
	return bv
}

// WithFsckDecrypt will return a context with the value for the decrypt
// during fsck flag set.
func WithFsckDecrypt(ctx context.Context, d bool) context.Context {
	return context.WithValue(ctx, ctxKeyFsckDecrypt, d)
}

// IsFsckDecrypt will return the value for the decrypt during fsck, defaulting
// to false.
func IsFsckDecrypt(ctx context.Context) bool {
	return is(ctx, ctxKeyFsckDecrypt, false)
}

// WithNoGitOps returns a context with the value for NoGitOps set.
// This will skip any git operations in concurrent goroutines.
func WithNoGitOps(ctx context.Context, d bool) context.Context {
	return context.WithValue(ctx, ctxKeyNoGitOps, d)
}

// IsNoGitOps returns the value for NoGitOps from the context
// or the default (false).
func IsNoGitOps(ctx context.Context) bool {
	return is(ctx, ctxKeyNoGitOps, false)
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
