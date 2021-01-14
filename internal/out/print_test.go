package out

import (
	"bytes"
	"context"
	"os"
	"testing"

	"github.com/fatih/color"
	"github.com/gopasspw/gopass/pkg/ctxutil"
	"github.com/stretchr/testify/assert"
)

func TestPrint(t *testing.T) {
	ctx := context.Background()
	buf := &bytes.Buffer{}
	Stdout = buf
	defer func() {
		Stdout = os.Stdout
	}()

	Print(ctx, "%s = %d", "foo", 42)
	assert.Equal(t, "foo = 42\n", buf.String())
	buf.Reset()

	Print(ctxutil.WithHidden(ctx, true), "%s = %d", "foo", 42)
	assert.Equal(t, "", buf.String())
	buf.Reset()

	Print(WithNewline(ctx, false), "%s = %d", "foo", 42)
	assert.Equal(t, "foo = 42", buf.String())
	buf.Reset()
}

func TestColor(t *testing.T) {
	ctx := context.Background()
	buf := &bytes.Buffer{}
	Stdout = buf
	defer func() {
		Stdout = os.Stdout
	}()
	color.NoColor = true

	for _, fn := range []func(context.Context, string, ...interface{}){
		Black,
		Blue,
		Cyan,
		Green,
		Magenta,
		Red,
		White,
		Yellow,
	} {
		buf.Reset()
		fn(ctx, "foobar")
		assert.Equal(t, "foobar\n", buf.String())
		buf.Reset()
		fn(ctxutil.WithHidden(ctx, true), "foobar")
		assert.Equal(t, "", buf.String())
	}
}
