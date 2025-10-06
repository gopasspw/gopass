package cli

import (
	"bytes"
	"context"
	"encoding/hex"
	"fmt"
	"os/exec"
	"strings"
	"text/template"

	"github.com/ProtonMail/go-crypto/openpgp"
	"github.com/gopasspw/gopass/internal/backend/crypto/gpg"
	"github.com/gopasspw/gopass/internal/backend/crypto/gpg/colons"
	"github.com/gopasspw/gopass/internal/out"
	"github.com/gopasspw/gopass/pkg/debug"
)

// listKey lists all keys of the given type and matching the search strings.
func (g *GPG) listKeys(ctx context.Context, typ string, search ...string) (gpg.KeyList, error) {
	debug.Log("listing %s keys for %v", typ, search)
	ctx, cancel := context.WithTimeout(ctx, Timeout)
	defer cancel()

	args := []string{"--with-colons", "--with-fingerprint", "--fixed-list-mode", "--list-" + typ + "-keys"}
	args = append(args, search...)
	if e, found := g.listCache.Get(strings.Join(args, ",")); found && gpg.UseCache(ctx) {
		debug.Log("listed cached keys: %q", strings.Join(e.Recipients(), ","))
		return e, nil
	}

	cmd := exec.CommandContext(ctx, g.binary, args...)
	errBuf := bytes.Buffer{}
	cmd.Stderr = &errBuf

	debug.V(1).Log("%s %+v\n", cmd.Path, cmd.Args)
	cmdout, err := cmd.Output()
	if err != nil {
		if bytes.Contains(cmdout, []byte("secret key not available")) || strings.Contains(errBuf.String(), "No secret key") {
			debug.Log("secret key not available for %v", search)
			return gpg.KeyList{}, nil
		}
		errStr := fmt.Errorf("%w: %s|%s", err, cmdout, errBuf.String())
		debug.Log("cmd error listing %s keys: %q", typ, errStr)
		return gpg.KeyList{}, errStr
	}

	kl := colons.Parse(bytes.NewBuffer(cmdout))
	g.listCache.Add(strings.Join(args, ","), kl)

	debug.Log("listed non-cached keys: %q", strings.Join(kl.Recipients(), ","))

	return kl, nil
}

// Fingerprint returns the fingerprint.
func (g *GPG) Fingerprint(ctx context.Context, id string) string {
	k, found := g.findKey(ctx, id)
	if !found {
		return ""
	}

	return k.Fingerprint
}

// FormatKey formats the details of a key id
// Examples:
// - NameFromKey: {{ .Name }}
// - EmailFromKey: {{ .Email }}.
func (g *GPG) FormatKey(ctx context.Context, id, tpl string) string {
	if tpl == "" {
		k, found := g.findKey(ctx, id)
		if !found {
			return ""
		}

		return k.OneLine()
	}

	tmpl, err := template.New(tpl).Parse(tpl)
	if err != nil {
		return ""
	}

	var gid gpg.Identity
	k, found := g.findKey(ctx, id)
	if found {
		gid = k.Identity()
	}

	buf := &bytes.Buffer{}
	if err := tmpl.Execute(buf, gid); err != nil {
		debug.Log("Failed to render template %q: %s", tpl, err)

		return ""
	}

	return buf.String()
}

// ReadNamesFromKey unmarshals and returns the names associated with the given public key.
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

// ImportPublicKey will import a key from the given location into the keyring.
func (g *GPG) ImportPublicKey(ctx context.Context, buf []byte) error {
	if len(buf) < 1 {
		return fmt.Errorf("empty input")
	}

	outBuf := &bytes.Buffer{}

	args := append(g.args, "--import")
	cmd := exec.CommandContext(ctx, g.binary, args...)
	cmd.Stdin = bytes.NewReader(buf)
	cmd.Stdout = outBuf
	cmd.Stderr = outBuf

	debug.Log("gpg.ImportPublicKey: %s %+v", cmd.Path, cmd.Args)
	if err := cmd.Run(); err != nil {
		out.Printf(ctx, "GPG import failed: %s", outBuf.String())

		return fmt.Errorf("failed to run command: '%s %+v': %w - %q", cmd.Path, cmd.Args, err, outBuf.String())
	}

	// clear key cache
	g.privKeys = nil
	g.pubKeys = nil

	return nil
}

// GetFingerprint returns the fingerprint of a key.
func (g *GPG) GetFingerprint(ctx context.Context, buf []byte) (string, error) {
	if len(buf) < 1 {
		return "", fmt.Errorf("empty input")
	}

	el, err := openpgp.ReadArmoredKeyRing(bytes.NewReader(buf))
	if err != nil {
		// maybe it's a non-armored key?
		el, err = openpgp.ReadKeyRing(bytes.NewReader(buf))
		if err != nil {
			return "", fmt.Errorf("failed to read key ring: %w", err)
		}
	}

	if len(el) != 1 {
		return "", fmt.Errorf("public Key must contain exactly one Entity")
	}

	return strings.ToUpper(hex.EncodeToString(el[0].PrimaryKey.Fingerprint[:])), nil
}

// ExportPublicKey will export the named public key to the location given.
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
