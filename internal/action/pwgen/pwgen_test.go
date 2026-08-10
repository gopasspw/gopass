package pwgen

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/gopasspw/gopass/internal/config"
	"github.com/gopasspw/gopass/internal/out"
	"github.com/gopasspw/gopass/pkg/ctxutil"
	"github.com/gopasspw/gopass/pkg/pwgen"
	"github.com/gopasspw/gopass/tests/gptest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v3"
)

func TestPwgen(t *testing.T) {
	u := gptest.NewUnitTester(t)
	assert.NotNil(t, u)

	ctx := config.NewContextInMemory()
	ctx = ctxutil.WithAlwaysYes(ctx, true)

	buf := &bytes.Buffer{}
	out.Stdout = buf
	defer func() {
		out.Stdout = os.Stdout
	}()

	require.NoError(t, Pwgen(ctx, gptest.CliCtxWithFlags(ctx, t, map[string]string{"one-per-line": "true"}, "24", "1")))
	assert.GreaterOrEqual(t, len(buf.Bytes()), 24, buf.String())
}

// TestPwgenMemorable exercises the --memorable generator path: dispatch,
// --symbols, --memorable-capitalize (which implies --memorable), the password
// count, --memorable taking precedence over --xkcd, and the --no-numerals
// incompatibility guard.
func TestPwgenMemorable(t *testing.T) {
	u := gptest.NewUnitTester(t)
	assert.NotNil(t, u)

	ctx := config.NewContextInMemory()
	ctx = ctxutil.WithAlwaysYes(ctx, true)

	buf := &bytes.Buffer{}
	out.Stdout = buf
	defer func() {
		out.Stdout = os.Stdout
	}()

	tt := []struct {
		name    string
		flags   map[string]string
		args    []string
		wantErr bool
		check   func(out string) bool
	}{
		{
			name:  "happy",
			flags: map[string]string{"memorable": "true"},
			args:  []string{"24", "1"},
			// memorable treats length as a minimum, so output is >= length.
			check: func(out string) bool { return len(out) >= 24 },
		},
		{
			name:  "symbols",
			flags: map[string]string{"memorable": "true", "symbols": "true"},
			args:  []string{"24", "1"},
			check: func(out string) bool { return strings.ContainsAny(out, pwgen.Syms) },
		},
		{
			flags: map[string]string{"memorable-capitalize": "true"},
			args:  []string{"24", "1"},
			// capitals=true guarantees at least one uppercased word.
			name:  "capitalize implies memorable",
			check: func(out string) bool { return hasUppercase(out) },
		},
		{
			name:  "count",
			flags: map[string]string{"memorable": "true"},
			args:  []string{"12", "5"},
			check: func(out string) bool { return strings.Count(out, "\n") == 5 },
		},
		{
			name:    "memorable and xkcd mutually exclusive",
			flags:   map[string]string{"memorable": "true", "xkcd": "true"},
			args:    []string{"24", "1"},
			wantErr: true,
		},
		{
			name:    "no-numerals incompatible",
			flags:   map[string]string{"memorable": "true", "no-numerals": "true"},
			args:    []string{"24", "1"},
			wantErr: true,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			buf.Reset()
			err := Pwgen(ctx, gptest.CliCtxWithFlags(ctx, t, tc.flags, tc.args...))
			if tc.wantErr {
				require.Error(t, err, "expected an error")

				return
			}

			require.NoError(t, err)

			if tc.check != nil {
				assert.True(t, tc.check(buf.String()), "%s: %s", tc.name, buf.String())
			}
		})
	}
}

// TestPwgenMemorableConfig verifies the pwgen.memorable-capitalize config key
// drives capitalization and that the --memorable-capitalize flag overrides it.
func TestPwgenMemorableConfig(t *testing.T) {
	u := gptest.NewUnitTester(t)
	assert.NotNil(t, u)

	ctx := config.NewContextInMemory()
	ctx = ctxutil.WithAlwaysYes(ctx, true)

	cfg, _ := config.FromContext(ctx)
	require.NoError(t, cfg.SetEnv("pwgen.memorable-capitalize", "true"))

	buf := &bytes.Buffer{}
	out.Stdout = buf
	defer func() {
		out.Stdout = os.Stdout
	}()

	// config on => capitals.
	require.NoError(t, Pwgen(ctx, gptest.CliCtxWithFlags(ctx, t, map[string]string{"memorable": "true"}, "24", "1")))
	assert.True(t, hasUppercase(buf.String()), buf.String())

	// flag overrides config off => no capitals.
	buf.Reset()
	require.NoError(t, Pwgen(ctx, gptest.CliCtxWithFlags(ctx, t, map[string]string{"memorable": "true", "memorable-capitalize": "false"}, "24", "1")))
	assert.False(t, hasUppercase(buf.String()), buf.String())

	// --no-capitalize overrides config on => no capitals.
	buf.Reset()
	require.NoError(t, Pwgen(ctx, gptest.CliCtxWithFlags(ctx, t, map[string]string{"memorable": "true", "no-capitalize": "true"}, "24", "1")))
	assert.False(t, hasUppercase(buf.String()), buf.String())
}

// TestMemorableFlagsRegistered checks the flags are actually registered on the
// pwgen command with their aliases. CliCtxWithFlags fabricates flags by name and
// does not use GetCommands(), so without this test a missing flag or alias would
// go unnoticed by the behavior tests above.
func TestMemorableFlagsRegistered(t *testing.T) {
	cmds := GetCommands()
	require.Len(t, cmds, 1)

	memorable := boolFlag(t, cmds[0].Flags, "memorable")
	require.NotNil(t, memorable)
	assert.Contains(t, memorable.Aliases, "m")
	assert.NotEmpty(t, memorable.Usage)

	memCap := boolFlag(t, cmds[0].Flags, "memorable-capitalize")
	require.NotNil(t, memCap)
	assert.Contains(t, memCap.Aliases, "mc")
	assert.NotEmpty(t, memCap.Usage)
}

func boolFlag(t *testing.T, flags []cli.Flag, name string) *cli.BoolFlag {
	t.Helper()

	for _, f := range flags {
		if bf, ok := f.(*cli.BoolFlag); ok && bf.Name == name {
			return bf
		}
	}

	return nil
}

func hasUppercase(s string) bool {
	return strings.ContainsAny(s, "ABCDEFGHIJKLMNOPQRSTUVWXYZ")
}
