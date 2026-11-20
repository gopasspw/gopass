package action

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"slices"
	"strings"
	"text/tabwriter"

	"github.com/gopasspw/gopass/internal/action/exit"
	"github.com/gopasspw/gopass/internal/cui"
	"github.com/gopasspw/gopass/internal/out"
	"github.com/gopasspw/gopass/internal/store/root"
	"github.com/gopasspw/gopass/internal/tree"
	"github.com/gopasspw/gopass/pkg/ctxutil"
	"github.com/gopasspw/gopass/pkg/debug"
	"github.com/gopasspw/gopass/pkg/gopass"
	"github.com/gopasspw/gopass/pkg/passkey"
	"github.com/gopasspw/gopass/pkg/termio"
	"github.com/urfave/cli/v3"
)

// passkeySecretStore adapts the root store to the interface the passkey
// package expects, which is the one of the public gopass API.
type passkeySecretStore struct {
	rs *root.Store
}

func (p passkeySecretStore) List(ctx context.Context) ([]string, error) {
	return p.rs.List(ctx, tree.INF) //nolint:wrapcheck
}

func (p passkeySecretStore) Get(ctx context.Context, name, _ string) (gopass.Secret, error) {
	return p.rs.Get(ctx, name) //nolint:wrapcheck
}

func (p passkeySecretStore) Set(ctx context.Context, name string, sec gopass.Byter) error {
	return p.rs.Set(ctx, name, sec) //nolint:wrapcheck
}

func (s *passkeyHandler) store(ctx context.Context) *passkey.Store {
	return passkey.NewStore(ctx, passkeySecretStore{rs: s.Store})
}

// confirm asks the user on the terminal to approve a ceremony.
func (s *passkeyHandler) confirm(ctx context.Context, req passkey.ConsentRequest) (bool, error) {
	user := req.UserName
	if user == "" {
		user = req.UserDisplayName
	}

	var question string

	switch req.Operation {
	case passkey.OperationRegister:
		question = fmt.Sprintf("Create a passkey for %q at %s?", user, req.RPID)
		if req.Entry != "" {
			question = fmt.Sprintf("Create another passkey for %q at %s? The account already has one in %s.", user, req.RPID, req.Entry)
		}
	case passkey.OperationAuthenticate:
		question = fmt.Sprintf("Sign in to %s as %q with the passkey in %s?", req.RPID, user, req.Entry)
	case passkey.OperationDisclose:
		question = fmt.Sprintf("A passkey for %q at %s is already stored in %s. Tell the site about it?", user, req.RPID, req.Entry)
	default:
		return false, fmt.Errorf("unknown operation %s", req.Operation)
	}

	return termio.AskForConfirmation(ctx, question), nil
}

// authenticator returns the authenticator that uses the passkeys in st and
// asks for confirmation on the terminal.
func (s *passkeyHandler) authenticator(st *passkey.Store) *passkey.Authenticator {
	return passkey.NewAuthenticator(st, passkey.ConsentByConfirmation(s.confirm))
}

// passkeyError maps a failed ceremony to an exit code.
func passkeyError(err error) error {
	switch {
	case errors.Is(err, passkey.ErrDeclined):
		return exit.Error(exit.Aborted, err, "Aborted: %s", err)
	case errors.Is(err, passkey.ErrNoCredentials):
		return exit.Error(exit.NotFound, err, "%s", err)
	case errors.Is(err, passkey.ErrSelectionRequired):
		return exit.Error(exit.Usage, err, "%s. Use --credential to select one", err)
	case passkey.NameOf(err) == passkey.SecurityError, passkey.NameOf(err) == passkey.TypeError:
		return exit.Error(exit.Usage, err, "%s", err)
	case passkey.NameOf(err) == passkey.NotSupportedError:
		return exit.Error(exit.Unsupported, err, "%s", err)
	default:
		return exit.Error(exit.Unknown, err, "%s", err)
	}
}

// printJSON writes v as indented JSON to stdout.
func printJSON(ctx context.Context, v any) error {
	var buf bytes.Buffer

	enc := json.NewEncoder(&buf)
	enc.SetIndent("", "  ")
	enc.SetEscapeHTML(false)

	if err := enc.Encode(v); err != nil {
		return exit.Error(exit.Unknown, err, "failed to encode response: %s", err)
	}

	out.Print(ctx, string(bytes.TrimRight(buf.Bytes(), "\n")))

	return nil
}

