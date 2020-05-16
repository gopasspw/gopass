package action

import (
	"bytes"
	"context"
	"os"
	"testing"

	"github.com/gopasspw/gopass/pkg/ctxutil"
	"github.com/gopasspw/gopass/pkg/out"
	"github.com/gopasspw/gopass/tests/gptest"

	"github.com/fatih/color"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCopy(t *testing.T) {
	u := gptest.NewUnitTester(t)
	defer u.Remove()

	ctx := context.Background()
	ctx = ctxutil.WithAlwaysYes(ctx, true)
	ctx = ctxutil.WithAutoClip(ctx, false)

	act, err := newMock(ctx, u)
	require.NoError(t, err)
	require.NotNil(t, act)

	buf := &bytes.Buffer{}
	out.Stdout = buf
	color.NoColor = true
	defer func() {
		out.Stdout = os.Stdout
	}()

	// copy foo bar
	c := clictx(ctx, t, "foo", "bar")
	assert.NoError(t, act.Copy(c))
	buf.Reset()

	// copy foo bar (again, should fail)
	{
		ctx := ctxutil.WithAlwaysYes(ctx, false)
		ctx = ctxutil.WithInteractive(ctx, false)
		c.Context = ctx
		assert.Error(t, act.Copy(c))
		buf.Reset()
	}

	// copy not-found still-not-there
	c = clictx(ctx, t, "not-found", "still-not-there")
	assert.Error(t, act.Copy(c))
	buf.Reset()

	// copy
	c = clictx(ctx, t)
	assert.Error(t, act.Copy(c))
	buf.Reset()

	// insert bam/baz
	assert.NoError(t, act.insertStdin(ctx, "bam/baz", []byte("foobar"), false))
	assert.NoError(t, act.insertStdin(ctx, "bam/zab", []byte("barfoo"), false))

	// recursive copy: bam/ -> zab
	c = clictx(ctx, t, "bam", "zab")
	assert.NoError(t, act.Copy(c))
	buf.Reset()

	assert.NoError(t, act.show(ctx, c, "zab/bam/zab", false))
	assert.Equal(t, "barfoo\n", buf.String())
	buf.Reset()
}
