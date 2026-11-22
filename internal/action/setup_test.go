package action

import (
	"bytes"
	"context"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gopasspw/gopass/internal/backend"
	"github.com/gopasspw/gopass/internal/backend/crypto/plain"
	"github.com/gopasspw/gopass/internal/config"
	"github.com/gopasspw/gopass/internal/out"
	"github.com/gopasspw/gopass/pkg/ctxutil"
	"github.com/gopasspw/gopass/pkg/termio"
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
	assert.NotContains(t, buf.String(), "We need to unlock your newly created private key now!")

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

type setupStubCrypto struct {
	*plain.Mocker
	name string
}

func (s *setupStubCrypto) Name() string {
	return s.name
}

func (s *setupStubCrypto) GenerateIdentity(context.Context, string, string, string) (string, error) {
	return "0xDEADBEEF", nil
}

func (s *setupStubCrypto) ListIdentities(context.Context) ([]string, error) {
	return []string{"0xDEADBEEF"}, nil
}

// TestInitGenerateIdentityUnlockNotice verifies that initGenerateIdentity only
// prints the unlock passphrase notice when the backend will actually prompt the
// user to unlock the newly generated key (age with a passphrase-protected
// keyring and no pre-configured passphrase/callback). See GH-2862.
func TestInitGenerateIdentityUnlockNotice(t *testing.T) {
	u := gptest.NewUnitTester(t)

	const unlockPrefix = "We need to unlock your newly created private key now!"

	baseCtx := config.NewContextInMemory()
	act, err := newMock(baseCtx, u.StoreDir(""))
	require.NoError(t, err)

	tests := []struct {
		name        string
		cryptoName  string
		setupCtx    func(context.Context) context.Context
		stdin       string
		wantContain string
		wantAbsent  string
	}{
		{
			name:       "gpgcli generated passphrase does not prompt to unlock",
			cryptoName: "gpgcli",
			setupCtx: func(ctx context.Context) context.Context {
				return ctxutil.WithAlwaysYes(ctx, true)
			},
			wantAbsent: unlockPrefix,
		},
		{
			name:       "gpgcli user passphrase does not prompt to unlock",
			cryptoName: "gpgcli",
			setupCtx: func(ctx context.Context) context.Context {
				ctx = ctxutil.WithInteractive(ctx, true)
				ctx = ctxutil.WithAlwaysYes(ctx, false)

				return termio.WithPassPromptFunc(ctx, func(context.Context, string) (string, error) {
					return "secret", nil
				})
			},
			// name, email, enter passphrase=y, continue=y, export pubkey=n
			stdin:      "\n\ny\ny\nn\n",
			wantAbsent: unlockPrefix,
		},
		{
			name:       "age generated passphrase prompts to unlock",
			cryptoName: "age",
			setupCtx: func(ctx context.Context) context.Context {
				return ctxutil.WithAlwaysYes(ctx, true)
			},
			wantContain: "Please enter the passphrase you just generated.",
		},
		{
			name:       "age user passphrase prompts without generated wording",
			cryptoName: "age",
			setupCtx: func(ctx context.Context) context.Context {
				ctx = ctxutil.WithInteractive(ctx, true)
				ctx = ctxutil.WithAlwaysYes(ctx, false)

				return termio.WithPassPromptFunc(ctx, func(context.Context, string) (string, error) {
					return "secret", nil
				})
			},
			// enter passphrase=y, export pubkey=n
			stdin:       "y\nn\n",
			wantContain: "Please enter your passphrase.",
			wantAbsent:  "you just generated",
		},
		{
			name:       "age with preset passphrase does not prompt to unlock",
			cryptoName: "age",
			setupCtx: func(ctx context.Context) context.Context {
				ctx = ctxutil.WithAlwaysYes(ctx, true)

				return ctxutil.WithAgePassphrase(ctx, "foobar")
			},
			wantAbsent: unlockPrefix,
		},
		{
			name:       "age with password callback does not prompt to unlock",
			cryptoName: "age",
			setupCtx: func(ctx context.Context) context.Context {
				ctx = ctxutil.WithAlwaysYes(ctx, true)

				return ctxutil.WithPasswordCallback(ctx, func(string, bool) ([]byte, error) {
					return []byte("foobar"), nil
				})
			},
			wantAbsent: unlockPrefix,
		},
		{
			name:       "age with keyring-recipients does not prompt for passphrase",
			cryptoName: "age",
			setupCtx: func(ctx context.Context) context.Context {
				ctx = ctxutil.WithAlwaysYes(ctx, true)
				cfg := config.NewInMemory()
				require.NoError(t, cfg.Set("", "age.keyring-recipients", "age1ql3z7hjy54pw3hyww5ayyfg7zqgvc7w3j2elw8zmrj2kg5sfn9aqmcac8p"))

				return cfg.WithConfig(ctx)
			},
			wantAbsent: unlockPrefix,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			buf := &bytes.Buffer{}
			oldOut, oldErr := out.Stdout, out.Stderr
			oldStdin, oldTermErr := termio.Stdin, termio.Stderr
			out.Stdout = buf
			out.Stderr = buf
			termio.Stderr = io.Discard
			if tc.stdin != "" {
				termio.Stdin = strings.NewReader(tc.stdin)
			}
			t.Cleanup(func() {
				out.Stdout = oldOut
				out.Stderr = oldErr
				termio.Stdin = oldStdin
				termio.Stderr = oldTermErr
			})

			ctx := config.NewContextInMemory()
			ctx = tc.setupCtx(ctx)
			crypto := &setupStubCrypto{Mocker: plain.New(), name: tc.cryptoName}

			require.NoError(t, act.initGenerateIdentity(ctx, crypto, "foo bar", "foo.bar@example.org"))

			got := buf.String()
			if tc.wantContain != "" {
				assert.Contains(t, got, tc.wantContain)
			}
			if tc.wantAbsent != "" {
				assert.NotContains(t, got, tc.wantAbsent)
			}
		})
	}
}
