package sub

import (
	"context"

	"github.com/justwatchcom/gopass/store"
)

type contextKey int

const (
	ctxKeyFsckCheck contextKey = iota
	ctxKeyFsckForce
	ctxKeyAutoSync
	ctxKeyReason
	ctxKeyImportFunc
	ctxKeyRecipientFunc
	ctxKeyFsckFunc
	ctxKeyCheckRecipients
)

// WithFsckCheck returns a context with the flag for fscks check set
func WithFsckCheck(ctx context.Context, check bool) context.Context {
	return context.WithValue(ctx, ctxKeyFsckCheck, check)
}

// IsFsckCheck returns the value of fsck check
func IsFsckCheck(ctx context.Context) bool {
	bv, ok := ctx.Value(ctxKeyFsckCheck).(bool)
	if !ok {
		return false
	}
	return bv
}

// WithFsckForce returns a context with the flag for fsck force set
func WithFsckForce(ctx context.Context, force bool) context.Context {
	return context.WithValue(ctx, ctxKeyFsckForce, force)
}

// IsFsckForce returns the value of fsck force
func IsFsckForce(ctx context.Context) bool {
	bv, ok := ctx.Value(ctxKeyFsckForce).(bool)
	if !ok {
		return false
	}
	return bv
}

// WithAutoSync returns a context with the flag for autosync set
func WithAutoSync(ctx context.Context, sync bool) context.Context {
	return context.WithValue(ctx, ctxKeyAutoSync, sync)
}

// IsAutoSync returns the value of autosync
func IsAutoSync(ctx context.Context) bool {
	bv, ok := ctx.Value(ctxKeyAutoSync).(bool)
	if !ok {
		return true
	}
	return bv
}

// WithReason returns a context with a commit/change reason set
func WithReason(ctx context.Context, msg string) context.Context {
	return context.WithValue(ctx, ctxKeyReason, msg)
}

// GetReason returns the value of reason
func GetReason(ctx context.Context) string {
	sv, ok := ctx.Value(ctxKeyReason).(string)
	if !ok {
		return ""
	}
	return sv
}

// WithImportFunc will return a context with the import callback set
func WithImportFunc(ctx context.Context, imf store.ImportCallback) context.Context {
	return context.WithValue(ctx, ctxKeyImportFunc, imf)
}

// GetImportFunc will return the import callback or a default one returning true
// Note: will never return nil
func GetImportFunc(ctx context.Context) store.ImportCallback {
	imf, ok := ctx.Value(ctxKeyImportFunc).(store.ImportCallback)
	if !ok || imf == nil {
		return func(context.Context, string) bool {
			return true
		}
	}
	return imf
}

// WithRecipientFunc will return a context with the recipient callback set
func WithRecipientFunc(ctx context.Context, imf store.RecipientCallback) context.Context {
	return context.WithValue(ctx, ctxKeyRecipientFunc, imf)
}

// GetRecipientFunc will return the recipient callback or a default one returning
// the unaltered list of recipients.
// Note: will never return nil
func GetRecipientFunc(ctx context.Context) store.RecipientCallback {
	imf, ok := ctx.Value(ctxKeyRecipientFunc).(store.RecipientCallback)
	if !ok || imf == nil {
		return func(ctx context.Context, msg string, rs []string) ([]string, error) {
			return rs, nil
		}
	}
	return imf
}

// WithFsckFunc will return a context with the fsck confirmation callback set
func WithFsckFunc(ctx context.Context, imf store.FsckCallback) context.Context {
	return context.WithValue(ctx, ctxKeyFsckFunc, imf)
}

// GetFsckFunc will return the fsck confirmation callback or a default one
// returning true.
// Note: will never return nil
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
// set
func WithCheckRecipients(ctx context.Context, cr bool) context.Context {
	return context.WithValue(ctx, ctxKeyCheckRecipients, cr)
}

// IsCheckRecipients will return the value of the check recipients flag or the
// default value (false)
func IsCheckRecipients(ctx context.Context) bool {
	bv, ok := ctx.Value(ctxKeyCheckRecipients).(bool)
	if !ok {
		return false
	}
	return bv
}
