package main

import (
	"bytes"
	"context"
	"os"
	"testing"

	"github.com/gopasspw/gopass/internal/gptest"
	"github.com/gopasspw/gopass/internal/out"
	"github.com/gopasspw/gopass/pkg/ctxutil"

	"github.com/stretchr/testify/assert"
)

func TestJSONAPI(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ctx = ctxutil.WithAlwaysYes(ctx, true)

	act := &jsonapiCLI{}

	buf := &bytes.Buffer{}
	out.Stdout = buf
	defer func() {
		out.Stdout = os.Stdout
	}()

	assert.NoError(t, act.listen(gptest.CliCtx(ctx, t)))
	buf.Reset()

	b, err := act.getBrowser(ctx, gptest.CliCtx(ctx, t))
	assert.NoError(t, err)
	assert.Equal(t, b, "chrome")
	buf.Reset()
}
