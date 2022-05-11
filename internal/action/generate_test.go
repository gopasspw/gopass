package action

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"os"
	"regexp"
	"runtime"
	"strings"
	"testing"

	"github.com/fatih/color"
	"github.com/gopasspw/gopass/internal/out"
	"github.com/gopasspw/gopass/pkg/ctxutil"
	"github.com/gopasspw/gopass/tests/gptest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v2"
)

func TestRuleLookup(t *testing.T) {
	t.Parallel()

	domain, _ := hasPwRuleForSecret("foo/gopass.pw")
	assert.Equal(t, "", domain)
}

func TestGenerate(t *testing.T) { //nolint:paralleltest
	u := gptest.NewUnitTester(t)
	defer u.Remove()

	ctx := context.Background()
	ctx = ctxutil.WithAlwaysYes(ctx, true)
	ctx = ctxutil.WithInteractive(ctx, false)

	act, err := newMock(ctx, u)
	require.NoError(t, err)
	require.NotNil(t, act)

	act.cfg.AutoClip = false

	buf := &bytes.Buffer{}
	out.Stdout = buf
	out.Stderr = buf
	defer func() {
		out.Stdout = os.Stdout
		out.Stderr = os.Stderr
	}()
	color.NoColor = true

	// generate
	t.Run("generate", func(t *testing.T) { //nolint:paralleltest
		assert.Error(t, act.Generate(gptest.CliCtx(ctx, t)))
		buf.Reset()
	})

	// generate foobar
	t.Run("generate foobar", func(t *testing.T) { //nolint:paralleltest
		if testing.Short() {
			t.Skip("skipping test in short mode.")
		}

		assert.NoError(t, act.Generate(gptest.CliCtx(ctx, t, "foobar")))
		buf.Reset()
	})

	// generate foobar
	// should succeed because of always yes
	t.Run("generate foobar again", func(t *testing.T) { //nolint:paralleltest
		if testing.Short() {
			t.Skip("skipping test in short mode.")
		}

		assert.NoError(t, act.Generate(gptest.CliCtx(ctx, t, "foobar")))
		buf.Reset()
	})

	// generate --edit foobar
	t.Run("generate --edit foobar", func(t *testing.T) { //nolint:paralleltest
		if testing.Short() || runtime.GOOS == "windows" {
			t.Skip("skipping test in short mode.")
		}

		assert.NoError(t, act.Generate(gptest.CliCtxWithFlags(ctx, t, map[string]string{"edit": "true", "editor": "/bin/cat"}, "foobar")))
		buf.Reset()
	})

	// generate --force foobar
	t.Run("generate --force foobar", func(t *testing.T) { //nolint:paralleltest
		if testing.Short() {
			t.Skip("skipping test in short mode.")
		}

		assert.NoError(t, act.Generate(gptest.CliCtxWithFlags(ctx, t, map[string]string{"force": "true"}, "foobar")))
		buf.Reset()
	})

	// generate --force foobar 32
	t.Run("generate --force foobar 32", func(t *testing.T) { //nolint:paralleltest
		if testing.Short() {
			t.Skip("skipping test in short mode.")
		}

		assert.NoError(t, act.Generate(gptest.CliCtxWithFlags(ctx, t, map[string]string{"force": "true"}, "foobar", "32")))
		buf.Reset()
	})

	// generate --force --symbols foobar 32
	t.Run("generate --force --symbols foobar 32", func(t *testing.T) { //nolint:paralleltest
		if testing.Short() {
			t.Skip("skipping test in short mode.")
		}

		assert.NoError(t, act.Generate(gptest.CliCtxWithFlags(ctx, t, map[string]string{"force": "true", "print": "true", "symbols": "true"}, "foobar", "32")))
		passIsAlphaNum(t, buf.String(), false)
		buf.Reset()
	})

	// generate --force --symbols=false foobar 32
	t.Run("generate --force --symbols=False foobar 32", func(t *testing.T) { //nolint:paralleltest
		if testing.Short() {
			t.Skip("skipping test in short mode.")
		}

		assert.NoError(t, act.Generate(gptest.CliCtxWithFlags(ctx, t, map[string]string{"force": "true", "print": "true", "symbols": "false"}, "foobar", "32")))
		passIsAlphaNum(t, buf.String(), true)
		buf.Reset()
	})

	// generate --force --xkcd foobar 32
	t.Run("generate --force --xkcd foobar 32", func(t *testing.T) { //nolint:paralleltest
		assert.NoError(t, act.Generate(gptest.CliCtxWithFlags(ctx, t, map[string]string{"force": "true", "xkcd": "true", "lang": "en"}, "foobar", "32")))
		buf.Reset()
	})

	// generate --force --xkcd foobar baz 32
	t.Run("generate --force --xkcd foobar baz 32", func(t *testing.T) { //nolint:paralleltest
		assert.NoError(t, act.Generate(gptest.CliCtxWithFlags(ctx, t, map[string]string{"force": "true", "xkcd": "true", "lang": "en"}, "foobar", "baz", "32")))
		buf.Reset()
	})

	// generate --force --xkcd foobar baz
	t.Run("generate --force --xkcd foobar baz", func(t *testing.T) { //nolint:paralleltest
		assert.NoError(t, act.Generate(gptest.CliCtxWithFlags(ctx, t, map[string]string{"force": "true", "xkcd": "true", "lang": "en"}, "foobar", "baz")))
		buf.Reset()
	})

	// generate --force --xkcd --print foobar baz
	t.Run("generate --force --xkcd --print foobar baz", func(t *testing.T) { //nolint:paralleltest
		assert.NoError(t, act.Generate(gptest.CliCtxWithFlags(ctx, t, map[string]string{"force": "true", "xkcd": "true", "print": "true", "lang": "en"}, "foobar", "baz")))
		buf.Reset()
	})

	// generate --force foobar 24 w/ autoclip and output redirection
	t.Run("generate --force foobar 24", func(t *testing.T) { //nolint:paralleltest
		ov := act.cfg.AutoClip
		defer func() {
			act.cfg.AutoClip = ov
		}()
		act.cfg.AutoClip = true
		ctx := ctxutil.WithTerminal(ctx, false)
		assert.NoError(t, act.Generate(gptest.CliCtxWithFlags(ctx, t, map[string]string{"force": "true"}, "foobar", "24")))
		assert.Contains(t, buf.String(), "Not printing secrets by default")
		buf.Reset()
	})

	// generate --force foobar 24 w/ autoclip and no output redirection
	t.Run("generate --force foobar 24", func(t *testing.T) { //nolint:paralleltest
		ov := act.cfg.AutoClip
		defer func() {
			act.cfg.AutoClip = ov
		}()
		act.cfg.AutoClip = true
		ctx := ctxutil.WithTerminal(ctx, true)
		assert.NoError(t, act.Generate(gptest.CliCtxWithFlags(ctx, t, map[string]string{"force": "true"}, "foobar", "24")))
		assert.Contains(t, buf.String(), "Copied to clipboard")
		buf.Reset()
	})

	// generate --force foobar w/ pw length set via env variable (42 chars)
	t.Run("generate --force foobar", func(t *testing.T) { //nolint:paralleltest
		t.Setenv("GOPASS_PW_DEFAULT_LENGTH", "42")

		if testing.Short() {
			t.Skip("skipping test in short mode.")
		}

		assert.NoError(t, act.Generate(gptest.CliCtxWithFlags(ctx, t, map[string]string{"force": "true", "print": "true", "symbols": "false"}, "foobar")))
		lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
		assert.Len(t, lines[3], 42)
		buf.Reset()
	})

	// generate --force foobar w/ pw length set via env variable to invalid value, fallback mechanism
	t.Run("generate --force foobar", func(t *testing.T) { //nolint:paralleltest
		t.Setenv("GOPASS_PW_DEFAULT_LENGTH", "0")

		if testing.Short() {
			t.Skip("skipping test in short mode.")
		}

		assert.NoError(t, act.Generate(gptest.CliCtxWithFlags(ctx, t, map[string]string{"force": "true", "print": "true", "symbols": "false"}, "foobar")))
		lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
		assert.Len(t, lines[3], 24) // 24 = default value used as fallback
		buf.Reset()
	})

}

