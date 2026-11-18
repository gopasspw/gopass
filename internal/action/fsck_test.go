package action

import (
	"bytes"
	"context"
	"os"
	"strings"
	"testing"

	"github.com/fatih/color"
	"github.com/gopasspw/gopass/internal/action/exit"
	"github.com/gopasspw/gopass/internal/backend"
	"github.com/gopasspw/gopass/internal/config"
	"github.com/gopasspw/gopass/internal/out"
	"github.com/gopasspw/gopass/pkg/ctxutil"
	"github.com/gopasspw/gopass/tests/can"
	"github.com/gopasspw/gopass/tests/gptest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v3"
)

func TestFsck(t *testing.T) {
	u := gptest.NewUnitTester(t)

	ctx := config.NewContextInMemory()
	ctx = ctxutil.WithTerminal(ctx, false)
	act, err := newMock(ctx, u.StoreDir(""))
	require.NoError(t, err)
	require.NotNil(t, act)
	ctx = act.cfg.WithConfig(ctx)

	buf := &bytes.Buffer{}
	out.Stdout = buf
	out.Stderr = buf
	stdout = buf
	defer func() {
		stdout = os.Stdout
		out.Stdout = os.Stdout
		out.Stderr = os.Stderr
	}()
	color.NoColor = true

	// generate foo/bar
	c := gptest.CliCtx(ctx, t, "foo/bar", "24")
	require.NoError(t, act.Generate(ctx, c), "gopass generate foo/bar 24")
	buf.Reset()

	// fsck
	require.NoError(t, act.Fsck(ctx, gptest.CliCtx(ctx, t)))
	output := strings.TrimSpace(buf.String())
	assert.Contains(t, output, "Checking password store integrity ...")
	assert.Contains(t, output, "extra recipients on foo: [0xFEEDBEEF]")
	buf.Reset()

	// fsck (hidden)
	require.NoError(t, act.Fsck(ctxutil.WithHidden(ctx, true), gptest.CliCtx(ctx, t)))
	output = strings.TrimSpace(buf.String())
	assert.NotContains(t, output, "Checking password store integrity ...")
	assert.NotContains(t, output, "extra recipients on foo: [0xFEEDBEEF]")
	buf.Reset()

	// fsck --decrypt
	require.NoError(t, act.Fsck(ctx, gptest.CliCtxWithFlags(ctx, t, map[string]string{"decrypt": "true"})))
	output = strings.TrimSpace(buf.String())
	assert.Contains(t, output, "Checking password store integrity ...")
	assert.Contains(t, output, "extra recipients on foo: [0xFEEDBEEF]")
	buf.Reset()

	// fsck foo
	require.NoError(t, act.Fsck(ctx, gptest.CliCtx(ctx, t, "foo")))
	output = strings.TrimSpace(buf.String())
	assert.Contains(t, output, "Checking password store integrity ...")
	assert.Contains(t, output, "extra recipients on foo: [0xFEEDBEEF]")
	buf.Reset()
}

func TestFsckGpg(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	u := gptest.NewGUnitTester(t)

	ctx := config.NewContextInMemory()
	ctx = ctxutil.WithTerminal(ctx, false)
	ctx = backend.WithCryptoBackend(ctx, backend.GPGCLI)

	act, err := newMock(ctx, u.StoreDir(""))
	require.NoError(t, err)
	require.NotNil(t, act)
	ctx = act.cfg.WithConfig(ctx)

	buf := &bytes.Buffer{}
	out.Stdout = buf
	out.Stderr = buf
	stdout = buf
	defer func() {
		stdout = os.Stdout
		out.Stdout = os.Stdout
		out.Stderr = os.Stderr
	}()
	color.NoColor = true

	sub, err := act.Store.GetSubStore("")
	require.NoError(t, err)
	require.NoError(t, sub.ImportMissingPublicKeys(ctx, can.KeyID()))

	// generate foo/bar
	c := gptest.CliCtx(ctx, t, "foo/bar", "24")
	require.NoError(t, act.Generate(ctx, c), "gopass generate foo/bar 24")
	buf.Reset()

	// fsck
	require.NoError(t, act.Fsck(ctx, gptest.CliCtx(ctx, t)))
	output := strings.TrimSpace(buf.String())
	assert.Contains(t, output, "Checking password store integrity ...")
	buf.Reset()

	// fsck (hidden)
	require.NoError(t, act.Fsck(ctxutil.WithHidden(ctx, true), gptest.CliCtx(ctx, t)))
	output = strings.TrimSpace(buf.String())
	assert.NotContains(t, output, "Checking password store integrity ...")
	assert.NotContains(t, output, "Extra recipients on foo: [0xFEEDBEEF]")
	buf.Reset()

	// fsck --decrypt
	require.NoError(t, act.Fsck(ctx, gptest.CliCtxWithFlags(ctx, t, map[string]string{"decrypt": "true"})))
	output = strings.TrimSpace(buf.String())
	assert.Contains(t, output, "Checking password store integrity ...")
	buf.Reset()

	// fsck foo
	require.NoError(t, act.Fsck(ctx, gptest.CliCtx(ctx, t, "foo")))
	output = strings.TrimSpace(buf.String())
	assert.Contains(t, output, "Checking password store integrity ...")
	buf.Reset()
}

