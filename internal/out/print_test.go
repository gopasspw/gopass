package out

import (
	"bytes"
	"os"
	"testing"

	"github.com/gopasspw/gopass/internal/config"
	"github.com/gopasspw/gopass/pkg/ctxutil"
	"github.com/stretchr/testify/assert"
)

func TestPrint(t *testing.T) {
	ctx := config.NewContextInMemory()
	buf := &bytes.Buffer{}
	Stdout = buf
	defer func() {
		Stdout = os.Stdout
	}()

	Printf(ctx, "%s = %d", "foo", 42)
	assert.Equal(t, "foo = 42\n", buf.String())
	buf.Reset()

	Printf(ctxutil.WithHidden(ctx, true), "%s = %d", "foo", 42)
	assert.Empty(t, buf.String())
	buf.Reset()

	Printf(WithNewline(ctx, false), "%s = %d", "foo", 42)
	assert.Equal(t, "foo = 42", buf.String())
	buf.Reset()
}

func TestPrintStderr(t *testing.T) {
	ctx := config.NewContextInMemory()
	buf := &bytes.Buffer{}
	Stderr = buf
	defer func() {
		Stderr = os.Stderr
	}()

	PrintStderrf(ctx, "%s = %d", "foo", 42)
	assert.Equal(t, "foo = 42\n", buf.String())
	buf.Reset()

	PrintStderr(ctxutil.WithHidden(ctx, true), "hidden")
	assert.Empty(t, buf.String())
	buf.Reset()

	PrintStderr(WithNewline(ctx, false), "foo")
	assert.Equal(t, "foo", buf.String())
}