func passIsAlphaNum(t *testing.T, buf string, want bool) {
	t.Helper()

	reAlphaNum := regexp.MustCompile(`^[A-Za-z0-9]+$`)
	lines := strings.Split(strings.TrimSpace(buf), "\n")
	if len(lines) < 1 {
		t.Errorf("buffer empty (no lines)")
	}
	line := strings.TrimSpace(lines[len(lines)-1])
	if reAlphaNum.MatchString(line) != want {
		t.Errorf("buffer did not match alpha num re: %s (%s)", line, buf)
	}
}

func TestKeyAndLength(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct { //nolint:paralleltest
		in     []string
		key    string
		length string
	}{
		{
			in:     []string{"32"},
			key:    "",
			length: "32",
		},
		{
			in:     []string{"baz"},
			key:    "baz",
			length: "",
		},
		{
			in:     []string{"baz", "32"},
			key:    "baz",
			length: "32",
		},
		{
			in:     []string{},
			key:    "",
			length: "",
		},
	} {
		tc := tc

		t.Run(fmt.Sprintf("%v", tc.in), func(t *testing.T) {
			t.Parallel()

			app := cli.NewApp()
			fs := flag.NewFlagSet("default", flag.ContinueOnError)
			assert.NoError(t, fs.Parse(append([]string{"foobar"}, tc.in...)))
			c := cli.NewContext(app, fs, nil)
			args, _ := parseArgs(c)
			k, l := keyAndLength(args)
			assert.Equal(t, tc.key, k, "Key from %+v", tc.in)
			assert.Equal(t, tc.length, l, "Length from %+v", tc.in)
		})
	}
}

