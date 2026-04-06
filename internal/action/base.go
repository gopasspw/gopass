package action

import (
	"context"

	"github.com/blang/semver/v4"
	"github.com/gopasspw/gopass/internal/backend"
	"github.com/gopasspw/gopass/internal/config"
	"github.com/gopasspw/gopass/internal/reminder"
	"github.com/gopasspw/gopass/internal/store/root"
	"github.com/urfave/cli/v2"
)

// base holds the fields common to all handler types. Every handler embeds
// *base so it can access the store, config, and other shared state.
type base struct {
	Name    string
	Store   *root.Store
	cfg     *config.Config
	version semver.Version
	rem     *reminder.Store
}

// secretHandler handles secret CRUD operations (show, insert, edit, delete,
// copy, move, link, merge).
type secretHandler struct {
	*base
	findFn           func(ctx context.Context, c *cli.Context, needle string, cb showFunc, fuzzy bool) error
	renderTemplateFn func(ctx context.Context, name string, content []byte) ([]byte, bool)
	listFn           func(c *cli.Context) error
	findFuzzyFn      func(c *cli.Context) error
}

// searchHandler handles search and list operations (find, grep, list, history).
type searchHandler struct {
	*base
	showFn showFunc
	editFn func(ctx context.Context, c *cli.Context, name string) error
	syncFn func(c *cli.Context) error
}

// generateHandler handles password generation (generate, create).
type generateHandler struct {
	*base
	editFn           func(c *cli.Context) error
	renderTemplateFn func(ctx context.Context, name string, content []byte) ([]byte, bool)
}

// mountHandler handles store mount management.
type mountHandler struct {
	*base
	initFn func(ctx context.Context, alias, path string, keys ...string) error
}

// recipientHandler handles recipient management.
type recipientHandler struct {
	*base
}

// setupHandler handles store initialisation and onboarding (init, setup, clone, rcs).
type setupHandler struct {
	*base
	autoSyncFn      func(ctx context.Context) error
	printReminderFn func(ctx context.Context)
}

// syncHandler handles synchronisation with remote storage (sync, git).
type syncHandler struct {
	*base
}

// auditHandler handles store auditing.
type auditHandler struct {
	*base
}

// templateHandler handles secret templates.
type templateHandler struct {
	*base
}

// binaryHandler handles binary secret operations.
type binaryHandler struct {
	*base
}

// envHandler handles environment-variable injection.
type envHandler struct {
	*base
}

// otpHandler handles one-time password operations.
type otpHandler struct {
	*base
	insertYAMLFn func(ctx context.Context, name, key string, content []byte, kvps map[string]string) error
	findFn       func(ctx context.Context, c *cli.Context, needle string, cb showFunc, fuzzy bool) error
}

// miscHandler handles miscellaneous operations that do not fit a narrower
// category (aliases, version, convert, reorg, process, unclip, update,
// reminder, repl, doctor, completion, config, otp-adjacent helpers).
type miscHandler struct {
	*base
	initCheckPrivateKeysFn func(ctx context.Context, crypto backend.Crypto) error
	recipientsListFn       func(ctx context.Context) []string
	templatesListFn        func(ctx context.Context) []string
}

// newBase constructs the shared base from a config.
func newBase(cfg *config.Config, sv semver.Version) *base {
	return &base{
		cfg:     cfg,
		version: sv,
		Store:   root.New(cfg),
	}
}

// Crypto backend type alias used by setup_handler shims.
type cryptoBackend = backend.Crypto
