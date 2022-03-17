package cli

import (
	"context"

	"github.com/gopasspw/gopass/internal/backend/crypto/gpg"
)

// ListIdentities returns a parsed list of GPG secret keys.
func (g *GPG) ListIdentities(ctx context.Context) ([]string, error) {
	if g.privKeys == nil {
		kl, err := g.listKeys(ctx, "secret")
		if err != nil {
			return nil, err
		}
		g.privKeys = kl
	}
	if gpg.IsAlwaysTrust(ctx) {
		return g.privKeys.Recipients(), nil
	}
	return g.privKeys.UseableKeys(gpg.IsAlwaysTrust(ctx)).Recipients(), nil
}

// FindIdentities searches for the given private keys.
func (g *GPG) FindIdentities(ctx context.Context, search ...string) ([]string, error) {
	kl, err := g.listKeys(ctx, "secret", search...)
	if err != nil || kl == nil {
		return nil, err
	}
	if gpg.IsAlwaysTrust(ctx) {
		return kl.Recipients(), nil
	}
	return kl.UseableKeys(gpg.IsAlwaysTrust(ctx)).Recipients(), nil
}

func (g *GPG) findKey(ctx context.Context, id string) (gpg.Key, bool) {
	kl, _ := g.listKeys(ctx, "secret", id)
	if len(kl) >= 1 {
		return kl[0], true
	}
	kl, _ = g.listKeys(ctx, "public", id)
	if len(kl) >= 1 {
		return kl[0], true
	}
	return gpg.Key{
		Fingerprint: id,
	}, false
}
