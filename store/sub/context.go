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

// HasFsckCheck returns true if a value for fsck check has been set in this
// context
func HasFsckCheck(ctx context.Context) bool {
	_, ok := ctx.Value(ctxKeyFsckCheck).(bool)
	return ok
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

// HasFsckForce returns true if a value for fsck force has been set in this
// context
func HasFsckForce(ctx context.Context) bool {
	_, ok := ctx.Value(ctxKeyFsckForce).(bool)
	return ok
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

// HasAutoSync has been set if a value for auto sync has been set in this
// context
func HasAutoSync(ctx context.Context) bool {
	_, ok := ctx.Value(ctxKeyAutoSync).(bool)
	return ok
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

// HasReason returns true if a value for reason has been set in this context
func HasReason(ctx context.Context) bool {
	_, ok := ctx.Value(ctxKeyReason).(bool)
	return ok
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

// HasRecipientFunc returns true if a recipient func has been set in this
// context
func HasRecipientFunc(ctx context.Context) bool {
	imf, ok := ctx.Value(ctxKeyRecipientFunc).(store.RecipientCallback)
	return ok && imf != nil
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

// HasFsckFunc returns true if a fsck func has been set in this context
func HasFsckFunc(ctx context.Context) bool {
	imf, ok := ctx.Value(ctxKeyFsckFunc).(store.FsckCallback)
	return ok && imf != nil
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

// HasCheckRecipients returns true if check recipients has been set in this
// context
func HasCheckRecipients(ctx context.Context) bool {
	_, ok := ctx.Value(ctxKeyCheckRecipients).(bool)
	return ok
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
