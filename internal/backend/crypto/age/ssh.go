package age

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"filippo.io/age"
	"filippo.io/age/agessh"
	"github.com/gopasspw/gopass/pkg/appdir"
	"github.com/gopasspw/gopass/pkg/ctxutil"
	"github.com/gopasspw/gopass/pkg/debug"
	"github.com/gopasspw/gopass/pkg/fsutil"
	"golang.org/x/crypto/ssh"
)

var (
	sshCache map[string]age.Identity
	// ErrNoSSHDir signals that no SSH dir was found. Callers
	// are usually expected to ignore this.
	ErrNoSSHDir = errors.New("no ssh directory")
)

// getSSHIdentities returns all SSH identities available for the current user.
func (a *Age) getSSHIdentities(ctx context.Context) (map[string]age.Identity, error) {
	if sshCache != nil {
		debug.Log("using sshCache")

		return sshCache, nil
	}

	// notice that this respects the GOPASS_HOMEDIR env variable, and won't
	// find a .ssh folder in your home directory if you set GOPASS_HOMEDIR
	uhd := appdir.UserHome()
	sshDir := filepath.Join(uhd, ".ssh")
	if !fsutil.IsDir(sshDir) {
		debug.Log("no .ssh directory found at %s. Ignoring SSH identities", sshDir)

		return nil, fmt.Errorf("no identities found: %w", ErrNoSSHDir)
	}

	files, err := os.ReadDir(sshDir)
	if err != nil {
		debug.Log("unable to read .ssh dir %s: %s", sshDir, err)

		return nil, fmt.Errorf("no identities found: %w", ErrNoSSHDir)
	}

	ids := make(map[string]age.Identity, len(files))
	for _, file := range files {
		fn := filepath.Join(sshDir, file.Name())
		if !strings.HasSuffix(fn, ".pub") {
			continue
		}

		recp, id, err := a.parseSSHIdentity(ctx, fn)
		if err != nil {
			continue
		}

		ids[recp] = id
	}
	sshCache = ids
	debug.Log("returned %d SSH Identities", len(ids))

	return ids, nil
}

// parseSSHIdentity parses a SSH public key file and returns the recipient and the identity.
func (a *Age) parseSSHIdentity(ctx context.Context, pubFn string) (string, age.Identity, error) {
	privFn := strings.TrimSuffix(pubFn, ".pub")
	_, err := os.Stat(privFn)
	if err != nil {
		return "", nil, err
	}

	pubBuf, err := os.ReadFile(pubFn)
	if err != nil {
		return "", nil, err
	}

	privBuf, err := os.ReadFile(privFn)
	if err != nil {
		return "", nil, err
	}

	pubkey, _, _, _, err := ssh.ParseAuthorizedKey(pubBuf) //nolint:dogsled
	if err != nil {
		return "", nil, err
	}

	recp := strings.TrimSuffix(string(ssh.MarshalAuthorizedKey(pubkey)), "\n")
	id, err := agessh.ParseIdentity(privBuf)
	if err != nil {
		// handle encrypted SSH identities here.
		var perr *ssh.PassphraseMissingError
		if errors.As(err, &perr) {
			id, err := agessh.NewEncryptedSSHIdentity(pubkey, privBuf, func() ([]byte, error) {
				return ctxutil.GetPasswordCallback(ctx)(pubFn, false)
			})

			return recp, id, err
		}

		return "", nil, err
	}

	return recp, id, nil
}
