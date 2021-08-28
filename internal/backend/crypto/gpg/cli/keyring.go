package cli

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"text/template"

	"github.com/gopasspw/gopass/internal/backend/crypto/gpg"
	"github.com/gopasspw/gopass/internal/backend/crypto/gpg/colons"
	"github.com/gopasspw/gopass/pkg/debug"

	//lint:ignore SA1019 we'll try to migrate away later
	"golang.org/x/crypto/openpgp"
)

// listKey lists all keys of the given type and matching the search strings
func (g *GPG) listKeys(ctx context.Context, typ string, search ...string) (gpg.KeyList, error) {
	args := []string{"--with-colons", "--with-fingerprint", "--fixed-list-mode", "--list-" + typ + "-keys"}
	args = append(args, search...)
	if e, found := g.listCache.Get(strings.Join(args, ",")); found && gpg.UseCache(ctx) {
		if ev, ok := e.(gpg.KeyList); ok {
			return ev, nil
		}
	}
	cmd := exec.CommandContext(ctx, g.binary, args...)
	var errBuf = bytes.Buffer{}
	cmd.Stderr = &errBuf

	debug.Log("%s %+v\n", cmd.Path, cmd.Args)
	cmdout, err := cmd.Output()
	if err != nil {
		if bytes.Contains(cmdout, []byte("secret key not available")) {
			return gpg.KeyList{}, nil
		}
		return gpg.KeyList{}, fmt.Errorf("%s: %s|%s", err, cmdout, errBuf.String())
	}

	kl := colons.Parse(bytes.NewBuffer(cmdout))
	g.listCache.Add(strings.Join(args, ","), kl)
	return kl, nil
}

// ListRecipients returns a parsed list of GPG public keys
func (g *GPG) ListRecipients(ctx context.Context) ([]string, error) {
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
	return g.pubKeys.UseableKeys(gpg.IsAlwaysTrust(ctx)).Recipients(), nil
}

// FindRecipients searches for the given public keys
func (g *GPG) FindRecipients(ctx context.Context, search ...string) ([]string, error) {
	kl, err := g.listKeys(ctx, "public", search...)
	if err != nil || kl == nil {
		return nil, err
	}

	recp := kl.UseableKeys(gpg.IsAlwaysTrust(ctx)).Recipients()
	if gpg.IsAlwaysTrust(ctx) {
		recp = kl.Recipients()
	}

	debug.Log("found useable keys for %+v: %+v (all: %+v)", search, recp, kl.Recipients())
	return recp, nil
}

// ListIdentities returns a parsed list of GPG secret keys
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

// FindIdentities searches for the given private keys
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

func (g *GPG) findKey(ctx context.Context, id string) gpg.Key {
	kl, _ := g.listKeys(ctx, "secret", id)
	if len(kl) >= 1 {
		return kl[0]
	}
	kl, _ = g.listKeys(ctx, "public", id)
	if len(kl) >= 1 {
		return kl[0]
	}
	return gpg.Key{
		Fingerprint: id,
	}
}

// Fingerprint returns the fingerprint
func (g *GPG) Fingerprint(ctx context.Context, id string) string {
	return g.findKey(ctx, id).Fingerprint
}

// FormatKey formats the details of a key id
// Examples:
// - NameFromKey: {{ .Name }}
// - EmailFromKey: {{ .Email }}
func (g *GPG) FormatKey(ctx context.Context, id, tpl string) string {
	if tpl == "" {
		return g.findKey(ctx, id).OneLine()
	}

	tmpl, err := template.New(tpl).Parse(tpl)
	if err != nil {
		return ""
	}

	buf := &bytes.Buffer{}
	if err := tmpl.Execute(buf, g.findKey(ctx, id).Identity()); err != nil {
		debug.Log("Failed to render template %q: %s", tpl, err)
		return ""
	}

	return buf.String()
}

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

// ImportPublicKey will import a key from the given location into the keyring
func (g *GPG) ImportPublicKey(ctx context.Context, buf []byte) error {
	if len(buf) < 1 {
		return fmt.Errorf("empty input")
	}

	args := append(g.args, "--import")
	cmd := exec.CommandContext(ctx, g.binary, args...)
	cmd.Stdin = bytes.NewReader(buf)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	debug.Log("gpg.ImportPublicKey: %s %+v", cmd.Path, cmd.Args)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to run command: '%s %+v': %w", cmd.Path, cmd.Args, err)
	}

	// clear key cache
	g.privKeys = nil
	g.pubKeys = nil
	return nil
}

// ExportPublicKey will export the named public key to the location given
func (g *GPG) ExportPublicKey(ctx context.Context, id string) ([]byte, error) {
	if id == "" {
		return nil, fmt.Errorf("id is empty")
	}

	args := append(g.args, "--armor", "--export", id)
	cmd := exec.CommandContext(ctx, g.binary, args...)

	debug.Log("%s %+v", cmd.Path, cmd.Args)
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to run command '%s %+v': %w", cmd.Path, cmd.Args, err)
	}

	if len(out) < 1 {
		return nil, fmt.Errorf("key not found")
	}

	return out, nil
}
