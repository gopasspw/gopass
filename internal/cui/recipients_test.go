package cui

import (
	"bytes"
	"context"
	"os"
	"testing"

	"github.com/gopasspw/gopass/internal/backend/crypto/plain"
	"github.com/gopasspw/gopass/internal/gptest"
	"github.com/gopasspw/gopass/pkg/ctxutil"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfirmRecipients(t *testing.T) {
	ctx := context.Background()
	ctx = ctxutil.WithTerminal(ctx, false)

	buf := &bytes.Buffer{}
	Stdout = buf
	defer func() {
		Stdout = os.Stdout
	}()

	// AlwaysYes true
	in := []string{"foo", "bar"}
	got, err := ConfirmRecipients(ctxutil.WithAlwaysYes(ctx, true), plain.New(), "test", in)
	require.NoError(t, err)
	assert.Equal(t, in, got)
	buf.Reset()

	// IsConfirm true
	in = []string{"foo", "bar"}
	got, err = ConfirmRecipients(ctxutil.WithConfirm(ctx, true), plain.New(), "test", in)
	require.NoError(t, err)
	assert.Equal(t, in, got)
	buf.Reset()

	// IsEditRecipients true
	in = []string{"foo", "bar"}
	_, err = ConfirmRecipients(ctxutil.WithEditRecipients(ctx, true), plain.New(), "test", in)
	assert.Error(t, err)
	buf.Reset()
}

func TestAskForPrivateKey(t *testing.T) {
	buf := &bytes.Buffer{}
	Stdout = buf
	defer func() {
		Stdout = os.Stdout
	}()

	ctx := context.Background()

	ctx = ctxutil.WithAlwaysYes(ctx, true)
	key, err := AskForPrivateKey(ctx, plain.New(), "test", "test")
	require.NoError(t, err)
	assert.Equal(t, "0xDEADBEEF", key)
	buf.Reset()
}

func TestAskForGitConfigUser(t *testing.T) {
	u := gptest.NewUnitTester(t)
	defer u.Remove()

	ctx := context.Background()
	ctx = ctxutil.WithTerminal(ctx, true)
	ctx = ctxutil.WithAlwaysYes(ctx, true)

	_, _, err := AskForGitConfigUser(ctx, plain.New(), "test")
	assert.NoError(t, err)
}

type fakeMountPointer []string

func (f fakeMountPointer) MountPoints() []string {
	return f
}

func TestAskForStore(t *testing.T) {
	ctx := context.Background()

	// test non-interactive
	ctx = ctxutil.WithInteractive(ctx, false)
	assert.Equal(t, "", AskForStore(ctx, fakeMountPointer{"foo", "bar"}))

	// test interactive
	ctx = ctxutil.WithInteractive(ctx, true)
	ctx = ctxutil.WithAlwaysYes(ctx, true)
	assert.Equal(t, "", AskForStore(ctx, fakeMountPointer{"foo", "bar"}))

	// test zero mps
	assert.Equal(t, "", AskForStore(ctx, fakeMountPointer{}))

	// test one mp
	assert.Equal(t, "", AskForStore(ctx, fakeMountPointer{"foo"}))

	// test two mps
	assert.Equal(t, "", AskForStore(ctx, fakeMountPointer{"foo", "bar"}))
}

func TestSorted(t *testing.T) {
	assert.Equal(t, []string{"a", "b", "c"}, sorted([]string{"c", "a", "b"}))
}
