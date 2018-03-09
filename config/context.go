package config

import (
	"context"

	"github.com/justwatchcom/gopass/store/sub"
	"github.com/justwatchcom/gopass/utils/ctxutil"
)

// WithContext returns a context with all config options set for this store
// config, iff they have not been already set in the context
func (c StoreConfig) WithContext(ctx context.Context) context.Context {
	if !ctxutil.HasAskForMore(ctx) {
		ctx = ctxutil.WithAskForMore(ctx, c.AskForMore)
	}
	if !c.AutoImport {
		ctx = sub.WithImportFunc(ctx, nil)
	}
	if !sub.HasAutoSync(ctx) {
		ctx = sub.WithAutoSync(ctx, c.AutoSync)
	}
	if !ctxutil.HasEditRecipients(ctx) {
		ctx = ctxutil.WithEditRecipients(ctx, c.EditRecipients)
	}
	if !ctxutil.HasClipTimeout(ctx) {
		ctx = ctxutil.WithClipTimeout(ctx, c.ClipTimeout)
	}
	if !ctxutil.HasNoConfirm(ctx) {
		ctx = ctxutil.WithNoConfirm(ctx, c.NoConfirm)
	}
	if !ctxutil.HasShowSafeContent(ctx) {
		ctx = ctxutil.WithShowSafeContent(ctx, c.SafeContent)
	}
	return ctx
}
