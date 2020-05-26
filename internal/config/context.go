package config

import (
	"context"

	"github.com/gopasspw/gopass/pkg/ctxutil"
)

// WithContext returns a context with all config options set for this store
// config, iff they have not been already set in the context
func (c *Config) WithContext(ctx context.Context) context.Context {
	if !ctxutil.HasAutoClip(ctx) {
		ctx = ctxutil.WithAutoClip(ctx, c.AutoClip)
	}
	if !c.AutoImport {
		ctx = ctxutil.WithImportFunc(ctx, nil)
	}
	if !ctxutil.HasClipTimeout(ctx) {
		ctx = ctxutil.WithClipTimeout(ctx, c.ClipTimeout)
	}
	if !ctxutil.HasEditRecipients(ctx) {
		ctx = ctxutil.WithEditRecipients(ctx, c.EditRecipients)
	}
	if !ctxutil.HasExportKeys(ctx) {
		ctx = ctxutil.WithExportKeys(ctx, c.ExportKeys)
	}
	if !ctxutil.HasConfirm(ctx) {
		ctx = ctxutil.WithConfirm(ctx, c.ConfirmRecipients)
	}
	if !ctxutil.HasNoPager(ctx) {
		ctx = ctxutil.WithNoPager(ctx, c.NoPager)
	}
	if !ctxutil.HasNotifications(ctx) {
		ctx = ctxutil.WithNotifications(ctx, c.Notifications)
	}
	if !ctxutil.HasShowSafeContent(ctx) {
		ctx = ctxutil.WithShowSafeContent(ctx, c.SafeContent)
	}
	// always disable autoclip when redirecting stdout
	if !ctxutil.IsTerminal(ctx) {
		ctx = ctxutil.WithAutoClip(ctx, false)
	}
	return ctx
}
