package updater

import (
	"bufio"
	"bytes"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/gopasspw/gopass/pkg/debug"
	"golang.org/x/crypto/openpgp"
)

// gpg --expert --full-generate-key
// (10) ECC (sign only)
// (5) NIST P-512 (see https://github.com/golang/crypto/blob/eec23a3978ad/openpgp/packet/public_key.go#L227)
// 2y
var pubkey = []byte(`
-----BEGIN PGP PUBLIC KEY BLOCK-----

mJMEYAcgPBMFK4EEACMEIwQBD62NnKvhnCoZ7ndmto068vdwGGuFKVH8UynNBNLN
DP4pXhjpH27NwtCd1BZrXE+4novBIrFBQ4oy4jq1ga0XuRoAenuFxV5DhJn+LlmL
XbfX8VTHeoJL2q8ykwnWl3kMaNSDlt0VSGXoh9K6457ykkBo+Ih1AReepEbCVbEV
jQiEihq0M0dvcGFzcyBSZWxlYXNlIFNpZ25pbmcgS2V5IDIwMjEgPHJlbGVhc2VA
Z29wYXNzLnB3PojZBBMTCgA+FiEE4lp1rxNisOb6NHc4If9NjzXRruUFAmAHIDwC
GwMFCQPCZwAFCwkIBwMFFQoJCAsFFgIDAQACHgECF4AACgkQIf9NjzXRruVxwwIF
HlLRWDl9xUJGFV5vo+Bl357WbttbGj6t8VDpEXueZzs9fxmFplRZaIprOUqOKbWN
E6A0EOEhjvDQzN6GbpjnrnkCCQGXAUpHw06KYggBdrmQ9Aof7T9Vr/4GdoC58Om6
NR5JhDwCn+0+V/vWEjvj2GjNBWQGstWpZwHmU0M80IG2ELrj2A==
=WiS/
-----END PGP PUBLIC KEY BLOCK-----
`)

func gpgVerify(data, sig []byte) (bool, error) {
	keyring, err := openpgp.ReadArmoredKeyRing(bytes.NewReader(pubkey))
	if err != nil {
		debug.Log("failed to read public key: %q", err)
		return false, err
	}

	_, err = openpgp.CheckArmoredDetachedSignature(keyring, bytes.NewReader(data), bytes.NewReader(sig))
	if err != nil {
		debug.Log("failed to validate detached GPG signature: %q", err)
		debug.Log("data: %q", string(data))
		debug.Log("sig: %q", string(sig))
		return false, err
	}
	return true, nil
}

// retrieve the hash for the given filename from a checksum file
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
