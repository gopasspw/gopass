package action

import (
	"context"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/justwatchcom/gopass/utils/ctxutil"
	"github.com/stretchr/testify/assert"
)

func TestConfirmRecipients(t *testing.T) {
	td, err := ioutil.TempDir("", "gopass-")
	assert.NoError(t, err)
	defer func() {
		_ = os.RemoveAll(td)
	}()

	ctx := context.Background()
	act, err := newMock(ctx, td)
	assert.NoError(t, err)

	ctx = ctxutil.WithAlwaysYes(ctx, true)

	in := []string{"foo", "bar"}
	got, err := act.ConfirmRecipients(ctx, "test", in)
	assert.NoError(t, err)
	assert.Equal(t, in, got)
}

func TestAskForPrivateKey(t *testing.T) {
	td, err := ioutil.TempDir("", "gopass-")
	assert.NoError(t, err)
	defer func() {
		_ = os.RemoveAll(td)
	}()

	ctx := context.Background()
	act, err := newMock(ctx, td)
	assert.NoError(t, err)

	ctx = ctxutil.WithAlwaysYes(ctx, true)
	key, err := act.askForPrivateKey(ctx, "test")
	assert.NoError(t, err)
	assert.Equal(t, "000000000000000000000000DEADBEEF", key)
}

func TestAskForGitConfigUser(t *testing.T) {
	td, err := ioutil.TempDir("", "gopass-")
	assert.NoError(t, err)
	defer func() {
		_ = os.RemoveAll(td)
	}()

	ctx := context.Background()
	act, err := newMock(ctx, td)
	assert.NoError(t, err)

	ctx = ctxutil.WithTerminal(ctx, true)
	ctx = ctxutil.WithAlwaysYes(ctx, true)

	_, _, err = act.askForGitConfigUser(ctx)
	assert.NoError(t, err)
}

func TestAskForGitConfigUserNonInteractive(t *testing.T) {
	td, err := ioutil.TempDir("", "gopass-")
	assert.NoError(t, err)
	defer func() {
		_ = os.RemoveAll(td)
	}()

	ctx := context.Background()
	act, err := newMock(ctx, td)
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
	td, err := ioutil.TempDir("", "gopass-")
	assert.NoError(t, err)
	defer func() {
		_ = os.RemoveAll(td)
	}()

	ctx := context.Background()
	act, err := newMock(ctx, td)
	assert.NoError(t, err)

	sd1 := filepath.Join(td, "sub1")
	sd2 := filepath.Join(td, "sub2")

	assert.NoError(t, newStore(sd1))
	assert.NoError(t, newStore(sd2))

	assert.NoError(t, act.Store.AddMount(ctx, "sub1", sd1))
	assert.NoError(t, act.Store.AddMount(ctx, "sub2", sd2))

	ctx = ctxutil.WithInteractive(ctx, false)
	sel := act.askForStore(ctx)
	if sel != "" {
		t.Errorf("Wrong selection: %s", sel)
	}
}
