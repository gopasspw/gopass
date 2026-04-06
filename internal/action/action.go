package action

import (
	"io"
	"os"
	"path/filepath"

	"github.com/blang/semver/v4"
	"github.com/gopasspw/gopass/internal/config"
	"github.com/gopasspw/gopass/internal/reminder"
	"github.com/gopasspw/gopass/pkg/debug"
)

var (
	stdin  io.Reader = os.Stdin
	stdout io.Writer = os.Stdout
	stderr io.Writer = os.Stderr //nolint:unused
)

// Action is the top-level CLI orchestrator. It owns one focused handler per
// domain and exposes every public command through thin delegation shims so that
// the urfave/cli command registrations in commands.go do not need to change.
type Action struct {
	*base

	secrets    *secretHandler
	search     *searchHandler
	generate   *generateHandler
	mounts     *mountHandler
	recipients *recipientHandler
	setup      *setupHandler
	syncH      *syncHandler
	audit      *auditHandler
	templates  *templateHandler
	binary     *binaryHandler
	envH       *envHandler
	otpH       *otpHandler
	misc       *miscHandler
}

// New returns a new Action wrapper.
func New(cfg *config.Config, sv semver.Version) (*Action, error) {
	return newAction(cfg, sv, true)
}

func newAction(cfg *config.Config, sv semver.Version, remind bool) (*Action, error) {
	name := "gopass"
	if len(os.Args) > 0 {
		name = filepath.Base(os.Args[0])
	}

	b := newBase(cfg, sv)
	b.Name = name

	if remind {
		r, err := reminder.New()
		if err != nil {
			debug.Log("failed to init reminder: %s", err)
		} else {
			b.rem = r
		}
	}

	// Construct each focused handler.
	sec := &secretHandler{base: b}
	tmpl := &templateHandler{base: b}
	srch := &searchHandler{base: b}
	gen := &generateHandler{base: b}
	mnt := &mountHandler{base: b}
	rec := &recipientHandler{base: b}
	setup := &setupHandler{base: b}
	syn := &syncHandler{base: b}
	aud := &auditHandler{base: b}
	bin := &binaryHandler{base: b}
	env := &envHandler{base: b}
	otp := &otpHandler{base: b}
	misc := &miscHandler{base: b}

	// Wire cross-handler dependencies through explicit function references so
	// that each handler only depends on the specific operations it needs.
	sec.findFn = srch.find
	sec.renderTemplateFn = tmpl.renderTemplate

	srch.showFn = sec.show
	srch.editFn = sec.edit
	srch.syncFn = syn.Sync

	gen.editFn = sec.Edit
	gen.renderTemplateFn = tmpl.renderTemplate

	mnt.initFn = setup.init

	setup.autoSyncFn = syn.autoSync
	setup.printReminderFn = misc.printReminder

	otp.insertYAMLFn = sec.insertYAML
	otp.findFn = srch.find

	sec.listFn = srch.List
	sec.findFuzzyFn = srch.FindFuzzy

	misc.initCheckPrivateKeysFn = setup.initCheckPrivateKeys
	misc.recipientsListFn = rec.recipientsList
	misc.templatesListFn = tmpl.templatesList

	return &Action{
		base:       b,
		secrets:    sec,
		search:     srch,
		generate:   gen,
		mounts:     mnt,
		recipients: rec,
		setup:      setup,
		syncH:      syn,
		audit:      aud,
		templates:  tmpl,
		binary:     bin,
		envH:       env,
		otpH:       otp,
		misc:       misc,
	}, nil
}

// String implements fmt.Stringer.
func (s *Action) String() string {
	return s.Store.String()
}
