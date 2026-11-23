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
	// 2026 key - 0x66AA0F3E
	[]byte(`-----BEGIN PGP PUBLIC KEY BLOCK-----

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
rvymRkY/fWNWAgkE0ue9ToxRpU362yE=
=tacB
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
	return gpgVerifyAt(data, sig, nil)
}

func gpgVerifyAt(data, sig []byte, nowFn func() time.Time) (bool, error) {
	if nowFn == nil {
		nowFn = time.Now
	}

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
