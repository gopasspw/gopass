package action

import (
	"bytes"
	"context"
	"os"
	"testing"

	"github.com/gopasspw/gopass/internal/backend"
	"github.com/gopasspw/gopass/internal/config"
	"github.com/gopasspw/gopass/internal/out"
	"github.com/gopasspw/gopass/pkg/ctxutil"
	"github.com/gopasspw/gopass/pkg/termio"
	"github.com/gopasspw/gopass/tests/gptest"

	"github.com/blang/semver/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHistory(t *testing.T) {
	u := gptest.NewUnitTester(t)
	defer u.Remove()

	r1 := gptest.UnsetVars(termio.NameVars...)
	r2 := gptest.UnsetVars(termio.EmailVars...)
	defer r1()
	defer r2()

	ctx := context.Background()
	ctx = ctxutil.WithAlwaysYes(ctx, true)
	ctx = ctxutil.WithInteractive(ctx, false)

	ctx = backend.WithCryptoBackend(ctx, backend.Plain)
	ctx = backend.WithStorageBackend(ctx, backend.GitFS)

	cfg := config.New()
	cfg.Path = u.StoreDir("")
	act, err := newAction(cfg, semver.Version{}, false)
	require.NoError(t, err)
	require.NotNil(t, act)

	t.Run("can initialize", func(t *testing.T) {
		require.NoError(t, act.IsInitialized(gptest.CliCtx(ctx, t)))
	})

	buf := &bytes.Buffer{}
	out.Stdout = buf
	defer func() {
		out.Stdout = os.Stdout
	}()

	t.Run("init git", func(t *testing.T) {
		defer buf.Reset()
		require.NoError(t, act.rcsInit(ctx, "", "foo bar", "foo.bar@example.org"))
		t.Logf("init git: %s", buf.String())
	})

	t.Run("insert bar", func(t *testing.T) {
		defer buf.Reset()
		assert.NoError(t, act.Insert(gptest.CliCtx(ctx, t, "bar")))
	})

	t.Run("history bar", func(t *testing.T) {
		defer buf.Reset()
		assert.NoError(t, act.History(gptest.CliCtx(ctx, t, "bar")))
	})

	t.Run("history --password bar", func(t *testing.T) {
		defer buf.Reset()
		assert.NoError(t, act.History(gptest.CliCtxWithFlags(ctx, t, map[string]string{"password": "true"}, "bar")))
	})
}
