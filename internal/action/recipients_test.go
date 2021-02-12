package action

import (
	"bytes"
	"context"
	"os"
	"testing"

	"github.com/gopasspw/gopass/internal/out"
	"github.com/gopasspw/gopass/pkg/ctxutil"
	"github.com/gopasspw/gopass/tests/gptest"

	"github.com/fatih/color"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRecipients(t *testing.T) {
	u := gptest.NewUnitTester(t)
	defer u.Remove()

	ctx := context.Background()
	ctx = ctxutil.WithAlwaysYes(ctx, true)
	ctx = ctxutil.WithInteractive(ctx, false)

	act, err := newMock(ctx, u)
	require.NoError(t, err)
	require.NotNil(t, act)

	buf := &bytes.Buffer{}
	out.Stdout = buf
	out.Stderr = buf
	stdout = buf
	color.NoColor = true
	defer func() {
		out.Stdout = os.Stdout
		out.Stderr = os.Stderr
		stdout = os.Stdout
	}()

	t.Run("print recipients tree", func(t *testing.T) {
		defer buf.Reset()
		assert.NoError(t, act.RecipientsPrint(gptest.CliCtx(ctx, t)))
		want := `Hint: run 'gopass sync' to import any missing public keys
gopass
└── 0xDEADBEEF

`

		assert.Equal(t, want, buf.String())
	})

	t.Run("complete recipients", func(t *testing.T) {
		defer buf.Reset()
		act.RecipientsComplete(gptest.CliCtx(ctx, t))
		want := "0xDEADBEEF\n"
		assert.Equal(t, want, buf.String())
	})

	t.Run("add recipients w/o args", func(t *testing.T) {
		defer buf.Reset()
		assert.Error(t, act.RecipientsAdd(gptest.CliCtx(ctx, t)))
	})

	t.Run("remove recipients w/o args", func(t *testing.T) {
		defer buf.Reset()
		assert.Error(t, act.RecipientsRemove(gptest.CliCtx(ctx, t)))
	})

	t.Run("add recipient 0xFEEDBEEF", func(t *testing.T) {
		defer buf.Reset()
		assert.NoError(t, act.RecipientsAdd(gptest.CliCtx(ctx, t, "0xFEEDBEEF")))
	})

	t.Run("add recipient 0xBEEFFEED", func(t *testing.T) {
		defer buf.Reset()
		assert.NoError(t, act.RecipientsAdd(gptest.CliCtx(ctx, t, "0xBEEFFEED")))
	})

	t.Run("remove recipient 0xDEADBEEF", func(t *testing.T) {
		defer buf.Reset()
		assert.NoError(t, act.RecipientsRemove(gptest.CliCtx(ctx, t, "0xDEADBEEF")))
	})
}
