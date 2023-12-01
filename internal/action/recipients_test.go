package action

import (
	"bytes"
	"os"
	"testing"

	"github.com/fatih/color"
	"github.com/gopasspw/gopass/internal/backend"
	"github.com/gopasspw/gopass/internal/config"
	"github.com/gopasspw/gopass/internal/out"
	"github.com/gopasspw/gopass/pkg/ctxutil"
	"github.com/gopasspw/gopass/tests/gptest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRecipients(t *testing.T) {
	u := gptest.NewUnitTester(t)

	ctx := config.NewContextInMemory()
	ctx = ctxutil.WithAlwaysYes(ctx, true)
	ctx = ctxutil.WithInteractive(ctx, false)

	act, err := newMock(ctx, u.StoreDir(""))
	require.NoError(t, err)
	require.NotNil(t, act)
	ctx = act.cfg.WithConfig(ctx)

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
		require.NoError(t, act.RecipientsPrint(gptest.CliCtx(ctx, t)))

		hint := `Hint: run 'gopass sync' to import any missing public keys`
		want := `gopass
└── 0xDEADBEEF`

		assert.Contains(t, buf.String(), hint)
		assert.Contains(t, buf.String(), want)
	})

	t.Run("complete recipients", func(t *testing.T) {
		defer buf.Reset()
		act.RecipientsComplete(gptest.CliCtx(ctx, t))
		want := "0xDEADBEEF\n"
		assert.Equal(t, want, buf.String())
	})

	t.Run("add recipients w/o args", func(t *testing.T) {
		defer buf.Reset()
		require.Error(t, act.RecipientsAdd(gptest.CliCtx(ctx, t)))
	})

	t.Run("remove recipients w/o args", func(t *testing.T) {
		defer buf.Reset()
		require.Error(t, act.RecipientsRemove(gptest.CliCtx(ctx, t)))
	})

	t.Run("add recipient 0xFEEDBEEF", func(t *testing.T) {
		defer buf.Reset()
		require.NoError(t, act.RecipientsAdd(gptest.CliCtx(ctx, t, "0xFEEDBEEF")))
	})

	t.Run("add recipient 0xBEEFFEED", func(t *testing.T) {
		defer buf.Reset()
		require.NoError(t, act.RecipientsAdd(gptest.CliCtx(ctx, t, "0xBEEFFEED")))
	})

	t.Run("remove recipient 0xDEADBEEF", func(t *testing.T) {
		defer buf.Reset()
		require.NoError(t, act.RecipientsRemove(gptest.CliCtx(ctx, t, "0xDEADBEEF")))
	})
}

func TestRecipientsGpg(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	u := gptest.NewGUnitTester(t)

	ctx := config.NewContextInMemory()
	ctx = ctxutil.WithAlwaysYes(ctx, true)
	ctx = ctxutil.WithInteractive(ctx, false)
	ctx = backend.WithCryptoBackend(ctx, backend.GPGCLI)

	act, err := newMock(ctx, u.StoreDir(""))
	require.NoError(t, err)
	require.NotNil(t, act)
	ctx = act.cfg.WithConfig(ctx)

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
		require.NoError(t, act.RecipientsPrint(gptest.CliCtx(ctx, t)))

		hint := `Hint: run 'gopass sync' to import any missing public keys`
		want := `gopass
└── BE73F104`

		assert.Contains(t, buf.String(), hint)
		assert.Contains(t, buf.String(), want)
	})

	t.Run("complete recipients", func(t *testing.T) {
		defer buf.Reset()
		act.RecipientsComplete(gptest.CliCtx(ctx, t))
		want := "BE73F104\n"
		assert.Equal(t, want, buf.String())
	})

	t.Run("add recipients w/o args", func(t *testing.T) {
		defer buf.Reset()
		require.Error(t, act.RecipientsAdd(gptest.CliCtx(ctx, t)))
	})

	t.Run("remove recipients w/o args", func(t *testing.T) {
		defer buf.Reset()
		require.Error(t, act.RecipientsRemove(gptest.CliCtx(ctx, t)))
	})

	t.Run("add recipient 0xFEEDBEEF", func(t *testing.T) {
		defer buf.Reset()
		require.NoError(t, act.RecipientsAdd(gptest.CliCtx(ctx, t, "0xFEEDBEEF")))
	})

	t.Run("add recipient 0xBEEFFEED", func(t *testing.T) {
		defer buf.Reset()
		require.NoError(t, act.RecipientsAdd(gptest.CliCtxWithFlags(ctx, t, map[string]string{"force": "true"}, "0xBEEFFEED")))
	})

	t.Run("remove recipient 0x82EBD945BE73F104", func(t *testing.T) {
		defer buf.Reset()
		require.NoError(t, act.RecipientsRemove(gptest.CliCtx(ctx, t, "0x82EBD945BE73F104")))
	})

	t.Run("add recipient 0xFEEDFEED", func(t *testing.T) {
		defer buf.Reset()
		require.NoError(t, act.RecipientsAdd(gptest.CliCtxWithFlags(ctx, t, map[string]string{"force": "true"}, "0xFEEDFEED")))
	})

	t.Run("remove recipient 0xFEEDFEED", func(t *testing.T) {
		defer buf.Reset()
		require.NoError(t, act.RecipientsRemove(gptest.CliCtx(ctx, t, "0xFEEDFEED")))
	})

	t.Run("add recipient 0xFEEDFEED", func(t *testing.T) {
		defer buf.Reset()
		require.NoError(t, act.RecipientsAdd(gptest.CliCtxWithFlags(ctx, t, map[string]string{"force": "true"}, "0xFEEDFEED")))
	})

	t.Run("remove recipient 0xFEEDFEED (force)", func(t *testing.T) {
		defer buf.Reset()
		require.NoError(t, act.RecipientsRemove(gptest.CliCtxWithFlags(ctx, t, map[string]string{"force": "true"}, "0xFEEDFEED")))
	})
}
