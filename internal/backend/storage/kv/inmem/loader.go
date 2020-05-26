package inmem

import (
	"context"
	"fmt"

	"github.com/gopasspw/gopass/internal/backend"
)

const (
	name = "inmem"
)

func init() {
	backend.RegisterStorage(backend.InMem, name, &loader{})
}

type loader struct{}

// New implements backend.StorageLoader
func (l loader) New(ctx context.Context, _ string) (backend.Storage, error) {
	return New(), nil
}

func (l loader) Init(ctx context.Context, path string) (backend.Storage, error) {
	return l.New(ctx, path)
}

func (l loader) Handles(path string) error {
	if path == "//gopass/inmem" {
		return nil
	}
	return fmt.Errorf("not supported")
}

func (l loader) Priority() int {
	return 1000
}
func (l loader) String() string {
	return name
}
