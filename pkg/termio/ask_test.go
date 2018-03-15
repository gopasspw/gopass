package termio

import (
	"bytes"
	"context"
	"os"
	"strings"
	"testing"

	"github.com/justwatchcom/gopass/pkg/ctxutil"
	"github.com/justwatchcom/gopass/pkg/out"
	"github.com/stretchr/testify/assert"
)

func TestAskForString(t *testing.T) {
	buf := &bytes.Buffer{}
	out.Stdout = buf
	Stdout = buf
	defer func() {
		out.Stdout = os.Stdout
		Stdout = os.Stdout
	}()

	ctx := context.Background()
	ctx = ctxutil.WithAlwaysYes(ctx, true)

	sv, err := AskForString(ctx, "test", "foobar")
	assert.NoError(t, err)
	assert.Equal(t, "foobar", sv)

	t.Logf("Out: %s", buf.String())
	buf.Reset()

	// provide value on redirected stdin
	input := `foobaz
bar

`
	Stdin = strings.NewReader(input)
	ctx = ctxutil.WithAlwaysYes(ctx, false)
	sv, err = AskForString(ctx, "test", "foobar")
	assert.NoError(t, err)
	assert.Equal(t, "foobaz", sv)

	sv, err = AskForString(ctx, "test", "foobar")
	assert.NoError(t, err)
	assert.Equal(t, "bar", sv)
	Stdin = os.Stdin

	sv, err = AskForString(ctx, "test", "foobar")
	assert.NoError(t, err)
	assert.Equal(t, "foobar", sv)
	Stdin = os.Stdin

	t.Logf("Out: %s", buf.String())
	buf.Reset()
}

func TestAskForBool(t *testing.T) {
	buf := &bytes.Buffer{}
	out.Stdout = buf
	Stdout = buf
	defer func() {
		out.Stdout = os.Stdout
		Stdout = os.Stdout
	}()

	ctx := context.Background()
	ctx = ctxutil.WithAlwaysYes(ctx, true)

	bv, err := AskForBool(ctx, "test", false)
	assert.NoError(t, err)
	assert.Equal(t, false, bv)

	// provide value on redirected stdin
	input := `n
y
N
Y


z
`
	Stdin = strings.NewReader(input)
	ctx = ctxutil.WithAlwaysYes(ctx, false)
	bv, err = AskForBool(ctx, "test", true)
	assert.NoError(t, err)
	assert.Equal(t, false, bv)

	bv, err = AskForBool(ctx, "test", false)
	assert.NoError(t, err)
	assert.Equal(t, true, bv)

	bv, err = AskForBool(ctx, "test", true)
	assert.NoError(t, err)
	assert.Equal(t, false, bv)

	bv, err = AskForBool(ctx, "test", false)
	assert.NoError(t, err)
	assert.Equal(t, true, bv)

	bv, err = AskForBool(ctx, "test", true)
	assert.NoError(t, err)
	assert.Equal(t, true, bv)

	bv, err = AskForBool(ctx, "test", false)
	assert.NoError(t, err)
	assert.Equal(t, false, bv)

	bv, err = AskForBool(ctx, "test", false)
	assert.Error(t, err)
	assert.Equal(t, false, bv)
}

func TestAskForInt(t *testing.T) {
	buf := &bytes.Buffer{}
	out.Stdout = buf
	Stdout = buf
	defer func() {
		out.Stdout = os.Stdout
		Stdout = os.Stdout
	}()

	ctx := context.Background()
	ctx = ctxutil.WithAlwaysYes(ctx, true)

	got, err := AskForInt(ctx, "test", 42)
	assert.NoError(t, err)
	assert.Equal(t, 42, got)

	// provide value on redirected stdin
	input := `23
-1
0xDEADBEEF
0755
0.123

`
	Stdin = strings.NewReader(input)
	ctx = ctxutil.WithAlwaysYes(ctx, false)

	iv, err := AskForInt(ctx, "test", 42)
	assert.NoError(t, err)
	assert.Equal(t, 23, iv)
}

func TestAskForConfirmation(t *testing.T) {
	buf := &bytes.Buffer{}
	out.Stdout = buf
	Stdout = buf
	defer func() {
		out.Stdout = os.Stdout
		Stdout = os.Stdout
	}()

	ctx := context.Background()
	ctx = ctxutil.WithAlwaysYes(ctx, true)
	assert.Equal(t, true, AskForConfirmation(ctx, "test"))

	// provide value on redirected stdin
	input := `y
n
`
	for i := 0; i < maxTries+1; i++ {
		input += "z\n"
	}

	Stdin = strings.NewReader(input)
	ctx = ctxutil.WithAlwaysYes(ctx, false)

	assert.Equal(t, true, AskForConfirmation(ctx, "test"))
	assert.Equal(t, false, AskForConfirmation(ctx, "test"))
	assert.Equal(t, false, AskForConfirmation(ctx, "test"))
}

func TestAskForKeyImport(t *testing.T) {
	buf := &bytes.Buffer{}
	out.Stdout = buf
	Stdout = buf
	defer func() {
		out.Stdout = os.Stdout
		Stdout = os.Stdout
	}()

	ctx := context.Background()
	ctx = ctxutil.WithAlwaysYes(ctx, true)

	assert.Equal(t, true, AskForKeyImport(ctx, "test", []string{}))

	// provide value on redirected stdin
	input := `y
n
z
`

	Stdin = strings.NewReader(input)
	ctx = ctxutil.WithAlwaysYes(ctx, false)
	assert.Equal(t, false, AskForKeyImport(ctxutil.WithInteractive(ctx, false), "", nil))
	assert.Equal(t, true, AskForKeyImport(ctx, "", nil))
	assert.Equal(t, false, AskForKeyImport(ctx, "", nil))
	assert.Equal(t, false, AskForKeyImport(ctx, "", nil))
}

func TestAskForPasswordNonInteractive(t *testing.T) {
	buf := &bytes.Buffer{}
	out.Stdout = buf
	Stdout = buf
	defer func() {
		out.Stdout = os.Stdout
		Stdout = os.Stdout
	}()

	ctx := context.Background()
	ctx = ctxutil.WithInteractive(ctx, false)

	_, err := AskForPassword(ctx, "test")
	assert.Error(t, err)

	// provide value on redirected stdin
	input := `foo
foo
foobar
foobaz
`

	Stdin = strings.NewReader(input)
	ctx = ctxutil.WithAlwaysYes(ctx, false)
	ctx = ctxutil.WithInteractive(ctx, true)
	ctx = ctxutil.WithTerminal(ctx, false)
	sv, err := AskForPassword(ctx, "test")
	assert.NoError(t, err)
	assert.Equal(t, "foo", sv)

	sv, err = AskForPassword(ctx, "test")
	assert.NoError(t, err)
	assert.Equal(t, "", sv)
}

func TestAskForPasswordInteractive(t *testing.T) {
	buf := &bytes.Buffer{}
	out.Stdout = buf
	Stdout = buf
	defer func() {
		out.Stdout = os.Stdout
		Stdout = os.Stdout
	}()

	ctx := context.Background()
	askFn := func(ctx context.Context, prompt string) (string, error) {
		return "test", nil
	}
	ctx = ctxutil.WithInteractive(ctx, true)
	ctx = WithPassPromptFunc(ctx, askFn)

	pw, err := AskForPassword(ctx, "test")
	assert.NoError(t, err)
	assert.Equal(t, "test", pw)
}
