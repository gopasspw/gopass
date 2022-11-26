package config

import (
	"context"

	"github.com/gopasspw/gopass/pkg/gitconfig"
)

type contextKey int

const (
	ctxKeyConfig contextKey = iota
)

func (c *Config) WithConfig(ctx context.Context) context.Context {
	return context.WithValue(ctx, ctxKeyConfig, c)
}

func FromContext(ctx context.Context) *Config {
	if c, found := ctx.Value(ctxKeyConfig).(*Config); found && c != nil {
		return c
	}

	c := &Config{
		root: newGitconfig().LoadAll(""),
	}
	c.root.Preset = gitconfig.NewFromMap(defaults)

	return c
}
