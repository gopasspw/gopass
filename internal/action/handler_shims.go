package action

// handler_shims.go — thin delegation methods on *Action.
//
// Every public command registered in commands.go, and every internal method
// called from test files, is available on *Action through a one-liner that
// forwards to the appropriate focused handler. This keeps commands.go and all
// existing tests unchanged while the real logic lives in the handler types.

import (
	"context"

	"github.com/gopasspw/gopass/internal/backend"
	"github.com/urfave/cli/v3"
)

// ── secretHandler shims ────────────────────────────────────────────────────

func (s *Action) Show(ctx context.Context, cmd *cli.Command) error { return s.secrets.Show(ctx, cmd) }

func (s *Action) Insert(ctx context.Context, cmd *cli.Command) error {
	return s.secrets.Insert(ctx, cmd)
}
func (s *Action) Edit(ctx context.Context, cmd *cli.Command) error { return s.secrets.Edit(ctx, cmd) }
func (s *Action) Delete(ctx context.Context, cmd *cli.Command) error {
	return s.secrets.Delete(ctx, cmd)
}
func (s *Action) Copy(ctx context.Context, cmd *cli.Command) error  { return s.secrets.Copy(ctx, cmd) }
func (s *Action) Move(ctx context.Context, cmd *cli.Command) error  { return s.secrets.Move(ctx, cmd) }
func (s *Action) Link(ctx context.Context, cmd *cli.Command) error  { return s.secrets.Link(ctx, cmd) }
func (s *Action) Merge(ctx context.Context, cmd *cli.Command) error { return s.secrets.Merge(ctx, cmd) }

// Internal methods accessed from tests.
func (s *Action) show(ctx context.Context, cmd *cli.Command, name string, recurse bool) error {
	return s.secrets.show(ctx, cmd, name, recurse)
}

func (s *Action) showHandleRevision(ctx context.Context, cmd *cli.Command, name, revision string) error {
	return s.secrets.showHandleRevision(ctx, cmd, name, revision)
}

func (s *Action) showHandleError(ctx context.Context, cmd *cli.Command, name string, recurse bool, err error) error {
	return s.secrets.showHandleError(ctx, cmd, name, recurse, err)
}

func (s *Action) showPrintQR(name, pw string) error {
	return s.secrets.showPrintQR(name, pw)
}

func (s *Action) hasAliasDomain(ctx context.Context, name string) string {
	return s.secrets.hasAliasDomain(ctx, name)
}

func (s *Action) insert(ctx context.Context, cmd *cli.Command, name, key string, echo, multiline, force, appending bool, kvps map[string]string) error {
	return s.secrets.insert(ctx, cmd, name, key, echo, multiline, force, appending, kvps)
}

func (s *Action) insertStdin(ctx context.Context, name string, content []byte, appendTo bool) error {
	return s.secrets.insertStdin(ctx, name, content, appendTo)
}

func (s *Action) insertYAML(ctx context.Context, name, key string, content []byte, kvps map[string]string) error {
	return s.secrets.insertYAML(ctx, name, key, content, kvps)
}

func (s *Action) editUpdate(ctx context.Context, name string, content, nContent []byte, changed bool, ed string) error {
	return s.secrets.editUpdate(ctx, name, content, nContent, changed, ed)
}

// ── searchHandler shims ────────────────────────────────────────────────────

func (s *Action) Find(ctx context.Context, cmd *cli.Command) error { return s.search.Find(ctx, cmd) }

func (s *Action) FindFuzzy(ctx context.Context, cmd *cli.Command) error {
	return s.search.FindFuzzy(ctx, cmd)
}
func (s *Action) Grep(ctx context.Context, cmd *cli.Command) error { return s.search.Grep(ctx, cmd) }
func (s *Action) List(ctx context.Context, cmd *cli.Command) error { return s.search.List(ctx, cmd) }
func (s *Action) History(ctx context.Context, cmd *cli.Command) error {
	return s.search.History(ctx, cmd)
}

// ── generateHandler shims ──────────────────────────────────────────────────

func (s *Action) Generate(ctx context.Context, cmd *cli.Command) error {
	return s.generate.Generate(ctx, cmd)
}

func (s *Action) Create(ctx context.Context, cmd *cli.Command) error {
	return s.generate.Create(ctx, cmd)
}

func (s *Action) CompleteGenerate(ctx context.Context, cmd *cli.Command) {
	s.generate.CompleteGenerate(ctx, cmd)
}

// ── mountHandler shims ─────────────────────────────────────────────────────

func (s *Action) MountRemove(ctx context.Context, cmd *cli.Command) error {
	return s.mounts.MountRemove(ctx, cmd)
}

func (s *Action) MountsPrint(ctx context.Context, cmd *cli.Command) error {
	return s.mounts.MountsPrint(ctx, cmd)
}

func (s *Action) MountsComplete(ctx context.Context, cmd *cli.Command) {
	s.mounts.MountsComplete(ctx, cmd)
}

func (s *Action) MountAdd(ctx context.Context, cmd *cli.Command) error {
	return s.mounts.MountAdd(ctx, cmd)
}

