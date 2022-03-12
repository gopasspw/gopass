package out

import (
	"bytes"
	"context"
	"os"
	"testing"

	"github.com/gopasspw/gopass/pkg/ctxutil"
	"github.com/stretchr/testify/assert"
)

func TestPrint(t *testing.T) { //nolint:paralleltest
	ctx := context.Background()
	buf := &bytes.Buffer{}
	Stdout = buf
	defer func() {
		Stdout = os.Stdout
	}()

	Printf(ctx, "%s = %d", "foo", 42)
	assert.Equal(t, "foo = 42\n", buf.String())
	buf.Reset()

	Printf(ctxutil.WithHidden(ctx, true), "%s = %d", "foo", 42)
	assert.Equal(t, "", buf.String())
	buf.Reset()

	Printf(WithNewline(ctx, false), "%s = %d", "foo", 42)
	assert.Equal(t, "foo = 42", buf.String())
	buf.Reset()
}
