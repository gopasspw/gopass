package action

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/gopasspw/gopass/internal/backend"
	"github.com/gopasspw/gopass/internal/backend/crypto/plain"
	"github.com/gopasspw/gopass/internal/config"
	"github.com/gopasspw/gopass/internal/out"
	"github.com/gopasspw/gopass/pkg/ctxutil"
	"github.com/gopasspw/gopass/tests/gptest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v3"
)

// TestSetupTeamCreateFlagsRegistered verifies the renamed --team/--create-team
// flags are registered on the setup command with their deprecated aliases
// (--alias/--create), so existing scripts keep working. See GH-3497.
func TestSetupTeamCreateFlagsRegistered(t *testing.T) {
	u := gptest.NewUnitTester(t)

	ctx := config.NewContextInMemory()
	act, err := newMock(ctx, u.StoreDir(""))
	require.NoError(t, err)
	require.NotNil(t, act)

	setupCmd := findCommand(act.GetCommands(), "setup")
	require.NotNil(t, setupCmd)

	var team *cli.StringFlag
	var createTeam *cli.BoolFlag
	for _, f := range setupCmd.Flags {
		switch v := f.(type) {
		case *cli.StringFlag:
			if v.Name == "team" {
				team = v
			}
		case *cli.BoolFlag:
			if v.Name == "create-team" {
				createTeam = v
			}
		}
	}

	require.NotNil(t, team)
	assert.Contains(t, team.Aliases, "alias")

	require.NotNil(t, createTeam)
	assert.Contains(t, createTeam.Aliases, "create")
}

// TestSetupCreateTeamRequiresName verifies that requesting team creation
// without a team name fails fast with an actionable error message when
// running non-interactively (e.g. scripted setups), instead of the old,
// confusing "can not create a team without a team name". See GH-3497.
func TestSetupCreateTeamRequiresName(t *testing.T) {
	u := gptest.NewUnitTester(t)

	ctx := config.NewContextInMemory()
	ctx = ctxutil.WithAlwaysYes(ctx, true)
	ctx = ctxutil.WithInteractive(ctx, false)
	ctx = backend.WithCryptoBackend(ctx, backend.Age)
	ctx = backend.WithStorageBackend(ctx, backend.GitFS)
	ctx = ctxutil.WithAgePassphrase(ctx, "foobar")

	act, err := newMock(ctx, u.StoreDir(""))
	require.ErrorContains(t, err, "not initialized")
	require.NotNil(t, act)

	require.NoError(t, os.RemoveAll(u.StoreDir("")))
	require.NoError(t, os.Remove(u.GPConfig()))

	c := gptest.CliCtxWithFlags(ctx, t, map[string]string{"storage": "gitfs", "crypto": "age", "create-team": "true"})
	err = act.Setup(ctx, c)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "team name")
}

func TestSetupAgeGitFS(t *testing.T) {
	u := gptest.NewUnitTester(t) //nolint:staticcheck

	ctx := config.NewContextInMemory()
	ctx = ctxutil.WithAlwaysYes(ctx, true)
	ctx = ctxutil.WithInteractive(ctx, false)
	ctx = backend.WithCryptoBackend(ctx, backend.Age)
	ctx = backend.WithStorageBackend(ctx, backend.GitFS)
	ctx = ctxutil.WithAgePassphrase(ctx, "foobar")

	act, err := newMock(ctx, u.StoreDir(""))
	require.ErrorContains(t, err, "not initialized")
	require.NotNil(t, act)

	buf := &bytes.Buffer{}
	out.Stdout = buf
	out.Stderr = buf
	defer func() {
		out.Stdout = os.Stdout
		out.Stderr = os.Stderr
	}()

	// remove existing config and store, we want to start fresh
	require.NoError(t, os.RemoveAll(u.StoreDir("")))
	require.NoError(t, os.Remove(u.GPConfig()))

	c := gptest.CliCtxWithFlags(ctx, t, map[string]string{"storage": "gitfs", "crypto": "age"})
	_, errIsInit := act.IsInitialized(ctx, c)

	require.Error(t, errIsInit)
	require.NoError(t, act.Setup(ctx, c))
	assert.Contains(t, buf.String(), "Welcome to gopass")

	crypto := act.Store.Crypto(ctx, "")
	require.NotNil(t, crypto)
	assert.Equal(t, "age", crypto.Name())
	assert.True(t, act.initHasUseablePrivateKeys(ctx, crypto))
	require.NoError(t, act.initGenerateIdentity(ctx, crypto, "foo bar", "foo.bar@example.org"))
	buf.Reset()

	act.printRecipients(ctx, "")
	assert.Contains(t, buf.String(), "age1")
	buf.Reset()
}

func TestSetupPlainFS(t *testing.T) {
	u := gptest.NewUnitTester(t)

	ctx := config.NewContextInMemory()
	ctx = ctxutil.WithAlwaysYes(ctx, true)
	ctx = ctxutil.WithInteractive(ctx, false)
	ctx = backend.WithCryptoBackend(ctx, backend.Plain)
	ctx = backend.WithStorageBackend(ctx, backend.FS)

	act, err := newMock(ctx, u.StoreDir(""))
	require.NoError(t, err)
	require.NotNil(t, act)

	buf := &bytes.Buffer{}
	out.Stdout = buf
	out.Stderr = buf
	defer func() {
		out.Stdout = os.Stdout
		out.Stderr = os.Stderr
	}()

	c := gptest.CliCtx(ctx, t, "foo.bar@example.org")
	_, errIsInit := act.IsInitialized(ctx, c)

	require.NoError(t, errIsInit)
	buf.Reset()

	require.Error(t, act.Init(ctx, c))
	assert.Contains(t, buf.String(), "already initialized")
	buf.Reset()

	// this will abort because the store is already initialized
	require.NoError(t, act.Setup(ctx, c))
	assert.Contains(t, buf.String(), "already initialized")
	buf.Reset()

	crypto := act.Store.Crypto(ctx, "")
	require.NotNil(t, crypto)
	assert.Equal(t, "plain", crypto.Name())
	assert.True(t, act.initHasUseablePrivateKeys(ctx, crypto))
	require.Error(t, act.initGenerateIdentity(ctx, crypto, "foo bar", "foo.bar@example.org"))
	buf.Reset()

	act.printRecipients(ctx, "")
	assert.Contains(t, buf.String(), "0xDEADBEEF")
	buf.Reset()

	// un-initialize the store
	require.NoError(t, os.Remove(filepath.Join(u.StoreDir(""), plain.IDFile)))
	_, errIsInit = act.IsInitialized(ctx, c)

	require.Error(t, errIsInit)
	buf.Reset()

	// remove existing config and store
	require.NoError(t, os.RemoveAll(u.StoreDir("")))
	require.NoError(t, os.Remove(u.GPConfig()))

	// re-initialize the store, i.e. test that a fresh setup with plain and fs works
	require.NoError(t, act.Setup(ctx, c))
	assert.Contains(t, buf.String(), "Welcome to gopass")
	buf.Reset()
}
