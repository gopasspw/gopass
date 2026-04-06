package action

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/gopasspw/gopass/internal/action/exit"
	"github.com/gopasspw/gopass/internal/out"
	"github.com/gopasspw/gopass/internal/tpl"
	"github.com/gopasspw/gopass/pkg/ctxutil"
	"github.com/gopasspw/gopass/pkg/gopass"
	"github.com/urfave/cli/v2"
)

// secretGetter is the minimal store interface required by the template engine.
type secretGetter interface {
	Get(context.Context, string) (gopass.Secret, error)
}

// pathRestrictedStore wraps a secretGetter and only allows Get() calls for
// secret names that have one of the configured path prefixes. All other
// accesses are denied with a generic error so untrusted templates cannot
// exfiltrate secrets outside the declared scope.
type pathRestrictedStore struct {
	inner      secretGetter
	allowPaths []string
}

func (p *pathRestrictedStore) Get(ctx context.Context, name string) (gopass.Secret, error) {
	for _, prefix := range p.allowPaths {
		if strings.HasPrefix(name, prefix) {
			return p.inner.Get(ctx, name)
		}
	}

	return nil, fmt.Errorf("access denied: %q is not within an allowed path", name)
}

// Process is a command to process a template and replace secrets contained in it.
func (s *miscHandler) Process(c *cli.Context) error {
	ctx := ctxutil.WithGlobalFlags(c)
	file := c.Args().First()
	if file == "" {
		return exit.Error(exit.Usage, nil, "Usage: %s process <FILE>", s.Name)
	}

	allowPaths := c.StringSlice("allow-path")

	buf, err := os.ReadFile(file)
	if err != nil {
		return exit.Error(exit.IO, err, "Failed to read file: %s", file)
	}

	// Decide which store view the template engine may access.
	var store secretGetter
	if len(allowPaths) > 0 {
		store = &pathRestrictedStore{inner: s.Store, allowPaths: allowPaths}
	} else {
		out.Warningf(ctx, "No --allow-path flag set. The template has unrestricted access to ALL secrets in the store. Only process templates from trusted sources.")
		store = s.Store
	}

	obuf, err := tpl.Execute(ctx, string(buf), file, nil, store)
	if err != nil {
		return exit.Error(exit.IO, err, "Failed to process file: %s", file)
	}

	out.Print(ctx, string(obuf))

	return nil
}
