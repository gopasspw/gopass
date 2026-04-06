package age

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"filippo.io/age"
	"github.com/blang/semver/v4"
	"github.com/cenkalti/backoff/v4"
	"github.com/gopasspw/gopass/internal/backend/crypto/age/agent"
	"github.com/gopasspw/gopass/internal/cache"
	"github.com/gopasspw/gopass/internal/cache/ghssh"
	"github.com/gopasspw/gopass/internal/config"
	"github.com/gopasspw/gopass/pkg/appdir"
	"github.com/gopasspw/gopass/pkg/ctxutil"
	"github.com/gopasspw/gopass/pkg/debug"
)

const (
	// Ext is the file extension for age encrypted secrets.
	Ext = "age"
	// IDFile is the name for age recipients.
	IDFile = ".age-recipients"
)

type githubSSHCacher interface {
	ListKeys(ctx context.Context, user string) ([]string, error)
	String() string
}

// Age is an age backend.
type Age struct {
	identity        string
	ghCache         githubSSHCacher
	askPass         *askPass
	recpCache       *cache.OnDisk
	sshKeyPath      string // custom SSH key or directory path
	pwCallback      func(string, bool) ([]byte, error)
	pwPurgeCallback func(string)
}

// New creates a new Age backend.
func New(ctx context.Context, sshKeyPath string) (*Age, error) {
	ghc, err := ghssh.New()
	if err != nil {
		return nil, err
	}

	rc, err := cache.NewOnDisk("age-identity-recipients", 30*time.Hour)
	if err != nil {
		return nil, err
	}

	a := &Age{
		ghCache:    ghc,
		recpCache:  rc,
		identity:   filepath.Join(appdir.UserConfig(), "age", "identities"),
		askPass:    newAskPass(ctx),
		sshKeyPath: sshKeyPath,
	}

	// Capture any pre-configured passphrase (e.g. from GOPASS_AGE_PASSWORD).
	if ap := ctxutil.GetAgePassphrase(ctx); ap != "" {
		debug.Log("age: using pre-configured passphrase from context")
		a.pwCallback = func(_ string, _ bool) ([]byte, error) { return []byte(ap), nil }
		a.pwPurgeCallback = func(_ string) {} // no-op for static passwords
	}

	a.tryStartAgent(ctx)

	debug.Log("age initialized (ghc: %s, recipients: %s, identity: %s, sshKeyPath: %s)", a.ghCache.String(), a.recpCache.String(), a.identity, a.sshKeyPath)

	return a, nil
}

// SetPasswordCallback configures an external callback for obtaining
// the password used to encrypt/decrypt the age identity file.
// When set it takes precedence over the built-in interactive askPass prompt.
func (a *Age) SetPasswordCallback(cb func(string, bool) ([]byte, error)) {
	a.pwCallback = cb
}

// SetPasswordPurgeCallback configures an external callback that is invoked
// when a cached password should be invalidated (e.g. after a decrypt failure).
func (a *Age) SetPasswordPurgeCallback(cb func(string)) {
	a.pwPurgeCallback = cb
}

// effectivePwCallback returns the password callback to use for the given
// operation hint. If an external callback was configured it is returned;
// otherwise a closure using the interactive askPass prompt is returned.
func (a *Age) effectivePwCallback(hint string) func(string, bool) ([]byte, error) {
	if a.pwCallback != nil {
		return a.pwCallback
	}

	return func(prompt string, confirm bool) ([]byte, error) {
		pw, err := a.askPass.Passphrase(prompt, hint, confirm)

		return []byte(pw), err
	}
}

// effectivePwPurgeCallback returns the purge callback to use.
// Falls back to a.askPass.Remove when no external callback is configured.
func (a *Age) effectivePwPurgeCallback() func(string) {
	if a.pwPurgeCallback != nil {
		return a.pwPurgeCallback
	}

	return a.askPass.Remove
}

func (a *Age) tryStartAgent(ctx context.Context) {
	if !config.Bool(ctx, "age.agent-enabled") {
		debug.Log("age agent disabled")

		return
	}

	client := agent.NewClient()
	if err := client.Ping(); err == nil {
		debug.Log("age agent already running")

		return
	}

	debug.Log("age agent not running, starting it...")
	if err := startAgent(ctx); err != nil {
		debug.Log("failed to start age agent: %s", err)

		return
	}

	bo := backoff.NewExponentialBackOff()
	bo.InitialInterval = 25 * time.Millisecond
	bo.MaxElapsedTime = 3 * time.Second
	op := func() error {
		return client.Ping()
	}
	if err := backoff.Retry(op, bo); err != nil {
		debug.Log("failed to ping age agent after starting: %s", err)

		return
	}

	// send identities to agent
	ids, err := a.getAllIdentities(ctx)
	if err != nil {
		debug.Log("failed to get identities: %s", err)

		return
	}

	idStrs := make([]string, 0, len(ids))
	for _, id := range ids {
		idStrs = append(idStrs, fmt.Sprintf("%s", id))
	}

	if err := client.SendIdentities(strings.Join(idStrs, "\n")); err != nil {
		debug.Log("failed to send identities to agent: %s", err)
	}

	// set timeout
	if timeout := config.AsInt(config.String(ctx, "age.agent-timeout")); timeout > 0 {
		if err := client.SetTimeout(timeout); err != nil {
			debug.Log("failed to set agent timeout: %s", err)
		}
	}
}

// Initialized returns nil.
func (a *Age) Initialized(ctx context.Context) error {
	if a == nil {
		return fmt.Errorf("Age not initialized")
	}

	return nil
}

// Name returns age.
func (a *Age) Name() string {
	return "age"
}

// Version returns the version of the age dependency being used.
func (a *Age) Version(ctx context.Context) semver.Version {
	return debug.ModuleVersion("filippo.io/age")
}

// Ext returns the extension.
func (a *Age) Ext() string {
	return Ext
}

// IDFile return the recipients file.
func (a *Age) IDFile() string {
	return IDFile
}

// Concurrency returns 1 for `age` since otherwise it prompts for the identity password for each worker.
func (a *Age) Concurrency() int {
	return 1
}

// NeedsPublicKeyImport returns false because age public keys are the recipient
// identifiers themselves and do not need to be imported into a separate keyring.
func (a *Age) NeedsPublicKeyImport() bool {
	return false
}

// GetFingerprint returns the fingerprint of a key.
func (a *Age) GetFingerprint(ctx context.Context, key []byte) (string, error) {
	return string(key), nil
}

// Lock flushes the password cache.
func (a *Age) Lock() {
	a.askPass.Lock()
}

func (a *Age) identitiesToString(ids []age.Identity) (string, error) {
	var sb strings.Builder
	for _, id := range ids {
		fmt.Fprintln(&sb, id)
	}

	return sb.String(), nil
}

// String implements fmt.Stringer.
func (a *Age) String() string {
	var sb strings.Builder
	sb.WriteString("Age(")
	if a == nil {
		sb.WriteString("<nil>)")

		return sb.String()
	}
	sb.WriteString("Identity: ")
	sb.WriteString(a.identity)
	if a.sshKeyPath != "" {
		sb.WriteString(", SSHKeyPath: ")
		sb.WriteString(a.sshKeyPath)
	}
	sb.WriteString(")")

	return sb.String()
}
