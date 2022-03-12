package config

import (
	"context"

	"github.com/gopasspw/gopass/pkg/ctxutil"
)

// WithContext returns a context with all config options set for this store
// config, iff they have not been already set in the context.
func (c *Config) WithContext(ctx context.Context) context.Context {
	if !c.AutoImport {
		ctx = ctxutil.WithImportFunc(ctx, nil)
	}

	if !ctxutil.HasExportKeys(ctx) {
		ctx = ctxutil.WithExportKeys(ctx, c.ExportKeys)
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

	if !ctxutil.HasShowParsing(ctx) {
		ctx = ctxutil.WithShowParsing(ctx, c.Parsing)
	}

	return ctx
}
