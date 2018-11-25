package client

import "context"

type contextKey int

const (
	ctxKeyClient contextKey = iota
)

// WithClient returns a context with a client instance set.
func WithClient(ctx context.Context, c *Client) context.Context {
	return context.WithValue(ctx, ctxKeyClient, c)
}

// GetClient returns a client instance, if set. May be nil.
func GetClient(ctx context.Context) *Client {
	c, ok := ctx.Value(ctxKeyClient).(*Client)
	if !ok {
		return nil
	}
	return c
}
