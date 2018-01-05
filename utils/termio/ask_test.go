package termio

import (
	"bytes"
	"context"
	"os"
	"strings"
	"testing"

	"github.com/justwatchcom/gopass/utils/ctxutil"
	"github.com/justwatchcom/gopass/utils/out"
	"github.com/stretchr/testify/assert"
)

func TestAskForConfirmation(t *testing.T) {
	ctx := context.Background()
	ctx = ctxutil.WithAlwaysYes(ctx, true)
	assert.Equal(t, true, AskForConfirmation(ctx, "test"))
}

func TestAskForBool(t *testing.T) {
	ctx := context.Background()
	ctx = ctxutil.WithAlwaysYes(ctx, true)

	bv, err := AskForBool(ctx, "test", false)
	assert.NoError(t, err)
	assert.Equal(t, false, bv)
}

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

	t.Logf("Out: %s", buf.String())
	buf.Reset()
}

func TestAskForInt(t *testing.T) {
	ctx := context.Background()
	ctx = ctxutil.WithAlwaysYes(ctx, true)

	got, err := AskForInt(ctx, "test", 42)
	assert.NoError(t, err)
	assert.Equal(t, 42, got)
}

func TestAskForPasswordNonInteractive(t *testing.T) {
	ctx := context.Background()
	ctx = ctxutil.WithInteractive(ctx, false)

	_, err := AskForPassword(ctx, "test")
	assert.Error(t, err)
}

func TestAskForPasswordInteractive(t *testing.T) {
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

func TestAskForKeyImport(t *testing.T) {
	ctx := context.Background()
	ctx = ctxutil.WithAlwaysYes(ctx, true)

	assert.Equal(t, true, AskForKeyImport(ctx, "test", []string{}))
}
