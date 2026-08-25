package action

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/fatih/color"
	"github.com/gopasspw/gopass/internal/config"
	"github.com/gopasspw/gopass/internal/out"
	"github.com/gopasspw/gopass/pkg/ctxutil"
	"github.com/gopasspw/gopass/pkg/gopass/secrets"
	"github.com/gopasspw/gopass/tests/gptest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExport(t *testing.T) {
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
	stdout = buf
	defer func() {
		stdout = os.Stdout
		out.Stdout = os.Stdout
	}()
	color.NoColor = true

	// add a second secret with a password and an attribute.
	sec := secrets.NewAKV()
	sec.SetPassword("123")
	require.NoError(t, sec.Set("url", "https://example.com"))
	require.NoError(t, act.Store.Set(ctx, "foo/bar", sec))

	byName := func(t *testing.T, data []byte) map[string]exportedSecret {
		t.Helper()
		var got []exportedSecret
		require.NoError(t, json.Unmarshal(data, &got))
		m := make(map[string]exportedSecret, len(got))
		for _, e := range got {
			m[e.Name] = e
		}

		return m
	}

	t.Run("export the whole store", func(t *testing.T) {
		buf.Reset()
		require.NoError(t, act.Export(ctx, gptest.CliCtx(ctx, t)))

		m := byName(t, buf.Bytes())
		require.Len(t, m, 2)
		require.Contains(t, m, "foo")
		require.Contains(t, m, "foo/bar")
		assert.Equal(t, "123", m["foo/bar"].Password)
		assert.Equal(t, []string{"https://example.com"}, m["foo/bar"].Attributes["url"])
	})

	t.Run("filter by prefix", func(t *testing.T) {
		buf.Reset()
		require.NoError(t, act.Export(ctx, gptest.CliCtx(ctx, t, "foo/")))

		m := byName(t, buf.Bytes())
		require.Len(t, m, 1)
		require.Contains(t, m, "foo/bar")
		assert.NotContains(t, m, "foo")
	})

	t.Run("write to file", func(t *testing.T) {
		buf.Reset()
		fn := filepath.Join(t.TempDir(), "export.json")
		require.NoError(t, act.Export(ctx, gptest.CliCtxWithFlags(ctx, t, map[string]string{"out": fn})))

		data, err := os.ReadFile(fn)
		require.NoError(t, err)
		m := byName(t, data)
		require.Contains(t, m, "foo")
		require.Contains(t, m, "foo/bar")
		// stdout stays clean when writing to a file.
		assert.NotContains(t, buf.String(), "foo/bar")
	})
}

func TestHasAnyPrefix(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		prefixes []string
		want     bool
	}{
		{name: "foo", prefixes: nil, want: true},
		{name: "foo", prefixes: []string{"foo"}, want: true},
		{name: "foo/bar", prefixes: []string{"foo"}, want: true},
		{name: "foo/bar", prefixes: []string{"foo/"}, want: true},
		{name: "foobar", prefixes: []string{"foo"}, want: false},
		{name: "baz", prefixes: []string{"foo", "baz"}, want: true},
	}

	for _, tc := range tests {
		assert.Equal(t, tc.want, hasAnyPrefix(tc.name, tc.prefixes), tc.name)
	}
}
