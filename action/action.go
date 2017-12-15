package action

import (
	"context"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/blang/semver"
	"github.com/justwatchcom/gopass/backend/gpg"
	gpgcli "github.com/justwatchcom/gopass/backend/gpg/cli"
	"github.com/justwatchcom/gopass/config"
	"github.com/justwatchcom/gopass/store/root"
	"github.com/justwatchcom/gopass/utils/out"
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
	name := "gopass"
	if len(os.Args) > 0 {
		name = filepath.Base(os.Args[0])
	}

	act := &Action{
		Name:    name,
		cfg:     cfg,
		version: sv,
	}

	var err error
	act.gpg, err = gpgcli.New(gpgcli.Config{
		Umask: umask(),
		Args:  gpgOpts(),
	})
	if err != nil {
		out.Red(ctx, "Warning: GPG not found: %s", err)
	}

	store, err := root.New(ctx, cfg, act.gpg)
	if err != nil {
		return nil, err
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
