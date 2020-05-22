package cui

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"

	"github.com/gopasspw/gopass/internal/backend"
	"github.com/gopasspw/gopass/internal/backend/crypto/gpg"
	"github.com/gopasspw/gopass/internal/editor"
	"github.com/gopasspw/gopass/internal/out"
	"github.com/gopasspw/gopass/internal/termio"
	"github.com/gopasspw/gopass/pkg/ctxutil"

	"github.com/pkg/errors"
)

var (
	// Stdin is exported for tests
	Stdin io.Reader = os.Stdin
	// Stdout is exported for tests
	Stdout io.Writer = os.Stdout
	// Stderr is exported for tests
	Stderr io.Writer = os.Stderr
)

const (
	maxTries = 42
)

type recipientInfo struct {
	id   string
	name string
	self bool
}

func (r recipientInfo) String() string {
	self := ""
	if r.self {
		self = " *"
	}
	return fmt.Sprintf("- %s - %s%s", r.id, r.name, self)
}

type recipientInfos []recipientInfo

func (r recipientInfos) Recipients() []string {
	rs := make([]string, 0, len(r))
	for _, ri := range r {
		rs = append(rs, ri.id)
	}
	return rs
}

func (r recipientInfos) Len() int {
	return len(r)
}

func (r recipientInfos) Swap(i, j int) {
	r[i], r[j] = r[j], r[i]
}

func (r recipientInfos) Less(i, j int) bool {
	if r[i].self {
		return true
	} else if r[j].self {
		return false
	}
	return r[i].name < r[j].name
}

func getRecipientInfo(ctx context.Context, crypto backend.Crypto, name string, recipients []string) (recipientInfos, error) {
	ris := make(recipientInfos, 0, len(recipients))

	for _, r := range recipients {
		// check for context cancelation
		select {
		case <-ctx.Done():
			return nil, errors.New("user aborted")
		default:
		}

		kl, err := crypto.FindPublicKeys(ctx, r)
		if err != nil {
			out.Error(ctx, "Failed to read public key for '%s': %s", r, err)
			continue
		}
		ri := recipientInfo{
			id:   r,
			name: "key not found",
		}
		if len(kl) > 0 {
			ri.name = crypto.FormatKey(ctx, kl[0])
		}
		ris = append(ris, ri)
	}
	sort.Sort(ris)
	return ris, nil
}

// ConfirmRecipients asks the user to confirm a given set of recipients
func ConfirmRecipients(ctx context.Context, crypto backend.Crypto, name string, recipients []string) ([]string, error) {
	if ctxutil.IsNoConfirm(ctx) || !ctxutil.IsInteractive(ctx) {
		return recipients, nil
	}

	ris, err := getRecipientInfo(ctx, crypto, name, recipients)
	if err != nil {
		return nil, err
	}

	if ctxutil.IsEditRecipients(ctx) {
		return confirmEditRecipients(ctx, name, ris)
	}

	return confirmAskRecipients(ctx, name, ris)
}

func confirmAskRecipients(ctx context.Context, name string, ris recipientInfos) ([]string, error) {
	fmt.Fprintf(Stdout, "gopass: Encrypting %s for these recipients:\n", name)
	for _, ri := range ris {
		fmt.Fprintf(Stdout, ri.String()+"\n")
	}
	fmt.Fprintln(Stdout, "")

	recipients := ris.Recipients()

	if ctxutil.IsForce(ctx) {
		return recipients, nil
	}

	yes, err := termio.AskForBool(ctx, "Do you want to continue?", true)
	if err != nil {
		return recipients, errors.Wrapf(err, "failed to read user input")
	}
	if yes {
		return recipients, nil
	}

	return recipients, errors.New("user aborted")
}

