package config

import (
	"context"

	"github.com/gopasspw/gitconfig"
	"github.com/gopasspw/gopass/pkg/debug"
)

type contextKey int

const (
	ctxKeyConfig contextKey = iota
	ctxKeyMountPoint
)

func (c *Config) WithConfig(ctx context.Context) context.Context {
	return context.WithValue(ctx, ctxKeyConfig, c)
}

func WithMount(ctx context.Context, mp string) context.Context {
	return context.WithValue(ctx, ctxKeyMountPoint, mp)
}

// FromContext returns a config from a context, as well as the current mount point (store name) if found.
func FromContext(ctx context.Context) (*Config, string) {
	mount := ""
	if m, found := ctx.Value(ctxKeyMountPoint).(string); found && m != "" {
		mount = m
	}

	if c, found := ctx.Value(ctxKeyConfig).(*Config); found && c != nil {
		return c, mount
	}

	debug.Log("no config in context, loading anew")

	cfg := &Config{
		root: newGitconfig().LoadAll(""),
	}
	cfg.root.Preset = gitconfig.NewFromMap(defaults)

	return cfg, mount
}
