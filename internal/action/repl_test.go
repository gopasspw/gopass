package action

import (
	"bytes"
	"os"
	"testing"

	"github.com/ergochat/readline"
	"github.com/fatih/color"
	"github.com/gopasspw/gopass/internal/config"
	"github.com/gopasspw/gopass/internal/out"
	"github.com/gopasspw/gopass/pkg/ctxutil"
	"github.com/gopasspw/gopass/tests/gptest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v2"
)

func TestREPL(t *testing.T) {
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

	rl, err := readline.NewEx(&readline.Config{
		Prompt: "gopass> ",
		Stdin:  bytes.NewBufferString("help\nquit\n"),
	})
	require.NoError(t, err)

	defer func() {
		_ = rl.Close()
	}()

	err = act.REPL(gptest.CliCtx(ctx, t))
	require.NoError(t, err)
	assert.Contains(t, buf.String(), "help")
}

func TestEntriesForCompleter(t *testing.T) {
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

	completers, err := act.entriesForCompleter(ctx)
	require.NoError(t, err)
	assert.Len(t, completers, 1)
}

func TestReplCompleteRecipients(t *testing.T) {
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

	cmd := &cli.Command{
		Name: "remove",
	}

	completers := act.replCompleteRecipients(ctx, cmd)
	assert.Len(t, completers, 1)
}

func TestReplCompleteTemplates(t *testing.T) {
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

	cmd := &cli.Command{
		Name: "templates",
	}

	completers := act.replCompleteTemplates(ctx, cmd)
	assert.Len(t, completers, 1)
}

func TestEscapeEntry(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "no special chars",
			input:    "simple",
			expected: "simple",
		},
		{
			name:     "spaces",
			input:    "this is a test",
			expected: `this\ is\ a\ test`,
		},
		{
			name:     "backslash",
			input:    `back\slash`,
			expected: `back\\slash`,
		},
		{
			name:     "single quotes",
			input:    "it's",
			expected: `it\'s`,
		},
		{
			name:     "double quotes",
			input:    `say "hello"`,
			expected: `say\ \"hello\"`,
		},
		{
			name:     "path separators preserved",
			input:    "folder/my entry",
			expected: `folder/my\ entry`,
		},
		{
			name:     "special chars",
			input:    "a<>&;#|*?()",
			expected: `a\<\>\&\;\#\|\*\?\(\)`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := escapeEntry(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestUnescapeEntry(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "no escapes",
			input:    "simple",
			expected: "simple",
		},
		{
			name:     "escaped spaces",
			input:    `this\ is\ a\ test`,
			expected: "this is a test",
		},
		{
			name:     "escaped backslash",
			input:    `back\\slash`,
			expected: `back\slash`,
		},
		{
			name:     "roundtrip",
			input:    escapeEntry("hello world"),
			expected: "hello world",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := unescapeEntry(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}