// decodeFlag decodes the base64url value of a flag. If the flag is not set
// n random bytes are returned.
func decodeFlag(cmd *cli.Command, name string, n int) ([]byte, error) {
	if cmd.IsSet(name) {
		// A flag that is given without a value is a mistake, not a request
		// for the default.
		if cmd.String(name) == "" {
			return nil, exit.Error(exit.Usage, nil, "--%s must not be empty", name)
		}

		buf, err := passkey.ParseBase64URL(cmd.String(name))
		if err != nil {
			return nil, exit.Error(exit.Usage, err, "invalid --%s: %s", name, err)
		}

		return buf, nil
	}

	buf := make([]byte, n)
	if _, err := rand.Read(buf); err != nil {
		return nil, exit.Error(exit.Unknown, err, "failed to generate random --%s: %s", name, err)
	}

	return buf, nil
}

// originFlag returns the origin a ceremony is performed for. It defaults to
// the HTTPS origin of the RP ID. The origin must be one that is allowed to
// use the RP ID.
func originFlag(cmd *cli.Command, rpID string) (passkey.Origin, error) {
	origin := passkey.Origin{Origin: "https://" + rpID}
	if cmd.IsSet("origin") {
		origin.Origin = cmd.String("origin")
	}

	if _, err := passkey.ResolveRPID(origin.Origin, &rpID); err != nil {
		return origin, exit.Error(exit.Usage, err, "%s", err)
	}

	return origin, nil
}

// PasskeyList prints the stored passkeys.
func (s *passkeyHandler) PasskeyList(ctx context.Context, cmd *cli.Command) error {
	// RP IDs are domain names in lower case.
	rpID := strings.ToLower(cmd.Args().First())

	entries, err := s.store(ctx).List(ctx, rpID)
	if err != nil {
		if len(entries) == 0 {
			return exit.Error(exit.List, err, "failed to list passkeys: %s", err)
		}
		out.Warningf(ctx, "Some passkeys could not be read: %s", err)
	}

	if len(entries) == 0 {
		if rpID != "" {
			return exit.Error(exit.NotFound, nil, "No passkeys found for %s", rpID)
		}
		out.Noticef(ctx, "No passkeys found. Create one with '%s passkey create'", s.Name)

		return nil
	}

	var buf bytes.Buffer

	w := tabwriter.NewWriter(&buf, 0, 4, 2, ' ', 0)
	fmt.Fprintln(w, "SECRET\tRP ID\tUSER\tALGORITHM\tCREATED\tCREDENTIAL ID")

	for _, e := range entries {
		created := "-"
		if !e.Credential.Created.IsZero() {
			created = e.Credential.Created.UTC().Format("2006-01-02")
		}

		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\n",
			e.Name,
			e.Credential.RPID,
			e.Credential.Label(),
			e.Credential.Algorithm(),
			created,
			passkey.Base64URL(e.Credential.ID),
		)
	}

	if err := w.Flush(); err != nil {
		return exit.Error(exit.IO, err, "failed to print passkeys: %s", err)
	}

	out.Print(ctx, string(bytes.TrimRight(buf.Bytes(), "\n")))

	return nil
}

// PasskeyCreate creates a new passkey and prints the registration response.
func (s *passkeyHandler) PasskeyCreate(ctx context.Context, cmd *cli.Command) error {
	ctx = ctxutil.WithGlobalFlags(ctx, cmd)

	rpID := strings.ToLower(cmd.Args().Get(0))
	user := cmd.Args().Get(1)

	if rpID == "" || user == "" {
		return exit.Error(exit.Usage, nil, "Usage: %s passkey create <rp-id> <user>", s.Name)
	}

	alg, err := passkey.ParseAlgorithm(cmd.String("algorithm"))
	if err != nil {
		return exit.Error(exit.Usage, err, "%s", err)
	}

	origin, err := originFlag(cmd, rpID)
	if err != nil {
		return err
	}

	st := s.store(ctx)

	handle, err := decodeFlag(cmd, "user-handle", 32)
	if err != nil {
		return err
	}

	if !cmd.IsSet("user-handle") {
		// The user handle is what identifies an account. Without one, a
		// passkey for a user that already has one is meant for the same
		// account.
		if h := s.passkeyUserHandle(ctx, st, rpID, user); h != nil {
			handle = h
		}
	}

	challenge, err := decodeFlag(cmd, "challenge", 32)
	if err != nil {
		return err
	}

	displayName := cmd.String("display-name")
	if displayName == "" {
		displayName = user
	}

	opts := &passkey.CreationOptions{
		RP:               passkey.RelyingParty{ID: &rpID, Name: rpID},
		User:             passkey.User{ID: handle, Name: user, DisplayName: displayName},
		Challenge:        challenge,
		PubKeyCredParams: []passkey.CredentialParameter{{Type: "public-key", Alg: alg}},
	}

	rsp, entry, err := s.authenticator(st).Register(ctx, origin, opts)
	if err != nil {
		return passkeyError(err)
	}

	if entry.Warning != nil {
		out.Warningf(ctx, "%s", entry.Warning)
	}

	// Standard output carries the response. It must stay parseable if it is
	// not read by a human.
	if ctxutil.IsTerminal(ctx) {
		out.OKf(ctx, "Stored passkey for %q at %s in %s", user, entry.Credential.RPID, entry.Name)
	}

	return printJSON(ctx, rsp)
}