func confirmEditRecipients(ctx context.Context, name string, ris recipientInfos) ([]string, error) {
	buf := &bytes.Buffer{}
	fmt.Fprintf(buf, "# gopass - Encrypting %s\n", name)
	fmt.Fprintf(buf, "# Please review and remove any recipient you don't want to include.\n")
	fmt.Fprintf(buf, "# Lines starting with # will be ignored.\n")
	fmt.Fprintf(buf, "# WARNING: Do not edit existing entries.\n")
	for _, ri := range ris {
		fmt.Fprint(buf, ri.String())
	}
	out, err := editor.Invoke(ctx, editor.Path(nil), buf.Bytes())
	if err != nil {
		return nil, err
	}

	recipients := make([]string, 0, len(ris))
	scanner := bufio.NewScanner(bytes.NewReader(out))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "#") {
			continue
		}
		line = strings.TrimPrefix(line, "- ")
		p := strings.SplitN(line, "-", 2)
		if len(p) < 2 {
			continue
		}
		recipients = append(recipients, strings.TrimSpace(p[0]))
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	if len(recipients) < 1 {
		return recipients, errors.New("user aborted")
	}

	return recipients, nil
}

// AskForPrivateKey promts the user to select from a list of private keys
func AskForPrivateKey(ctx context.Context, crypto backend.Crypto, name, prompt string) (string, error) {
	if !ctxutil.IsInteractive(ctx) {
		return "", errors.New("can not select private key without terminal")
	}
	if crypto == nil {
		return "", errors.New("can not select private key without valid crypto backend")
	}

	kl, err := crypto.ListPrivateKeyIDs(gpg.WithAlwaysTrust(ctx, false))
	if err != nil {
		return "", err
	}
	if len(kl) < 1 {
		return "", errors.New("no useable private keys found. make sure you have valid private keys with sufficient trust")
	}

	for i := 0; i < maxTries; i++ {
		if !ctxutil.IsTerminal(ctx) {
			return kl[0], nil
		}
		// check for context cancelation
		select {
		case <-ctx.Done():
			return "", errors.New("user aborted")
		default:
		}

		fmt.Fprintln(Stdout, prompt)
		for i, k := range kl {
			fmt.Fprintf(Stdout, "[%d] %s - %s\n", i, crypto.Name(), crypto.FormatKey(ctx, k))
		}
		iv, err := termio.AskForInt(ctx, fmt.Sprintf("Please enter the number of a key (0-%d, [q]uit)", len(kl)-1), 0)
		if err != nil {
			if err.Error() == "user aborted" {
				return "", err
			}

			continue
		}
		if iv >= 0 && iv < len(kl) {
			return kl[iv], nil
		}
	}
	return "", errors.New("no valid user input")
}

// AskForGitConfigUser will iterate over GPG private key identities and prompt
// the user for selecting one identity whose name and email address will be used as
// git config user.name and git config user.email, respectively.
// On error or no selection, name and email will be empty.
// If s.isTerm is false (i.e., the user cannot be prompted), however,
// the first identity's name/email pair found is returned.
func AskForGitConfigUser(ctx context.Context, crypto backend.Crypto, name string) (string, string, error) {
	var useCurrent bool

	keyList, err := crypto.ListPrivateKeyIDs(ctx)
	if err != nil {
		return "", "", err
	}
	if len(keyList) < 1 {
		return "", "", errors.New("no usable private keys found")
	}

	for _, key := range keyList {
		// check for context cancelation
		select {
		case <-ctx.Done():
			return "", "", errors.New("user aborted")
		default:
		}

		name := crypto.NameFromKey(ctx, key)
		email := crypto.EmailFromKey(ctx, key)

		if name == "" && email == "" {
			continue
		}

		useCurrent, err = termio.AskForBool(
			ctx,
			fmt.Sprintf("Use %s (%s) for password store git config?", name, email),
			true,
		)
		if err != nil {
			return "", "", err
		}
		if useCurrent {
			return name, email, nil
		}
	}

	return "", "", nil
}

type mountPointer interface {
	MountPoints() []string
}

func sorted(s []string) []string {
	sort.Strings(s)
	return s
}

// AskForStore shows a store / mount point selection
func AskForStore(ctx context.Context, s mountPointer) string {
	if !ctxutil.IsInteractive(ctx) {
		return ""
	}

	stores := []string{"<root>"}
	stores = append(stores, sorted(s.MountPoints())...)
	if len(stores) < 2 {
		return ""
	}

	act, sel := GetSelection(ctx, "Please select the store you would like to use", stores)
	switch act {
	case "default":
		fallthrough
	case "show":
		store := stores[sel]
		if store == "<root>" {
			store = ""
		}
		return store
	default:
		return "" // root store
	}
}
