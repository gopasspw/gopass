package xc

import (
	"context"
	"fmt"
)

// ExportPublicKey exports a given public key
func (x *XC) ExportPublicKey(ctx context.Context, id string) ([]byte, error) {
	if x.pubring.Contains(id) {
		return x.pubring.Export(id)
	}
	if x.secring.Contains(id) {
		return x.secring.Export(id, false)
	}
	return nil, fmt.Errorf("key not found")
}

// ExportPrivateKey exports a given private key
func (x *XC) ExportPrivateKey(id string) ([]byte, error) {
	return x.secring.Export(id, true)
}
