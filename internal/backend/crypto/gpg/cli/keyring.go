package cli

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"
	"text/template"

	"github.com/gopasspw/gopass/internal/backend/crypto/gpg"
	"github.com/gopasspw/gopass/pkg/debug"
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

	kl := parseColons(bytes.NewBuffer(cmdout))
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
	if gpg.IsAlwaysTrust(ctx) {
		return kl.Recipients(), nil
	}
	return kl.UseableKeys(gpg.IsAlwaysTrust(ctx)).Recipients(), nil
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
