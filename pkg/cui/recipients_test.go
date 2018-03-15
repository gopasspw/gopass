package cui

import (
	"bytes"
	"context"
	"os"
	"testing"

	"github.com/justwatchcom/gopass/pkg/backend/crypto/plain"
	"github.com/justwatchcom/gopass/pkg/ctxutil"
	"github.com/justwatchcom/gopass/tests/gptest"
	"github.com/stretchr/testify/assert"
)

func TestConfirmRecipients(t *testing.T) {
	ctx := context.Background()

	buf := &bytes.Buffer{}
	Stdout = buf
	defer func() {
		Stdout = os.Stdout
	}()

	// AlwaysYes true
	in := []string{"foo", "bar"}
	got, err := ConfirmRecipients(ctxutil.WithAlwaysYes(ctx, true), plain.New(), "test", in)
	assert.NoError(t, err)
	assert.Equal(t, in, got)
	buf.Reset()

	// IsNoConfirm true
	in = []string{"foo", "bar"}
	got, err = ConfirmRecipients(ctxutil.WithNoConfirm(ctx, true), plain.New(), "test", in)
	assert.NoError(t, err)
	assert.Equal(t, in, got)
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
	assert.NoError(t, err)
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

type fakeMountPointer struct{}

func (f *fakeMountPointer) MountPoints() []string {
	return []string{"foo", "bar"}
}

func TestAskForStore(t *testing.T) {
	ctx := context.Background()

	ctx = ctxutil.WithInteractive(ctx, false)
	assert.Equal(t, "", AskForStore(ctx, &fakeMountPointer{}))

	ctx = ctxutil.WithInteractive(ctx, true)
	ctx = ctxutil.WithAlwaysYes(ctx, true)
	assert.Equal(t, "", AskForStore(ctx, &fakeMountPointer{}))
}
