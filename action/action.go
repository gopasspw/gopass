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
)

type gpger interface {
	FindPublicKeys(context.Context, ...string) (gpg.KeyList, error)
	FindPrivateKeys(context.Context, ...string) (gpg.KeyList, error)
	ListPublicKeys(context.Context) (gpg.KeyList, error)
	ListPrivateKeys(context.Context) (gpg.KeyList, error)
	CreatePrivateKeyBatch(context.Context, string, string, string) error
	CreatePrivateKey(context.Context) error
	ExportPublicKey(context.Context, string, string) error
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
func New(ctx context.Context, cfg *config.Config, sv semver.Version) *Action {
	name := "gopass"
	if len(os.Args) > 0 {
		name = filepath.Base(os.Args[0])
	}

	act := &Action{
		Name:    name,
		cfg:     cfg,
		version: sv,
	}

	store, err := root.New(ctx, cfg)
	if err != nil {
		panic(err)
	}
	act.Store = store

	act.gpg = gpgcli.New(gpgcli.Config{
		Umask: umask(),
		Args:  gpgOpts(),
	})

	return act
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
