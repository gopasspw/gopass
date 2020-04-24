package main

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"runtime"
	"testing"

	"github.com/blang/semver"
	"github.com/gopasspw/gopass/pkg/ctxutil"
	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli/v2"
)

func TestGlobalFlags(t *testing.T) {
	ctx := context.Background()
	app := cli.NewApp()

	fs := flag.NewFlagSet("default", flag.ContinueOnError)
	sf := cli.BoolFlag{
		Name:  "yes",
		Usage: "yes",
	}
	assert.NoError(t, sf.Apply(fs))
	assert.NoError(t, fs.Parse([]string{"--yes"}))
	c := cli.NewContext(app, fs, nil)

	assert.Equal(t, true, ctxutil.IsAlwaysYes(withGlobalFlags(ctx, c)))
}

func TestVersionPrinter(t *testing.T) {
	buf := &bytes.Buffer{}
	vp := makeVersionPrinter(buf, semver.Version{Major: 1})
	vp(nil)
	assert.Equal(t, fmt.Sprintf("gopass 1.0.0 %s %s %s\n", runtime.Version(), runtime.GOOS, runtime.GOARCH), buf.String())
}
