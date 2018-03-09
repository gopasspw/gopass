package action

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/justwatchcom/gopass/backend/crypto/gpg"
	"github.com/justwatchcom/gopass/utils/ctxutil"
	"github.com/justwatchcom/gopass/utils/cui"
	"github.com/justwatchcom/gopass/utils/out"
	"github.com/justwatchcom/gopass/utils/termio"
	"github.com/pkg/errors"
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

func (s *Action) getRecipientInfo(ctx context.Context, name string, recipients []string) (recipientInfos, error) {
	crypto := s.Store.Crypto(ctx, name)
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
			out.Red(ctx, "Failed to read public key for '%s': %s", r, err)
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
func (s *Action) ConfirmRecipients(ctx context.Context, name string, recipients []string) ([]string, error) {
	if ctxutil.IsNoConfirm(ctx) || !ctxutil.IsInteractive(ctx) {
		return recipients, nil
	}

	ris, err := s.getRecipientInfo(ctx, name, recipients)
	if err != nil {
		return nil, err
	}

	if ctxutil.IsEditRecipients(ctx) {
		return s.confirmEditRecipients(ctx, name, ris)
	}

	return s.confirmAskRecipients(ctx, name, ris)
}

func (s *Action) confirmAskRecipients(ctx context.Context, name string, ris recipientInfos) ([]string, error) {
	fmt.Fprintf(stdout, "gopass: Encrypting %s for these recipients:\n", name)
	for _, ri := range ris {
		fmt.Fprintf(stdout, ri.String())
	}
	fmt.Fprintln(stdout, "")

	recipients := ris.Recipients()

	yes, err := termio.AskForBool(ctx, "Do you want to continue?", true)
	if err != nil {
		return recipients, errors.Wrapf(err, "failed to read user input")
	}
	if yes {
		return recipients, nil
	}

	return recipients, errors.New("user aborted")
}

func (s *Action) confirmEditRecipients(ctx context.Context, name string, ris recipientInfos) ([]string, error) {
	buf := &bytes.Buffer{}
	fmt.Fprintf(buf, "# gopass - Encrypting %s\n", name)
	fmt.Fprintf(buf, "# Please review and remove any recipient you don't want to include.\n")
	fmt.Fprintf(buf, "# Lines starting with # will be ignored.\n")
	fmt.Fprintf(buf, "# WARNING: Do not edit existing entries.\n")
	for _, ri := range ris {
		fmt.Fprintf(buf, ri.String())
	}
	out, err := s.editor(ctx, getEditor(nil), buf.Bytes())
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

// askforPrivateKey promts the user to select from a list of private keys
func (s *Action) askForPrivateKey(ctx context.Context, name, prompt string) (string, error) {
	if !ctxutil.IsInteractive(ctx) {
		return "", errors.New("no interaction without terminal")
	}

	crypto := s.Store.Crypto(ctx, name)
	kl, err := crypto.ListPrivateKeyIDs(gpg.WithAlwaysTrust(ctx, false))
	if err != nil {
		return "", err
	}
	if len(kl) < 1 {
		return "", errors.New("No useable private keys found")
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

		fmt.Fprintln(stdout, prompt)
		for i, k := range kl {
			fmt.Fprintf(stdout, "[%d] %s\n", i, crypto.FormatKey(ctx, k))
		}
		iv, err := termio.AskForInt(ctx, fmt.Sprintf("Please enter the number of a key (0-%d, [q]uit)", len(kl)-1), 0)
		if err != nil {
			continue
		}
		if iv >= 0 && iv < len(kl) {
			return kl[iv], nil
		}
	}
	return "", errors.New("no valid user input")
}

// askForGitConfigUser will iterate over GPG private key identities and prompt
// the user for selecting one identity whose name and email address will be used as
// git config user.name and git config user.email, respectively.
// On error or no selection, name and email will be empty.
// If s.isTerm is false (i.e., the user cannot be prompted), however,
// the first identity's name/email pair found is returned.
func (s *Action) askForGitConfigUser(ctx context.Context, name string) (string, string, error) {
	var useCurrent bool

	crypto := s.Store.Crypto(ctx, name)
	keyList, err := crypto.ListPrivateKeyIDs(ctx)
	if err != nil {
		return "", "", err
	}
	if len(keyList) < 1 {
		return "", "", errors.New("No usable private keys found")
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

func (s *Action) askForStore(ctx context.Context) string {
	if !ctxutil.IsInteractive(ctx) {
		return ""
	}

	mps := s.Store.MountPoints()
	if len(mps) < 1 {
		return ""
	}

	stores := []string{"<root>"}
	stores = append(stores, mps...)
	act, sel := cui.GetSelection(ctx, "Please select the store you would like to use", "<↑/↓> to change the selection, <→> to select, <ESC> to quit", stores)
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