func TestExtractEmails(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct { //nolint:paralleltest
		in  []string
		out []string
	}{
		{
			out: []string{},
		},
		{
			in:  []string{"some/mount/gmail.com/john.doe@example.org", "example.com/user@example.org"},
			out: []string{"john.doe@example.org", "user@example.org"},
		},
	} {
		tc := tc
		t.Run(fmt.Sprintf("%v", tc.in), func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tc.out, extractEmails(tc.in))
		})
	}
}

func TestExtractDomains(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct { //nolint:paralleltest
		in  []string
		out []string
	}{
		{
			out: []string{},
		},
		{
			in:  []string{"websites/gmail.com", "live.com", "some/mount/websites/web.de"},
			out: []string{"gmail.com", "live.com", "web.de"},
		},
	} {
		tc := tc
		t.Run(fmt.Sprintf("%v", tc.in), func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tc.out, extractDomains(tc.in))
		})
	}
}

func TestUniq(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct { //nolint:paralleltest
		in  []string
		out []string
	}{
		{
			out: []string{},
		},
		{
			in:  []string{"foo", "foo", "bar"},
			out: []string{"bar", "foo"},
		},
	} {
		tc := tc
		t.Run(fmt.Sprintf("%v", tc.in), func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tc.out, uniq(tc.in))
		})
	}
}

func TestFilterPrefix(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct { //nolint:paralleltest
		in     []string
		prefix string
		out    []string
	}{
		{
			out: []string{},
		},
		{
			in:     []string{"foo", "bar", "baz"},
			prefix: "foo",
			out:    []string{"foo"},
		},
		{
			in:     []string{"foo/bar", "foo/baz", "bar/foo"},
			prefix: "foo/",
			out:    []string{"foo/bar", "foo/baz"},
		},
	} {
		tc := tc
		t.Run(fmt.Sprintf("%v", tc.in), func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tc.out, filterPrefix(tc.in, tc.prefix))
		})
	}
}
