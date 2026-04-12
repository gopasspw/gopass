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
	"github.com/urfave/cli/v2"
)

// ── secretHandler shims ────────────────────────────────────────────────────

func (s *Action) Show(c *cli.Context) error   { return s.secrets.Show(c) }
func (s *Action) Insert(c *cli.Context) error { return s.secrets.Insert(c) }
func (s *Action) Edit(c *cli.Context) error   { return s.secrets.Edit(c) }
func (s *Action) Delete(c *cli.Context) error { return s.secrets.Delete(c) }
func (s *Action) Copy(c *cli.Context) error   { return s.secrets.Copy(c) }
func (s *Action) Move(c *cli.Context) error   { return s.secrets.Move(c) }
func (s *Action) Link(c *cli.Context) error   { return s.secrets.Link(c) }
func (s *Action) Merge(c *cli.Context) error  { return s.secrets.Merge(c) }

// Internal methods accessed from tests.
func (s *Action) show(ctx context.Context, c *cli.Context, name string, recurse bool) error {
	return s.secrets.show(ctx, c, name, recurse)
}

func (s *Action) showHandleRevision(ctx context.Context, c *cli.Context, name, revision string) error {
	return s.secrets.showHandleRevision(ctx, c, name, revision)
}

func (s *Action) showHandleError(ctx context.Context, c *cli.Context, name string, recurse bool, err error) error {
	return s.secrets.showHandleError(ctx, c, name, recurse, err)
}

func (s *Action) showPrintQR(name, pw string) error {
	return s.secrets.showPrintQR(name, pw)
}

func (s *Action) hasAliasDomain(ctx context.Context, name string) string {
	return s.secrets.hasAliasDomain(ctx, name)
}

