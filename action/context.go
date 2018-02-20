package action

import "context"

type contextKey int

const (
	ctxKeyClip contextKey = iota
	ctxKeyForce
	ctxKeyPasswordOnly
	ctxKeyPrintQR
	ctxKeyRevision
)

// WithClip returns a context with the value for clip (for copy to clipboard)
// set
func WithClip(ctx context.Context, clip bool) context.Context {
	return context.WithValue(ctx, ctxKeyClip, clip)
}

// IsClip returns the value of clip or the default (false)
func IsClip(ctx context.Context) bool {
	bv, ok := ctx.Value(ctxKeyClip).(bool)
	if !ok {
		return false
	}
	return bv
}

// WithForce returns a context with the value for force set
func WithForce(ctx context.Context, force bool) context.Context {
	return context.WithValue(ctx, ctxKeyForce, force)
}

// IsForce returns the value of force or the default (false)
func IsForce(ctx context.Context) bool {
	bv, ok := ctx.Value(ctxKeyForce).(bool)
	if !ok {
		return false
	}
	return bv
}

// WithPasswordOnly returns a context with the value of password only set
func WithPasswordOnly(ctx context.Context, pw bool) context.Context {
	return context.WithValue(ctx, ctxKeyPasswordOnly, pw)
}

// IsPasswordOnly returns the value of password only or the default (false)
func IsPasswordOnly(ctx context.Context) bool {
	bv, ok := ctx.Value(ctxKeyPasswordOnly).(bool)
	if !ok {
		return false
	}
	return bv
}

// WithPrintQR returns a context with the value of print QR set
func WithPrintQR(ctx context.Context, qr bool) context.Context {
	return context.WithValue(ctx, ctxKeyPrintQR, qr)
}

// IsPrintQR returns the value of print QR or the default (false)
func IsPrintQR(ctx context.Context) bool {
	bv, ok := ctx.Value(ctxKeyPrintQR).(bool)
	if !ok {
		return false
	}
	return bv
}

// WithRevision returns a context withe the value of revision set
func WithRevision(ctx context.Context, rev string) context.Context {
	return context.WithValue(ctx, ctxKeyRevision, rev)
}

// HasRevision returns true if a value for revision was set in this context
func HasRevision(ctx context.Context) bool {
	sv, ok := ctx.Value(ctxKeyRevision).(string)
	return ok && sv != ""
}

// GetRevision returns the revison set in this context or an empty string
func GetRevision(ctx context.Context) string {
	sv, ok := ctx.Value(ctxKeyRevision).(string)
	if !ok {
		return ""
	}
	return sv
}