func (s *Action) MountsVersions(ctx context.Context, cmd *cli.Command) error {
	return s.mounts.MountsVersions(ctx, cmd)
}

// ── recipientHandler shims ─────────────────────────────────────────────────

func (s *Action) RecipientsPrint(ctx context.Context, cmd *cli.Command) error {
	return s.recipients.RecipientsPrint(ctx, cmd)
}

func (s *Action) RecipientsComplete(ctx context.Context, cmd *cli.Command) {
	s.recipients.RecipientsComplete(ctx, cmd)
}

func (s *Action) RecipientsAck(ctx context.Context, cmd *cli.Command) error {
	return s.recipients.RecipientsAck(ctx, cmd)
}

func (s *Action) RecipientsAdd(ctx context.Context, cmd *cli.Command) error {
	return s.recipients.RecipientsAdd(ctx, cmd)
}

func (s *Action) RecipientsRemove(ctx context.Context, cmd *cli.Command) error {
	return s.recipients.RecipientsRemove(ctx, cmd)
}

// ── setupHandler shims ─────────────────────────────────────────────────────

func (s *Action) IsInitialized(ctx context.Context, cmd *cli.Command) (context.Context, error) {
	return s.setup.IsInitialized(ctx, cmd)
}
func (s *Action) Init(ctx context.Context, cmd *cli.Command) error  { return s.setup.Init(ctx, cmd) }
func (s *Action) Setup(ctx context.Context, cmd *cli.Command) error { return s.setup.Setup(ctx, cmd) }
func (s *Action) Clone(ctx context.Context, cmd *cli.Command) error { return s.setup.Clone(ctx, cmd) }
func (s *Action) RCSInit(ctx context.Context, cmd *cli.Command) error {
	return s.setup.RCSInit(ctx, cmd)
}

func (s *Action) RCSAddRemote(ctx context.Context, cmd *cli.Command) error {
	return s.setup.RCSAddRemote(ctx, cmd)
}

func (s *Action) RCSRemoveRemote(ctx context.Context, cmd *cli.Command) error {
	return s.setup.RCSRemoveRemote(ctx, cmd)
}

func (s *Action) RCSPull(ctx context.Context, cmd *cli.Command) error {
	return s.setup.RCSPull(ctx, cmd)
}

func (s *Action) RCSPush(ctx context.Context, cmd *cli.Command) error {
	return s.setup.RCSPush(ctx, cmd)
}

func (s *Action) RCSStatus(ctx context.Context, cmd *cli.Command) error {
	return s.setup.RCSStatus(ctx, cmd)
}

// Internal methods accessed from tests.
func (s *Action) clone(ctx context.Context, repo, mount, path string) error {
	return s.setup.clone(ctx, repo, mount, path)
}

func (s *Action) cloneGetGitConfig(ctx context.Context, name string) (string, string, error) {
	return s.setup.cloneGetGitConfig(ctx, name)
}

func (s *Action) initHasUseablePrivateKeys(ctx context.Context, crypto backend.Crypto) bool {
	return s.setup.initHasUseablePrivateKeys(ctx, crypto)
}

func (s *Action) initGenerateIdentity(ctx context.Context, crypto backend.Crypto, name, email string) error {
	return s.setup.initGenerateIdentity(ctx, crypto, name, email)
}

func (s *Action) printRecipients(ctx context.Context, alias string) {
	s.setup.printRecipients(ctx, alias)
}

func (s *Action) rcsInit(ctx context.Context, store, un, ue string) error {
	return s.setup.rcsInit(ctx, store, un, ue)
}

func (s *Action) getUserData(ctx context.Context, store, name, email string) (string, string) {
	return s.setup.getUserData(ctx, store, name, email)
}

// ── syncHandler shims ──────────────────────────────────────────────────────

func (s *Action) Sync(ctx context.Context, cmd *cli.Command) error { return s.syncH.Sync(ctx, cmd) }
func (s *Action) Git(ctx context.Context, cmd *cli.Command) error  { return s.syncH.Git(ctx, cmd) }

// ── auditHandler shims ─────────────────────────────────────────────────────

func (s *Action) Audit(ctx context.Context, cmd *cli.Command) error { return s.audit.Audit(ctx, cmd) }
func (s *Action) Fsck(ctx context.Context, cmd *cli.Command) error  { return s.audit.Fsck(ctx, cmd) }

// ── templateHandler shims ──────────────────────────────────────────────────

func (s *Action) TemplatesPrint(ctx context.Context, cmd *cli.Command) error {
	return s.templates.TemplatesPrint(ctx, cmd)
}

func (s *Action) TemplatePrint(ctx context.Context, cmd *cli.Command) error {
	return s.templates.TemplatePrint(ctx, cmd)
}

func (s *Action) TemplateEdit(ctx context.Context, cmd *cli.Command) error {
	return s.templates.TemplateEdit(ctx, cmd)
}

func (s *Action) TemplateRemove(ctx context.Context, cmd *cli.Command) error {
	return s.templates.TemplateRemove(ctx, cmd)
}

