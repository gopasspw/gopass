package main

import (
	"bytes"
	"context"
	//"io"
	"os"
	//"strings"
	"testing"

	"github.com/fatih/color"
	"github.com/gopasspw/gopass/internal/gptest"
	"github.com/gopasspw/gopass/internal/out"
	"github.com/gopasspw/gopass/internal/termio"
	//"github.com/gopasspw/gopass/pkg/ctxutil"
	"github.com/gopasspw/gopass/pkg/gopass/apimock"
	//"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)


func TestSummonProviderOutput(t *testing.T) {
	ctx := context.Background()
	act := &gc{
		gp: apimock.New(),
	}
	require.NoError(t, act.gp.Set(ctx, "foo", &apimock.Secret{Buf: []byte("bar")}))

	stdout := &bytes.Buffer{}
	out.Stdout = stdout
	color.NoColor = true
	defer func() {
		out.Stdout = os.Stdout
		termio.Stdin = os.Stdin
	}()

	c := gptest.CliCtx(ctx, t)
	_ = c
}
