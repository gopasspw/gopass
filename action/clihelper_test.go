package action

import (
	"bytes"
	"context"
	"os"
	"testing"

	"github.com/justwatchcom/gopass/tests/gptest"
	"github.com/justwatchcom/gopass/utils/ctxutil"
	"github.com/stretchr/testify/assert"
)

func TestConfirmRecipients(t *testing.T) {
	u := gptest.NewUnitTester(t)
	defer u.Remove()

	buf := &bytes.Buffer{}
	stdout = buf
	defer func() {
		stdout = os.Stdout
	}()

	ctx := context.Background()
	act, err := newMock(ctx, u)
	assert.NoError(t, err)

	// AlwaysYes true
	in := []string{"foo", "bar"}
	got, err := act.ConfirmRecipients(ctxutil.WithAlwaysYes(ctx, true), "test", in)
	assert.NoError(t, err)
	assert.Equal(t, in, got)
	buf.Reset()

	// IsNoConfirm true
	in = []string{"foo", "bar"}
	got, err = act.ConfirmRecipients(ctxutil.WithNoConfirm(ctx, true), "test", in)
	assert.NoError(t, err)
	assert.Equal(t, in, got)
	buf.Reset()
}

func TestAskForPrivateKey(t *testing.T) {
	u := gptest.NewUnitTester(t)
	defer u.Remove()

	buf := &bytes.Buffer{}
	stdout = buf
	defer func() {
		stdout = os.Stdout
	}()

	ctx := context.Background()
	act, err := newMock(ctx, u)
	assert.NoError(t, err)

	ctx = ctxutil.WithAlwaysYes(ctx, true)
	key, err := act.askForPrivateKey(ctx, "test", "test")
	assert.NoError(t, err)
	assert.Equal(t, "0xDEADBEEF", key)
	buf.Reset()
}

func TestAskForGitConfigUser(t *testing.T) {
	u := gptest.NewUnitTester(t)
	defer u.Remove()

	ctx := context.Background()
	act, err := newMock(ctx, u)
	assert.NoError(t, err)

	ctx = ctxutil.WithTerminal(ctx, true)
	ctx = ctxutil.WithAlwaysYes(ctx, true)

	_, _, err = act.askForGitConfigUser(ctx, "test")
	assert.NoError(t, err)
}

func TestAskForStore(t *testing.T) {
	u := gptest.NewUnitTester(t)
	defer u.Remove()

	ctx := context.Background()
	act, err := newMock(ctx, u)
	assert.NoError(t, err)

	assert.NoError(t, u.InitStore("sub1"))
	assert.NoError(t, u.InitStore("sub2"))

	assert.NoError(t, act.Store.AddMount(ctx, "sub1", u.StoreDir("sub1")))
	assert.NoError(t, act.Store.AddMount(ctx, "sub2", u.StoreDir("sub2")))

	ctx = ctxutil.WithInteractive(ctx, false)
	assert.Equal(t, "", act.askForStore(ctx))

	ctx = ctxutil.WithInteractive(ctx, true)
	ctx = ctxutil.WithAlwaysYes(ctx, true)
	assert.Equal(t, "", act.askForStore(ctx))
}
