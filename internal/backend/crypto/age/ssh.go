package age

import (
	"context"
	"os"
	"path/filepath"
	"strings"

	"filippo.io/age"
	"filippo.io/age/agessh"
	"github.com/gopasspw/gopass/pkg/ctxutil"
	"golang.org/x/crypto/ssh"
)

var sshCache map[string]age.Identity

// getSSHIdentities returns all SSH identities available for the current user.
func (a *Age) getSSHIdentities(ctx context.Context) (map[string]age.Identity, error) {
	if sshCache != nil {
		return sshCache, nil
	}
	uhd, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	sshDir := filepath.Join(uhd, ".ssh")
	files, err := os.ReadDir(sshDir)
	if err != nil {
		return nil, err
	}
	ids := make(map[string]age.Identity, len(files))
	for _, file := range files {
		fn := filepath.Join(sshDir, file.Name())
		if !strings.HasSuffix(fn, ".pub") {
			continue
		}
		recp, id, err := a.parseSSHIdentity(ctx, fn)
		if err != nil {
			// debug.Log("Failed to parse SSH identity %s: %s", fn, err)
			continue
		}
		// debug.Log("parsed SSH identity %s from %s", recp, fn)
		ids[recp] = id
	}
	sshCache = ids
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
	pubkey, _, _, _, err := ssh.ParseAuthorizedKey(pubBuf)
	if err != nil {
		return "", nil, err
	}
	recp := strings.TrimSuffix(string(ssh.MarshalAuthorizedKey(pubkey)), "\n")
	id, err := agessh.ParseIdentity(privBuf)
	if err != nil {
		// handle encrypted SSH identities here.
		if _, ok := err.(*ssh.PassphraseMissingError); ok {
			id, err := agessh.NewEncryptedSSHIdentity(pubkey, privBuf, func() ([]byte, error) {
				return ctxutil.GetPasswordCallback(ctx)(pubFn, false)
			})
			return recp, id, err
		}
		return "", nil, err
	}
	return recp, id, nil
}
