package updater

import (
	"bufio"
	"bytes"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/ProtonMail/go-crypto/openpgp"
	"github.com/gopasspw/gopass/pkg/debug"
)

// To generate the private key use:
// ```
// gpg --expert --full-generate-key
// (1) RSA
// 3072
// 2y
// ```
// .
var pubkeys = [][]byte{
	[]byte(`-----BEGIN PGP PUBLIC KEY BLOCK-----

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
jxn7SFTOuQxBJAsmB7Q0aGs=
=HkIB
-----END PGP PUBLIC KEY BLOCK-----
`),
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

	_, err := openpgp.CheckArmoredDetachedSignature(keyring, bytes.NewReader(data), bytes.NewReader(sig), nil)
	if err != nil {
		debug.Log("failed to validate detached GPG signature: %q", err)
		debug.Log("data: %q", string(data))
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
