package action

import (
	"context"
	"fmt"

	"github.com/gopasspw/gopass/internal/action/exit"
	"github.com/gopasspw/gopass/internal/backend"
	"github.com/gopasspw/gopass/internal/config"
	"github.com/gopasspw/gopass/internal/cui"
	"github.com/gopasspw/gopass/internal/out"
	"github.com/gopasspw/gopass/pkg/ctxutil"
	"github.com/gopasspw/gopass/pkg/debug"
	"github.com/gopasspw/gopass/pkg/fsutil"
	"github.com/gopasspw/gopass/pkg/termio"
	"github.com/urfave/cli/v2"
)

const logo = `
   __     _    _ _      _ _   ___   ___
 /'_ '\ /'_'\ ( '_'\  /'_' )/',__)/',__)
( (_) |( (_) )| (_) )( (_| |\__, \\__, \
'\__  |'\___/'| ,__/''\__,_)(____/(____/
( )_) |       | |
 \___/'       (_)
`

// IsInitialized returns an error if the store is not properly
// prepared.
func (s *Action) IsInitialized(c *cli.Context) error {
	ctx := ctxutil.WithGlobalFlags(c)
	inited, err := s.Store.IsInitialized(ctx)
	if err != nil {
		return exit.Error(exit.Unknown, err, "Failed to initialize store: %s", err)
	}

	if inited {
		debug.Log("Store is fully initialized and ready to go\n\nAll systems go. 🚀\n")
		s.printReminder(ctx)
		_ = s.autoSync(ctx)

		return nil
	}

	debug.Log("Store needs to be initialized.\n\nAbort. Abort. Abort. 🚫\n")
	if !ctxutil.IsInteractive(ctx) {
		return exit.Error(exit.NotInitialized, nil, "password-store is not initialized. Try '%s init'", s.Name)
	}

	out.Printf(ctx, logo)
	out.Printf(ctx, "🌟 Welcome to gopass!")
	out.Noticef(ctx, "No existing configuration found.")
	out.Printf(ctx, "☝ Please run 'gopass setup'")

	return exit.Error(exit.NotInitialized, err, "not initialized")
}

// Init a new password store with a first gpg id.
func (s *Action) Init(c *cli.Context) error {
	ctx := ctxutil.WithGlobalFlags(c)
	path := c.String("path")
	alias := c.String("store")

	ctx = initParseContext(ctx, c)
	out.Printf(ctx, "🍭 Initializing a new password store ...")

	if name := termio.DetectName(c.Context, c); name != "" {
		ctx = ctxutil.WithUsername(ctx, name)
	}

	if email := termio.DetectEmail(c.Context, c); email != "" {
		ctx = ctxutil.WithEmail(ctx, email)
	}

	inited, err := s.Store.IsInitialized(ctx)
	if err != nil {
		return exit.Error(exit.Unknown, err, "Failed to initialized store: %s", err)
	}

	if inited {
		out.Errorf(ctx, "Store is already initialized!")
	}

	if err := s.init(ctx, alias, path, c.Args().Slice()...); err != nil {
		return exit.Error(exit.Unknown, err, "Failed to initialize store: %s", err)
	}

	return nil
}

func initParseContext(ctx context.Context, c *cli.Context) context.Context {
	if c.IsSet("crypto") {
		ctx = backend.WithCryptoBackendString(ctx, c.String("crypto"))
	}

	if c.IsSet("storage") {
		ctx = backend.WithStorageBackendString(ctx, c.String("storage"))
	}

	if !backend.HasCryptoBackend(ctx) {
		debug.Log("Using default Crypto Backend (GPGCLI)")
		ctx = backend.WithCryptoBackend(ctx, backend.GPGCLI)
	}

	if !backend.HasStorageBackend(ctx) {
		debug.Log("Using default storage backend (GitFS)")
		ctx = backend.WithStorageBackend(ctx, backend.GitFS)
	}

	return ctx
}

func (s *Action) init(ctx context.Context, alias, path string, keys ...string) error {
	if path == "" {
		if alias != "" {
			path = config.PwStoreDir(alias)
		} else {
			path = s.Store.Path()
		}
	}
	path = fsutil.CleanPath(path)

	debug.Log("Initializing Store %q in %q for %+v", alias, path, keys)

	out.Printf(ctx, "🔑 Searching for usable private Keys ...")
	debug.Log("Checking private keys for: %+v", keys)
	crypto := s.getCryptoFor(ctx, alias)

	// private key selection doesn't matter for plain. save one question.
	// TODO should ask the backend
	if crypto.Name() == "plain" {
		keys, _ = crypto.ListIdentities(ctx)
	}

	if len(keys) < 1 {
		out.Notice(ctx, "Hint: Use 'gopass init <subkey> to use subkeys!'")
		nk, err := cui.AskForPrivateKey(ctx, crypto, "🎮 Please select a private key for encrypting secrets:")
		if err != nil {
			return fmt.Errorf("failed to read user input: %w", err)
		}
		keys = []string{nk}
	}

	debug.Log("Initializing sub store - Alias: %q - Path: %q - Keys: %+v", alias, path, keys)
	if err := s.Store.Init(ctx, alias, path, keys...); err != nil {
		return fmt.Errorf("failed to init store %q at %q: %w", alias, path, err)
	}

	if alias != "" && path != "" {
		debug.Log("Mounting sub store %q -> %q", alias, path)
		if err := s.Store.AddMount(ctx, alias, path); err != nil {
			return fmt.Errorf("failed to add mount %q: %w", alias, err)
		}
	}

	if backend.HasStorageBackend(ctx) {
		bn := backend.StorageBackendName(backend.GetStorageBackend(ctx))
		debug.Log("Initializing RCS (%s) ...", bn)
		if err := s.rcsInit(ctx, alias, ctxutil.GetUsername(ctx), ctxutil.GetEmail(ctx)); err != nil {
			debug.Log("Stacktrace: %+v\n", err)
			out.Errorf(ctx, "❌ Failed to init Version Control (%s): %s", bn, err)
		}
		debug.Log("RCS initialized as %s", s.Store.Storage(ctx, alias).Name())
	} else {
		debug.Log("not initializing RCS backend ...")
	}

	out.Printf(ctx, "🏁 Password store %s initialized for:", path)
	s.printRecipients(ctx, alias)

	return nil
}

func (s *Action) printRecipients(ctx context.Context, alias string) {
	crypto := s.Store.Crypto(ctx, alias)
	for _, recipient := range s.Store.ListRecipients(ctx, alias) {
		r := "0x" + recipient
		if kl, err := crypto.FindRecipients(ctx, recipient); err == nil && len(kl) > 0 {
			r = crypto.FormatKey(ctx, kl[0], "")
		}
		out.Printf(ctx, "📩 "+r)
	}
}

func (s *Action) getCryptoFor(ctx context.Context, name string) backend.Crypto {
	return s.Store.Crypto(ctx, name)
}
