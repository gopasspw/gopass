package cli

import (
	"bytes"
	"context"
	"fmt"

	"golang.org/x/crypto/openpgp"
)

// ReadNamesFromKey unmarshals and returns the names associated with the given public key
func (g *GPG) ReadNamesFromKey(ctx context.Context, buf []byte) ([]string, error) {
	el, err := openpgp.ReadArmoredKeyRing(bytes.NewReader(buf))
	if err != nil {
		return nil, fmt.Errorf("failed to read key ring: %w", err)
	}
	if len(el) != 1 {
		return nil, fmt.Errorf("public Key must contain exactly one Entity")
	}
	names := make([]string, 0, len(el[0].Identities))
	for _, v := range el[0].Identities {
		names = append(names, v.Name)
	}
	return names, nil
}
