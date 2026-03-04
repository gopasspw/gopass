//go:build darwin

package env

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
)

var (
	// Stdin is exported for tests.
	Stdin io.Reader = os.Stdin
	// Stderr is exported for tests.
	Stderr io.Writer = os.Stderr
)

// Check validates the runtime environment on MacOS.
// It checks if the keychain is used.
func Check(ctx context.Context) (string, error) {
	buf := &bytes.Buffer{}

	cmd := exec.CommandContext(ctx, "defaults", "read", "org.gpgtools.common", "UseKeychain")
	cmd.Stdin = Stdin
	cmd.Stdout = buf
	cmd.Stderr = Stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("`default read org.gpgtools.common UseKeychain` failed: %w", err)
	}

	// if the keychain is not used, we can skip the rest
	if strings.ToUpper(strings.TrimSpace(buf.String())) == "NO" {
		return "", nil
	}

	// gpg uses the keychain to store the passphrase, warn once in a while that users
	// might want to change that because it's not secure.
	return "pinentry-mac will use the MacOS Keychain to store your passphrase indefinitely. Consider running 'defaults write org.gpgtools.common UseKeychain NO' to disable that.", nil
}
