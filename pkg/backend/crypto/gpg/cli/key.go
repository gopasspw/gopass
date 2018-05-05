package cli

import (
	"bytes"
	"context"

	"golang.org/x/crypto/openpgp"

	"github.com/pkg/errors"
)

// ReadNamesFromKey unmarshals and returns the names associated with the given public key
func (g *GPG) ReadNamesFromKey(ctx context.Context, buf []byte) ([]string, error) {
	el, err := openpgp.ReadArmoredKeyRing(bytes.NewReader(buf))
	if err != nil {
		return nil, errors.Wrapf(err, "failed to read key ring")
	}
	if len(el) != 1 {
		return nil, errors.Errorf("Public Key must contain exactly one Entity")
	}
	names := make([]string, 0, len(el[0].Identities))
	for _, v := range el[0].Identities {
		names = append(names, v.Name)
	}
	return names, nil
}
