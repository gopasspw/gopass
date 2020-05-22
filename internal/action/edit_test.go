package action

import (
	"bytes"
	"context"
	"os"
	"testing"

	"github.com/gopasspw/gopass/internal/out"
	"github.com/gopasspw/gopass/pkg/ctxutil"
	"github.com/gopasspw/gopass/tests/gptest"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEdit(t *testing.T) {
	u := gptest.NewUnitTester(t)
	defer u.Remove()

	ctx := context.Background()
	ctx = ctxutil.WithAlwaysYes(ctx, true)
	ctx = ctxutil.WithTerminal(ctx, false)
	act, err := newMock(ctx, u)
	require.NoError(t, err)
	require.NotNil(t, act)

	buf := &bytes.Buffer{}
	out.Stdout = buf
	defer func() {
		out.Stdout = os.Stdout
	}()

	// edit
	assert.Error(t, act.Edit(clictx(ctx, t)))
	buf.Reset()

	// edit foo (existing)
	assert.Error(t, act.Edit(clictx(ctx, t, "foo")))
	buf.Reset()

	// edit bar (new)
	assert.Error(t, act.Edit(clictx(ctx, t, "foo")))
	buf.Reset()
}

func TestEditUpdate(t *testing.T) {
	u := gptest.NewUnitTester(t)
	defer u.Remove()

	ctx := context.Background()
	ctx = ctxutil.WithAlwaysYes(ctx, true)
	ctx = ctxutil.WithTerminal(ctx, false)
	act, err := newMock(ctx, u)
	require.NoError(t, err)
	require.NotNil(t, act)

	buf := &bytes.Buffer{}
	out.Stdout = buf
	defer func() {
		out.Stdout = os.Stdout
	}()

	content := []byte("foobar")
	// no changes
	assert.NoError(t, act.editUpdate(ctx, "foo", content, content, false, "test"))
	buf.Reset()

	// changes
	nContent := []byte("barfoo")
	assert.NoError(t, act.editUpdate(ctx, "foo", content, nContent, false, "test"))
	buf.Reset()
}