// passkeyUserHandle returns the user handle of the passkey that is stored
// below the prefix for the user with the given name, or nil.
func (s *passkeyHandler) passkeyUserHandle(ctx context.Context, st *passkey.Store, rpID, user string) []byte {
	entries, err := st.List(ctx, rpID)
	if err != nil {
		debug.Log("failed to read some passkeys of %s: %s", rpID, err)
	}

	for _, e := range entries {
		if e.Credential.UserName == user && st.Owns(e.Name) {
			return e.Credential.UserHandle
		}
	}

	return nil
}

// PasskeyAssert signs a challenge with a passkey and prints the
// authentication response.
func (s *passkeyHandler) PasskeyAssert(ctx context.Context, cmd *cli.Command) error {
	ctx = ctxutil.WithGlobalFlags(ctx, cmd)

	name := cmd.Args().First()
	if name == "" {
		return exit.Error(exit.Usage, nil, "Usage: %s passkey assert <rp-id|secret> --challenge <challenge>", s.Name)
	}

	if !cmd.IsSet("challenge") {
		return exit.Error(exit.Usage, nil, "--challenge is required")
	}

	challenge, err := decodeFlag(cmd, "challenge", 0)
	if err != nil {
		return err
	}

	var credentialID []byte
	if cmd.IsSet("credential") {
		credentialID, err = decodeFlag(cmd, "credential", 0)
		if err != nil {
			return err
		}
	}

	verification := cmd.String("user-verification")
	if !slices.Contains([]string{"", "required", "preferred", "discouraged"}, verification) {
		return exit.Error(exit.Usage, nil, "invalid --user-verification %q, use one of required, preferred or discouraged", verification)
	}

	st := s.store(ctx)
	rpID := strings.ToLower(name)

	// The argument is either an RP ID or the name of a secret with passkeys.
	if s.Store.Exists(ctx, name) {
		rpID, credentialID, err = s.passkeyFromSecret(ctx, name, credentialID)
		if err != nil {
			return err
		}
	}

	opts := &passkey.RequestOptions{
		Challenge:        challenge,
		RPID:             &rpID,
		UserVerification: verification,
	}

	origin, err := originFlag(cmd, rpID)
	if err != nil {
		return err
	}

	auth := s.authenticator(st)

	if len(credentialID) == 0 {
		credentialID, err = s.passkeySelect(ctx, auth, origin, opts)
		if err != nil {
			return err
		}
	}

	rsp, entry, err := auth.Authenticate(ctx, origin, opts, credentialID)
	if err != nil {
		return passkeyError(err)
	}

	if entry.Warning != nil {
		out.Warningf(ctx, "%s", entry.Warning)
	}

	return printJSON(ctx, rsp)
}

// passkeyFromSecret returns the RP ID and the ID of the passkey stored in
// the secret with the given name. If the secret has more than one passkey
// it must be selected by its ID.
func (s *passkeyHandler) passkeyFromSecret(ctx context.Context, name string, credentialID []byte) (string, []byte, error) {
	sec, err := s.Store.Get(ctx, name)
	if err != nil {
		return "", nil, exit.Error(exit.Decrypt, err, "failed to read %s: %s", name, err)
	}

	creds, err := passkey.FromSecret(sec)
	if err != nil {
		if len(creds) == 0 {
			return "", nil, exit.Error(exit.Unknown, err, "failed to parse the passkeys in %s: %s", name, err)
		}
		out.Warningf(ctx, "Some passkeys in %s could not be parsed: %s", name, err)
	}

	if len(credentialID) > 0 {
		creds = slices.DeleteFunc(creds, func(c *passkey.Credential) bool { return !bytes.Equal(c.ID, credentialID) })
	}

	switch len(creds) {
	case 0:
		return "", nil, exit.Error(exit.NotFound, nil, "No matching passkey found in %s", name)
	case 1:
		return creds[0].RPID, creds[0].ID, nil
	default:
		return "", nil, exit.Error(exit.Usage, nil, "%s holds %d passkeys. Use --credential to select one", name, len(creds))
	}
}

