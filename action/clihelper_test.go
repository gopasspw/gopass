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

	ctx = ctxutil.WithAlwaysYes(ctx, true)

	in := []string{"foo", "bar"}
	got, err := act.ConfirmRecipients(ctx, "test", in)
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
	key, err := act.askForPrivateKey(ctx, "test")
	assert.NoError(t, err)
	assert.Equal(t, "000000000000000000000000DEADBEEF", key)
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

	_, _, err = act.askForGitConfigUser(ctx)
	assert.NoError(t, err)
}

func TestAskForGitConfigUserNonInteractive(t *testing.T) {
	u := gptest.NewUnitTester(t)
	defer u.Remove()

	ctx := context.Background()
	act, err := newMock(ctx, u)
	assert.NoError(t, err)

	ctx = ctxutil.WithTerminal(ctx, false)

	keyList, err := act.gpg.ListPrivateKeys(ctx)
	assert.NoError(t, err)

	name, email, _ := act.askForGitConfigUser(ctx)

	// unit tests cannot know whether keyList returned empty or not.
	// a better distinction would require mocking/patching
	// calls to s.gpg.ListPrivateKeys()
	if len(keyList) > 0 {
		assert.NotEqual(t, "", name)
		assert.NotEqual(t, "", email)
	} else {
		assert.Equal(t, "", name)
		assert.Equal(t, "", email)
	}
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
}
