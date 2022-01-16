package action

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/gopasspw/gopass/internal/backend"
	"github.com/gopasspw/gopass/internal/out"
	"github.com/gopasspw/gopass/tests/gptest"
	"github.com/stretchr/testify/assert"
)

func TestInit(t *testing.T) {
	t.Skip("TODO: Add an init test")
}

func TestInitParseContext(t *testing.T) {
	buf := &bytes.Buffer{}
	out.Stdout = buf
	out.Stderr = buf
	defer func() {
		out.Stdout = os.Stdout
		out.Stderr = os.Stderr
	}()

	for _, tc := range []struct {
		name  string
		flags map[string]string
		check func(context.Context) error
	}{
		{
			name:  "crypto age",
			flags: map[string]string{"crypto": "age"},
			check: func(ctx context.Context) error {
				if be := backend.GetCryptoBackend(ctx); be != backend.Age {
					return fmt.Errorf("wrong backend: %d", be)
				}
				return nil
			},
		},
		{
			name: "default",
			check: func(ctx context.Context) error {
				if backend.GetStorageBackend(ctx) != backend.GitFS {
					return fmt.Errorf("wrong backend")
				}
				return nil
			},
		},
	} {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			c := gptest.CliCtxWithFlags(context.Background(), t, tc.flags)
			assert.NoError(t, tc.check(initParseContext(c.Context, c)), tc.name)
			buf.Reset()
		})
	}
}
