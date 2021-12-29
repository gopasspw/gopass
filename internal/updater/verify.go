package updater

import (
	"bufio"
	"bytes"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/gopasspw/gopass/pkg/debug"

	//lint:ignore SA1019 we'll try to migrate away later
	"golang.org/x/crypto/openpgp"
)

// To generate the private key use:
// ```
// gpg --expert --full-generate-key
// (1) RSA
// 3072
// 2y
// ```
//.
var pubkey = []byte(`
-----BEGIN PGP PUBLIC KEY BLOCK-----

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
jxn7SFTOuQxBJAsmB7Q0aGs=
=HkIB
-----END PGP PUBLIC KEY BLOCK-----
`)

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
	keyring, err := openpgp.ReadArmoredKeyRing(bytes.NewReader(pubkey))
	if err != nil {
		debug.Log("failed to read public key: %q", err)
		return false, err
	}

	debug.Log("Keyring: %q", &krLogger{keyring})

	_, err = openpgp.CheckArmoredDetachedSignature(keyring, bytes.NewReader(data), bytes.NewReader(sig))
	if err != nil {
		debug.Log("failed to validate detached GPG signature: %q", err)
		debug.Log("data: %q", string(data))
		debug.Log("sig: %q", string(sig))
		return false, err
	}
	return true, nil
}

// retrieve the hash for the given filename from a checksum file.
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
			return nil, err
		}
		return h, nil
	}

	return nil, fmt.Errorf("hash for file %q not found", filename)
}
