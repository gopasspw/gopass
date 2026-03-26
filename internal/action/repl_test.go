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

func TestNewGopassCompleter(t *testing.T) {
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

	gc := act.newGopassCompleter(gptest.CliCtx(ctx, t))
	require.NotNil(t, gc)
	assert.NotEmpty(t, gc.entries)
}

func TestGopassCompleterDo(t *testing.T) {
	gc := &gopassCompleter{
		cmdSpecs: map[string]completionSpec{
			"show":   completeEntries,
			"cat":    completeEntries,
			"config": completeConfig,
		},
		subCmds:    map[string][]string{},
		entries:    []string{"foo", "bar", "this is just a test", "folder/my entry"},
		configKeys: []string{"autosync", "autopush"},
		commands:   []string{"cat", "config", "show"},
	}

	t.Run("complete command", func(t *testing.T) {
		line := []rune("sh")
		matches, length := gc.Do(line, len(line))
		require.Len(t, matches, 1)
		assert.Equal(t, "ow ", string(matches[0]))
		assert.Equal(t, 2, length)
	})

	t.Run("complete entry without spaces", func(t *testing.T) {
		line := []rune("show fo")
		matches, length := gc.Do(line, len(line))
		require.NotEmpty(t, matches)
		assert.GreaterOrEqual(t, len(matches), 1)
		assert.Equal(t, 2, length)
	})

	t.Run("complete entry with spaces", func(t *testing.T) {
		line := []rune(`show this`)
		matches, length := gc.Do(line, len(line))
		require.Len(t, matches, 1)
		assert.Equal(t, `\ is\ just\ a\ test `, string(matches[0]))
		assert.Equal(t, 4, length)
	})

	t.Run("complete entry with partial escaped space", func(t *testing.T) {
		line := []rune(`show this\ is`)
		matches, length := gc.Do(line, len(line))
		require.Len(t, matches, 1)
		assert.Equal(t, `\ just\ a\ test `, string(matches[0]))
		assert.Equal(t, len([]rune(`this\ is`)), length)
	})

	t.Run("complete config keys", func(t *testing.T) {
		line := []rune("config auto")
		matches, length := gc.Do(line, len(line))
		require.Len(t, matches, 2)
		assert.Equal(t, 4, length)
	})

	t.Run("complete with flags present", func(t *testing.T) {
		line := []rune("show -c fo")
		matches, length := gc.Do(line, len(line))
		require.NotEmpty(t, matches)
		assert.Equal(t, 2, length)
	})

	t.Run("empty line completes commands", func(t *testing.T) {
		line := []rune("")
		matches, length := gc.Do(line, len(line))
		require.Len(t, matches, 3)
		assert.Equal(t, 0, length)
	})

	t.Run("command with trailing space shows all entries", func(t *testing.T) {
		line := []rune("show ")
		matches, length := gc.Do(line, len(line))
		require.Len(t, matches, 4)
		assert.Equal(t, 0, length)
	})
}