// passkeySelect lets the user pick one of several passkeys that can be used
// to sign in. It returns nil if there is nothing to choose from.
func (s *passkeyHandler) passkeySelect(ctx context.Context, auth *passkey.Authenticator, origin passkey.Origin, opts *passkey.RequestOptions) ([]byte, error) {
	entries, err := auth.Candidates(ctx, origin, opts)
	if err != nil {
		return nil, passkeyError(err)
	}

	if len(entries) < 2 || !ctxutil.IsInteractive(ctx) || !ctxutil.IsTerminal(ctx) || ctxutil.IsAlwaysYes(ctx) {
		return nil, nil
	}

	choices := make([]string, 0, len(entries))
	for _, e := range entries {
		choices = append(choices, fmt.Sprintf("%s - %s", e.Credential.Label(), e.Name))
	}

	act, sel := cui.GetSelection(ctx, "Please select the passkey to sign in with", choices)
	if act != "default" || sel < 0 || sel >= len(entries) {
		return nil, exit.Error(exit.Aborted, nil, "user aborted")
	}

	return entries[sel].Credential.ID, nil
}

// PasskeyImport imports the passkeys of a Credential Exchange Format file.
func (s *passkeyHandler) PasskeyImport(ctx context.Context, cmd *cli.Command) error {
	file := cmd.Args().First()
	if file == "" {
		return exit.Error(exit.Usage, nil, "Usage: %s passkey import <file>", s.Name)
	}

	in := stdin

	if file != "-" {
		fh, err := os.Open(file) //nolint:gosec
		if err != nil {
			return exit.Error(exit.IO, err, "failed to open %s: %s", file, err)
		}
		defer fh.Close() //nolint:errcheck

		in = fh
	}

	creds, err := passkey.ImportCXF(in)
	if err != nil {
		if len(creds) == 0 {
			return exit.Error(exit.Unknown, err, "failed to import %s: %s", file, err)
		}
		out.Warningf(ctx, "Some passkeys could not be imported: %s", err)
	}

	st := s.store(ctx)

	for _, cred := range creds {
		name, err := st.Put(ctx, cred)
		if err != nil {
			if !errors.Is(err, passkey.ErrNotSynced) {
				return exit.Error(exit.Encrypt, err, "failed to store the passkey for %q at %s: %s", cred.Label(), cred.RPID, err)
			}
			out.Warningf(ctx, "%s", err)
		}

		out.Printf(ctx, "Imported passkey for %q at %s into %s", cred.Label(), cred.RPID, name)
	}

	out.OKf(ctx, "Imported %d passkeys", len(creds))

	return nil
}

// PasskeyExport exports passkeys as a Credential Exchange Format document.
func (s *passkeyHandler) PasskeyExport(ctx context.Context, cmd *cli.Command) error {
	rpID := strings.ToLower(cmd.Args().First())

	entries, err := s.store(ctx).List(ctx, rpID)
	if err != nil {
		// An export that silently lacks passkeys is worse than none.
		return exit.Error(exit.List, err, "failed to read all passkeys: %s", err)
	}

	if len(entries) == 0 {
		return exit.Error(exit.NotFound, nil, "No passkeys found")
	}

	var buf bytes.Buffer

	skipped, err := passkey.ExportCXF(&buf, entries)
	if err != nil {
		return exit.Error(exit.Unknown, err, "failed to export passkeys: %s", err)
	}

	for _, e := range skipped {
		out.Warningf(ctx, "Not exported, the format does not support signature counters: %s (%s)", e.Name, e.Credential.Label())
	}

	if len(skipped) == len(entries) {
		return exit.Error(exit.Unsupported, nil, "None of the passkeys can be exported")
	}

	out.Warningf(ctx, "The export contains the private keys of %d passkeys and is not encrypted", len(entries)-len(skipped))

	file := cmd.String("output")
	if cmd.IsSet("output") && file == "" {
		return exit.Error(exit.Usage, nil, "--output must not be empty")
	}

	if file == "" || file == "-" {
		// The document holds private keys. It is printed, but not logged.
		out.Print(ctx, out.Secret(bytes.TrimRight(buf.Bytes(), "\n")))

		return nil
	}

	if err := writeNewFile(file, buf.Bytes()); err != nil {
		return exit.Error(exit.IO, err, "failed to write %s: %s", file, err)
	}

	out.OKf(ctx, "Exported %d passkeys to %s", len(entries)-len(skipped), file)

	return nil
}

// writeNewFile writes data to a new file that only the user can read. A file
// that is already there is not replaced, and none is left behind if writing
// fails: the data is a set of private keys.
func writeNewFile(name string, data []byte) error {
	fh, err := os.OpenFile(name, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600) //nolint:gosec
	if err != nil {
		return err //nolint:wrapcheck
	}

	_, err = fh.Write(data)
	if cerr := fh.Close(); err == nil {
		err = cerr
	}

	if err != nil {
		_ = os.Remove(name)

		return err //nolint:wrapcheck
	}

	return nil
}
