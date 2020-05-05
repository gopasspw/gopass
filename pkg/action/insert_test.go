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

func TestInsert(t *testing.T) {
	u := gptest.NewUnitTester(t)
	defer u.Remove()

	ctx := context.Background()
	ctx = ctxutil.WithAlwaysYes(ctx, true)
	ctx = ctxutil.WithTerminal(ctx, false)
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

	// insert bar
	assert.NoError(t, act.Insert(clictx(ctx, t, "bar")))

	// insert bar baz
	assert.NoError(t, act.Insert(clictx(ctx, t, "bar", "baz")))

	// insert baz via stdin
	assert.NoError(t, act.insertStdin(ctx, "baz", []byte("foobar"), false))
	buf.Reset()

	assert.NoError(t, act.show(ctx, clictx(ctx, t), "baz", false))
	assert.Equal(t, "foobar", buf.String())
	buf.Reset()

	// insert zab#key
	assert.NoError(t, act.insertYAML(ctx, "zab", "key", []byte("foobar"), nil))

	// insert --multiline bar baz
	assert.NoError(t, act.Insert(clictxf(ctx, t, map[string]string{"multiline": "true"}, "bar", "baz")))
}
