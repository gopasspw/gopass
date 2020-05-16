package action

import (
	"context"
	"io"
	"os"
	"path/filepath"

	"github.com/gopasspw/gopass/pkg/config"
	"github.com/gopasspw/gopass/pkg/out"
	"github.com/gopasspw/gopass/pkg/store/root"

	"github.com/blang/semver"
)

var (
	stdin  io.Reader = os.Stdin
	stdout io.Writer = os.Stdout
)

// Action knows everything to run gopass CLI actions
type Action struct {
	Name    string
	Store   *root.Store
	cfg     *config.Config
	version semver.Version
}

// New returns a new Action wrapper
func New(ctx context.Context, cfg *config.Config, sv semver.Version) (*Action, error) {
	return newAction(ctx, cfg, sv)
}

func newAction(ctx context.Context, cfg *config.Config, sv semver.Version) (*Action, error) {
	name := "gopass"
	if len(os.Args) > 0 {
		name = filepath.Base(os.Args[0])
	}

	act := &Action{
		Name:    name,
		cfg:     cfg,
		version: sv,
	}

	ctx = out.AddPrefix(ctx, "[action] ")

	store, err := root.New(ctx, cfg)
	if err != nil {
		return nil, ExitError(ctx, ExitUnknown, err, "failed to init root store: %s", err)
	}
	act.Store = store

	return act, nil
}

// String implement fmt.Stringer
func (s *Action) String() string {
	return s.Store.String()
}
