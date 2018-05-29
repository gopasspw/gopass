package cli

import (
	"bytes"
	"context"
	"os/exec"
	"strings"

	"github.com/gopasspw/gopass/pkg/backend/crypto/gpg"
	"github.com/gopasspw/gopass/pkg/out"
)

// listKey lists all keys of the given type and matching the search strings
func (g *GPG) listKeys(ctx context.Context, typ string, search ...string) (gpg.KeyList, error) {
	args := []string{"--with-colons", "--with-fingerprint", "--fixed-list-mode", "--list-" + typ + "-keys"}
	args = append(args, search...)
	if e, found := g.listCache.Get(strings.Join(args, ",")); found {
		if ev, ok := e.(gpg.KeyList); ok {
			return ev, nil
		}
	}
	cmd := exec.CommandContext(ctx, g.binary, args...)
	cmd.Stderr = nil

	out.Debug(ctx, "gpg.listKeys: %s %+v\n", cmd.Path, cmd.Args)
	cmdout, err := cmd.Output()
	if err != nil {
		if bytes.Contains(cmdout, []byte("secret key not available")) {
			return gpg.KeyList{}, nil
		}
		return gpg.KeyList{}, err
	}

	kl := parseColons(bytes.NewBuffer(cmdout))
	g.listCache.Add(strings.Join(args, ","), kl)
	return kl, nil
}

// ListPublicKeyIDs returns a parsed list of GPG public keys
func (g *GPG) ListPublicKeyIDs(ctx context.Context) ([]string, error) {
	if g.pubKeys == nil {
		kl, err := g.listKeys(ctx, "public")
		if err != nil {
			return nil, err
		}
		g.pubKeys = kl
	}
	if gpg.IsAlwaysTrust(ctx) {
		return g.pubKeys.Recipients(), nil
	}
	return g.pubKeys.UseableKeys().Recipients(), nil
}

// FindPublicKeys searches for the given public keys
func (g *GPG) FindPublicKeys(ctx context.Context, search ...string) ([]string, error) {
	kl, err := g.listKeys(ctx, "public", search...)
	if err != nil || kl == nil {
		return nil, err
	}
	if gpg.IsAlwaysTrust(ctx) {
		return kl.Recipients(), nil
	}
	return kl.UseableKeys().Recipients(), nil
}

// ListPrivateKeyIDs returns a parsed list of GPG secret keys
func (g *GPG) ListPrivateKeyIDs(ctx context.Context) ([]string, error) {
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
	return g.privKeys.UseableKeys().Recipients(), nil
}

// FindPrivateKeys searches for the given private keys
func (g *GPG) FindPrivateKeys(ctx context.Context, search ...string) ([]string, error) {
	kl, err := g.listKeys(ctx, "secret", search...)
	if err != nil || kl == nil {
		return nil, err
	}
	if gpg.IsAlwaysTrust(ctx) {
		return kl.Recipients(), nil
	}
	return kl.UseableKeys().Recipients(), nil
}

func (g *GPG) findKey(ctx context.Context, id string) gpg.Key {
	kl, _ := g.listKeys(ctx, "secret", id)
	if len(kl) == 1 {
		return kl[0]
	}
	kl, _ = g.listKeys(ctx, "public", id)
	if len(kl) == 1 {
		return kl[0]
	}
	return gpg.Key{
		Fingerprint: id,
	}
}

// EmailFromKey extracts the email from a key id
func (g *GPG) EmailFromKey(ctx context.Context, id string) string {
	return g.findKey(ctx, id).Identity().Email
}

// NameFromKey extracts the name from a key id
func (g *GPG) NameFromKey(ctx context.Context, id string) string {
	return g.findKey(ctx, id).Identity().Name
}

// FormatKey formats the details of a key id
func (g *GPG) FormatKey(ctx context.Context, id string) string {
	return g.findKey(ctx, id).OneLine()
}

// Fingerprint returns the full-length native fingerprint
func (g *GPG) Fingerprint(ctx context.Context, id string) string {
	return g.findKey(ctx, id).Fingerprint
}