func (s *Action) insert(ctx context.Context, c *cli.Context, name, key string, echo, multiline, force, appending bool, kvps map[string]string) error {
	return s.secrets.insert(ctx, c, name, key, echo, multiline, force, appending, kvps)
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

func (s *Action) Find(c *cli.Context) error      { return s.search.Find(c) }
func (s *Action) FindFuzzy(c *cli.Context) error { return s.search.FindFuzzy(c) }
func (s *Action) Grep(c *cli.Context) error      { return s.search.Grep(c) }
func (s *Action) List(c *cli.Context) error      { return s.search.List(c) }
func (s *Action) History(c *cli.Context) error   { return s.search.History(c) }

// ── generateHandler shims ──────────────────────────────────────────────────

func (s *Action) Generate(c *cli.Context) error   { return s.generate.Generate(c) }
func (s *Action) Create(c *cli.Context) error     { return s.generate.Create(c) }
func (s *Action) CompleteGenerate(c *cli.Context) { s.generate.CompleteGenerate(c) }

// ── mountHandler shims ─────────────────────────────────────────────────────

func (s *Action) MountRemove(c *cli.Context) error    { return s.mounts.MountRemove(c) }
func (s *Action) MountsPrint(c *cli.Context) error    { return s.mounts.MountsPrint(c) }
func (s *Action) MountsComplete(c *cli.Context)       { s.mounts.MountsComplete(c) }
func (s *Action) MountAdd(c *cli.Context) error       { return s.mounts.MountAdd(c) }
func (s *Action) MountsVersions(c *cli.Context) error { return s.mounts.MountsVersions(c) }

// ── recipientHandler shims ─────────────────────────────────────────────────

func (s *Action) RecipientsPrint(c *cli.Context) error  { return s.recipients.RecipientsPrint(c) }
func (s *Action) RecipientsComplete(c *cli.Context)     { s.recipients.RecipientsComplete(c) }
func (s *Action) RecipientsAck(c *cli.Context) error    { return s.recipients.RecipientsAck(c) }
func (s *Action) RecipientsAdd(c *cli.Context) error    { return s.recipients.RecipientsAdd(c) }
func (s *Action) RecipientsRemove(c *cli.Context) error { return s.recipients.RecipientsRemove(c) }

// ── setupHandler shims ─────────────────────────────────────────────────────

func (s *Action) IsInitialized(c *cli.Context) error   { return s.setup.IsInitialized(c) }
func (s *Action) Init(c *cli.Context) error            { return s.setup.Init(c) }
func (s *Action) Setup(c *cli.Context) error           { return s.setup.Setup(c) }
func (s *Action) Clone(c *cli.Context) error           { return s.setup.Clone(c) }
func (s *Action) RCSInit(c *cli.Context) error         { return s.setup.RCSInit(c) }
func (s *Action) RCSAddRemote(c *cli.Context) error    { return s.setup.RCSAddRemote(c) }
func (s *Action) RCSRemoveRemote(c *cli.Context) error { return s.setup.RCSRemoveRemote(c) }
func (s *Action) RCSPull(c *cli.Context) error         { return s.setup.RCSPull(c) }
func (s *Action) RCSPush(c *cli.Context) error         { return s.setup.RCSPush(c) }
func (s *Action) RCSStatus(c *cli.Context) error       { return s.setup.RCSStatus(c) }

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

func (s *Action) Sync(c *cli.Context) error { return s.syncH.Sync(c) }
func (s *Action) Git(c *cli.Context) error  { return s.syncH.Git(c) }

// ── auditHandler shims ─────────────────────────────────────────────────────

func (s *Action) Audit(c *cli.Context) error { return s.audit.Audit(c) }
func (s *Action) Fsck(c *cli.Context) error  { return s.audit.Fsck(c) }

// ── templateHandler shims ──────────────────────────────────────────────────

func (s *Action) TemplatesPrint(c *cli.Context) error { return s.templates.TemplatesPrint(c) }
func (s *Action) TemplatePrint(c *cli.Context) error  { return s.templates.TemplatePrint(c) }
func (s *Action) TemplateEdit(c *cli.Context) error   { return s.templates.TemplateEdit(c) }
func (s *Action) TemplateRemove(c *cli.Context) error { return s.templates.TemplateRemove(c) }
func (s *Action) TemplatesComplete(c *cli.Context)    { s.templates.TemplatesComplete(c) }

// ── binaryHandler shims ────────────────────────────────────────────────────

func (s *Action) Cat(c *cli.Context) error        { return s.binary.Cat(c) }
func (s *Action) BinaryCopy(c *cli.Context) error { return s.binary.BinaryCopy(c) }
func (s *Action) BinaryMove(c *cli.Context) error { return s.binary.BinaryMove(c) }
func (s *Action) Sum(c *cli.Context) error        { return s.binary.Sum(c) }

// Internal methods accessed from tests.
func (s *Action) binaryCopy(ctx context.Context, c *cli.Context, from, to string, deleteSource bool) error {
	return s.binary.binaryCopy(ctx, c, from, to, deleteSource)
}

func (s *Action) binaryGet(ctx context.Context, name string) ([]byte, error) {
	return s.binary.binaryGet(ctx, name)
}

// ── envHandler shims ───────────────────────────────────────────────────────

func (s *Action) Env(c *cli.Context) error { return s.envH.Env(c) }

// ── otpHandler shims ───────────────────────────────────────────────────────

func (s *Action) OTP(c *cli.Context) error { return s.otpH.OTP(c) }

// Internal methods accessed from tests.
func (s *Action) otp(ctx context.Context, name, qrf string, clip, pw, recurse, chained, alsoClip bool) error {
	return s.otpH.otp(ctx, name, qrf, clip, pw, recurse, chained, alsoClip)
}

// ── miscHandler shims ──────────────────────────────────────────────────────

func (s *Action) AliasesPrint(c *cli.Context) error { return s.misc.AliasesPrint(c) }
func (s *Action) Version(c *cli.Context) error      { return s.misc.Version(c) }
func (s *Action) Convert(c *cli.Context) error      { return s.misc.Convert(c) }
func (s *Action) Reorg(c *cli.Context) error        { return s.misc.Reorg(c) }
func (s *Action) ReorgAfterEdit(ctx context.Context, initial, modified []string) error {
	return s.misc.ReorgAfterEdit(ctx, initial, modified)
}
func (s *Action) Process(c *cli.Context) error          { return s.misc.Process(c) }
func (s *Action) Unclip(c *cli.Context) error           { return s.misc.Unclip(c) }
func (s *Action) Update(c *cli.Context) error           { return s.misc.Update(c) }
func (s *Action) REPL(c *cli.Context) error             { return s.misc.REPL(c) }
func (s *Action) Doctor(c *cli.Context) error           { return s.misc.Doctor(c) }
func (s *Action) Complete(c *cli.Context)               { s.misc.Complete(c) }
func (s *Action) CompletionOpenBSDKsh(a *cli.App) error { return s.misc.CompletionOpenBSDKsh(a) }
func (s *Action) CompletionBash(c *cli.Context) error   { return s.misc.CompletionBash(c) }
func (s *Action) CompletionFish(a *cli.App) error       { return s.misc.CompletionFish(a) }
func (s *Action) CompletionZSH(a *cli.App) error        { return s.misc.CompletionZSH(a) }
func (s *Action) Config(c *cli.Context) error           { return s.misc.Config(c) }
func (s *Action) ConfigComplete(c *cli.Context)         { s.misc.ConfigComplete(c) }

func (s *Action) newGopassCompleter(c *cli.Context) *gopassCompleter {
	return s.misc.newGopassCompleter(c)
}

func (s *Action) setConfigValue(ctx context.Context, store, key, value string) error {
	return s.misc.setConfigValue(ctx, store, key, value)
}

func (s *Action) printConfigValues(ctx context.Context, store string, needles ...string) {
	s.misc.printConfigValues(ctx, store, needles...)
}

// ── searchHandler internal shims ───────────────────────────────────────────

func (s *Action) find(ctx context.Context, c *cli.Context, needle string, cb showFunc, fuzzy bool) error {
	return s.search.find(ctx, c, needle, cb, fuzzy)
}

func (s *Action) findSelection(ctx context.Context, c *cli.Context, choices []string, needle string, cb showFunc) error {
	return s.search.findSelection(ctx, c, choices, needle, cb)
}
