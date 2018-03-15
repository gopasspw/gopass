package main

import (
	"context"
	"flag"
	"testing"

	"github.com/justwatchcom/gopass/pkg/ctxutil"
	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli"
)

func TestGlobalFlags(t *testing.T) {
	ctx := context.Background()
	app := cli.NewApp()

	fs := flag.NewFlagSet("default", flag.ContinueOnError)
	sf := cli.BoolFlag{
		Name:  "yes",
		Usage: "yes",
	}
	assert.NoError(t, sf.ApplyWithError(fs))
	assert.NoError(t, fs.Parse([]string{"--yes"}))
	c := cli.NewContext(app, fs, nil)

	assert.Equal(t, true, ctxutil.IsAlwaysYes(withGlobalFlags(ctx, c)))
}
