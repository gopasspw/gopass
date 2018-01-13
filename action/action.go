package action

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/blang/semver"
	"github.com/justwatchcom/gopass/backend/crypto/gpg"
	gpgcli "github.com/justwatchcom/gopass/backend/crypto/gpg/cli"
	"github.com/justwatchcom/gopass/config"
	"github.com/justwatchcom/gopass/store/root"
	"github.com/justwatchcom/gopass/utils/out"
)

var (
	stdin  io.Reader = os.Stdin
	stdout io.Writer = os.Stdout
	stderr io.Writer = os.Stderr
)

type gpger interface {
	Binary() string
	ListPublicKeys(context.Context) (gpg.KeyList, error)
	FindPublicKeys(context.Context, ...string) (gpg.KeyList, error)
	ListPrivateKeys(context.Context) (gpg.KeyList, error)
	CreatePrivateKeyBatch(context.Context, string, string, string) error
	CreatePrivateKey(context.Context) error
	FindPrivateKeys(context.Context, ...string) (gpg.KeyList, error)
	GetRecipients(context.Context, string) ([]string, error)
	Encrypt(context.Context, string, []byte, []string) error
	Decrypt(context.Context, string) ([]byte, error)
	ExportPublicKey(context.Context, string, string) error
	ImportPublicKey(context.Context, string) error
	Version(context.Context) semver.Version
}

// Action knows everything to run gopass CLI actions
type Action struct {
	Name    string
	Store   *root.Store
	cfg     *config.Config
	gpg     gpger
	version semver.Version
}

// New returns a new Action wrapper
func New(ctx context.Context, cfg *config.Config, sv semver.Version) (*Action, error) {
	gpg, err := gpgcli.New(ctx, gpgcli.Config{
		Umask: umask(),
		Args:  gpgOpts(),
	})
	if err != nil {
		out.Red(ctx, "Warning: GPG not found: %s", err)
	}
	return newAction(ctx, cfg, sv, gpg)
}

func newAction(ctx context.Context, cfg *config.Config, sv semver.Version, gpg gpger) (*Action, error) {
	name := "gopass"
	if len(os.Args) > 0 {
		name = filepath.Base(os.Args[0])
	}

	act := &Action{
		Name:    name,
		cfg:     cfg,
		version: sv,
		gpg:     gpg,
	}

	store, err := root.New(ctx, cfg, act.gpg)
	if err != nil {
		return nil, exitError(ctx, ExitUnknown, err, "failed to init root store: %s", err)
	}
	act.Store = store

	return act, nil
}

func umask() int {
	for _, en := range []string{"GOPASS_UMASK", "PASSWORD_STORE_UMASK"} {
		if um := os.Getenv(en); um != "" {
			if iv, err := strconv.ParseInt(um, 8, 32); err == nil && iv >= 0 && iv <= 0777 {
				return int(iv)
			}
		}
	}
	return 077
}

func gpgOpts() []string {
	for _, en := range []string{"GOPASS_GPG_OPTS", "PASSWORD_STORE_GPG_OPTS"} {
		if opts := os.Getenv(en); opts != "" {
			return strings.Fields(opts)
		}
	}
	return nil
}

// String implement fmt.Stringer
func (s *Action) String() string {
	return s.Store.String()
}

// HasGPG returns true if the GPG wrapper is initialized
func (s *Action) HasGPG() bool {
	return s.gpg != nil
}