func setupFsckMultiMount(t *testing.T) (*Action, context.Context, *gptest.Unit, *bytes.Buffer) {
	t.Helper()

	u := gptest.NewUnitTester(t)

	ctx := config.NewContextInMemory()
	ctx = ctxutil.WithTerminal(ctx, false)
	ctx = ctxutil.WithAlwaysYes(ctx, true)
	act, err := newMock(ctx, u.StoreDir(""))
	require.NoError(t, err)
	require.NotNil(t, act)
	ctx = act.cfg.WithConfig(ctx)

	buf := &bytes.Buffer{}
	out.Stdout = buf
	out.Stderr = buf
	stdout = buf
	t.Cleanup(func() {
		stdout = os.Stdout
		out.Stdout = os.Stdout
		out.Stderr = os.Stderr
	})
	color.NoColor = true

	require.NoError(t, u.InitStore("sub1"))
	require.NoError(t, u.InitStore("sub2"))
	require.NoError(t, act.Store.AddMount(ctx, "sub1", u.StoreDir("sub1")))
	require.NoError(t, act.Store.AddMount(ctx, "sub2", u.StoreDir("sub2")))

	for _, name := range []string{"rootsec", "sub1/sec1", "sub2/sec2"} {
		require.NoError(t, act.Generate(ctx, gptest.CliCtx(ctx, t, name, "24")))
	}
	buf.Reset()

	return act, ctx, u, buf
}

func TestFsckMultiMount(t *testing.T) {
	act, ctx, u, buf := setupFsckMultiMount(t)

	for _, tc := range []struct {
		name       string
		flags      map[string]string
		args       []string
		want       []string
		wantAbsent []string
	}{
		{
			name:  "fsck --store=sub1 --decrypt",
			flags: map[string]string{"store": "sub1", "decrypt": "true"},
			want: []string{
				"Checking password store integrity ...",
				"[sub1] Re-encrypting sec1",
				"[sub1] Re-encrypting foo",
			},
			wantAbsent: []string{
				"[] Re-encrypting",
				"rootsec",
				"[sub2]",
				"sec2",
			},
		},
		{
			name:  "fsck --store=<root> --decrypt",
			flags: map[string]string{"store": "<root>", "decrypt": "true"},
			want: []string{
				"[] Re-encrypting rootsec",
				"[] Re-encrypting foo",
			},
			wantAbsent: []string{
				"[sub1]",
				"sec1",
				"[sub2]",
				"sec2",
			},
		},
		{
			name:  "fsck --store=sub1 --decrypt with relative filter",
			flags: map[string]string{"store": "sub1", "decrypt": "true"},
			args:  []string{"sec1"},
			want: []string{
				"[sub1] Re-encrypting sec1",
			},
			wantAbsent: []string{
				"[sub1] Re-encrypting foo",
				"[] Re-encrypting",
				"[sub2]",
			},
		},
		{
			name:  "fsck --store=sub1 --decrypt with mount-prefixed filter",
			flags: map[string]string{"store": "sub1", "decrypt": "true"},
			args:  []string{"sub1/sec1"},
			want: []string{
				"[sub1] Re-encrypting sec1",
			},
			wantAbsent: []string{
				"[sub1] Re-encrypting foo",
				"[] Re-encrypting",
				"[sub2]",
			},
		},
		{
			name:  "fsck positional mount filter without trailing slash",
			flags: map[string]string{"decrypt": "true"},
			args:  []string{"sub1"},
			want: []string{
				"[sub1] Re-encrypting sec1",
			},
			wantAbsent: []string{
				"[] Re-encrypting",
				"[sub2]",
			},
		},
		{
			name:  "fsck positional mount filter with trailing slash",
			flags: map[string]string{"decrypt": "true"},
			args:  []string{"sub1/"},
			want: []string{
				"[sub1] Re-encrypting sec1",
			},
			wantAbsent: []string{
				"[] Re-encrypting",
				"[sub2]",
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			buf.Reset()
			require.NoError(t, act.Fsck(ctx, gptest.CliCtxWithFlags(ctx, t, tc.flags, tc.args...)))
			output := buf.String()

			for _, w := range tc.want {
				assert.Contains(t, output, w)
			}

			for _, a := range tc.wantAbsent {
				assert.NotContains(t, output, a)
			}
		})
	}

	t.Run("unknown --store returns exit.NotFound", func(t *testing.T) {
		buf.Reset()
		err := act.Fsck(ctx, gptest.CliCtxWithFlags(ctx, t, map[string]string{"store": "nonexistent"}))
		require.Error(t, err)

		var exitErr cli.ExitCoder
		require.ErrorAs(t, err, &exitErr)
		assert.Equal(t, exit.NotFound, exitErr.ExitCode())
	})

	t.Run("broken unrelated mount does not block --store=sub1", func(t *testing.T) {
		buf.Reset()
		require.NoError(t, os.RemoveAll(u.StoreDir("sub2")))

		err := act.Fsck(ctx, gptest.CliCtxWithFlags(ctx, t, map[string]string{"store": "sub1", "decrypt": "true"}))
		require.NoError(t, err)
		output := buf.String()
		assert.Contains(t, output, "[sub1] Re-encrypting sec1")
		assert.NotContains(t, output, "[sub2]")
		assert.NotContains(t, output, "[] Re-encrypting")
	})
}
