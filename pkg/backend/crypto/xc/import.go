package xc

import "context"

// ImportPublicKey imports a given public key into the keyring
func (x *XC) ImportPublicKey(ctx context.Context, buf []byte) error {
	if err := x.pubring.Import(buf); err != nil {
		return err
	}
	return x.pubring.Save()
}

// ImportPrivateKey imports a given private key into the keyring
func (x *XC) ImportPrivateKey(ctx context.Context, buf []byte) error {
	if err := x.secring.Import(buf); err != nil {
		return err
	}
	return x.secring.Save()
}
