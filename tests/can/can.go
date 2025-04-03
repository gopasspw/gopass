// Package can provides access to the embedded key material used for testing.
// The key material is embedded in the binary and is used for testing
// purposes only.

package can

import (
	"embed"
	"fmt"
	"os"
	"path/filepath"

	"github.com/ProtonMail/go-crypto/openpgp"
)

//go:embed gnupg/*
var can embed.FS

func EmbeddedKeyRing() openpgp.EntityList {
	fh, err := can.Open("gnupg/pubring.gpg")
	if err != nil {
		// This must not happen. Since the key material is embedded into the
		// binary, it must be available in in the correct format. If it is not
		// we ca only panic. Since this is used for tests only this shouldn't
		// affect users.
		panic(err)
	}
	defer fh.Close() //nolint:errcheck

	el, err := openpgp.ReadKeyRing(fh)
	if err != nil {
		// See reasoning above.
		panic(err)
	}

	return el
}

func KeyID() string {
	el := EmbeddedKeyRing()
	if len(el) != 1 {
		panic("pubring.gpg must contain exactly one key")
	}

	return el[0].PrimaryKey.KeyIdShortString()
}

// WriteTo writes the embedded content to the given output
// directory.
func WriteTo(path string) error {
	fes, err := can.ReadDir("gnupg")
	if err != nil {
		return fmt.Errorf("failed to read can dir: %w", err)
	}
	for _, fe := range fes {
		from := "gnupg/" + fe.Name()
		to := filepath.Join(path, fe.Name())
		buf, err := can.ReadFile(from)
		if err != nil {
			return fmt.Errorf("failed to read %s: %w", from, err)
		}

		if err := os.MkdirAll(filepath.Dir(to), 0o700); err != nil {
			return fmt.Errorf("failed to create dir %s: %w", filepath.Dir(to), err)
		}

		if err := os.WriteFile(to, buf, 0o600); err != nil {
			return fmt.Errorf("failed to write %s: %w", to, err)
		}
	}

	return nil
}
