package updater

import (
	"bufio"
	"bytes"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/ProtonMail/go-crypto/openpgp"
	"github.com/ProtonMail/go-crypto/openpgp/packet"
	"github.com/gopasspw/gopass/internal/hashsum"
	"github.com/gopasspw/gopass/pkg/debug"
)

// To update see README.md.
var pubkeys = [][]byte{
	// 2025 key - 0x67E6E8D2
	[]byte(`-----BEGIN PGP PUBLIC KEY BLOCK-----

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
fT1TaQbilUs+PL7VNI6d2uAPwQ==
=bs0I
-----END PGP PUBLIC KEY BLOCK-----
`),
	// 2023 key - 0x560D8522
	[]byte(`-----BEGIN PGP PUBLIC KEY BLOCK-----

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
4e8VuujUJ7JmdEbfFhsx7sk=
=XyXN
-----END PGP PUBLIC KEY BLOCK-----
`),
}

type krLogger struct {
	r openpgp.EntityList
}

func (k *krLogger) Str() string {
	var out strings.Builder

	for _, e := range k.r {
		for k := range e.Identities {
			out.WriteString(k)
			out.WriteString(", ")
		}
	}

	return out.String()[:out.Len()-2]
}

func gpgVerify(data, sig []byte) (bool, error) {
	return gpgVerifyAt(data, sig, func() time.Time { return time.Now() })
}

func gpgVerifyAt(data, sig []byte, nowFn func() time.Time) (bool, error) {
	var keyring openpgp.EntityList
	for _, pubkey := range pubkeys {
		k, err := openpgp.ReadArmoredKeyRing(bytes.NewReader(pubkey))
		if err != nil {
			debug.Log("failed to read public key: %q", err)

			return false, fmt.Errorf("failed to read public key: %w", err)
		}
		keyring = append(keyring, k...)
	}

	debug.Log("Keyring: %q", &krLogger{keyring})

	_, err := openpgp.CheckArmoredDetachedSignature(keyring, bytes.NewReader(data), bytes.NewReader(sig), &packet.Config{
		Time: nowFn,
	})
	if err != nil {
		debug.Log("failed to validate detached GPG signature: %q", err)
		debug.Log("data: %q, %s", string(data), hashsum.MD5Hex(string(data)))
		debug.Log("sig: %q", string(sig))

		return false, fmt.Errorf("failed to validated detached GPG signature: %w", err)
	}

	return true, nil
}

// retrieve the hash for the given filename from a checksum file.
//
//nolint:goerr113
func findHashForFile(buf []byte, filename string) ([]byte, error) {
	s := bufio.NewScanner(bytes.NewReader(buf))
	for s.Scan() {
		p := strings.Split(s.Text(), "  ")
		if len(p) < 2 {
			continue
		}

		if p[1] != filename {
			continue
		}

		h, err := hex.DecodeString(p[0])
		if err != nil {
			return nil, fmt.Errorf("failed to decode hash: %w", err)
		}

		return h, nil
	}

	return nil, fmt.Errorf("hash for file %q not found", filename)
}