func (s *Action) TemplatesComplete(ctx context.Context, cmd *cli.Command) {
	s.templates.TemplatesComplete(ctx, cmd)
}

// ── binaryHandler shims ────────────────────────────────────────────────────

func (s *Action) Cat(ctx context.Context, cmd *cli.Command) error { return s.binary.Cat(ctx, cmd) }

func (s *Action) BinaryCopy(ctx context.Context, cmd *cli.Command) error {
	return s.binary.BinaryCopy(ctx, cmd)
}

func (s *Action) BinaryMove(ctx context.Context, cmd *cli.Command) error {
	return s.binary.BinaryMove(ctx, cmd)
}
func (s *Action) Sum(ctx context.Context, cmd *cli.Command) error { return s.binary.Sum(ctx, cmd) }

// Internal methods accessed from tests.
func (s *Action) binaryCopy(ctx context.Context, cmd *cli.Command, from, to string, deleteSource bool) error {
	return s.binary.binaryCopy(ctx, cmd, from, to, deleteSource)
}

func (s *Action) binaryGet(ctx context.Context, name string) ([]byte, error) {
	return s.binary.binaryGet(ctx, name)
}

// ── envHandler shims ───────────────────────────────────────────────────────

func (s *Action) Env(ctx context.Context, cmd *cli.Command) error { return s.envH.Env(ctx, cmd) }

// ── otpHandler shims ───────────────────────────────────────────────────────

func (s *Action) OTP(ctx context.Context, cmd *cli.Command) error { return s.otpH.OTP(ctx, cmd) }

// Internal methods accessed from tests.
func (s *Action) otp(ctx context.Context, name, qrf string, clip, pw, recurse, chained, alsoClip bool) error {
	return s.otpH.otp(ctx, name, qrf, clip, pw, recurse, chained, alsoClip)
}

// ── miscHandler shims ──────────────────────────────────────────────────────

func (s *Action) AliasesPrint(ctx context.Context, cmd *cli.Command) error {
	return s.misc.AliasesPrint(ctx, cmd)
}

func (s *Action) Version(ctx context.Context, cmd *cli.Command) error {
	return s.misc.Version(ctx, cmd)
}

func (s *Action) Convert(ctx context.Context, cmd *cli.Command) error {
	return s.misc.Convert(ctx, cmd)
}
func (s *Action) Reorg(ctx context.Context, cmd *cli.Command) error { return s.misc.Reorg(ctx, cmd) }
func (s *Action) ReorgAfterEdit(ctx context.Context, initial, modified []string) error {
	return s.misc.ReorgAfterEdit(ctx, initial, modified)
}

func (s *Action) Process(ctx context.Context, cmd *cli.Command) error {
	return s.misc.Process(ctx, cmd)
}
func (s *Action) Unclip(ctx context.Context, cmd *cli.Command) error { return s.misc.Unclip(ctx, cmd) }
func (s *Action) Update(ctx context.Context, cmd *cli.Command) error { return s.misc.Update(ctx, cmd) }
func (s *Action) REPL(ctx context.Context, cmd *cli.Command) error   { return s.misc.REPL(ctx, cmd) }
func (s *Action) Doctor(ctx context.Context, cmd *cli.Command) error { return s.misc.Doctor(ctx, cmd) }
func (s *Action) Complete(ctx context.Context, cmd *cli.Command)     { s.misc.Complete(ctx, cmd) }
func (s *Action) CompletionOpenBSDKsh(a *cli.Command) error          { return s.misc.CompletionOpenBSDKsh(a) }
func (s *Action) CompletionBash(ctx context.Context, cmd *cli.Command) error {
	return s.misc.CompletionBash(ctx, cmd)
}
func (s *Action) CompletionFish(a *cli.Command) error                { return s.misc.CompletionFish(a) }
func (s *Action) CompletionZSH(a *cli.Command) error                 { return s.misc.CompletionZSH(a) }
func (s *Action) Config(ctx context.Context, cmd *cli.Command) error { return s.misc.Config(ctx, cmd) }
func (s *Action) ConfigComplete(ctx context.Context, cmd *cli.Command) {
	s.misc.ConfigComplete(ctx, cmd)
}

func (s *Action) newGopassCompleter(ctx context.Context, cmd *cli.Command) *gopassCompleter {
	return s.misc.newGopassCompleter(ctx, cmd)
}

func (s *Action) setConfigValue(ctx context.Context, store, key, value string) error {
	return s.misc.setConfigValue(ctx, store, key, value)
}

func (s *Action) printConfigValues(ctx context.Context, store string, needles ...string) {
	s.misc.printConfigValues(ctx, store, needles...)
}

// ── searchHandler internal shims ───────────────────────────────────────────

func (s *Action) find(ctx context.Context, cmd *cli.Command, needle string, cb showFunc, fuzzy bool) error {
	return s.search.find(ctx, cmd, needle, cb, fuzzy)
}

func (s *Action) findSelection(ctx context.Context, cmd *cli.Command, choices []string, needle string, cb showFunc) error {
	return s.search.findSelection(ctx, cmd, choices, needle, cb)
}
