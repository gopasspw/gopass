package age

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/gopasspw/gopass/internal/backend/crypto/age/agent"
	"github.com/gopasspw/gopass/internal/config"
	"github.com/gopasspw/gopass/pkg/fsutil"
)

const sessionFileLimit = 1 << 24

// sessionFingerprint binds cached identities to their encrypted source and
// bootstrap configuration without invoking any hardware authentication.
func (a *Age) sessionFingerprint(ctx context.Context) (string, error) {
	h := sha256.New()
	paths := []string{a.identity}
	bootstrap := strings.TrimSpace(config.String(ctx, "age.keyring-identities"))
	if bootstrap != "" {
		paths = append(paths, fsutil.CleanPath(bootstrap))
	}
	for _, key := range []string{"age.keyring-identities", "age.keyring-recipients", "age.identities", "age.agent-timeout"} {
		_, _ = fmt.Fprintf(h, "%s\x00%s\x00", key, config.String(ctx, key))
	}
	for _, path := range paths {
		_, _ = fmt.Fprintf(h, "%s\x00", path)
		file, err := os.Open(path)
		if err != nil {
			return "", fmt.Errorf("cannot read age session source: %w", err)
		}
		n, err := io.Copy(h, io.LimitReader(file, sessionFileLimit+1))
		_ = file.Close()
		if err != nil {
			return "", fmt.Errorf("cannot fingerprint age session source: %w", err)
		}
		if n > sessionFileLimit {
			return "", errors.New("age session source exceeds size limit")
		}
		_, _ = h.Write([]byte{0})
	}

	return hex.EncodeToString(h.Sum(nil)), nil
}

func (a *Age) unlockSession(ctx context.Context, client *agent.Client) (string, error) {
	generation, err := client.BeginSession()
	if err != nil {
		return "", fmt.Errorf("failed to begin age session; restart the agent if it predates session support: %w", err)
	}
	source, err := a.sessionFingerprint(ctx)
	if err != nil {
		return "", err
	}
	native, err := a.getNativeIdentities(ctx)
	if err != nil {
		return "", err
	}
	serialized, err := a.identitiesToString(orderedIdentities(ctx, native))
	if err != nil {
		return "", err
	}
	if serialized == "" {
		return "", errors.New("no serializable identities in age keyring")
	}
	current, err := a.sessionFingerprint(ctx)
	if err != nil {
		return "", err
	}
	if source != current {
		return "", errors.New("age keyring changed while unlocking; retry with the current keyring")
	}
	timeout := max(0, config.AsInt(config.String(ctx, "age.agent-timeout")))
	if err := client.LoadIdentitiesWithGeneration(serialized, source, timeout, generation); err != nil {
		return "", fmt.Errorf("failed to load age session; restart the agent if it predates session support: %w", err)
	}

	return source, nil
}

func (a *Age) decryptWithSession(ctx context.Context, ciphertext []byte) ([]byte, error) {
	client := agent.NewClient()
	source, err := a.sessionFingerprint(ctx)
	if err != nil {
		_ = client.Lock()

		return nil, err
	}
	plaintext, err := client.DecryptSession(ciphertext, source)
	if err == nil {
		return plaintext, nil
	}
	if !strings.Contains(err.Error(), "agent is locked") && !strings.Contains(err.Error(), "no identities specified") {
		return nil, fmt.Errorf("failed to decrypt with age session: %w", err)
	}
	source, err = a.unlockSession(ctx, client)
	if err != nil {
		return nil, err
	}

	return client.DecryptSession(ciphertext, source)
}

func (a *Age) usesKeyringSession(ctx context.Context) (bool, error) {
	if config.String(ctx, "age.keyring-identities") != "" {
		return true, nil
	}

	return a.recipientKeyring()
}
