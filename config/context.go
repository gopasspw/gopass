package config

import (
	"context"

	"github.com/justwatchcom/gopass/store/sub"
	"github.com/justwatchcom/gopass/utils/ctxutil"
)

// WithContext returns a context with all config options set for this store
// config
func (c StoreConfig) WithContext(ctx context.Context) context.Context {
	ctx = ctxutil.WithAskForMore(ctx, c.AskForMore)
	if !c.AutoImport {
		ctx = sub.WithImportFunc(ctx, nil)
	}
	ctx = sub.WithAutoSync(ctx, c.AutoSync)
	ctx = ctxutil.WithClipTimeout(ctx, c.ClipTimeout)
	ctx = ctxutil.WithNoConfirm(ctx, c.NoConfirm)
	ctx = ctxutil.WithShowSafeContent(ctx, c.SafeContent)
	return ctx
}
