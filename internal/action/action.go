package action

import (
	"io"
	"os"
	"path/filepath"

	"github.com/gopasspw/gopass/internal/config"
	"github.com/gopasspw/gopass/internal/store/root"

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
func New(cfg *config.Config, sv semver.Version) (*Action, error) {
	return newAction(cfg, sv)
}

func newAction(cfg *config.Config, sv semver.Version) (*Action, error) {
	name := "gopass"
	if len(os.Args) > 0 {
		name = filepath.Base(os.Args[0])
	}

	act := &Action{
		Name:    name,
		cfg:     cfg,
		version: sv,
		Store:   root.New(cfg),
	}

	return act, nil
}

// String implement fmt.Stringer
func (s *Action) String() string {
	return s.Store.String()
}
